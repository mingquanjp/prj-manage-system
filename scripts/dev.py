#!/usr/bin/env python3
"""Chạy FE + Go auth local; không đọc dotenv, không dùng DATABASE_URL/Neon."""
import os
from pathlib import Path
import signal
import socket
import subprocess
import sys
import tempfile
import time
from urllib.request import build_opener, ProxyHandler

from authdb import manage as authdb

ROOT = Path(__file__).resolve().parents[1]


def clean_env(source):
    return {key: value for key, value in source.items()
            if key not in ('DATABASE_URL', 'APP_ENV', 'TEST_DATABASE_URL',
                           'AUTH_TEST_DB_SYSTEM_ID', 'AUTH_LOCAL_DATABASE_URL',
                           'AUTH_LOCAL_SYSTEM_ID', 'TASKASSIGN_API_TARGET')
            and not key.startswith('PG')}


def require_free(port):
    try:
        with socket.socket() as probe:
            probe.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            probe.bind(('127.0.0.1', port))
    except OSError:
        raise authdb.Refused(f'Cổng {port} đang bận hoặc không mở được. '
                             'Dừng server cũ bằng Ctrl+C; không tự kill tiến trình khác.') from None


class Processes:
    def __init__(self):
        self.children = []
        self.stopping = False

    def stop_signal(self, *_):
        self.stopping = True

    def start(self, args, cwd, env):
        process = subprocess.Popen(args, cwd=cwd, env=env, start_new_session=True)
        self.children.append(process)
        return process

    def close(self):
        # Mỗi group được tạo bởi lượt chạy này; không tìm/kill theo tên hoặc cổng.
        for process in reversed(self.children):
            if process.poll() is None:
                try:
                    os.killpg(process.pid, signal.SIGTERM)
                except ProcessLookupError:
                    pass
        for process in reversed(self.children):
            try:
                process.wait(timeout=25)
            except subprocess.TimeoutExpired:
                print('Tiến trình local chưa dừng; kết thúc đúng group đã tạo. '
                      'DB tạm có thể còn lại, cần đối soát; không tự FORCE/drop.', file=sys.stderr)
                try:
                    os.killpg(process.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                process.wait()


def wait_for_backend(process, processes):
    # Không dùng proxy hệ thống cho probe loopback; không gửi credential HTTP.
    opener = build_opener(ProxyHandler({}))
    deadline = time.monotonic() + 100
    while not processes.stopping and time.monotonic() < deadline:
        if process.poll() is not None:
            raise authdb.Refused('Backend local dừng khi khởi tạo; FE chưa được mở.')
        try:
            with opener.open('http://127.0.0.1:8080/api/healthz', timeout=1) as response:
                if response.status == 200:
                    return
        except OSError:
            pass
        time.sleep(0.2)
    if not processes.stopping:
        raise authdb.Refused('Backend local chưa sẵn sàng; không mở FE giả thành công.')


def main():
    processes = Processes()
    for signum in (signal.SIGINT, signal.SIGTERM):
        signal.signal(signum, processes.stop_signal)
    require_free(5174)
    require_free(8080)
    print('Đang kiểm tra PostgreSQL local sở hữu. Docker Desktop cần đang chạy.', flush=True)
    # Giữ nguyên mọi guard Docker/URL/owner/major của runner đã review.
    authdb.up()
    if processes.stopping:
        return
    state = authdb.load()
    metadata = authdb.probe(state)
    authdb.host_probe(state, metadata)
    env = clean_env(os.environ)
    try:
        with tempfile.TemporaryDirectory(prefix='taskassign-local-') as directory:
            binary = str(Path(directory) / 'localdev')
            print('Đang build Go backend local…', flush=True)
            build = processes.start(['go', 'build', '-o', binary, './cmd/localdev'], ROOT, env)
            while build.poll() is None and not processes.stopping:
                time.sleep(0.1)
            if processes.stopping:
                return
            if build.returncode:
                raise authdb.Refused('Build backend local không thành công.')
            backend_env = dict(env, AUTH_LOCAL_DATABASE_URL=state['url'],
                               AUTH_LOCAL_SYSTEM_ID=metadata['system_id'])
            backend = processes.start([binary], ROOT, backend_env)
            wait_for_backend(backend, processes)
            if processes.stopping:
                return
            frontend_env = dict(env, VITE_AUTH_MODE='server',
                                TASKASSIGN_API_TARGET='http://127.0.0.1:8080')
            frontend = processes.start(['node', 'node_modules/vite/bin/vite.js'], ROOT / 'web', frontend_env)
            print('FE: http://127.0.0.1:5174 — Swagger: http://127.0.0.1:8080/docs/; Ctrl+C dừng cả FE/BE. '
                  'Dữ liệu giả của lượt chạy sẽ được dọn khi dừng.', flush=True)
            while not processes.stopping:
                if backend.poll() is not None or frontend.poll() is not None:
                    raise authdb.Refused('Một tiến trình local đã dừng; đang dừng tiến trình còn lại.')
                time.sleep(0.2)
    finally:
        processes.close()


if __name__ == '__main__':
    try:
        main()
    except (authdb.Refused, OSError, ValueError, KeyError, subprocess.TimeoutExpired) as error:
        # Chỉ xuất lỗi đã kiểm soát; exception hệ thống có thể chứa thông tin nhạy cảm.
        message = str(error) if isinstance(error, authdb.Refused) else 'Môi trường local chưa sẵn sàng; kiểm Docker/Go/Node.'
        print('DỪNG: ' + message, file=sys.stderr)
        sys.exit(1)
