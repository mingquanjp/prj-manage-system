"""Điều phối đợt nhập DATN; dùng transport/guard/ADF đã kiểm của auth, không sửa global auth."""
import json
import re

from auth_import import (SITE, PROJECT, BOARD, require, Stop, digest, adf_text,
                         preflight, atomic_json, actual_links as shared_links, expected_links,
                         wait_read, mutate, clear_pending, link_payload)


def description_signature(document):
    # Jira có thể thêm localId/format; giữ nội dung, URL nguồn và cấu trúc bảng để đối soát.
    links, tables = [], []
    def walk(node):
        for mark in node.get('marks', []):
            if mark.get('type') == 'link':
                links.append((node.get('text', ''), mark.get('attrs', {}).get('href')))
        if node.get('type') == 'table':
            tables.append([[cell.get('type') for cell in row.get('content', [])]
                           for row in node.get('content', [])])
        for child in node.get('content', []):
            walk(child)
    if isinstance(document, dict):
        walk(document)
    return adf_text(document), links, tables


def actual_links(issues, mapping, link_type, allow_external=False):
    for key in mapping.values():
        seen = set()
        for link in issues[key]['fields'].get('issuelinks', []):
            if str(link['type']['id']) != str(link_type):
                continue
            for direction in ('outwardIssue', 'inwardIssue'):
                if direction in link:
                    identity = (direction, link[direction]['key'])
                    require(identity not in seen, f'{key}: Blocks trùng cùng chiều; review trước.')
                    seen.add(identity)
    return shared_links(issues, mapping, link_type, allow_external)


class BatchImporter:
    def __init__(self, manifest, render_description):
        self.epic_id = manifest['records'][0]['source_id']
        self.order = manifest['order']
        self.import_label = manifest['import_label']
        self.render_description = render_description

    def load_state(self, path, fingerprint):
        if not path.exists():
            return {'version': 1, 'batch': self.import_label, 'site': SITE, 'project': PROJECT, 'board': BOARD, 'fingerprint': fingerprint,
                    'mapping': {}, 'pending': None, 'rank_before': None}
        state = json.loads(path.read_text())
        require(state.get('version') == 1 and state.get('batch') == self.import_label and state.get('fingerprint') == fingerprint
                and state.get('site') == SITE and state.get('project') == PROJECT and state.get('board') == BOARD,
                'Nguồn/cấu hình đã đổi hoặc journal sai. Review journal trước; không xóa để chạy lại mù.')
        require(set(state.get('mapping', {})) <= {self.epic_id, *self.order}, 'Mapping có mã lạ.')
        require(all(re.fullmatch(r'DATN-\d+', k) for k in state['mapping'].values())
                and len(set(state['mapping'].values())) == len(state['mapping']), 'Mapping key sai/trùng.')
        return state

    def payload(self, record, meta, parent, source_base):
        fields = {'project': {'key': PROJECT}, 'issuetype': {'id': meta['types'][record['type']]},
                  'summary': record['summary'], 'description': self.render_description(record, source_base),
                  'assignee': {'accountId': meta['user']}, 'labels': record['labels']}
        if record['type'] == 'Task':
            require(parent is not None, 'Task chưa có parent Epic.')
            fields['parent'] = {'key': parent}
        # Không gửi status, priority, thời gian, sprint, reporter hoặc field tùy biến ngoài scope.
        return {'fields': fields}

    def validate_issue(self, issue, record, meta, parent, source_base):
        f = issue['fields']
        require(f.get('project', {}).get('key') == PROJECT and str(f.get('issuetype', {}).get('id')) == str(meta['types'][record['type']]),
                f'{issue["key"]}: sai project/type.')
        require(f.get('summary') == record['summary'] and set(f.get('labels', [])) == set(record['labels']), f'{issue["key"]}: lệch title/label.')
        require((f.get('assignee') or {}).get('accountId') == meta['user'], f'{issue["key"]}: assignee đã khác.')
        require(description_signature(f.get('description')) == description_signature(self.render_description(record, source_base)), f'{issue["key"]}: description khác nguồn; không ghi đè.')
        if record['type'] == 'Task':
            require((f.get('parent') or {}).get('key') == parent, f'{issue["key"]}: sai parent.')

    def discover_mapping(self, jira, manifest, meta, state, source_base, direct_known=True):
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
                          or (sid == self.epic_id and i['fields'].get('summary') == record['summary'])]
            require(len(candidates) <= 1, f'{sid}: nhiều ticket trùng; không tạo thêm.')
            if sid in mapping:
                require(candidates and candidates[0]['key'] == mapping[sid], f'{sid}: mapping lệch Jira.')
            elif candidates:
                mapping[sid] = candidates[0]['key']
            if sid in mapping:
                self.validate_issue(by_key[mapping[sid]], record, meta, mapping.get(self.epic_id), source_base)
        owned = {i['key'] for i in by_key.values() if self.import_label in i['fields'].get('labels', [])}
        require(owned <= set(mapping.values()), 'Có ticket mang label import nhưng không thuộc manifest; dừng.')
        return mapping, by_key

    def report(self, manifest, mapping, edges):
        print(f"{PROJECT} / board {BOARD}: {len(mapping)}/{len(manifest['records'])} ticket đã đối soát; còn tạo {len(manifest['records'])-len(mapping)}.")
        print('Epic: ' + manifest['records'][0]['summary'] + ' → ' + mapping.get(self.epic_id, 'chưa tạo'))
        for pos, r in enumerate(manifest['records'][1:], 1):
            sid = r['source_id']
            print(f'{pos:02}. {sid} → {mapping.get(sid, "chưa tạo")} | {r["summary"]}')
            print('    Parent: ' + mapping.get(self.epic_id, 'Epic sẽ tạo') + '; is blocked by: ' + (', '.join(r['depends_on']) or '—'))
        print(f'Blocks: {len(edges)}/{len(expected_links(manifest))}; chỉ tạo một link cho mỗi cặp.')
        print('Không gửi estimate/remaining/status/priority/sprint/deadline; không tự chuyển Done.')

    def run_online(self, jira, manifest, state_path, mode, source_base=None, retry_create=None):
        require(mode in ('plan', 'apply', 'check'), 'Mode ngoài phạm vi nhập Jira.')
        fingerprint = digest({'manifest': manifest, 'source_base': source_base})
        state = self.load_state(state_path, fingerprint)
        meta = preflight(jira)
        if state.get('account'):
            require(state['account'] == meta['user'], 'Tài khoản khác journal; dừng.')
        state['account'] = meta['user']
        context = {k: meta[k] for k in ['user', 'types', 'link_type', 'rank', 'project_id', 'filter_jql'] if k in meta}
        require(state.get('context', context) == context, 'Metadata Jira khác journal; review trước khi tiếp tục.')
        state['context'] = context
        mapping, issues = self.discover_mapping(jira, manifest, meta, state, source_base)
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
        self.report(manifest, mapping, edges)
        if mode == 'plan':
            print('PLAN chỉ đọc; chưa thay Jira hoặc lưu mapping mới.')
            return
        if mode == 'check':
            require(len(mapping) == len(manifest['records']) and edges == expected_links(manifest), 'Chưa đủ ticket/link.')
            self.verify_rank(jira, manifest, meta, mapping)
            print('CHECK đạt nội dung/parent/assignee/labels/link/Rank; không chứng minh ứng dụng runtime.')
            return
        require(input(f'Tạo phần còn thiếu và sắp Rank trên {PROJECT} board {BOARD}. Gõ YES: ') == 'YES', 'Đã hủy apply.')
        atomic_json(state_path, state)
        for record in manifest['records']:
            sid = record['source_id']
            if sid in mapping:
                continue
            # Đối soát ngay trước mỗi POST, không dùng snapshot đầu lượt để quyết định tạo.
            refreshed, _ = self.discover_mapping(jira, manifest, meta, state, source_base, direct_known=False)
            require(refreshed == mapping, 'Jira thay đổi trong lượt chạy; chạy plan lại.')
            result = mutate(jira, state, state_path, {'kind': 'create', 'source_id': sid}, 'POST', '/rest/api/3/issue',
                            self.payload(record, meta, mapping.get(self.epic_id), source_base))
            require(isinstance(result, dict) and re.fullmatch(r'DATN-\d+', result.get('key', '')), 'Create chưa có key hợp lệ; giữ journal.')
            key = result['key']
            mapping[sid] = key
            atomic_json(state_path, state)  # Lưu key trước read-back để không mất dấu nếu GET lỗi.
            self.validate_issue(jira.issue(key), record, meta, mapping.get(self.epic_id), source_base)
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
        self.reorder(jira, manifest, meta, mapping, state, state_path)
        # Đọc lại toàn bộ trước khi tuyên bố hoàn tất.
        mapping2, issues2 = self.discover_mapping(jira, manifest, meta, state, source_base)
        require(mapping2 == mapping and actual_links(issues2, mapping, meta['link_type']) == expected_links(manifest), 'Đối soát cuối chưa khớp.')
        self.verify_rank(jira, manifest, meta, mapping)
        print(f"APPLY hoàn tất: 1 Epic + {len(manifest['order'])} Task, {len(expected_links(manifest))} Blocks, Rank đúng; mapping ở {state_path}.")

    def ranked(self, jira, meta, mapping):
        if '_reconcile_ids' not in meta:
            meta['_reconcile_ids'] = [int(jira.issue(k)['id']) for k in mapping.values()]
        return [i for i in jira.issues(meta['rank'], meta['_reconcile_ids']) if i['key'] in set(mapping.values()) - {mapping[self.epic_id]}]

    def verify_rank(self, jira, manifest, meta, mapping):
        target = [mapping[k] for k in manifest['order']]
        wait_read(lambda: [i['key'] for i in self.ranked(jira, meta, mapping)], lambda keys: keys == target)

    def reorder(self, jira, manifest, meta, mapping, state, path):
        target = [mapping[k] for k in manifest['order']]
        baseline = {k: jira.issue(k)['fields'] for k in mapping.values()}
        initial = [i['key'] for i in self.ranked(jira, meta, mapping)]
        require(set(initial) == set(target), 'Search chưa đủ task của đợt để rank; dừng.')
        if state['rank_before'] is None:
            state['rank_before'] = initial
            atomic_json(path, state)
        for index, key in enumerate(target):
            current = [i['key'] for i in self.ranked(jira, meta, mapping)]
            require(current[:index] == target[:index], 'Rank đã bị người/lượt khác thay đổi; dừng.')
            if current[index] == key:
                continue
            body = {'issues': [key], 'rankCustomFieldId': meta['rank'],
                    'rankBeforeIssue' if index == 0 else 'rankAfterIssue': current[0] if index == 0 else target[index - 1]}
            mutate(jira, state, path, {'kind': 'rank', 'key': key}, 'PUT', '/rest/agile/1.0/issue/rank', body)
            wait_read(lambda: [i['key'] for i in self.ranked(jira, meta, mapping)], lambda keys: keys[:index+1] == target[:index+1])
            require(all(jira.issue(k)['fields'] == f for k, f in baseline.items()), 'Field ngoài Rank đổi; không tự khôi phục.')
            clear_pending(state, path)
