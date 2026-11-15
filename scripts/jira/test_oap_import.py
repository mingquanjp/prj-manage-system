"""Kiểm nguồn OAP và điều phối bằng Jira giả; không token/mạng thật."""
import copy
import io
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import auth_import
import jira_batch
import oap_import as app
from test_auth_import import FakeJira, META


class OapImportTests(unittest.TestCase):
    def setUp(self):
        self.manifest = app.load_manifest()
        self.engine = jira_batch.BatchImporter(self.manifest, app.render_description)
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.state = Path(self.tmp.name) / 'state.json'
        self.jira = FakeJira()

    def run_mode(self, mode, answers=None):
        answer = patch('builtins.input', side_effect=answers) if answers else patch('builtins.input', return_value='YES')
        with patch.object(jira_batch, 'preflight', return_value=copy.deepcopy(META)), answer, patch('sys.stdout', new=io.StringIO()):
            self.engine.run_online(self.jira, self.manifest, self.state, mode)

    def test_real_source_has_14_records_15_wbs_17_links_without_estimates(self):
        self.assertEqual(len(self.manifest['records']), 14)
        self.assertEqual(len(app.expected_links(self.manifest)), 17)
        ws = [w for record in self.manifest['records'][1:] for w in record['wbs']]
        self.assertEqual(len(ws), 15)
        self.assertEqual(set(ws), {f'OAP-{n:02}' for n in range(1, 16)})
        self.assertTrue(all('hours_baseline' not in r for r in self.manifest['records']))
        self.assertEqual(self.manifest['records'][0]['source_id'], 'OAP-EPIC')
        self.assertIn(app.SQL_PLAN.relative_to(app.ROOT).as_posix(), self.manifest['source_hashes'])

    def test_descriptions_preserve_two_column_tables_checkboxes_and_sources(self):
        for record in self.manifest['records']:
            doc = app.render_description(record)
            text = auth_import.adf_text(doc)
            self.assertIn(record['source_id'], text)
            self.assertEqual(text.count('Mã nguồn ticket:'), 1)
            self.assertIn('☐', text)
            self.assertNotIn('AUTH-', text)
            if record['type'] == 'Task':
                tables = [n for n in doc['content'] if n['type'] == 'table']
                expected_widths = [3, 2] if record['source_id'] == 'OAP-T04' else [2]
                self.assertEqual(len(tables), len(expected_widths))
                for table, width in zip(tables, expected_widths):
                    self.assertTrue(all(len(row['content']) == width for row in table['content']))
                self.assertIn('Đầu ra bàn giao', text)
                self.assertIn('link remote chờ bàn giao', text)

    def test_rendered_tickets_include_checks_endpoints_and_navigation_checkpoint(self):
        rendered = {r['source_id']: auth_import.adf_text(app.render_description(r))
                    for r in self.manifest['records'][1:]}
        for sid, text in rendered.items():
            self.assertIn('node scripts/check-harness.mjs', text, sid)
            self.assertIn('git diff --check', text, sid)
        for number in (2, 3, 4, 6, 8, 9, 10, 11, 13):
            text = rendered[f'OAP-T{number:02}']
            for command in ('go vet ./...', 'go test ./... -count=1', 'go build ./...'):
                self.assertIn(command, text)
        for number in (2, 3, 4, 6, 8, 9, 11, 13):
            text = rendered[f'OAP-T{number:02}']
            self.assertIn('python3 scripts/authdb/manage.py check', text)
            self.assertIn('run -- go test -tags=integration -p 1', text)
            self.assertIn('Không fallback DATABASE_URL/.env/Neon', text)
        for number in (5, 7, 12, 13):
            self.assertIn('pnpm check', rendered[f'OAP-T{number:02}'])
        for collection in ('departments', 'positions', 'job-levels'):
            self.assertIn(f'/api/{collection}', rendered['OAP-T04'])
            self.assertIn(f'/api/{collection}/{{id}}', rendered['OAP-T04'])
        self.assertIn('GET và POST', rendered['OAP-T04'])
        self.assertIn('PATCH và DELETE', rendered['OAP-T04'])
        self.assertIn('Checkpoint trước sửa UI', rendered['OAP-T05'])
        self.assertIn('checkpoint kỹ thuật chưa hoàn tất', rendered['OAP-T07'])
        self.assertNotIn('theo brief đã chốt', rendered['OAP-T07'])

    def test_description_remote_links_point_to_same_commit(self):
        base = 'https://github.com/mingquanjp/ai-staffing/blob/' + 'a'*40
        doc = app.render_description(self.manifest['records'][1], base)
        links = []
        def walk(node):
            links.extend(mark['attrs']['href'] for mark in node.get('marks', []) if mark['type'] == 'link')
            for child in node.get('content', []):
                walk(child)
        walk(doc)
        self.assertTrue(links)
        self.assertTrue(all(link.startswith(base + '/') for link in links))
        self.assertTrue(any(link.endswith('backend-organization-accounts-profile-tickets.md') for link in links))

    def test_plan_and_incomplete_check_do_not_write_state_or_jira(self):
        self.run_mode('plan')
        self.assertFalse(self.state.exists())
        with self.assertRaises(app.Stop):
            self.run_mode('check')
        self.assertFalse(self.jira.writes)
        self.assertFalse(self.state.exists())

    def test_apply_is_idempotent_and_only_sends_allowed_fields(self):
        self.run_mode('apply')
        self.assertEqual(len(self.jira.data), 14)
        self.assertEqual(sum(path.endswith('/issueLink') for _, path, _ in self.jira.writes), 17)
        for method, path, body in self.jira.writes:
            if path == '/rest/api/3/issue':
                self.assertLessEqual(set(body['fields']), {'project', 'issuetype', 'summary', 'description', 'assignee', 'labels', 'parent'})
                self.assertIn('datn-oap-v1', body['fields']['labels'])
                self.assertNotIn('datn-auth-v1', body['fields']['labels'])
        self.assertEqual(self.state.stat().st_mode & 0o777, 0o600)
        state = json.loads(self.state.read_text())
        self.assertEqual(state['batch'], 'datn-oap-v1')
        self.assertIsNone(state['pending'])
        self.assertIsNotNone(state['rank_before'])
        before = len(self.jira.writes)
        self.run_mode('apply')
        self.run_mode('check')
        self.assertEqual(len(self.jira.writes), before)

    def test_existing_auth_tickets_fields_and_relative_order_are_untouched(self):
        with patch.object(auth_import, 'preflight', return_value=copy.deepcopy(META)), patch('builtins.input', return_value='YES'), patch('sys.stdout', new=io.StringIO()):
            auth_import.run_online(self.jira, auth_import.load_manifest(), Path(self.tmp.name)/'auth-state.json', 'apply')
        auth_data = copy.deepcopy(self.jira.data)
        auth_order = list(self.jira.order)
        write_start = len(self.jira.writes)
        self.run_mode('apply')
        self.run_mode('check')
        self.assertEqual({k: self.jira.data[k] for k in auth_data}, auth_data)
        self.assertEqual([k for k in self.jira.order if k in auth_data], auth_order)
        for method, path, body in self.jira.writes[write_start:]:
            if path.endswith('/rank'):
                self.assertTrue(set(body['issues']).isdisjoint(auth_data))
            elif path.endswith('/issueLink'):
                self.assertNotIn(body['inwardIssue']['key'], auth_data)
                self.assertNotIn(body['outwardIssue']['key'], auth_data)
        self.assertEqual(len(self.jira.data), 35)

    def test_auth_state_or_foreign_batch_is_rejected_before_preflight(self):
        state = {'version': 1, 'batch': 'datn-auth-v1', 'fingerprint': 'foreign'}
        self.state.write_text(json.dumps(state))
        with patch.object(jira_batch, 'preflight') as preflight:
            with self.assertRaisesRegex(app.Stop, 'journal sai'):
                self.run_mode_no_patch()
            preflight.assert_not_called()
        self.assertEqual(json.loads(self.state.read_text()), state)

    def run_mode_no_patch(self):
        self.engine.run_online(self.jira, self.manifest, self.state, 'apply')

    def test_lost_create_is_reconciled_without_duplicate(self):
        self.jira.lose_create_response = True
        with self.assertRaises(app.Stop):
            self.run_mode('apply')
        self.assertEqual(json.loads(self.state.read_text())['pending']['kind'], 'create')
        self.run_mode('apply')
        self.assertEqual(sum(path == '/rest/api/3/issue' for _, path, _ in self.jira.writes), 14)

    def test_unknown_create_stops_without_blind_replay(self):
        self.jira.lose_create_response = True
        with self.assertRaises(app.Stop):
            self.run_mode('apply')
        self.jira.search_hidden = True
        with self.assertRaisesRegex(app.Stop, 'chưa rõ'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), 1)

    def test_lost_link_is_read_before_resume(self):
        self.jira.fail_link = True
        with self.assertRaises(app.Stop):
            self.run_mode('apply')
        self.run_mode('apply')
        self.assertEqual(sum(path.endswith('/issueLink') for _, path, _ in self.jira.writes), 17)

    def test_link_not_written_requires_separate_retry_confirmation(self):
        original = self.jira.request
        def request(method, path, body):
            if path.endswith('/issueLink'):
                raise app.Stop('Mất kết nối trước gửi link.')
            return original(method, path, body)
        with patch.object(self.jira, 'request', side_effect=request):
            with self.assertRaises(app.Stop):
                self.run_mode('apply')
        with self.assertRaisesRegex(app.Stop, 'hủy retry link'):
            self.run_mode('apply')
        self.run_mode('apply', ['RETRY LINK', 'YES'])
        self.assertEqual(sum(path.endswith('/issueLink') for _, path, _ in self.jira.writes), 17)

    def test_lost_rank_response_is_not_retried_in_same_run(self):
        original = self.jira.request
        lost = [False]
        def request(method, path, body):
            result = original(method, path, body)
            if path.endswith('/rank') and not lost[0]:
                lost[0] = True
                raise app.Stop('Mất phản hồi sau Rank.')
            return result
        with patch.object(self.jira, 'request', side_effect=request):
            with self.assertRaises(app.Stop):
                self.run_mode('apply')
        self.assertEqual(json.loads(self.state.read_text())['pending']['kind'], 'rank')
        self.run_mode('plan')
        self.run_mode('apply')
        self.run_mode('check')

    def test_duplicate_and_content_drift_stop_without_overwrite(self):
        record = self.manifest['records'][0]
        for _ in range(2):
            self.jira.request('POST', '/rest/api/3/issue', self.engine.payload(record, META, None, None))
        before = len(self.jira.writes)
        with self.assertRaisesRegex(app.Stop, 'trùng'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), before)
        self.jira = FakeJira()
        self.run_mode('apply')
        self.jira.data['DATN-2']['fields']['description'] = auth_import.adf('Sửa tay')
        before = len(self.jira.writes)
        with self.assertRaisesRegex(app.Stop, 'description khác'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), before)

    def test_source_change_or_changed_rank_metadata_stops(self):
        self.run_mode('apply')
        before = len(self.jira.writes)
        self.manifest['source_hashes']['changed'] = 'new'
        with self.assertRaisesRegex(app.Stop, 'Nguồn/cấu hình'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), before)

    def test_reversed_external_links_and_foreign_import_label_stop(self):
        self.run_mode('apply')
        state = json.loads(self.state.read_text())
        a, b = state['mapping']['OAP-T01'], state['mapping']['OAP-T02']
        self.jira.request('POST', '/rest/api/3/issueLink', auth_import.link_payload(b, a, '3'))
        with self.assertRaisesRegex(app.Stop, 'sai/dư/đảo chiều'):
            self.run_mode('plan')
        self.jira.data[a]['fields']['issuelinks'].append({'type': {'id': '3'}, 'outwardIssue': {'key': 'SCRUM-1'}})
        with self.assertRaisesRegex(app.Stop, 'ngoài manifest'):
            self.run_mode('plan')

    def test_link_url_drift_and_duplicate_blocks_are_not_accepted(self):
        self.run_mode('apply')
        state = json.loads(self.state.read_text())
        key = state['mapping']['OAP-T01']
        record = self.manifest['records'][1]
        base = 'https://github.com/mingquanjp/ai-staffing/blob/' + 'a'*40
        doc = app.render_description(record, base)
        changed = copy.deepcopy(doc)
        def alter(node):
            for mark in node.get('marks', []):
                if mark.get('type') == 'link':
                    mark['attrs']['href'] = 'https://other.invalid'
            for child in node.get('content', []):
                alter(child)
        alter(changed)
        self.assertEqual(auth_import.adf_text(doc), auth_import.adf_text(changed))
        self.assertNotEqual(jira_batch.description_signature(doc), jira_batch.description_signature(changed))
        original = self.jira.data[key]['fields']['description']
        self.jira.data[key]['fields']['description'] = changed
        with self.assertRaisesRegex(app.Stop, 'description khác'):
            self.engine.validate_issue(self.jira.issue(key), record, META, state['mapping']['OAP-EPIC'], base)
        self.jira.data[key]['fields']['description'] = original
        link = self.jira.data[key]['fields']['issuelinks'][0]
        self.jira.data[key]['fields']['issuelinks'].append(copy.deepcopy(link))
        before = len(self.jira.writes)
        with self.assertRaisesRegex(app.Stop, 'Blocks trùng'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), before)

    def test_metadata_or_account_change_stops_without_write(self):
        self.run_mode('apply')
        before = len(self.jira.writes)
        for field, value in [('rank', 99999), ('user', 'another-user')]:
            meta = dict(META, **{field: value})
            with patch.object(jira_batch, 'preflight', return_value=meta), self.assertRaises(app.Stop):
                self.run_mode_no_patch()
        self.assertEqual(len(self.jira.writes), before)

    def test_manifest_parser_stops_on_mapping_or_description_drift(self):
        source = app.SOURCE.read_text()
        for changed in [source.replace('| OAP-04 | organization |', '| OAP-03 | organization |', 1),
                        source.replace('- Is blocked by: OAP-T01.', '- Is blocked by: OAP-T13.', 1),
                        source.replace('### [OAP-T13]', '### [OAP-T99]', 1)]:
            fake = Path(self.tmp.name)/'tickets.md'
            fake.write_text(changed)
            with patch.object(app, 'SOURCE', fake), self.assertRaises(app.Stop):
                app.load_manifest()

    def test_exclusive_export_and_journal_path_isolation(self):
        export = Path(self.tmp.name)/'manifest.json'
        app.export_manifest(export, self.manifest)
        self.assertEqual(export.stat().st_mode & 0o777, 0o600)
        before = export.read_bytes()
        with self.assertRaises(FileExistsError):
            app.export_manifest(export, {})
        self.assertEqual(export.read_bytes(), before)
        with self.assertRaisesRegex(app.Stop, 'journal OAP riêng'):
            app.safe_state_path(app.ROOT/'.local/jira-auth/state.json')
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root/'.local').mkdir()
            (root/'.local/jira-oap').symlink_to(self.tmp.name, target_is_directory=True)
            with patch.object(app, 'ROOT', root), self.assertRaisesRegex(app.Stop, 'symlink'):
                app.safe_state_path(root/'.local/jira-oap/state.json')

    def test_offline_cli_and_diagnose_do_not_read_or_write_journal(self):
        with patch('sys.argv', ['oap_import.py', '--mode', 'plan', '--offline']), \
             patch.object(app, 'credentials') as credentials, patch.object(app, 'state_lock') as lock, \
             patch('sys.stdout', new=io.StringIO()):
            app.main()
        credentials.assert_not_called()
        lock.assert_not_called()
        with patch('sys.argv', ['oap_import.py', '--mode', 'diagnose']), \
             patch.object(app, 'credentials', return_value=('example@example.invalid', 'SECRET_MARKER')), \
             patch.object(app, 'board_diagnostics') as diagnose, patch.object(app, 'load_manifest') as manifest, \
             patch.object(app, 'safe_state_path') as state:
            app.main()
        diagnose.assert_called_once()
        manifest.assert_not_called()
        state.assert_not_called()

    def test_token_input_requires_tty_and_source_url_fixed_repo(self):
        with patch('sys.stdin.isatty', return_value=False), self.assertRaises(app.Stop):
            app.credentials()
        for url in ['https://github.com/other/repo/blob/'+'a'*40, 'https://other.invalid/SECRET_MARKER',
                    'https://github.com/mingquanjp/ai-staffing/blob/main']:
            with self.assertRaises(app.Stop):
                app.validate_source_base(url)


if __name__ == '__main__':
    unittest.main()
