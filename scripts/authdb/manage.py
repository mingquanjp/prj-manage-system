#!/usr/bin/env python3
"""Môi trường PostgreSQL test sở hữu riêng; không đọc dotenv hoặc DB phát triển."""
import argparse
import json
import os
from pathlib import Path
import secrets
import subprocess
import sys
import tempfile
import time
from urllib.parse import urlsplit

ROOT = Path(__file__).resolve().parents[2]
STATE = ROOT / '.local/authdb/state.json'
NAME = 'ai-staffing-auth-test'
VOLUME = NAME + '-data'
DATABASE = 'ai_staffing_auth_test'
USER = 'ai_staffing_test'
PORT = '55438'
LABEL = 'vn.ai-staffing.auth-test-owner'


class Refused(Exception):
    pass


def docker(*args, input=None, env=None):
    # Không đưa stderr Docker/psql vào lỗi: có thể chứa credential hoặc SQL.
    result = subprocess.run(['docker', *args], input=input, text=True,
                            capture_output=True, env=env, timeout=180)
    if result.returncode:
        raise Refused('Docker/SQL không thành công; không xuất lỗi thô. Kiểm tra tài nguyên test riêng.')
    return result.stdout.strip()


def engine():
    if os.environ.get('DOCKER_HOST') or os.environ.get('DOCKER_CONTEXT'):
        raise Refused('Không nhận DOCKER_HOST/DOCKER_CONTEXT override.')
    context = json.loads(docker('context', 'inspect'))[0]
    if not context['Endpoints']['docker']['Host'].startswith('unix://'):
        raise Refused('Chỉ dùng Docker local qua Unix socket.')
    return docker('info', '--format', '{{.ID}}')


def load():
    if not STATE.is_file() or STATE.is_symlink() or STATE.stat().st_mode & 0o077:
        raise Refused('Thiếu state riêng hoặc quyền file không phải 0600; chạy up trước.')
    return json.loads(STATE.read_text())


def validate_url(raw, expected):
    # So khớp toàn bộ URL cấp bởi runner, không chỉ tin tên DB hoặc loopback.
    if not raw or raw != expected:
        raise Refused('TEST_DATABASE_URL không khớp tài nguyên test sở hữu.')
    u = urlsplit(raw)
    if (u.scheme != 'postgresql' or u.hostname != '127.0.0.1'
            or str(u.port) != PORT or u.username != USER
            or u.path != '/' + DATABASE or u.query != 'sslmode=disable' or u.fragment):
        raise Refused('Đích test không hợp lệ.')


def checked(state, running=True):
    if state['engine'] != engine():
        raise Refused('Docker engine khác engine đã tạo tài nguyên.')
    validate_url(state['url'], state['url'])
    if 'TEST_DATABASE_URL' in os.environ:
        validate_url(os.environ['TEST_DATABASE_URL'], state['url'])
    c = json.loads(docker('inspect', NAME))[0]
    v = json.loads(docker('volume', 'inspect', VOLUME))[0]
    binds = c['HostConfig']['PortBindings']
    mounts = c['Mounts']
    if (c['Id'] != state['container_id'] or c['Image'] != state['image_id']
            or c['Config']['Labels'].get(LABEL) != state['owner']
            or (v.get('Labels') or {}).get(LABEL) != state['owner']
            or binds != {'5432/tcp': [{'HostIp': '127.0.0.1', 'HostPort': PORT}]}
            or len(mounts) != 1 or mounts[0].get('Name') != VOLUME
            or mounts[0]['Destination'] != '/var/lib/postgresql'):
        raise Refused('Container/volume/port không đúng tài nguyên test sở hữu; không thao tác.')
    if running and not c['State']['Running']:
        raise Refused('Container test chưa chạy; dùng up.')
    return c


def probe(state):
    checked(state)
    env = dict(os.environ, PGPASSWORD=state['password'])
    # TCP bên trong đúng container, không dùng trust/socket; không đọc psqlrc.
    sql = """SELECT json_build_object(
        'major', current_setting('server_version_num')::int / 10000,
        'database', current_database(), 'user', current_user,
        'owner', pg_get_userbyid(datdba) = current_user,
        'public_tables', (SELECT count(*) FROM pg_tables WHERE schemaname='public'),
        'system_id', (SELECT system_identifier::text FROM pg_control_system()))
        FROM pg_database WHERE datname=current_database();"""
    result = json.loads(docker('exec', '-i', '-e', 'PGPASSWORD', NAME,
                              'psql', '-X', '-w', '-h', '127.0.0.1', '-U', USER,
                              '-d', DATABASE, '-At', '-v', 'ON_ERROR_STOP=1',
                              input=sql, env=env))
    if (result['major'], result['database'], result['user'], result['owner']) != (18, DATABASE, USER, True):
        raise Refused('Metadata DB không đúng major/name/user/owner.')
    return result


def up():
    engine_id = engine()
    if STATE.exists():
        state = load()
        checked(state, running=False)
        docker('start', NAME)
    else:
        # Không nhận tài nguyên cùng tên đã tồn tại, kể cả có label giống.
        containers = docker('ps', '-a', '--format', '{{.Names}}').splitlines()
        volumes = docker('volume', 'ls', '--format', '{{.Name}}').splitlines()
        if NAME in containers or VOLUME in volumes:
            raise Refused('Tên container/volume đã tồn tại nhưng không có state; không tái sử dụng/xóa.')
        docker('pull', 'postgres:18')
        image_id = docker('image', 'inspect', 'postgres:18', '--format', '{{.Id}}')
        owner, password = secrets.token_hex(16), secrets.token_hex(32)
        STATE.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
        state = dict(engine=engine_id, owner=owner, image_id=image_id, password=password,
                     url=f'postgresql://{USER}:{password}@127.0.0.1:{PORT}/{DATABASE}?sslmode=disable')
        docker('volume', 'create', '--label', f'{LABEL}={owner}', VOLUME)
        # File env tạm 0600, không xuất giá trị lên command line/log.
        with tempfile.NamedTemporaryFile(mode='w', prefix='authdb-', delete=True) as f:
            f.write(f'POSTGRES_DB={DATABASE}\nPOSTGRES_USER={USER}\nPOSTGRES_PASSWORD={password}\n')
            f.flush()
            state['container_id'] = docker('run', '-d', '--name', NAME,
                '--label', f'{LABEL}={owner}', '--publish', f'127.0.0.1:{PORT}:5432',
                '--mount', f'type=volume,src={VOLUME},dst=/var/lib/postgresql',
                '--env-file', f.name, image_id)
        fd = os.open(STATE, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(fd, 'w') as f:
            json.dump(state, f)
    for _ in range(30):
        try:
            result = probe(state)
            host_probe(state, result)
            print(json.dumps(result, ensure_ascii=False))
            return
        except Refused:
            time.sleep(1)
    raise Refused('DB chưa sẵn sàng; giữ tài nguyên/state để kiểm tra, không tự dọn.')


def host_probe(state, expected):
    # Đối chiếu cluster bên trong container với kết nối TCP từ máy host.
    env = {k: v for k, v in os.environ.items()
           if k not in ('DATABASE_URL', 'APP_ENV') and not k.startswith('PG')}
    env['TEST_DATABASE_URL'] = state['url']
    result = subprocess.run(['go', 'run', './scripts/authdb/probe.go'], cwd=ROOT,
                            env=env, text=True, capture_output=True, timeout=120)
    if result.returncode or json.loads(result.stdout) != expected:
        raise Refused('Kết nối TCP host không đúng cluster/DB/owner test.')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['up', 'check', 'stop', 'destroy', 'run'])
    parser.add_argument('command', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    if args.mode == 'up':
        up()
        return
    state = load()
    checked(state, running=args.mode not in ('destroy', 'stop'))
    if args.mode == 'check':
        result = probe(state)
        host_probe(state, result)
        print(json.dumps(result, ensure_ascii=False))
    elif args.mode == 'stop':
        docker('stop', NAME)
        print('Đã dừng container test; giữ volume/state.')
    elif args.mode == 'destroy':
        if args.command != ['--confirm', NAME]:
            raise Refused('Cần --confirm ai-staffing-auth-test để xóa đúng container/volume test.')
        docker('rm', '-f', NAME)
        docker('volume', 'rm', VOLUME)
        STATE.unlink()
        print('Đã xóa container/volume test sở hữu; không xóa image.')
    elif args.mode == 'run':
        host_probe(state, probe(state))
        command = args.command
        if command[:1] == ['--']:
            command = command[1:]
        if not command:
            raise Refused('Thiếu lệnh test.')
        # Không đưa DATABASE_URL/APP_ENV/PG* từ shell vào test, không tự nạp dotenv.
        env = {k: v for k, v in os.environ.items()
               if k not in ('DATABASE_URL', 'APP_ENV') and not k.startswith('PG')}
        env['TEST_DATABASE_URL'] = state['url']
        result = subprocess.run(command, cwd=ROOT, env=env)
        if result.returncode:
            raise Refused('Lệnh test không đạt; runner không tự migration/reset.')


if __name__ == '__main__':
    try:
        main()
    except Refused as error:
        print('DỪNG: ' + str(error), file=sys.stderr)
        sys.exit(1)
    except (OSError, ValueError, KeyError, subprocess.TimeoutExpired):
        # Không in exception tùy ý: lỗi URL/JSON/subprocess có thể chứa bí mật.
        print('DỪNG: môi trường/lệnh test chưa được xác minh hoặc không thành công. '
              'Không thay DB khác; kiểm state/quyền/engine/đích test.', file=sys.stderr)
        sys.exit(1)
