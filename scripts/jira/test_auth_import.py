"""Kiểm thử offline bằng Jira giả; không dùng token hoặc mạng thật."""
import copy
import io
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import urllib.error

import auth_import as app

META = {'user': 'test-user', 'types': {'Epic': '1', 'Task': '2'}, 'link_type': '3', 'rank': 10019}


class FakeJira:
    def __init__(self):
        self.data = {}
        self.order = []
        self.writes = []
        self.lose_create_response = False
        self.search_hidden = False
        self.fail_link = False

    def issue(self, key):
        return copy.deepcopy(self.data[key])

    def issues(self, rank, reconcile=()):
        return [] if self.search_hidden else [self.issue(k) for k in self.order]

    def request(self, method, path, body):
        self.writes.append((method, path, copy.deepcopy(body)))
        if path == '/rest/api/3/issue':
            key = 'DATN-' + str(len(self.data) + 1)
            f = copy.deepcopy(body['fields'])
            f['status'] = {'name': 'To Do'}
            f['priority'] = {'name': 'Medium'}
            f['issuelinks'] = []
            self.data[key] = {'id': str(len(self.data) + 1), 'key': key, 'fields': f}
            self.order.insert(0, key)
            if self.lose_create_response:
                self.lose_create_response = False
                raise app.Stop('Phản hồi create bị mất sau khi server đã lưu.')
            return {'key': key, 'id': self.data[key]['id']}
        if path == '/rest/api/3/issueLink':
            a, b = body['inwardIssue']['key'], body['outwardIssue']['key']
            link_type = {'id': body['type']['id'], 'outward': 'blocks', 'inward': 'is blocked by'}
            self.data[a]['fields']['issuelinks'].append({'type': link_type, 'outwardIssue': {'key': b}})
            self.data[b]['fields']['issuelinks'].append({'type': link_type, 'inwardIssue': {'key': a}})
            if self.fail_link:
                self.fail_link = False
                raise app.Stop('Mất phản hồi link sau khi server đã ghi.')
            return None
        if path == '/rest/agile/1.0/issue/rank':
            key = body['issues'][0]
            self.order.remove(key)
            if 'rankAfterIssue' in body:
                index = self.order.index(body['rankAfterIssue']) + 1
            else:
                index = self.order.index(body['rankBeforeIssue'])
            self.order.insert(index, key)
            return None
        raise AssertionError(path)


class ImportTests(unittest.TestCase):
    def setUp(self):
        self.manifest = app.load_manifest()
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.state_path = Path(self.tmp.name) / 'state.json'
        self.jira = FakeJira()

    def run_mode(self, mode):
        with patch.object(app, 'preflight', return_value=copy.deepcopy(META)), patch('builtins.input', return_value='YES'), patch('sys.stdout', new=io.StringIO()):
            return app.run_online(self.jira, self.manifest, self.state_path, mode)

    def test_real_source_mapping_order_and_links(self):
        self.assertEqual(len(self.manifest['records']), 21)
        self.assertEqual(len(app.expected_links(self.manifest)), 47)
        self.assertEqual(sum(r['hours_baseline'] for r in self.manifest['records'][1:]), 105)
        self.assertLess(app.ORDER.index('AUTH-T15'), app.ORDER.index('AUTH-T14'))

    def test_adf_preserves_headings_tables_criteria_and_all_text(self):
        for record in self.manifest['records']:
            doc = app.render_description(record)
            self.assertEqual(doc['type'], 'doc')
            self.assertIn(record['source_id'], app.adf_text(doc))
            if record['type'] == 'Task':
                self.assertTrue(any(n['type'] == 'table' for n in doc['content']))
                self.assertIn('☐', app.adf_text(doc))
                self.assertIn('Tiêu chí hoàn thành', app.adf_text(doc))
            self.assertNotIn('](backend-auth', app.adf_text(doc))

    def test_payload_allowlist_and_no_estimate_or_permission_fields(self):
        record = self.manifest['records'][1]
        body = app.payload(record, META, 'DATN-1', None)
        self.assertEqual(set(body['fields']), {'project', 'issuetype', 'summary', 'description', 'assignee', 'labels', 'parent'})
        self.assertEqual(body['fields']['parent']['key'], 'DATN-1')

    def test_offline_plan_online_plan_and_check_never_mutate(self):
        self.run_mode('plan')
        self.assertFalse(self.jira.writes)
        self.assertFalse(self.state_path.exists())
        with self.assertRaises(app.Stop):
            self.run_mode('check')
        self.assertFalse(self.jira.writes)

    def test_apply_creates_21_issues_47_links_ranks_then_rerun_no_writes(self):
        self.run_mode('apply')
        self.assertEqual(sum(p == '/rest/api/3/issue' for _, p, _ in self.jira.writes), 21)
        self.assertEqual(sum(p == '/rest/api/3/issueLink' for _, p, _ in self.jira.writes), 47)
        state = json.loads(self.state_path.read_text())
        self.assertEqual(len(state['mapping']), 21)
        self.assertIsNone(state['pending'])
        self.assertIsNotNone(state['rank_before'])
        self.assertEqual(self.state_path.stat().st_mode & 0o777, 0o600)
        before = len(self.jira.writes)
        self.run_mode('apply')
        self.run_mode('check')
        self.assertEqual(len(self.jira.writes), before)

    def test_lost_create_response_found_on_resume_not_created_twice(self):
        self.jira.lose_create_response = True
        with self.assertRaises(app.Stop):
            self.run_mode('apply')
        self.assertEqual(json.loads(self.state_path.read_text())['pending']['kind'], 'create')
        self.run_mode('apply')
        self.assertEqual(len(self.jira.data), 21)
        self.assertEqual(sum(p == '/rest/api/3/issue' for _, p, _ in self.jira.writes), 21)

    def test_unknown_create_not_visible_stops_not_blind_retry(self):
        self.jira.lose_create_response = True
        with self.assertRaises(app.Stop):
            self.run_mode('apply')
        self.jira.search_hidden = True
        with self.assertRaisesRegex(app.Stop, 'chưa rõ'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), 1)

    def test_lost_link_response_read_before_resume(self):
        self.jira.fail_link = True
        with self.assertRaises(app.Stop):
            self.run_mode('apply')
        self.run_mode('apply')
        self.assertEqual(sum(p == '/rest/api/3/issueLink' for _, p, _ in self.jira.writes), 47)

    def test_duplicate_ticket_blocks_before_any_new_write(self):
        record = self.manifest['records'][0]
        for _ in range(2):
            self.jira.request('POST', '/rest/api/3/issue', app.payload(record, META, None, None))
        before = len(self.jira.writes)
        with self.assertRaisesRegex(app.Stop, 'trùng'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), before)

    def test_content_drift_never_overwritten(self):
        self.run_mode('apply')
        self.jira.data['DATN-2']['fields']['description'] = app.adf('Đã sửa tay')
        before = len(self.jira.writes)
        with self.assertRaisesRegex(app.Stop, 'description khác'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), before)

    def test_missing_assignee_or_parent_stops_cleanly(self):
        self.run_mode('apply')
        for field in ['assignee', 'parent']:
            before = copy.deepcopy(self.jira.data['DATN-2']['fields'])
            self.jira.data['DATN-2']['fields'][field] = None
            with self.assertRaises(app.Stop):
                self.run_mode('plan')
            self.jira.data['DATN-2']['fields'] = before

    def test_link_failed_before_write_needs_explicit_retry_confirmation(self):
        real_request = self.jira.request
        def fail_before_link(method, path, body):
            if path == '/rest/api/3/issueLink':
                raise app.Stop('Chưa ghi link; mô phỏng lỗi mạng trước gửi.')
            return real_request(method, path, body)
        with patch.object(self.jira, 'request', side_effect=fail_before_link):
            with self.assertRaises(app.Stop):
                self.run_mode('apply')
        with self.assertRaisesRegex(app.Stop, 'hủy retry link'):
            self.run_mode('apply')  # YES không thay được xác nhận RETRY LINK.
        with patch.object(app, 'preflight', return_value=copy.deepcopy(META)), patch('builtins.input', side_effect=['RETRY LINK', 'YES']), patch('sys.stdout', new=io.StringIO()):
            app.run_online(self.jira, self.manifest, self.state_path, 'apply')
        self.assertEqual(sum(p == '/rest/api/3/issueLink' for _, p, _ in self.jira.writes), 47)

    def test_source_digest_change_stops_before_remote_write(self):
        self.run_mode('apply')
        self.manifest['records'][1]['markdown'] += '\nKhác nguồn'
        with self.assertRaisesRegex(app.Stop, 'Nguồn/cấu hình'):
            self.run_mode('apply')

    def test_reversed_and_external_blocks_rejected(self):
        self.run_mode('apply')
        state = json.loads(self.state_path.read_text())
        mapping = state['mapping']
        provider, consumer = mapping['AUTH-T01'], mapping['AUTH-T02']
        body = app.link_payload(consumer, provider, META['link_type'])
        self.jira.request('POST', '/rest/api/3/issueLink', body)
        with self.assertRaisesRegex(app.Stop, 'sai/dư/đảo chiều'):
            self.run_mode('plan')
        self.jira.data[provider]['fields']['issuelinks'].append({'type': {'id': '3'}, 'outwardIssue': {'key': 'SCRUM-1'}})
        with self.assertRaisesRegex(app.Stop, 'ngoài manifest'):
            self.run_mode('plan')

    def test_lock_prevents_concurrent_apply(self):
        with app.state_lock(self.state_path):
            with self.assertRaisesRegex(app.Stop, 'lượt chạy khác'):
                with app.state_lock(self.state_path):
                    pass

    def test_read_retries_only(self):
        values = iter([False, False, True])
        with patch.object(app.time, 'sleep'):
            self.assertTrue(app.wait_read(lambda: next(values), bool))
        self.assertEqual(app.link_payload('DATN-2', 'DATN-3', '3'),
                         {'type': {'id': '3'}, 'inwardIssue': {'key': 'DATN-2'}, 'outwardIssue': {'key': 'DATN-3'}})


class TransportTests(unittest.TestCase):
    def test_allowlist_blocks_deletion_updates_and_foreign_urls(self):
        jira = app.Jira('test@example.invalid', 'SECRET_MARKER')
        for method, path in [('DELETE', '/rest/api/3/issue/DATN-1'), ('PUT', '/rest/api/3/issue/DATN-1'),
                             ('POST', '/rest/api/3/user'), ('GET', 'https://other.invalid')]:
            with self.assertRaises(app.Stop):
                jira.request(method, path)

    def test_http_errors_do_not_echo_secrets_or_retry(self):
        jira = app.Jira('test@example.invalid', 'SECRET_MARKER')
        error = urllib.error.HTTPError('https://example.invalid/SECRET_MARKER', 429, 'SECRET_MARKER', {}, None)
        with patch.object(jira.opener, 'open', side_effect=error) as request:
            with self.assertRaises(app.Stop) as caught:
                jira.request('POST', '/rest/api/3/issue', {'fields': {}})
            self.assertNotIn('SECRET_MARKER', str(caught.exception))
            self.assertEqual(request.call_count, 1)

    def test_network_unknown_write_no_retry(self):
        jira = app.Jira('test@example.invalid', 'SECRET_MARKER')
        with patch.object(jira.opener, 'open', side_effect=TimeoutError('SECRET_MARKER')) as request:
            with self.assertRaises(app.Stop) as caught:
                jira.request('POST', '/rest/api/3/issue', {'fields': {}})
            self.assertNotIn('SECRET_MARKER', str(caught.exception))
            self.assertEqual(request.call_count, 1)

    def test_redirect_disabled(self):
        self.assertIsNone(app.NoRedirect().redirect_request(None, None, 302, '', {}, 'https://other.invalid'))

    def test_metadata_missing_pagination_not_treated_complete(self):
        jira = app.Jira('test@example.invalid', 'SECRET_MARKER')
        with patch.object(jira, 'get', return_value={'fields': []}):
            with self.assertRaises(app.Stop):
                jira.pages('/rest/api/3/issue/createmeta/DATN/issuetypes/2', 'fields')


class PreflightTests(unittest.TestCase):
    def metadata_jira(self, jql='project = DATN ORDER BY Rank ASC', missing_parent=False, allowed=True):
        jira = app.Jira('test@example.invalid', 'SECRET_MARKER')
        def get(path, **query):
            if path.endswith('/serverInfo'):
                return {'baseUrl': app.SITE}
            if path.endswith('/issue/DATN-12'):
                return {'key': 'DATN-12', 'fields': {'project': {'key': 'DATN'}}}
            if path.endswith('/myself'):
                return {'accountId': 'test-user'}
            if path.endswith('/project/DATN'):
                return {'id': '10001', 'key': 'DATN'}
            if path.endswith('/board/1/configuration'):
                return {'id': 1, 'ranking': {'rankCustomFieldId': 10019}, 'filter': {'id': '10'}}
            if path.endswith('/board/1'):
                return {'id': 1, 'type': 'kanban'}
            if path.endswith('/filter/10'):
                return {'jql': jql}
            if path.endswith('/mypermissions'):
                return {'permissions': {k: {'havePermission': allowed} for k in query['permissions'].split(',')}}
            if path.endswith('/user/assignable/search'):
                return [{'accountId': 'test-user', 'active': True}]
            if path.endswith('/issueLinkType'):
                return {'issueLinkTypes': [{'id': '3', 'outward': 'blocks', 'inward': 'is blocked by'}]}
            if path.endswith('/issuetypes'):
                return {'total': 2, 'issueTypes': [{'id': '1', 'name': 'Epic'}, {'id': '2', 'name': 'Task'}]}
            if '/issuetypes/' in path:
                fields = ['project', 'issuetype', 'summary', 'description', 'assignee', 'labels']
                if path.endswith('/2') and not missing_parent:
                    fields.append('parent')
                return {'total': len(fields), 'fields': [{'fieldId': f, 'required': True} for f in fields]}
            raise AssertionError(path)
        jira.get = get
        return jira

    def test_metadata_and_assignability_with_no_guessed_ids(self):
        self.assertEqual(app.preflight(self.metadata_jira())['rank'], 10019)
        self.assertEqual(app.preflight(self.metadata_jira(jql='project = 10001 ORDER BY Rank ASC'))['project_id'], '10001')

    def test_wrong_board_filter_missing_parent_or_permission_stop(self):
        for kwargs in [{'jql': 'project = SCRUM ORDER BY Rank ASC'}, {'missing_parent': True}, {'allowed': False}]:
            with self.assertRaises(app.Stop):
                app.preflight(self.metadata_jira(**kwargs))

    def test_preflight_stops_before_metadata_if_site_or_issue_is_wrong(self):
        jira = self.metadata_jira()
        with patch.object(jira, 'get', return_value={'baseUrl': 'https://other.atlassian.net'}) as get:
            with self.assertRaisesRegex(app.Stop, 'baseUrl'):
                app.preflight(jira)
            get.assert_called_once_with('/rest/api/3/serverInfo')

        wrong_issue = {'key': 'DATN-12', 'fields': {'project': {'key': 'OTHER'}}}
        with patch.object(jira, 'get', side_effect=[{'baseUrl': app.SITE}, wrong_issue]) as get:
            with self.assertRaisesRegex(app.Stop, 'DATN-12'):
                app.preflight(jira)
            self.assertEqual(get.call_count, 2)

    def test_simple_board_without_sprints_accepted(self):
        jira = self.metadata_jira()
        original = jira.get
        def get(path, **query):
            if path.endswith('/board/1'):
                return {'id': 1, 'type': 'simple'}
            if path.endswith('/board/1/features'):
                return {'features': [{'boardFeature': 'SPRINTS', 'boardId': 1, 'state': 'DISABLED'}]}
            return original(path, **query)
        jira.get = get
        with patch('sys.stdout', new=io.StringIO()):
            self.assertEqual(app.preflight(jira)['rank'], 10019)

    def test_simple_does_not_bypass_sprints_ids_or_missing_features(self):
        jira = self.metadata_jira()
        board = {'id': 1, 'type': 'simple'}
        config = {'id': 1}
        cases = [
            {'features': [{'boardFeature': 'SPRINTS', 'boardId': 1, 'state': 'ENABLED'}]},
            {'features': []},
            {'features': [{'featureId': 'sprints', 'state': 'DISABLED'}]},
            {'features': [None]},
            {'features': [{'feature': 'jsw.agility.sprints', 'state': 'ENABLED'}]},
            {'features': [{'feature': 'jsw.agility.sprints', 'state': 'COMING_SOON'}]},
            {'features': [{'feature': 'jsw.agility.sprints', 'state': 'DISABLED', 'boardId': 2}]},
            {'features': [{'feature': 'unknown.sprints', 'state': 'DISABLED'}]},
            {'features': [{'boardFeature': 'BACKLOG', 'feature': 'jsw.agility.sprints', 'state': 'DISABLED'}]},
            {'features': [{'boardFeature': 'SPRINTS', 'feature': 'jsw.agility.backlog', 'state': 'DISABLED'}]},
            {'features': [{'boardFeature': 'SPRINTS', 'state': 'DISABLED'},
                          {'feature': 'jsw.agility.sprints', 'state': 'DISABLED'}]},
            {'features': [{'boardFeature': 'SPRINTS', 'boardId': 2, 'state': 'DISABLED'}]},
            {'features': [{'boardFeature': 'SPRINTS', 'boardId': 1, 'state': 'COMING_SOON'}]},
        ]
        for response in cases:
            with patch.object(jira, 'get', return_value=response):
                with self.assertRaises(app.Stop):
                    app.validate_board(jira, board, config)
        for b, c in [({'id': 2, 'type': 'kanban'}, config), (board, {'id': 2}), ({'id': 1, 'type': 'scrum'}, config)]:
            with self.assertRaises(app.Stop):
                app.validate_board(jira, b, c)

    def test_simple_runtime_feature_key_accepted_with_other_features(self):
        jira = self.metadata_jira()
        original = jira.get
        def get(path, **query):
            if path.endswith('/board/1'):
                return {'id': 1, 'type': 'simple'}
            if path.endswith('/board/1/features'):
                return {'features': [
                    {'feature': 'jsw.agility.board', 'boardId': 1, 'state': 'ENABLED'},
                    {'feature': 'jsw.agility.sprints', 'boardId': 1, 'state': 'DISABLED'},
                ]}
            return original(path, **query)
        jira.get = get
        with patch('sys.stdout', new=io.StringIO()):
            self.assertEqual(app.preflight(jira)['rank'], 10019)

    def test_board_diagnostics_selects_fields_without_writes_or_raw_response(self):
        jira = self.metadata_jira()
        responses = [
            {'baseUrl': app.SITE},
            {'key': 'DATN-12', 'fields': {'project': {'key': 'DATN'}}},
            {'id': 1, 'type': 'simple', 'location': {'projectId': '10001', 'projectKey': 'DATN'},
             'secret': 'SECRET_MARKER', 'name': 'PRIVATE_NAME'},
            {'id': 1, 'type': 'simple', 'filter': {'id': '10002'},
             'ranking': {'rankCustomFieldId': 10019}},
            {'features': [{'featureId': 'example.sprints', 'feature': 'jsw.agility.sprints', 'state': 'DISABLED',
                           'localisedDescription': 'SECRET_MARKER'}, None]},
        ]
        output = io.StringIO()
        with patch.object(jira, 'get', side_effect=responses) as get, \
             patch.object(jira, 'request') as request, patch('sys.stdout', new=output):
            app.board_diagnostics(jira)
        self.assertEqual([c.args[0] for c in get.call_args_list], [
            '/rest/api/3/serverInfo', '/rest/api/3/issue/DATN-12',
            '/rest/agile/1.0/board/1', '/rest/agile/1.0/board/1/configuration',
            '/rest/agile/1.0/board/1/features'])
        self.assertIn('https://datn-mingquanjp.atlassian.net', output.getvalue())
        request.assert_not_called()
        self.assertIn('example.sprints', output.getvalue())
        self.assertIn('jsw.agility.sprints', output.getvalue())
        self.assertIn('fieldNames', output.getvalue())
        self.assertIn('invalidEntry', output.getvalue())
        self.assertNotIn('SECRET_MARKER', output.getvalue())
        self.assertNotIn('PRIVATE_NAME', output.getvalue())

    def test_diagnose_bypasses_manifest_journal_and_preflight(self):
        with patch('sys.argv', ['auth_import.py', '--mode', 'diagnose']), \
             patch('sys.stdin.isatty', return_value=True), patch('sys.stderr.isatty', return_value=True), \
             patch('builtins.input', return_value='example@example.invalid'), \
             patch.object(app.getpass, 'getpass', return_value='SECRET_MARKER'), \
             patch.object(app, 'Jira') as factory, patch.object(app, 'board_diagnostics') as diagnose, \
             patch.object(app, 'load_manifest') as manifest, patch.object(app, 'state_lock') as lock, \
             patch.object(app, 'run_online') as online:
            app.main()
        diagnose.assert_called_once_with(factory.return_value)
        manifest.assert_not_called()
        lock.assert_not_called()
        online.assert_not_called()

    def test_rank_207_stops_without_retry(self):
        jira = app.Jira('test@example.invalid', 'SECRET_MARKER')
        response = unittest.mock.MagicMock()
        response.__enter__.return_value.status = 207
        with patch.object(jira.opener, 'open', return_value=response) as request:
            with self.assertRaisesRegex(app.Stop, '207'):
                jira.request('PUT', '/rest/agile/1.0/issue/rank', {'issues': ['DATN-1']})
            self.assertEqual(request.call_count, 1)


if __name__ == '__main__':
    unittest.main()
