#!/usr/bin/env python3
"""Chuẩn bị, tạo và đối soát ticket DATN; chỉ dùng thư viện chuẩn Python."""
import argparse
import base64
import contextlib
import getpass
import hashlib
import json
import os
from pathlib import Path
import re
import sys
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request
import warnings

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / 'docs/proposals/backend-auth-session-tickets.md'
PROPOSAL = ROOT / 'docs/proposals/backend-auth-session.md'
SITE = 'https://datn-mingquanjp.atlassian.net'
PROJECT = 'DATN'
BOARD = 1
VERIFY_ISSUE = 'DATN-12'
IMPORT_LABEL = 'datn-auth-v1'
ORDER = [f'AUTH-T{i:02}' for i in [*range(1, 14), 15, 14, *range(16, 21)]]
FIELDS = 'project,summary,description,issuetype,parent,assignee,labels,issuelinks,status,priority,duedate,timetracking'


class Stop(RuntimeError):
    pass


def require(condition, message):
    if not condition:
        raise Stop(message)


def digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True).encode()).hexdigest()


def inline(text):
    """Chỉ hỗ trợ inline Markdown được dùng trong bộ ticket đã duyệt."""
    result = []
    pattern = r'(\[[^\]]+\]\([^)]+\)|\*\*[^*]+\*\*|`[^`]+`)'
    for part in re.split(pattern, text):
        if not part:
            continue
        marks = []
        if part.startswith('**') and part.endswith('**'):
            part = part[2:-2]
            marks = [{'type': 'strong'}]
        elif part.startswith('`') and part.endswith('`'):
            part = part[1:-1]
            marks = [{'type': 'code'}]
        elif re.fullmatch(r'\[[^\]]+\]\([^)]+\)', part):
            label, href = re.fullmatch(r'\[([^\]]+)\]\(([^)]+)\)', part).groups()
            if href.startswith('https://'):
                part, marks = label, [{'type': 'link', 'attrs': {'href': href}}]
            else:
                part = f'{label} ({href})'
        node = {'type': 'text', 'text': part}
        if marks:
            node['marks'] = marks
        result.append(node)
    return result


def paragraph(text):
    return {'type': 'paragraph', 'content': inline(text)}


def adf(markdown):
    """Bảng là table ADF; checkbox chỉ là tiêu chí, không tự tick Done."""
    lines = markdown.strip().splitlines()
    nodes, i = [], 0
    while i < len(lines):
        line = lines[i].strip()
        if not line:
            i += 1
            continue
        if line.startswith('|'):
            rows = []
            while i < len(lines) and lines[i].strip().startswith('|'):
                cells = [x.strip() for x in lines[i].strip().strip('|').split('|')]
                i += 1
                if all(re.fullmatch(r':?-+:?', c) for c in cells):
                    continue
                kind = 'tableHeader' if not rows else 'tableCell'
                rows.append({'type': 'tableRow', 'content': [
                    {'type': kind, 'content': [paragraph(c)]} for c in cells]})
            require(rows and len({len(r['content']) for r in rows}) == 1, 'Bảng Markdown không đều.')
            nodes.append({'type': 'table', 'content': rows})
            continue
        heading = re.fullmatch(r'\*\*([^*]+)\*\*', line)
        if heading:
            nodes.append({'type': 'heading', 'attrs': {'level': 3}, 'content': inline(heading[1])})
        elif re.match(r'^(?:- |\d+\. )', line):
            ordered = bool(re.match(r'^\d+\. ', line))
            items = []
            while i < len(lines) and re.match(r'^\d+\. ' if ordered else r'^- ', lines[i].strip()):
                value = re.sub(r'^\d+\. |^- ', '', lines[i].strip())
                value = value.replace('[ ] ', '☐ ', 1)
                items.append({'type': 'listItem', 'content': [paragraph(value)]})
                i += 1
            nodes.append({'type': 'orderedList' if ordered else 'bulletList', 'content': items})
            continue
        else:
            require(not line.startswith('```'), 'Có fenced code mới: review bộ chuyển ADF trước.')
            nodes.append(paragraph(line))
        i += 1
    return {'type': 'doc', 'version': 1, 'content': nodes}


def adf_text(value):
    """So nội dung theo đoạn/cell, không coi localId/format do Jira thêm là lệch nghiệp vụ."""
    if isinstance(value, dict):
        if value.get('type') == 'text':
            return value.get('text', '')
        chunks = [adf_text(x) for x in value.get('content', [])]
        return ('' if value.get('type') in ('paragraph', 'heading') else '\n').join(chunks)
    return ''


def load_manifest():
    text = SOURCE.read_text()
    source_rows = re.findall(r'^\| (AUTH-T\d+) \| (\[AUTH-T\d+\] .*?) \| (.*?) \| (\d+)h \| (.*?) \| (.*?) \|$', text, re.M)
    require(len(source_rows) == 20, 'Nguồn phải có đúng 20 dòng ticket.')
    wbs = dict(re.findall(r'^\| (P\d{2}\.\d+) \|.*?\| (\d+) \|$', PROPOSAL.read_text(), re.M))
    tasks, mapped = {}, []
    required = ['Mục tiêu', 'Đầu vào và điều kiện bắt đầu', 'Phạm vi thực hiện',
                'Hành vi và tình huống cần đạt', 'Tiêu chí hoàn thành', 'Kiểm tra và bằng chứng',
                'Ngoài phạm vi', 'Tham chiếu']
    for sid, title, ws, hours, deps, blocks in source_rows:
        marker = f'### {title}\n'
        require(text.count(marker) == 1 and title.startswith(f'[{sid}] '), f'{sid}: tiêu đề không duy nhất.')
        body = re.split(r'\n### \[|\n## 5\.', text.split(marker, 1)[1], maxsplit=1)[0].strip()
        require(all(f'**{h}**' in body for h in required) and body.count('- [ ]') >= 3, f'{sid}: thiếu description/AC.')
        depends = re.findall(r'AUTH-T\d+', deps)
        for field, expected in [('Bị chặn bởi', depends), ('Blocks', re.findall(r'AUTH-T\d+', blocks))]:
            match = re.search(rf'^- {field}: (.*)$', body, re.M)
            require(match is not None and re.findall(r'AUTH-T\d+', match[1]) == expected, f'{sid}: lệch {field}.')
        units = re.findall(r'P\d{2}\.\d+', ws)
        require(all(w in wbs for w in units) and sum(int(wbs[w]) for w in units) == int(hours), f'{sid}: lệch giờ WBS.')
        mapped.extend(units)
        require(sid not in tasks, 'Mã ticket bị trùng.')
        tasks[sid] = {'source_id': sid, 'summary': title, 'markdown': body, 'type': 'Task',
                      'depends_on': depends, 'blocks': re.findall(r'AUTH-T\d+', blocks), 'hours_baseline': int(hours),
                      'wbs': units, 'labels': ['auth-session', IMPORT_LABEL, 'source-' + sid.lower()]
                      + (['learn-go'] if sid in ['AUTH-T14', 'AUTH-T15', 'AUTH-T16'] else [])}
    require(len(mapped) == len(set(mapped)) == 33 and set(mapped) == set(wbs), 'Mapping phải đủ 33 WBS đúng một lần.')
    require(set(tasks) == set(ORDER) and len(ORDER) == len(set(ORDER)), 'Thứ tự task thiếu/trùng.')
    positions = {k: i for i, k in enumerate(ORDER)}
    for sid, t in tasks.items():
        require(all(d in tasks and positions[d] < positions[sid] for d in t['depends_on']), f'{sid}: vòng phụ thuộc hoặc Rank sai.')
        require(set(t['blocks']) == {k for k, v in tasks.items() if sid in v['depends_on']}, f'{sid}: hai chiều link lệch.')
    epic_body = text.split('## 2. Epic\n', 1)[1].split('\n## 3.', 1)[0].strip()
    summary = re.search(r'^\*\*Summary:\*\* `([^`]+)`', epic_body, re.M)
    require(summary is not None, 'Thiếu summary Epic.')
    epic = {'source_id': 'AUTH-EPIC', 'summary': summary[1], 'markdown': epic_body.split('\n', 1)[1].strip(),
            'type': 'Epic', 'labels': ['auth-session', IMPORT_LABEL, 'source-auth-epic']}
    records = [epic] + [tasks[k] for k in ORDER]
    return {'site': SITE, 'project': PROJECT, 'board': BOARD, 'order': ORDER, 'records': records,
            'source_hashes': {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest()
                              for p in [SOURCE, PROPOSAL, ROOT / 'docs/api/auth.openapi.json', ROOT / 'docs/product/functional-requirements.md']}}


def render_description(record, source_base=None):
    text = record['markdown']
    # Giải link theo vị trí tài liệu nguồn; không sinh URL tới file chưa push.
    def replace(match):
        label, href = match.groups()
        if re.match(r'^https://', href):
            return match[0]
        target, _, anchor = href.partition('#')
        resolved = (SOURCE.parent / target).resolve()
        require(resolved.is_relative_to(ROOT) and resolved.is_file(), 'Link nguồn không hợp lệ.')
        path = resolved.relative_to(ROOT).as_posix()
        if source_base:
            return f'[{label}]({source_base}/{path}' + (f'#{anchor}' if anchor else '') + ')'
        return f'{label} (`{path}` — link remote chờ bàn giao)'
    text = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', replace, text)
    text += f'\n\n**Truy vết nhập Jira**\n\nMã nguồn ticket: `{record["source_id"]}`. Nguồn: `docs/proposals/backend-auth-session-tickets.md`.\n'
    return adf(text)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


class Jira:
    def __init__(self, email, token, cloud_id=None):
        self.base = SITE
        if cloud_id:
            require(bool(re.fullmatch(r'[a-fA-F0-9-]{36}', cloud_id)), 'Cloud ID không hợp lệ.')
            self.base = f'https://api.atlassian.com/ex/jira/{cloud_id}'
        self.authorization = 'Basic ' + base64.b64encode(f'{email}:{token}'.encode()).decode()
        self.opener = urllib.request.build_opener(NoRedirect)

    def request(self, method, path, body=None):
        board_path = f'/rest/agile/1.0/board/{BOARD}'
        allowed_get = path.startswith('/rest/api/3/') or path == board_path or path.startswith(board_path + '/')
        allowed_write = (method, path) in {('POST', '/rest/api/3/issue'), ('POST', '/rest/api/3/issueLink'),
                                         ('PUT', '/rest/agile/1.0/issue/rank')}
        require((method == 'GET' and allowed_get) or allowed_write, 'Endpoint ngoài phạm vi script.')
        require('://' not in path and path.startswith('/') and '..' not in path, 'Đường dẫn API không hợp lệ.')
        req = urllib.request.Request(self.base + path, method=method,
                                     data=None if body is None else json.dumps(body).encode(),
                                     headers={'Authorization': self.authorization, 'Accept': 'application/json',
                                              'Content-Type': 'application/json', 'Cache-Control': 'no-cache'})
        try:
            with self.opener.open(req, timeout=30) as response:
                expected = 200 if method == 'GET' else (201 if method == 'POST' else 204)
                require(response.status == expected, f'Jira HTTP {response.status}; dừng để đối soát, không tự gửi lại.')
                data = response.read()
                return json.loads(data) if data else None
        except urllib.error.HTTPError as exc:
            raise Stop(f'Jira HTTP {exc.code}; kiểm quyền/token/scope. Không in body lỗi; lệnh ghi có journal để đối soát.') from None
        except (urllib.error.URLError, TimeoutError, OSError, ValueError):
            raise Stop('Mất kết nối/phản hồi không hợp lệ; không tự gửi lại lệnh ghi. Chạy plan để đối soát journal.') from None

    def get(self, path, **query):
        return self.request('GET', path + ('?' + urllib.parse.urlencode(query, doseq=True) if query else ''))

    def pages(self, path, key, **query):
        values, start = [], 0
        while True:
            page = self.get(path, startAt=start, maxResults=100, **query)
            batch = page.get(key)
            require(isinstance(batch, list), 'Metadata phân trang không đúng định dạng.')
            values.extend(batch)
            start += len(batch)
            if page.get('isLast') is True or ('total' in page and start >= page['total']):
                return values
            require(batch and start < 10000, 'Không xác minh được trang metadata cuối.')

    def issue(self, key):
        require(bool(re.fullmatch(r'DATN-\d+', key)), 'Key ngoài project DATN.')
        return self.get('/rest/api/3/issue/' + key, fields=FIELDS)

    def issues(self, rank_field, reconcile=()):
        result, tokens, token = [], set(), None
        while True:
            query = {'jql': f'project = {PROJECT} ORDER BY cf[{rank_field}] ASC', 'maxResults': 100, 'fields': FIELDS}
            if reconcile:
                query['reconcileIssues'] = list(reconcile)
            if token:
                query['nextPageToken'] = token
            page = self.get('/rest/api/3/search/jql', **query)
            require(isinstance(page.get('issues'), list), 'Không đọc được danh sách issue.')
            result.extend(page['issues'])
            token = page.get('nextPageToken')
            if not token:
                require(page.get('isLast') is not False, 'Thiếu nextPageToken.')
                require(len({i['key'] for i in result}) == len(result), 'Issue lặp giữa các trang.')
                return result
            require(token not in tokens, 'Token phân trang lặp.')
            tokens.add(token)


def verify_target(jira):
    server = jira.get('/rest/api/3/serverInfo')
    base_url = server.get('baseUrl')
    require(isinstance(base_url, str) and base_url.rstrip('/') == SITE,
            'Jira trả baseUrl khác site DATN; kiểm lại cloud ID.')
    issue = jira.get(f'/rest/api/3/issue/{VERIFY_ISSUE}', fields='project')
    require(issue.get('key') == VERIFY_ISSUE and issue.get('fields', {}).get('project', {}).get('key') == PROJECT,
            f'Không xác minh được {VERIFY_ISSUE} thuộc project {PROJECT}.')


def board_diagnostics(jira):
    """Chỉ đọc metadata chọn lọc; không in response thô, tên người hoặc credential."""
    def identifier(value):
        if isinstance(value, (str, int)) and not isinstance(value, bool):
            value = str(value)
            if re.fullmatch(r'[A-Za-z0-9_.:-]{1,160}', value):
                return value
        return None

    verify_target(jira)
    board = jira.get(f'/rest/agile/1.0/board/{BOARD}')
    config = jira.get(f'/rest/agile/1.0/board/{BOARD}/configuration')
    response = jira.get(f'/rest/agile/1.0/board/{BOARD}/features')
    features = response.get('features')
    report = {
        'expected': {'site': SITE, 'project': PROJECT, 'board': BOARD},
        'verifiedIssue': VERIFY_ISSUE,
        'board': {k: identifier(board.get(k)) for k in ('id', 'type')},
        'location': {k: identifier((board.get('location') or {}).get(k))
                     for k in ('projectId', 'projectKey')},
        'configuration': {'id': identifier(config.get('id')), 'type': identifier(config.get('type')),
                          'filterId': identifier((config.get('filter') or {}).get('id')),
                          'rankCustomFieldId': identifier((config.get('ranking') or {}).get('rankCustomFieldId'))},
        'featuresIsList': isinstance(features, list),
        'featuresCount': len(features) if isinstance(features, list) else None,
        'features': [{**{k: identifier(f.get(k)) for k in ('boardFeature', 'feature', 'featureId', 'boardId', 'state')},
                      'fieldNames': [identifier(k) for k in f]}
                     if isinstance(f, dict) else {'invalidEntry': True}
                     for f in features[:100]] if isinstance(features, list) else [],
    }
    print(json.dumps(report, ensure_ascii=False, indent=2))
    print('DIAGNOSE: chỉ đọc site, ticket xác minh và metadata board.')


def validate_board(jira, board, config):
    require(str(board.get('id')) == str(BOARD), f'GET board trả ID khác board {BOARD}; dừng.')
    require(str(config.get('id')) == str(BOARD), f'Configuration trả ID khác board {BOARD}; dừng.')
    kind = board.get('type')
    require(kind in ('kanban', 'simple'), 'Loại board không hỗ trợ; chỉ nhận kanban hoặc simple không bật Sprint.')
    if kind == 'simple':
        # Team-managed không phân biệt Kanban/Scrum bằng type; đọc feature theo Atlassian.
        features = jira.get(f'/rest/agile/1.0/board/{BOARD}/features').get('features')
        require(isinstance(features, list), 'Không đọc được features của board team-managed.')
        require(all(isinstance(f, dict) for f in features), 'Features chứa mục sai định dạng; chạy --mode diagnose.')
        require(all(f.get('boardId') is None or str(f['boardId']) == str(BOARD) for f in features),
                f'Features trả board ID khác board {BOARD}.')
        # API thực tế còn dùng feature=jsw.agility.sprints thay enum trong OpenAPI.
        # Chỉ nhận đúng định danh, không tìm chuỗi gần giống hoặc nhãn dịch.
        sprints = []
        for feature in features:
            enum, key = feature.get('boardFeature'), feature.get('feature')
            if enum == 'SPRINTS' or key == 'jsw.agility.sprints':
                require(enum in (None, 'SPRINTS') and key in (None, 'jsw.agility.sprints'),
                        'Định danh Sprint mâu thuẫn giữa boardFeature và feature; dừng.')
                sprints.append(feature)
        require(len(sprints) == 1, 'Chưa xác minh được tính năng SPRINTS; chạy --mode diagnose để lấy metadata chọn lọc. Không suy thiếu SPRINTS là đã tắt.')
        require(sprints[0].get('state') == 'DISABLED', 'Board team-managed đang bật Sprint hoặc trạng thái chưa rõ; cần review trước.')
        print(f'Board {BOARD}: team-managed (API type=simple), SPRINTS=DISABLED; tiếp tục kiểm filter DATN/Rank.')
    else:
        print(f'Board {BOARD}: API type=kanban; tiếp tục kiểm filter DATN/Rank.')


def preflight(jira):
    verify_target(jira)
    user = jira.get('/rest/api/3/myself')
    project = jira.get('/rest/api/3/project/DATN')
    require(project.get('key') == PROJECT, 'Sai project.')
    board = jira.get(f'/rest/agile/1.0/board/{BOARD}')
    config = jira.get(f'/rest/agile/1.0/board/{BOARD}/configuration')
    validate_board(jira, board, config)
    rank = config.get('ranking', {}).get('rankCustomFieldId')
    require(isinstance(rank, int), 'Board chưa có Rank field.')
    filt = jira.get('/rest/api/3/filter/' + str(config['filter']['id']))
    project_selector = r'(?:DATN|\"DATN\"|' + re.escape(str(project['id'])) + r')'
    require(re.fullmatch(r'\s*project\s*=\s*' + project_selector
                         + r'\s+ORDER\s+BY\s+(?:Rank|cf\[' + str(rank) + r'\])\s+ASC\s*',
                         filt.get('jql', ''), re.I),
            'Chỉ hỗ trợ board filter project DATN ORDER BY Rank ASC; filter khác cần review trước.')
    permissions = jira.get('/rest/api/3/mypermissions', projectKey=PROJECT,
                           permissions='BROWSE_PROJECTS,CREATE_ISSUES,ASSIGN_ISSUES,LINK_ISSUES,SCHEDULE_ISSUES,EDIT_ISSUES')['permissions']
    require(all(permissions.get(k, {}).get('havePermission') for k in
                ['BROWSE_PROJECTS', 'CREATE_ISSUES', 'ASSIGN_ISSUES', 'LINK_ISSUES', 'SCHEDULE_ISSUES', 'EDIT_ISSUES']),
            'Thiếu quyền browse/create/assign/link/rank/edit trong DATN.')
    assignees = jira.get('/rest/api/3/user/assignable/search', project=PROJECT, accountId=user['accountId'], maxResults=100)
    require(any(u.get('accountId') == user['accountId'] and u.get('active') for u in assignees), 'Tài khoản hiện tại không assign được trong DATN.')
    types = jira.pages('/rest/api/3/issue/createmeta/DATN/issuetypes', 'issueTypes')
    selected = {}
    for name in ['Epic', 'Task']:
        matches = [t for t in types if t.get('name') == name and not t.get('subtask')]
        require(len(matches) == 1, f'Không có duy nhất loại {name}.')
        selected[name] = matches[0]['id']
        meta = jira.pages('/rest/api/3/issue/createmeta/DATN/issuetypes/' + str(selected[name]), 'fields')
        fields = {f.get('fieldId', f.get('key')): f for f in meta}
        needed = {'project', 'issuetype', 'summary', 'description', 'assignee', 'labels'} | ({'parent'} if name == 'Task' else set())
        require(needed <= set(fields), f'{name}: thiếu field tạo issue/parent.')
        require(not {k for k, f in fields.items() if f.get('required') and k not in needed and not f.get('hasDefaultValue')},
                f'{name}: có trường bắt buộc mới; cần review trước ghi.')
    links = jira.get('/rest/api/3/issueLinkType')['issueLinkTypes']
    blocks = [t for t in links if t.get('outward', '').lower() == 'blocks' and t.get('inward', '').lower() == 'is blocked by']
    require(len(blocks) == 1, 'Không có duy nhất link blocks/is blocked by.')
    return {'user': user['accountId'], 'types': selected, 'link_type': blocks[0]['id'], 'rank': rank,
            'project_id': str(project['id']), 'filter_jql': filt['jql']}


def atomic_json(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    require(not path.is_symlink(), 'Không ghi journal qua symlink.')
    fd, name = tempfile.mkstemp(prefix='.jira-', dir=path.parent)
    try:
        with os.fdopen(fd, 'w') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)
            f.write('\n')
        os.replace(name, path)
    finally:
        if os.path.exists(name):
            os.unlink(name)


@contextlib.contextmanager
def state_lock(path):
    path.parent.mkdir(parents=True, exist_ok=True)
    lock = path.with_suffix('.lock')
    try:
        fd = os.open(lock, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    except FileExistsError:
        raise Stop('Có lượt chạy khác hoặc lock cũ. Xác minh process đã dừng trước khi gỡ lock.') from None
    try:
        os.close(fd)
        yield
    finally:
        lock.unlink()


def load_state(path, fingerprint):
    if not path.exists():
        return {'version': 1, 'site': SITE, 'project': PROJECT, 'board': BOARD, 'fingerprint': fingerprint,
                'mapping': {}, 'pending': None, 'rank_before': None}
    state = json.loads(path.read_text())
    require(state.get('version') == 1 and state.get('fingerprint') == fingerprint
            and state.get('site') == SITE and state.get('project') == PROJECT and state.get('board') == BOARD,
            'Nguồn/cấu hình đã đổi hoặc journal sai. Review journal trước; không xóa để chạy lại mù.')
    require(set(state.get('mapping', {})) <= {'AUTH-EPIC', *ORDER}, 'Mapping có mã lạ.')
    require(all(re.fullmatch(r'DATN-\d+', k) for k in state['mapping'].values())
            and len(set(state['mapping'].values())) == len(state['mapping']), 'Mapping key sai/trùng.')
    return state


def payload(record, meta, parent, source_base):
    fields = {'project': {'key': PROJECT}, 'issuetype': {'id': meta['types'][record['type']]},
              'summary': record['summary'], 'description': render_description(record, source_base),
              'assignee': {'accountId': meta['user']}, 'labels': record['labels']}
    if record['type'] == 'Task':
        require(parent is not None, 'Task chưa có parent Epic.')
        fields['parent'] = {'key': parent}
    # Không gửi status, priority, thời gian, sprint, reporter hoặc field tùy biến ngoài scope.
    return {'fields': fields}


def validate_issue(issue, record, meta, parent, source_base):
    f = issue['fields']
    require(f.get('project', {}).get('key') == PROJECT and str(f.get('issuetype', {}).get('id')) == str(meta['types'][record['type']]),
            f'{issue["key"]}: sai project/type.')
    require(f.get('summary') == record['summary'] and set(f.get('labels', [])) == set(record['labels']), f'{issue["key"]}: lệch title/label.')
    require((f.get('assignee') or {}).get('accountId') == meta['user'], f'{issue["key"]}: assignee đã khác.')
    require(adf_text(f.get('description')) == adf_text(render_description(record, source_base)), f'{issue["key"]}: description khác nguồn; không ghi đè.')
    if record['type'] == 'Task':
        require((f.get('parent') or {}).get('key') == parent, f'{issue["key"]}: sai parent.')


def discover_mapping(jira, manifest, meta, state, source_base, direct_known=True):
    issues = jira.issues(meta['rank'])
    by_key = {i['key']: i for i in issues}
    # Direct GET lấy được issue đã tạo nhưng chưa lên search index.
    for key in state['mapping'].values():
        if direct_known or key not in by_key:
            by_key[key] = jira.issue(key)
    mapping = dict(state['mapping'])
    for record in manifest['records']:
        sid = record['source_id']
        label = 'source-' + sid.lower()
        candidates = [i for i in by_key.values() if label in i['fields'].get('labels', [])
                      or i['fields'].get('summary', '').startswith(f'[{sid}]')
                      or (sid == 'AUTH-EPIC' and i['fields'].get('summary') == record['summary'])]
        require(len(candidates) <= 1, f'{sid}: nhiều ticket trùng; không tạo thêm.')
        if sid in mapping:
            require(candidates and candidates[0]['key'] == mapping[sid], f'{sid}: mapping lệch Jira.')
        elif candidates:
            mapping[sid] = candidates[0]['key']
        if sid in mapping:
            validate_issue(by_key[mapping[sid]], record, meta, mapping.get('AUTH-EPIC'), source_base)
    owned = {i['key'] for i in by_key.values() if IMPORT_LABEL in i['fields'].get('labels', [])}
    require(owned <= set(mapping.values()), 'Có ticket mang label import nhưng không thuộc manifest; dừng.')
    return mapping, by_key


def actual_links(issues, mapping, link_type, allow_external=False):
    reverse = {key: sid for sid, key in mapping.items()}
    edges = set()
    for sid, key in mapping.items():
        for link in issues[key]['fields'].get('issuelinks', []):
            if str(link['type']['id']) != str(link_type):
                continue
            # Trên provider, outwardIssue là consumer và UI hiển thị "blocks".
            if 'outwardIssue' in link:
                other = link['outwardIssue']['key']
                if allow_external and other not in reverse:
                    continue
                require(other in reverse, f'{key}: Blocks tới issue ngoài manifest; review trước.')
                edges.add((sid, reverse[other]))
            if 'inwardIssue' in link:
                other = link['inwardIssue']['key']
                if allow_external and other not in reverse:
                    continue
                require(other in reverse, f'{key}: bị block bởi issue ngoài manifest; review trước.')
                edges.add((reverse[other], sid))
    return edges


def expected_links(manifest):
    return {(d, r['source_id']) for r in manifest['records'][1:] for d in r['depends_on']}


def link_payload(provider, consumer, link_type):
    # Jira trả outwardIssue=consumer khi GET provider; không đảo chiều body.
    return {'type': {'id': link_type}, 'inwardIssue': {'key': provider}, 'outwardIssue': {'key': consumer}}


def wait_read(read, good):
    for delay in [0, 1, 2, 4, 8]:
        if delay:
            time.sleep(delay)
        result = read()
        if good(result):
            return result
    raise Stop('Đọc lại chưa khớp sau 5 lần; không gửi lại mutation. Chạy plan/check để đối soát.')


def mutate(jira, state, state_path, pending, method, path, body):
    require(state['pending'] is None, 'Có mutation chưa đối soát; không ghi tiếp.')
    state['pending'] = pending
    atomic_json(state_path, state)
    return jira.request(method, path, body)


def clear_pending(state, path):
    state['pending'] = None
    atomic_json(path, state)


def report(manifest, mapping, edges):
    print(f'{PROJECT} / board {BOARD}: {len(mapping)}/21 ticket đã đối soát; còn tạo {21-len(mapping)}.')
    print('Epic: ' + manifest['records'][0]['summary'] + ' → ' + mapping.get('AUTH-EPIC', 'chưa tạo'))
    for pos, r in enumerate(manifest['records'][1:], 1):
        sid = r['source_id']
        print(f'{pos:02}. {sid} → {mapping.get(sid, "chưa tạo")} | {r["summary"]}')
        print('    Parent: ' + mapping.get('AUTH-EPIC', 'Epic sẽ tạo') + '; is blocked by: ' + (', '.join(r['depends_on']) or '—'))
    print(f'Blocks: {len(edges)}/{len(expected_links(manifest))}; chỉ tạo một link cho mỗi cặp.')
    print('Không gửi estimate/remaining/status/priority/sprint/deadline. 105h chỉ là cơ sở WBS.')


def run_online(jira, manifest, state_path, mode, source_base=None, retry_create=None):
    fingerprint = digest({'manifest': manifest, 'source_base': source_base})
    state = load_state(state_path, fingerprint)
    meta = preflight(jira)
    if state.get('account'):
        require(state['account'] == meta['user'], 'Tài khoản khác journal; dừng.')
    state['account'] = meta['user']
    context = {k: meta[k] for k in ['user', 'types', 'link_type', 'rank']}
    require(state.get('context', context) == context, 'Metadata Jira khác journal; review trước khi tiếp tục.')
    state['context'] = context
    mapping, issues = discover_mapping(jira, manifest, meta, state, source_base)
    edges = actual_links(issues, mapping, meta['link_type'])
    require(edges <= expected_links(manifest), 'Có Blocks sai/dư/đảo chiều; không tự xóa.')
    state['mapping'] = mapping
    pending = state['pending']
    if pending:
        kind = pending['kind']
        if kind == 'create' and pending['source_id'] in mapping:
            state['pending'] = None
        elif kind == 'link' and tuple(pending['edge']) in edges:
            state['pending'] = None
        elif kind == 'link' and mode == 'apply':
            a, b = pending['edge']
            require(input(f'Đã đọc Jira, chưa thấy link {a} blocks {b}. Gõ RETRY LINK để thử có chủ ý: ') == 'RETRY LINK',
                    'Đã hủy retry link.')
            state['pending'] = None
        elif kind == 'rank':
            # Rank là phép đặt vị trí; chỉ cho chạy lại sau đọc và xác nhận apply mới.
            state['pending'] = None
        elif kind == 'create' and retry_create == pending['source_id'] and mode == 'apply':
            require(input('Đã kiểm Jira và xác nhận chưa tạo. Gõ RETRY ' + retry_create + ': ') == 'RETRY ' + retry_create,
                    'Đã hủy retry create.')
            state['pending'] = None
        else:
            raise Stop('Journal còn mutation chưa rõ kết quả. Chưa tự retry. Nếu create: tìm issue và chờ search index; chỉ --retry-create khi đã xác minh không tồn tại. Link thiếu: kiểm Jira trước khi xử lý journal.')
    report(manifest, mapping, edges)
    if mode == 'plan':
        print('PLAN chỉ đọc; chưa thay Jira hoặc lưu mapping mới.')
        return
    if mode == 'check':
        require(len(mapping) == 21 and edges == expected_links(manifest), 'Chưa đủ ticket/link.')
        verify_rank(jira, manifest, meta, mapping)
        print('CHECK đạt nội dung/parent/assignee/labels/link/Rank; không chứng minh auth runtime.')
        return
    require(input(f'Tạo phần còn thiếu và sắp Rank trên {PROJECT} board {BOARD}. Gõ YES: ') == 'YES', 'Đã hủy apply.')
    atomic_json(state_path, state)
    for record in manifest['records']:
        sid = record['source_id']
        if sid in mapping:
            continue
        # Đối soát ngay trước mỗi POST, không dùng snapshot đầu lượt để quyết định tạo.
        refreshed, _ = discover_mapping(jira, manifest, meta, state, source_base, direct_known=False)
        require(refreshed == mapping, 'Jira thay đổi trong lượt chạy; chạy plan lại.')
        result = mutate(jira, state, state_path, {'kind': 'create', 'source_id': sid}, 'POST', '/rest/api/3/issue',
                        payload(record, meta, mapping.get('AUTH-EPIC'), source_base))
        require(isinstance(result, dict) and re.fullmatch(r'DATN-\d+', result.get('key', '')), 'Create chưa có key hợp lệ; giữ journal.')
        key = result['key']
        mapping[sid] = key
        atomic_json(state_path, state)  # Lưu key trước read-back để không mất dấu nếu GET lỗi.
        validate_issue(jira.issue(key), record, meta, mapping.get('AUTH-EPIC'), source_base)
        clear_pending(state, state_path)
        print(f'{sid}: đã tạo và đọc lại {key}.', flush=True)
    known_edges = set(edges)
    for edge in sorted(expected_links(manifest)):
        if edge in known_edges:
            continue
        def read_pair():
            # Chỉ đọc hai đầu để giảm số request; đối soát toàn bộ lại ở cuối.
            pairs = {mapping[sid]: jira.issue(mapping[sid]) for sid in edge}
            return actual_links(pairs, {sid: mapping[sid] for sid in edge}, meta['link_type'], allow_external=True)
        existing = read_pair()
        require((edge[1], edge[0]) not in existing, 'Link đảo chiều; dừng.')
        if edge not in existing:
            mutate(jira, state, state_path, {'kind': 'link', 'edge': list(edge)}, 'POST', '/rest/api/3/issueLink',
                   link_payload(mapping[edge[0]], mapping[edge[1]], meta['link_type']))
            wait_read(read_pair, lambda es: edge in es)
            clear_pending(state, state_path)
        known_edges.add(edge)
    reorder(jira, manifest, meta, mapping, state, state_path)
    # Đọc lại toàn bộ trước khi tuyên bố hoàn tất.
    mapping2, issues2 = discover_mapping(jira, manifest, meta, state, source_base)
    require(mapping2 == mapping and actual_links(issues2, mapping, meta['link_type']) == expected_links(manifest), 'Đối soát cuối chưa khớp.')
    verify_rank(jira, manifest, meta, mapping)
    print(f'APPLY hoàn tất: 1 Epic + 20 Task, {len(expected_links(manifest))} Blocks, Rank đúng; mapping ở {state_path}.')


def ranked(jira, meta, mapping):
    if '_reconcile_ids' not in meta:
        meta['_reconcile_ids'] = [int(jira.issue(k)['id']) for k in mapping.values()]
    return [i for i in jira.issues(meta['rank'], meta['_reconcile_ids']) if i['key'] in set(mapping.values()) - {mapping['AUTH-EPIC']}]


def verify_rank(jira, manifest, meta, mapping):
    target = [mapping[k] for k in manifest['order']]
    wait_read(lambda: [i['key'] for i in ranked(jira, meta, mapping)], lambda keys: keys == target)


def reorder(jira, manifest, meta, mapping, state, path):
    target = [mapping[k] for k in manifest['order']]
    baseline = {k: jira.issue(k)['fields'] for k in mapping.values()}
    initial = [i['key'] for i in ranked(jira, meta, mapping)]
    require(set(initial) == set(target), 'Search chưa đủ 20 task để rank; dừng.')
    if state['rank_before'] is None:
        state['rank_before'] = initial
        atomic_json(path, state)
    for index, key in enumerate(target):
        current = [i['key'] for i in ranked(jira, meta, mapping)]
        require(current[:index] == target[:index], 'Rank đã bị người/lượt khác thay đổi; dừng.')
        if current[index] == key:
            continue
        body = {'issues': [key], 'rankCustomFieldId': meta['rank'],
                'rankBeforeIssue' if index == 0 else 'rankAfterIssue': current[0] if index == 0 else target[index - 1]}
        mutate(jira, state, path, {'kind': 'rank', 'key': key}, 'PUT', '/rest/agile/1.0/issue/rank', body)
        wait_read(lambda: [i['key'] for i in ranked(jira, meta, mapping)], lambda keys: keys[:index+1] == target[:index+1])
        require(all(jira.issue(k)['fields'] == f for k, f in baseline.items()), 'Field ngoài Rank đổi; không tự khôi phục.')
        clear_pending(state, path)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--mode', choices=['plan', 'apply', 'check', 'diagnose'], default='plan')
    parser.add_argument('--offline', action='store_true', help='Chỉ so nguồn, không mạng/credential.')
    parser.add_argument('--state', type=Path, default=ROOT / '.local/jira-auth/state.json')
    parser.add_argument('--manifest-out', type=Path, help='Xuất bản sinh từ nguồn; không chỉnh tay.')
    parser.add_argument('--source-base', help='URL GitHub blob tới commit đã push, không kết thúc bằng /.')
    parser.add_argument('--cloud-id', help='UUID site nếu dùng token có scope qua api.atlassian.com.')
    parser.add_argument('--retry-create', choices=['AUTH-EPIC', *ORDER], help='Chỉ sau đối soát thủ công mutation chưa rõ.')
    args = parser.parse_args()
    require(not args.offline or args.mode == 'plan', '--offline chỉ dùng với plan.')
    if args.mode == 'diagnose':
        require(not (args.source_base or args.manifest_out or args.retry_create),
                'diagnose chỉ đọc site, ticket xác minh và metadata board.')
        require(sys.stdin.isatty() and sys.stderr.isatty(), 'Chạy trong terminal tương tác để nhập token ẩn; không pipe token.')
        email = input('Email Atlassian: ').strip()
        with warnings.catch_warnings():
            warnings.simplefilter('error', getpass.GetPassWarning)
            token = getpass.getpass('API token (ẩn khi nhập): ')
        require(email and token, 'Email/token không được trống.')
        board_diagnostics(Jira(email, token, args.cloud_id))
        return
    if args.source_base:
        require(bool(re.fullmatch(r'https://github.com/mingquanjp/ai-staffing/blob/[a-f0-9]{40}', args.source_base)), 'source-base phải là commit URL đúng repo.')
        sha = args.source_base.rsplit('/', 1)[1]
        import subprocess
        for file in [SOURCE, PROPOSAL, ROOT / 'docs/api/auth.openapi.json', ROOT / 'docs/product/functional-requirements.md']:
            try:
                committed = subprocess.check_output(['git', 'show', f'{sha}:{file.relative_to(ROOT)}'], cwd=ROOT, stderr=subprocess.DEVNULL)
            except subprocess.CalledProcessError:
                raise Stop('Commit chưa chứa đủ nguồn. Commit/push riêng theo quyền trước khi thêm link.') from None
            require(committed == file.read_bytes(), 'Nguồn local khác commit được dẫn.')
    manifest = load_manifest()
    if args.manifest_out:
        require(not args.manifest_out.exists(), 'Manifest-out đã tồn tại; chọn file mới, không ghi đè nguồn.')
        atomic_json(args.manifest_out, manifest)
        print(f'Manifest sinh từ nguồn: {args.manifest_out}')
    if args.offline:
        print('OFFLINE: 1 Epic + 20 Task; 33 WBS/105h; Rank hợp phụ thuộc; ' + str(len(expected_links(manifest))) + ' links.')
        for n, sid in enumerate(manifest['order'], 1):
            print(f'{n:02}. {sid}')
        for record in manifest['records']:
            render_description(record, args.source_base)
        print('Description đã chuyển ADF. Chưa gọi Jira, chưa tạo/đổi dữ liệu.')
        return
    require(sys.stdin.isatty() and sys.stderr.isatty(), 'Chạy trong terminal tương tác để nhập token ẩn; không pipe token.')
    email = input('Email Atlassian: ').strip()
    with warnings.catch_warnings():
        warnings.simplefilter('error', getpass.GetPassWarning)
        token = getpass.getpass('API token (ẩn khi nhập): ')
    require(email and token, 'Email/token không được trống.')
    jira = Jira(email, token, args.cloud_id)
    with state_lock(args.state):
        run_online(jira, manifest, args.state, args.mode, args.source_base, args.retry_create)


if __name__ == '__main__':
    try:
        main()
    except (Stop, OSError, ValueError, KeyError, getpass.GetPassWarning):
        # Không phát traceback/body/header chứa credential do driver hoặc input lỗi.
        exc = sys.exc_info()[1]
        print('DỪNG: ' + (str(exc) if isinstance(exc, Stop) else 'Input/file/phản hồi không hợp lệ; kiểm tra bằng dữ liệu giả, không in bí mật.'), file=sys.stderr)
        sys.exit(1)
    except (KeyboardInterrupt, EOFError):
        print('\nĐã dừng; nếu có mutation pending hãy chạy plan để đối soát.', file=sys.stderr)
        sys.exit(1)
