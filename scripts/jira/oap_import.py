#!/usr/bin/env python3
"""Chuẩn bị, tạo và đối soát OAP trên Jira DATN; token chỉ nhập ẩn trong terminal."""
import argparse
import getpass
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import warnings

from auth_import import (SITE, PROJECT, BOARD, ROOT, Stop, require, adf, Jira,
                         board_diagnostics, expected_links, state_lock)
from jira_batch import BatchImporter

SOURCE = ROOT / 'docs/proposals/backend-organization-accounts-profile-tickets.md'
PROPOSAL = ROOT / 'docs/proposals/backend-organization-accounts-profile.md'
SQL_PLAN = ROOT / 'docs/proposals/backend-organization-accounts-profile-sql-plan.md'
SOURCES = [SOURCE, PROPOSAL, SQL_PLAN, ROOT / 'docs/api/organization-accounts-profile.openapi.json',
           ROOT / 'docs/product/functional-requirements.md', ROOT / 'docs/ai/workflow.md']
IMPORT_LABEL = 'datn-oap-v1'
ORDER = [f'OAP-T{i:02}' for i in range(1, 14)]
STATE = ROOT / '.local/jira-oap/state.json'


def load_manifest():
    text = SOURCE.read_text(encoding='utf-8')
    rows = re.findall(r'^\| (\d{2}) \| (\[OAP-T\d{2}\] [^|]+) \| ([^|]+) \| ([a-z]+) \| ([^|]+) \| ([^|]+) \|$', text, re.M)
    require(len(rows) == 13, 'Nguồn OAP phải có đúng 13 dòng ticket.')
    wbs = re.findall(r'^\| (OAP-\d{2}) \|', PROPOSAL.read_text(encoding='utf-8'), re.M)
    require(len(wbs) == len(set(wbs)) == 15, 'Proposal phải có đúng 15 WBS duy nhất.')
    tasks, mapped = {}, []
    required = ['Mục tiêu', 'Đầu vào và điều kiện bắt đầu', 'Phạm vi thực hiện',
                'Hành vi và tình huống cần đạt', 'Tiêu chí hoàn thành', 'Phụ thuộc và cách thực hiện',
                'Kiểm tra và bằng chứng', 'Đầu ra bàn giao', 'Ngoài phạm vi', 'Tham chiếu']
    groups = {'design', 'organization', 'profile', 'accounts', 'mail', 'handoff'}
    for pos, (rank, title, ws, group, deps, blocks) in enumerate(rows, 1):
        sid = re.match(r'\[(OAP-T\d{2})\] ', title)[1]
        require(sid == ORDER[pos-1] and int(rank) == pos and sid not in tasks, 'Rank/mã OAP thiếu, trùng hoặc khác thứ tự đã duyệt.')
        require(group in groups, f'{sid}: nhóm chưa được review.')
        marker = f'### {title}\n'
        require(text.count(marker) == 1, f'{sid}: tiêu đề description không duy nhất.')
        body = re.split(r'\n### \[|\n## 6\.', text.split(marker, 1)[1], maxsplit=1)[0].strip()
        require(all(f'**{h}**' in body for h in required) and body.count('- [ ]') >= 3, f'{sid}: thiếu description/AC.')
        depends, children = re.findall(r'OAP-T\d{2}', deps), re.findall(r'OAP-T\d{2}', blocks)
        for field, expected in [('Is blocked by', depends), ('Blocks', children)]:
            value = re.search(rf'^- {field}: (.*)$', body, re.M)
            require(value is not None and re.findall(r'OAP-T\d{2}', value[1]) == expected, f'{sid}: lệch {field}.')
            require(len(expected) == len(set(expected)), f'{sid}: cạnh bị trùng.')
        units = re.findall(r'OAP-\d{2}', ws)
        mapped.extend(units)
        require(all(w in wbs for w in units) and units, f'{sid}: WBS không hợp lệ.')
        body_wbs = re.search(r'^- WBS: ([^.]+)\.', body, re.M)
        require(body_wbs is not None and re.findall(r'OAP-\d{2}', body_wbs[1]) == units, f'{sid}: description lệch WBS.')
        # Description đã có dòng truy vết nội bộ; thay bằng một mục chuẩn khi render.
        body = re.sub(r'\n\n\*\*Truy vết nhập Jira:\*\*[^\n]+$', '', body)
        tasks[sid] = {'source_id': sid, 'summary': title, 'markdown': body, 'type': 'Task',
                      'depends_on': depends, 'blocks': children, 'wbs': units,
                      'labels': ['organization-accounts-profile', IMPORT_LABEL, 'source-' + sid.lower(), group]}
    require(len(mapped) == len(set(mapped)) == 15 and set(mapped) == set(wbs), 'Mapping OAP phải đủ 15 WBS đúng một lần.')
    positions = {sid: n for n, sid in enumerate(ORDER)}
    for sid, record in tasks.items():
        require(all(d in tasks and positions[d] < positions[sid] for d in record['depends_on']), f'{sid}: vòng phụ thuộc/Rank sai.')
        require(set(record['blocks']) == {k for k, v in tasks.items() if sid in v['depends_on']}, f'{sid}: hai chiều Blocks lệch.')
    epic_body = text.split('## 2. Epic\n', 1)[1].split('\n## 3.', 1)[0].strip()
    summary = re.search(r'^\*\*Summary:\*\* `([^`]+)`', epic_body, re.M)
    require(summary is not None, 'Thiếu Summary Epic OAP.')
    epic = {'source_id': 'OAP-EPIC', 'summary': summary[1], 'markdown': epic_body.split('\n', 1)[1].strip(),
            'type': 'Epic', 'labels': ['organization-accounts-profile', IMPORT_LABEL, 'source-oap-epic']}
    manifest = {'site': SITE, 'project': PROJECT, 'board': BOARD, 'import_label': IMPORT_LABEL,
                'order': ORDER, 'records': [epic] + [tasks[sid] for sid in ORDER],
                'source_hashes': {p.relative_to(ROOT).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in SOURCES}}
    require(len(expected_links(manifest)) == 17, 'Tập phụ thuộc phải có đúng 17 Blocks đã review.')
    return manifest


def render_description(record, source_base=None):
    def replace(match):
        label, href = match.groups()
        if href.startswith('https://'):
            return match[0]
        target, _, anchor = href.partition('#')
        resolved = (SOURCE.parent / target).resolve()
        require(resolved.is_relative_to(ROOT) and resolved.is_file(), 'Link nguồn OAP không hợp lệ.')
        path = resolved.relative_to(ROOT).as_posix()
        if source_base:
            return f'[{label}]({source_base}/{path}' + (f'#{anchor}' if anchor else '') + ')'
        return f'{label} (`{path}` — link remote chờ bàn giao)'
    text = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', replace, record['markdown'])
    source_path = SOURCE.relative_to(ROOT).as_posix()
    source = f'[Bộ ticket OAP]({source_base}/{source_path})' if source_base else f'`{source_path}` — link remote chờ bàn giao'
    text += f'\n\n**Truy vết nhập Jira**\n\nMã nguồn ticket: `{record["source_id"]}`. Nguồn: {source}.\n'
    return adf(text)


def validate_source_base(source_base):
    if not source_base:
        return
    require(bool(re.fullmatch(r'https://github.com/mingquanjp/ai-staffing/blob/[a-f0-9]{40}', source_base)), 'source-base phải là commit URL đúng repo.')
    sha = source_base.rsplit('/', 1)[1]
    for file in SOURCES:
        try:
            data = subprocess.check_output(['git', 'show', f'{sha}:{file.relative_to(ROOT)}'], cwd=ROOT, stderr=subprocess.DEVNULL)
        except subprocess.CalledProcessError:
            raise Stop('Commit chưa chứa đủ nguồn OAP; bàn giao nguồn trước khi thêm link.') from None
        require(data == file.read_bytes(), 'Nguồn OAP local khác commit được dẫn.')


def safe_state_path(path=STATE):
    # Đích cố định; không nhận --state trỏ nhầm journal auth hoặc đường dẫn qua symlink.
    require(path == ROOT / '.local/jira-oap/state.json', 'Chỉ dùng journal OAP riêng.')
    for item in [ROOT / '.local', path.parent, path, path.with_suffix('.lock')]:
        require(not item.is_symlink(), 'Journal/lock OAP không được đi qua symlink.')
    return path


def export_manifest(path, manifest):
    # Chỉ tạo file mới, không ghi đè nguồn hoặc journal, kể cả khi có process chạy song song.
    resolved = path.resolve()
    if resolved.is_relative_to(ROOT):
        require('.local' not in resolved.relative_to(ROOT).parts, 'Không xuất manifest vào journal/config local.')
    with open(path, 'x', encoding='utf-8', opener=lambda name, flags: os.open(name, flags, 0o600)) as output:
        json.dump(manifest, output, ensure_ascii=False, indent=2)
        output.write('\n')


def credentials():
    require(sys.stdin.isatty() and sys.stderr.isatty(), 'Chạy trong terminal tương tác; không pipe token.')
    email = input('Email Atlassian: ').strip()
    with warnings.catch_warnings():
        warnings.simplefilter('error', getpass.GetPassWarning)
        token = getpass.getpass('API token (ẩn khi nhập): ')
    require(email and token, 'Email/token không được trống.')
    return email, token


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--mode', choices=['plan', 'apply', 'check', 'diagnose'], default='plan')
    parser.add_argument('--offline', action='store_true')
    parser.add_argument('--source-base', help='URL GitHub blob tới commit đầy đủ đã push; giữ cố định cả lượt import.')
    parser.add_argument('--manifest-out', type=Path, help='File mới sinh từ nguồn; không chỉnh tay.')
    parser.add_argument('--cloud-id', help='UUID site DATN mới khi dùng token có scope.')
    parser.add_argument('--retry-create', choices=['OAP-EPIC', *ORDER])
    args = parser.parse_args()
    require(not args.offline or args.mode == 'plan', '--offline chỉ dùng với plan.')
    require(not args.retry_create or args.mode == 'apply', '--retry-create chỉ dùng với apply sau đối soát thủ công.')
    if args.mode == 'diagnose':
        require(not (args.source_base or args.manifest_out or args.retry_create), 'diagnose chỉ đọc metadata, không dùng manifest/journal.')
        email, token = credentials()
        board_diagnostics(Jira(email, token, args.cloud_id))
        return
    validate_source_base(args.source_base)
    manifest = load_manifest()
    # Render trước mọi lệnh mạng/ghi; lỗi ADF/link phải lộ ở plan offline.
    for record in manifest['records']:
        render_description(record, args.source_base)
    if args.manifest_out:
        export_manifest(args.manifest_out, manifest)
        print(f'Manifest sinh từ nguồn: {args.manifest_out}')
    if args.offline:
        print('OFFLINE: 1 Epic + 13 Task; 15 WBS; Rank hợp phụ thuộc; 17 Blocks.')
        BatchImporter(manifest, render_description).report(manifest, {}, set())
        print('Description đã chuyển ADF; chưa gọi Jira hoặc ghi journal.')
        return
    email, token = credentials()
    path = safe_state_path()
    with state_lock(path):
        BatchImporter(manifest, render_description).run_online(Jira(email, token, args.cloud_id), manifest, path,
                                                              args.mode, args.source_base, args.retry_create)


if __name__ == '__main__':
    try:
        main()
    except (Stop, OSError, ValueError, KeyError, TypeError, getpass.GetPassWarning):
        error = sys.exc_info()[1]
        print('DỪNG: ' + (str(error) if isinstance(error, Stop) else 'Input/file/phản hồi không hợp lệ; không in dữ liệu nhạy cảm.'), file=sys.stderr)
        sys.exit(1)
    except (KeyboardInterrupt, EOFError):
        print('\nĐã dừng; chạy plan để đối soát mutation pending trước khi tiếp tục.', file=sys.stderr)
        sys.exit(1)
