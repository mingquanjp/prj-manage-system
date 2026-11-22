#!/usr/bin/env python3
"""Chuẩn bị, tạo và đối soát PCA trên Jira DATN; token chỉ nhập ẩn trong terminal."""
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

SOURCE = ROOT / 'docs/proposals/backend-profile-change-approval-tickets.md'
PROPOSAL = ROOT / 'docs/proposals/backend-profile-change-approval.md'
SQL_PLAN = ROOT / 'docs/proposals/backend-profile-change-approval-sql-plan.md'
SOURCES = [SOURCE, PROPOSAL, SQL_PLAN, ROOT / 'docs/api/profile-change-approval.openapi.json',
           ROOT / 'docs/product/functional-requirements.md', ROOT / 'docs/ai/workflow.md', ROOT / '.claude/docs/employee_profile_workflow.md']
IMPORT_LABEL = 'datn-pca-v1'
ORDER = [f'PCA-T{i:02}' for i in range(1, 13)]
STATE = ROOT / '.local/jira-pca/state.json'


def codes(value, prefix):
    """Kiểm toàn bộ token để không bỏ qua mã sai bên cạnh mã hợp lệ."""
    if value.strip() in ('—', '—.'):
        return []
    tokens = [token.strip() for token in value.strip().removesuffix('.').split(',')]
    require(all(re.fullmatch(prefix + r'\d{2}', token) for token in tokens), 'Mã WBS/phụ thuộc PCA không hợp lệ.')
    require(len(tokens) == len(set(tokens)), 'Mã WBS/phụ thuộc PCA bị trùng.')
    return tokens


def load_manifest():
    text = SOURCE.read_text(encoding='utf-8')
    require(all(text.count(heading) == 1 for heading in (
        '## 2. Epic\n', '## 3. Rank, WBS và phụ thuộc\n',
        '## 4. Quy ước nghiệm thu\n', '## 5. Description từng Task\n')),
        'Nguồn PCA thiếu hoặc trùng mục Epic/Rank/description.')
    table = text.split('## 3. Rank, WBS và phụ thuộc\n', 1)[1].split('\n## 4.', 1)[0]
    lines = [line for line in table.splitlines() if line.startswith('|')][2:]
    rows = []
    for line in lines:
        cells = [cell.strip() for cell in line.strip('|').split('|')]
        require(len(cells) == 5, 'Dòng Rank PCA phải có đúng năm cột.')
        rows.append(cells)
    require(len(rows) == 12, 'Nguồn PCA phải có đúng 12 dòng ticket.')
    proposal = PROPOSAL.read_text(encoding='utf-8')
    require(proposal.count('## 9. WBS') == 1 and proposal.count('\n## 10.') == 1,
            'Proposal PCA thiếu hoặc trùng mục WBS/bàn giao.')
    wbs_table = proposal.split('## 9. WBS', 1)[1].split('\n## 10.', 1)[0]
    wbs = re.findall(r'^\| (PCA-[^ |]+) \|', wbs_table, re.M)
    require(len(wbs) == len(set(wbs)) == 12 and set(wbs) == {f'PCA-{n:02}' for n in range(1, 13)},
            'Proposal phải có đúng 12 WBS PCA duy nhất.')
    sections = text.split('## 5. Description từng Task\n', 1)[1].split('\n## 6.', 1)[0]
    chunks = re.split(r'^### ', sections, flags=re.M)
    require(len(chunks) == 13, 'Phải có đúng 12 description PCA.')
    descriptions = {}
    for chunk in chunks[1:]:
        title, _, body = chunk.partition('\n')
        match = re.fullmatch(r'\[(PCA-T\d{2})\] .+', title)
        require(match is not None and match[1] not in descriptions, 'Heading PCA sai mã hoặc bị trùng.')
        descriptions[match[1]] = (title, body.strip())
    tasks, mapped = {}, []
    required = ['Mục tiêu', 'Đầu vào và điều kiện bắt đầu', 'Phạm vi thực hiện',
                'Hành vi và tình huống cần đạt', 'Tiêu chí hoàn thành', 'Phụ thuộc và cách thực hiện',
                'Kiểm tra và bằng chứng', 'Đầu ra bàn giao', 'Ngoài phạm vi', 'Tham chiếu']
    for pos, (rank, sid, ws, deps, group) in enumerate(rows, 1):
        require(sid == ORDER[pos-1] and rank == str(pos), 'Rank/mã PCA khác thứ tự đã duyệt.')
        require(group in {'design', 'db', 'http', 'web', 'e2e'}, f'{sid}: nhóm chưa được review.')
        require(sid in descriptions, f'{sid}: thiếu description.')
        title, body = descriptions[sid]
        require(all(f'**{h}**' in body for h in required) and body.count('- [ ]') >= 2,
                f'{sid}: thiếu description/AC.')
        depends = codes(deps, 'PCA-T')
        value = re.findall(r'^- Is blocked by: (.*)$', body, re.M)
        require(len(value) == 1 and codes(value[0], 'PCA-T') == depends, f'{sid}: lệch Is blocked by.')
        value = re.findall(r'^- Blocks: (.*)$', body, re.M)
        require(len(value) == 1, f'{sid}: thiếu/trùng Blocks.')
        children = codes(value[0], 'PCA-T')
        units = codes(ws, 'PCA-')
        mapped.extend(units)
        value = re.findall(r'^- WBS: ([^.]+)\.', body, re.M)
        require(len(value) == 1 and codes(value[0], 'PCA-') == units, f'{sid}: description lệch WBS.')
        trace = re.findall(r'^\*\*Truy vết nhập Jira:\*\* `([^`]+)`; nguồn `([^`]+)`\.$', body, re.M)
        require(trace == [(sid, 'docs/proposals/backend-profile-change-approval-tickets.md')], f'{sid}: truy vết nguồn sai.')
        body = re.sub(r'\n\n\*\*Truy vết nhập Jira:\*\*[^\n]+$', '', body)
        tasks[sid] = {'source_id': sid, 'summary': title, 'markdown': body, 'type': 'Task',
                      'depends_on': depends, 'blocks': children, 'wbs': units,
                      'labels': ['profile-change-approval', IMPORT_LABEL, 'source-' + sid.lower(), group]}
    require(len(mapped) == len(set(mapped)) == 12 and set(mapped) == set(wbs), 'Mapping PCA phải đủ 12 WBS đúng một lần.')
    positions = {sid: n for n, sid in enumerate(ORDER)}
    for sid, record in tasks.items():
        require(all(d in tasks and positions[d] < positions[sid] for d in record['depends_on']), f'{sid}: vòng phụ thuộc/Rank sai.')
        require(set(record['blocks']) == {k for k, v in tasks.items() if sid in v['depends_on']}, f'{sid}: hai chiều Blocks lệch.')
    epic_body = text.split('## 2. Epic\n', 1)[1].split('\n## 3.', 1)[0].strip()
    summary = re.search(r'^\*\*Summary:\*\* (.+)$', epic_body, re.M)
    require(summary is not None, 'Thiếu Summary Epic PCA.')
    epic = {'source_id': 'PCA-EPIC', 'summary': summary[1], 'markdown': epic_body.split('\n', 1)[1].strip(),
            'type': 'Epic', 'labels': ['profile-change-approval', IMPORT_LABEL, 'source-pca-epic']}
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
        require(resolved.is_relative_to(ROOT) and resolved.is_file(), 'Link nguồn PCA không hợp lệ.')
        path = resolved.relative_to(ROOT).as_posix()
        if source_base:
            return f'[{label}]({source_base}/{path}' + (f'#{anchor}' if anchor else '') + ')'
        return f'{label} (`{path}` — link remote chờ bàn giao)'
    text = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', replace, record['markdown'])
    source_path = SOURCE.relative_to(ROOT).as_posix()
    source = f'[Bộ ticket PCA]({source_base}/{source_path})' if source_base else f'`{source_path}` — link remote chờ bàn giao'
    text += f'\n\n**Truy vết nhập Jira**\n\nMã nguồn ticket: `{record["source_id"]}`. Nguồn: {source}.\n'
    return adf(text)


def validate_source_base(source_base):
    if not source_base:
        return
    require(bool(re.fullmatch(r'https://github.com/mingquanjp/(?:ai-staffing|prj-manage-system)/blob/[a-f0-9]{40}', source_base)), 'source-base phải là commit URL đúng repo.')
    sha = source_base.rsplit('/', 1)[1]
    for file in SOURCES:
        try:
            data = subprocess.check_output(['git', 'show', f'{sha}:{file.relative_to(ROOT)}'], cwd=ROOT, stderr=subprocess.DEVNULL)
        except subprocess.CalledProcessError:
            raise Stop('Commit chưa chứa đủ nguồn PCA; bàn giao nguồn trước khi thêm link.') from None
        require(data == file.read_bytes(), 'Nguồn PCA local khác commit được dẫn.')


def safe_state_path(path=STATE):
    # Đích cố định; không nhận --state trỏ nhầm journal auth hoặc đường dẫn qua symlink.
    require(path == ROOT / '.local/jira-pca/state.json', 'Chỉ dùng journal PCA riêng.')
    for item in [ROOT / '.local', path.parent, path, path.with_suffix('.lock')]:
        require(not item.is_symlink(), 'Journal/lock PCA không được đi qua symlink.')
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
    parser.add_argument('--retry-create', choices=['PCA-EPIC', *ORDER])
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
        print('OFFLINE: 1 Epic + 12 Task; 12 WBS; Rank hợp phụ thuộc; 17 Blocks.')
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
