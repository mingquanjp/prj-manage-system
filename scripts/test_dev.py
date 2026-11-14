import importlib.util
from pathlib import Path
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('local_dev', Path(__file__).with_name('dev.py'))
dev = importlib.util.module_from_spec(spec)
spec.loader.exec_module(dev)


class DevRunnerTests(unittest.TestCase):
    def test_env_never_inherits_database_or_target(self):
        supplied = {key: 'synthetic' for key in ('DATABASE_URL', 'APP_ENV', 'TEST_DATABASE_URL',
                    'AUTH_TEST_DB_SYSTEM_ID', 'AUTH_LOCAL_DATABASE_URL', 'AUTH_LOCAL_SYSTEM_ID',
                    'TASKASSIGN_API_TARGET', 'PGPASSWORD', 'PATH')}
        self.assertEqual(dev.clean_env(supplied), {'PATH': 'synthetic'})

    def test_busy_port_does_not_kill_or_start_process(self):
        with patch.object(dev.socket, 'socket') as factory, patch.object(dev.subprocess, 'Popen') as start:
            factory.return_value.__enter__.return_value.bind.side_effect = OSError()
            with self.assertRaises(dev.authdb.Refused):
                dev.require_free(5174)
            start.assert_not_called()

    def test_child_failure_stops_before_frontend(self):
        class Exited:
            def poll(self):
                return 1
        class Running:
            stopping = False
        with self.assertRaises(dev.authdb.Refused):
            dev.wait_for_backend(Exited(), Running())

    def test_signals_only_change_supervisor_state(self):
        processes = dev.Processes()
        processes.stop_signal()
        self.assertTrue(processes.stopping)
        self.assertEqual(processes.children, [])


if __name__ == '__main__':
    unittest.main()
