"""Kiểm nguồn MSK và điều phối qua Jira giả; không gọi mạng hoặc dùng token thật."""
import copy
import io
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import auth_import
import jira_batch
import msk_import as app
from test_auth_import import FakeJira, META


class MskImportTests(unittest.TestCase):
    def setUp(self):
        self.manifest = app.load_manifest()
        self.engine = jira_batch.BatchImporter(self.manifest, app.render_description)
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.state = Path(self.tmp.name) / 'state.json'
        self.jira = FakeJira()

    def run_mode(self, mode):
        with patch.object(jira_batch, 'preflight', return_value=copy.deepcopy(META)), \
             patch('builtins.input', return_value='YES'), patch('sys.stdout', new=io.StringIO()):
            self.engine.run_online(self.jira, self.manifest, self.state, mode)

    def test_source_has_13_records_12_wbs_and_19_forward_links(self):
        self.assertEqual(len(self.manifest['records']), 13)
        self.assertEqual(len(app.expected_links(self.manifest)), 19)
        tasks = self.manifest['records'][1:]
        self.assertEqual([r['source_id'] for r in tasks], app.ORDER)
        self.assertEqual([r['wbs'] for r in tasks], [[f'MSK-{i:02}'] for i in range(1, 13)])
        self.assertTrue(all('hours_baseline' not in r for r in tasks))
        self.assertEqual(self.manifest['records'][0]['source_id'], 'MSK-EPIC')
        self.assertEqual(len(self.manifest['source_hashes']), 8)
        self.assertIn('docs/api/profile-change-approval.openapi.json', self.manifest['source_hashes'])

    def test_adf_keeps_tables_checks_and_reviewed_behavior(self):
        rendered = {}
        for record in self.manifest['records']:
            document = app.render_description(record)
            text = auth_import.adf_text(document)
            self.assertEqual(text.count('Mã nguồn ticket:'), 1)
            self.assertIn(record['source_id'], text)
            self.assertIn('☐', text)
            self.assertIn('link remote chờ bàn giao', text)
            if record['type'] == 'Task':
                tables = [node for node in document['content'] if node['type'] == 'table']
                self.assertEqual(len(tables), 1)
                self.assertTrue(all(len(row['content']) == 2 for row in tables[0]['content']))
                self.assertIn('Đầu ra bàn giao', text)
                self.assertIn('git diff --check', text)
            rendered[record['source_id']] = text
        self.assertIn('before.note', rendered['MSK-T06'])
        self.assertIn('không thêm audit trùng', rendered['MSK-T07'])
        self.assertIn('không Number/toFixed', rendered['MSK-T09'])
        for number in (2, 3, 5, 6, 7, 8, 12):
            self.assertIn('go vet ./...', rendered[f'MSK-T{number:02}'])
            self.assertIn('python3 scripts/authdb/manage.py check', rendered[f'MSK-T{number:02}'])
        for number in (4, 9, 10, 11, 12):
            self.assertIn('pnpm check', rendered[f'MSK-T{number:02}'])

    def test_plan_and_incomplete_check_do_not_write(self):
        self.run_mode('plan')
        self.assertFalse(self.state.exists())
        with self.assertRaises(app.Stop):
            self.run_mode('check')
        self.assertFalse(self.state.exists())
        self.assertFalse(self.jira.writes)

    def test_apply_rerun_and_check_use_only_msk_fields(self):
        self.run_mode('apply')
        self.assertEqual(len(self.jira.data), 13)
        self.assertEqual(sum(path.endswith('/issueLink') for _, path, _ in self.jira.writes), 19)
        for _, path, body in self.jira.writes:
            if path == '/rest/api/3/issue':
                self.assertLessEqual(set(body['fields']),
                                     {'project', 'issuetype', 'summary', 'description', 'assignee', 'labels', 'parent'})
                self.assertEqual(len(body['fields']['labels']), 3)
                self.assertIn('datn-msk-v1', body['fields']['labels'])
                self.assertIn('manual-skills', body['fields']['labels'])
        state = json.loads(self.state.read_text())
        self.assertEqual(state['batch'], 'datn-msk-v1')
        self.assertEqual(self.state.stat().st_mode & 0o777, 0o600)
        before = len(self.jira.writes)
        self.run_mode('apply')
        self.run_mode('check')
        self.assertEqual(len(self.jira.writes), before)

    def test_prior_oap_and_pca_fields_and_order_are_preserved(self):
        import oap_import
        import pca_import
        for previous in (oap_import, pca_import):
            manifest = previous.load_manifest()
            engine = jira_batch.BatchImporter(manifest, previous.render_description)
            with patch.object(jira_batch, 'preflight', return_value=copy.deepcopy(META)), \
                 patch('builtins.input', return_value='YES'), patch('sys.stdout', new=io.StringIO()):
                engine.run_online(self.jira, manifest, Path(self.tmp.name) / (previous.IMPORT_LABEL + '.json'), 'apply')
        before_data, before_order = copy.deepcopy(self.jira.data), list(self.jira.order)
        self.run_mode('apply')
        self.run_mode('check')
        self.assertEqual({key: self.jira.data[key] for key in before_data}, before_data)
        self.assertEqual([key for key in self.jira.order if key in before_data], before_order)

    def test_lost_create_reconciles_but_hidden_result_stops(self):
        self.jira.lose_create_response = True
        with self.assertRaises(app.Stop):
            self.run_mode('apply')
        self.assertEqual(json.loads(self.state.read_text())['pending']['kind'], 'create')
        before = len(self.jira.writes)
        self.jira.search_hidden = True
        with self.assertRaisesRegex(app.Stop, 'chưa rõ'):
            self.run_mode('apply')
        self.assertEqual(len(self.jira.writes), before)
        self.jira.search_hidden = False
        self.run_mode('apply')
        self.assertEqual(sum(path == '/rest/api/3/issue' for _, path, _ in self.jira.writes), 13)

    def test_parser_rejects_dependency_wbs_heading_and_trace_drift(self):
        source = app.SOURCE.read_text()
        for changed in [
            source.replace('| MSK-T04 | MSK-04 |', '| MSK-T04 | MSK-03 |', 1),
            source.replace('- Is blocked by: MSK-T01.', '- Is blocked by: MSK-T12.', 1),
            source.replace('### [MSK-T12]', '### [MSK-T99]', 1),
            source.replace('| MSK-01 | — |', '| MSK-01 | MSK-T12 |', 1),
            source.replace('**Truy vết nhập Jira:** `MSK-T01`', '**Truy vết nhập Jira:** `MSK-T99`', 1),
        ]:
            self.assertNotEqual(changed, source)
            fixture = Path(self.tmp.name) / 'tickets.md'
            fixture.write_text(changed)
            with patch.object(app, 'SOURCE', fixture), self.assertRaises(app.Stop):
                app.load_manifest()

    def test_source_link_and_commit_match_all_eight_inputs(self):
        base = 'https://github.com/mingquanjp/prj-manage-system/blob/' + 'a' * 40
        document = app.render_description(self.manifest['records'][1], base)
        links = jira_batch.description_signature(document)[1]
        self.assertTrue(links)
        self.assertTrue(all(href.startswith(base + '/') for _, href in links))
        blobs = [file.read_bytes() for file in app.SOURCES]
        with patch.object(app.subprocess, 'check_output', side_effect=blobs) as show:
            app.validate_source_base(base)
        self.assertEqual(show.call_count, 8)
        with patch.object(app.subprocess, 'check_output', return_value=b'changed'), self.assertRaises(app.Stop):
            app.validate_source_base(base)
        for invalid in ['https://other.invalid/secret',
                        'https://github.com/mingquanjp/prj-manage-system/blob/main']:
            with self.assertRaises(app.Stop):
                app.validate_source_base(invalid)

    def test_foreign_journal_and_symlinks_are_rejected(self):
        for batch in ('datn-auth-v1', 'datn-oap-v1', 'datn-pca-v1'):
            state = {'version': 1, 'batch': batch, 'fingerprint': 'foreign'}
            self.state.write_text(json.dumps(state))
            with patch.object(jira_batch, 'preflight') as preflight, self.assertRaises(app.Stop):
                self.engine.run_online(self.jira, self.manifest, self.state, 'apply')
            preflight.assert_not_called()
            self.assertEqual(json.loads(self.state.read_text()), state)
        with self.assertRaisesRegex(app.Stop, 'journal MSK riêng'):
            app.safe_state_path(app.ROOT / '.local/jira-pca/state.json')
        root = Path(self.tmp.name) / 'root'
        (root / '.local').mkdir(parents=True)
        (root / '.local/jira-msk').symlink_to(self.tmp.name, target_is_directory=True)
        with patch.object(app, 'ROOT', root), self.assertRaisesRegex(app.Stop, 'symlink'):
            app.safe_state_path(root / '.local/jira-msk/state.json')

    def test_offline_does_not_construct_jira_or_touch_credentials_and_journal(self):
        with patch('sys.argv', ['msk_import.py', '--mode', 'plan', '--offline']), \
             patch.object(app, 'Jira') as jira, patch.object(app, 'credentials') as credentials, \
             patch.object(app, 'state_lock') as lock, patch('sys.stdout', new=io.StringIO()):
            app.main()
        jira.assert_not_called()
        credentials.assert_not_called()
        lock.assert_not_called()
        with patch('sys.argv', ['msk_import.py', '--mode', 'diagnose']), \
             patch.object(app, 'credentials', return_value=('example@example.invalid', 'SECRET_MARKER')), \
             patch.object(app, 'board_diagnostics') as diagnose, patch.object(app, 'load_manifest') as manifest, \
             patch.object(app, 'safe_state_path') as state:
            app.main()
        diagnose.assert_called_once()
        manifest.assert_not_called()
        state.assert_not_called()


if __name__ == '__main__':
    unittest.main()
