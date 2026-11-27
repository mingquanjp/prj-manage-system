#!/usr/bin/env python3
"""Chuẩn bị, tạo và đối soát MSK trên Jira DATN từ bộ ticket được review."""
import argparse
import getpass
import hashlib
import re
import subprocess
import sys
from pathlib import Path

from auth_import import (SITE, PROJECT, BOARD, ROOT, Stop, require, adf, Jira,
                         board_diagnostics, expected_links, state_lock)
from jira_batch import BatchImporter
from pca_import import credentials, export_manifest

SOURCE = ROOT / 'docs/proposals/backend-manual-skills-tickets.md'
PROPOSAL = ROOT / 'docs/proposals/backend-manual-skills.md'
SQL_PLAN = ROOT / 'docs/proposals/backend-manual-skills-sql-plan.md'
SOURCES = [SOURCE, PROPOSAL, SQL_PLAN, ROOT / 'docs/api/manual-skills.openapi.json',
           ROOT / 'docs/api/profile-change-approval.openapi.json',
           ROOT / 'docs/product/functional-requirements.md', ROOT / 'docs/ai/workflow.md',
           ROOT / '.claude/docs/skill_approval_workflow.md']
IMPORT_LABEL = 'datn-msk-v1'
ORDER = [f'MSK-T{i:02}' for i in range(1, 13)]
STATE = ROOT / '.local/jira-msk/state.json'


def codes(value, prefix):
    if value.strip() in ('—', '—.'):
        return []
    tokens = [token.strip() for token in value.strip().removesuffix('.').split(',')]
    require(all(re.fullmatch(prefix + r'\d{2}', token) for token in tokens),
            'Mã WBS/phụ thuộc MSK không hợp lệ.')
    require(len(tokens) == len(set(tokens)), 'Mã WBS/phụ thuộc MSK bị trùng.')
    return tokens


def load_manifest():
    text = SOURCE.read_text(encoding='utf-8')
    require(all(text.count(heading) == 1 for heading in (
        '## 1. Epic\n', '## 2. Rank, WBS và dependency\n',
        '## 3. Quy ước kiểm và bàn giao\n', '## 4. Description từng task\n')),
        'Nguồn MSK thiếu hoặc trùng mục Epic/Rank/description.')
    table = text.split('## 2. Rank, WBS và dependency\n', 1)[1].split('\n## 3.', 1)[0]
    rows = []
    for line in [line for line in table.splitlines() if line.startswith('|')][2:]:
        cells = [cell.strip() for cell in line.strip('|').split('|')]
        require(len(cells) == 4, 'Dòng Rank MSK phải có đúng bốn cột.')
        rows.append(cells)
    require(len(rows) == 12, 'Nguồn MSK phải có đúng 12 dòng ticket.')

    proposal = PROPOSAL.read_text(encoding='utf-8')
    require(proposal.count('## 9. WBS') == 1 and proposal.count('\n## 10.') == 1,
            'Proposal MSK thiếu hoặc trùng mục WBS/bàn giao.')
    wbs_table = proposal.split('## 9. WBS', 1)[1].split('\n## 10.', 1)[0]
    wbs = re.findall(r'^\| (MSK-[^ |]+) \|', wbs_table, re.M)
    require(len(wbs) == len(set(wbs)) == 12 and set(wbs) == {f'MSK-{n:02}' for n in range(1, 13)},
            'Proposal phải có đúng 12 WBS MSK duy nhất.')

    sections = text.split('## 4. Description từng task\n', 1)[1]
    chunks = re.split(r'^### ', sections, flags=re.M)
    require(len(chunks) == 13, 'Phải có đúng 12 description MSK.')
    descriptions = {}
    for chunk in chunks[1:]:
        title, _, body = chunk.partition('\n')
        match = re.fullmatch(r'\[(MSK-T\d{2})\] .+', title)
        require(match is not None and match[1] not in descriptions, 'Heading MSK sai mã hoặc bị trùng.')
        descriptions[match[1]] = (title, body.strip())

    tasks, mapped = {}, []
    required = ['Mục tiêu', 'Đầu vào và điều kiện bắt đầu', 'Phạm vi thực hiện',
                'Hành vi và tình huống cần đạt', 'Tiêu chí hoàn thành', 'Phụ thuộc và cách thực hiện',
                'Kiểm tra và bằng chứng', 'Đầu ra bàn giao', 'Ngoài phạm vi', 'Tham chiếu']
    for pos, (rank, sid, ws, deps) in enumerate(rows, 1):
        require(sid == ORDER[pos-1] and rank == str(pos), 'Rank/mã MSK khác thứ tự đã duyệt.')
        require(sid in descriptions, f'{sid}: thiếu description.')
        title, body = descriptions[sid]
        require(all(body.count(f'**{heading}**') == 1 for heading in required) and body.count('- [ ]') >= 2,
                f'{sid}: thiếu/trùng description hoặc thiếu AC.')
        depends = codes(deps, 'MSK-T')
        value = re.findall(r'^- Is blocked by: (.*)$', body, re.M)
        require(len(value) == 1 and codes(value[0], 'MSK-T') == depends, f'{sid}: lệch Is blocked by.')
        value = re.findall(r'^- Blocks: (.*)$', body, re.M)
        require(len(value) == 1, f'{sid}: thiếu/trùng Blocks.')
        children = codes(value[0], 'MSK-T')
        units = codes(ws, 'MSK-')
        mapped.extend(units)
        value = re.findall(r'^- WBS: ([^.]+)\.', body, re.M)
        require(len(value) == 1 and codes(value[0], 'MSK-') == units, f'{sid}: description lệch WBS.')
        trace = re.findall(r'^\*\*Truy vết nhập Jira:\*\* `([^`]+)`; nguồn `([^`]+)`\.$', body, re.M)
        require(trace == [(sid, 'docs/proposals/backend-manual-skills-tickets.md')], f'{sid}: truy vết nguồn sai.')
        body = re.sub(r'\n\n\*\*Truy vết nhập Jira:\*\*[^\n]+$', '', body)
        tasks[sid] = {'source_id': sid, 'summary': title, 'markdown': body, 'type': 'Task',
                      'depends_on': depends, 'blocks': children, 'wbs': units,
                      'labels': ['manual-skills', IMPORT_LABEL, 'source-' + sid.lower()]}
    require(len(mapped) == len(set(mapped)) == 12 and set(mapped) == set(wbs),
            'Mapping MSK phải đủ 12 WBS đúng một lần.')
    positions = {sid: n for n, sid in enumerate(ORDER)}
    for sid, record in tasks.items():
        require(all(d in tasks and positions[d] < positions[sid] for d in record['depends_on']),
                f'{sid}: vòng phụ thuộc/Rank sai.')
        require(set(record['blocks']) == {k for k, v in tasks.items() if sid in v['depends_on']},
                f'{sid}: hai chiều Blocks lệch.')
    epic_body = text.split('## 1. Epic\n', 1)[1].split('\n## 2.', 1)[0].strip()
    summary = re.search(r'^\*\*Summary:\*\* (.+)$', epic_body, re.M)
    require(summary is not None, 'Thiếu Summary Epic MSK.')
    epic = {'source_id': 'MSK-EPIC', 'summary': summary[1], 'markdown': epic_body.split('\n', 1)[1].strip(),
            'type': 'Epic', 'labels': ['manual-skills', IMPORT_LABEL, 'source-msk-epic']}
    manifest = {'site': SITE, 'project': PROJECT, 'board': BOARD, 'import_label': IMPORT_LABEL,
                'order': ORDER, 'records': [epic] + [tasks[sid] for sid in ORDER],
                'source_hashes': {p.relative_to(ROOT).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in SOURCES}}
    require(len(expected_links(manifest)) == 19, 'Tập phụ thuộc phải có đúng 19 Blocks đã review.')
    return manifest


def render_description(record, source_base=None):
    def replace(match):
        label, href = match.groups()
        if href.startswith('https://'):
            return match[0]
        target, _, anchor = href.partition('#')
        resolved = (SOURCE.parent / target).resolve()
        require(resolved.is_relative_to(ROOT) and resolved.is_file(), 'Link nguồn MSK không hợp lệ.')
        path = resolved.relative_to(ROOT).as_posix()
        if source_base:
            return f'[{label}]({source_base}/{path}' + (f'#{anchor}' if anchor else '') + ')'
        return f'{label} (`{path}` — link remote chờ bàn giao)'
    text = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', replace, record['markdown'])
    source_path = SOURCE.relative_to(ROOT).as_posix()
    if source_base:
        source = f'[Bộ ticket MSK]({source_base}/{source_path})'
    else:
        source = f'`{source_path}` — link remote chờ bàn giao'
    text += f'\n\n**Truy vết nhập Jira**\n\nMã nguồn ticket: `{record["source_id"]}`. Nguồn: {source}.\n'
    return adf(text)


def validate_source_base(source_base):
    if not source_base:
        return
    require(bool(re.fullmatch(r'https://github.com/mingquanjp/(?:ai-staffing|prj-manage-system)/blob/[a-f0-9]{40}', source_base)),
            'source-base phải là commit URL đúng repo.')
    sha = source_base.rsplit('/', 1)[1]
    for file in SOURCES:
        try:
            data = subprocess.check_output(['git', 'show', f'{sha}:{file.relative_to(ROOT)}'],
                                           cwd=ROOT, stderr=subprocess.DEVNULL)
        except subprocess.CalledProcessError:
            raise Stop('Commit chưa chứa đủ nguồn MSK; bàn giao nguồn trước khi thêm link.') from None
        require(data == file.read_bytes(), 'Nguồn MSK local khác commit được dẫn.')


def safe_state_path(path=STATE):
    # Journal cố định ngăn dùng nhầm mapping đợt khác hoặc ghi qua symlink.
    require(path == ROOT / '.local/jira-msk/state.json', 'Chỉ dùng journal MSK riêng.')
    for item in [ROOT / '.local', path.parent, path, path.with_suffix('.lock')]:
        require(not item.is_symlink(), 'Journal/lock MSK không được đi qua symlink.')
    return path


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--mode', choices=['plan', 'apply', 'check', 'diagnose'], default='plan')
    parser.add_argument('--offline', action='store_true')
    parser.add_argument('--source-base', help='URL GitHub blob tới commit đầy đủ đã push; giữ cố định cả lượt import.')
    parser.add_argument('--manifest-out', type=Path, help='File mới sinh từ nguồn; không chỉnh tay.')
    parser.add_argument('--cloud-id', help='UUID site DATN mới khi dùng token có scope.')
    parser.add_argument('--retry-create', choices=['MSK-EPIC', *ORDER])
    args = parser.parse_args()
    require(not args.offline or args.mode == 'plan', '--offline chỉ dùng với plan.')
    require(not args.retry_create or args.mode == 'apply', '--retry-create chỉ dùng với apply sau đối soát thủ công.')
    if args.mode == 'diagnose':
        require(not (args.source_base or args.manifest_out or args.retry_create),
                'diagnose chỉ đọc metadata, không dùng manifest/journal.')
        email, token = credentials()
        board_diagnostics(Jira(email, token, args.cloud_id))
        return
    validate_source_base(args.source_base)
    manifest = load_manifest()
    # Lỗi nội dung/link phải được phát hiện trước khi hỏi token hoặc gọi Jira.
    for record in manifest['records']:
        render_description(record, args.source_base)
    if args.manifest_out:
        export_manifest(args.manifest_out, manifest)
        print(f'Manifest sinh từ nguồn: {args.manifest_out}')
    if args.offline:
        print('OFFLINE: 1 Epic + 12 Task; 12 WBS; Rank hợp phụ thuộc; 19 Blocks.')
        BatchImporter(manifest, render_description).report(manifest, {}, set())
        print('Description đã chuyển ADF; chưa gọi Jira hoặc ghi journal.')
        return
    email, token = credentials()
    path = safe_state_path()
    with state_lock(path):
        BatchImporter(manifest, render_description).run_online(
            Jira(email, token, args.cloud_id), manifest, path, args.mode, args.source_base, args.retry_create)


if __name__ == '__main__':
    try:
        main()
    except (Stop, OSError, ValueError, KeyError, TypeError, getpass.GetPassWarning):
        error = sys.exc_info()[1]
        if isinstance(error, Stop):
            message = str(error)
        else:
            message = 'Input/file/phản hồi không hợp lệ; không in dữ liệu nhạy cảm.'
        print('DỪNG: ' + message, file=sys.stderr)
        sys.exit(1)
    except (KeyboardInterrupt, EOFError):
        print('\nĐã dừng; chạy plan để đối soát mutation pending trước khi tiếp tục.', file=sys.stderr)
        sys.exit(1)
