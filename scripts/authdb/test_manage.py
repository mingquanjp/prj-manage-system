"""Kiểm guard bằng Docker giả; không phải bằng chứng DB thật."""
import copy
import importlib.util
import json
import os
from pathlib import Path
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('authdb', Path(__file__).with_name('manage.py'))
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)


class GuardTests(unittest.TestCase):
    def setUp(self):
        self.state = dict(engine='engine-owned', owner='nonce', container_id='cid',
                          image_id='sha256:owned', password='SECRET_MARKER',
                          url=f'postgresql://{m.USER}:SECRET_MARKER@127.0.0.1:{m.PORT}/{m.DATABASE}?sslmode=disable')
        self.container = {'Id': 'cid', 'Image': 'sha256:owned',
            'Config': {'Labels': {m.LABEL: 'nonce'}},
            'HostConfig': {'PortBindings': {'5432/tcp': [{'HostIp': '127.0.0.1', 'HostPort': m.PORT}]}},
            'Mounts': [{'Name': m.VOLUME, 'Destination': '/var/lib/postgresql'}],
            'State': {'Running': True}}
        self.volume = {'Labels': {m.LABEL: 'nonce'}}

    def check(self, container=None, volume=None):
        def fake(*args, **kwargs):
            return json.dumps([container or self.container] if args[0] == 'inspect' else [volume or self.volume])
        with patch.dict(os.environ, {}, clear=True), patch.object(m, 'engine', return_value='engine-owned'), patch.object(m, 'docker', side_effect=fake):
            return m.checked(self.state)

    def test_owned_target(self):
        self.assertEqual(self.check()['Id'], 'cid')

    def test_foreign_container_volume_port_image_and_mount_are_rejected(self):
        mutations = [lambda c: c.update(Id='foreign'), lambda c: c.update(Image='other'),
            lambda c: c['Config'].update(Labels={m.LABEL: 'foreign'}),
            lambda c: c['HostConfig']['PortBindings']['5432/tcp'][0].update(HostIp='0.0.0.0'),
            lambda c: c['HostConfig']['PortBindings']['5432/tcp'][0].update(HostPort='5432'),
            lambda c: c['Mounts'][0].update(Name='development'),
            lambda c: c['Mounts'][0].update(Destination='/var/lib/postgresql/data'),
            lambda c: c['State'].update(Running=False)]
        for mutation in mutations:
            c = copy.deepcopy(self.container)
            mutation(c)
            with self.subTest(mutation=mutation), self.assertRaises(m.Refused):
                self.check(container=c)
        with self.assertRaises(m.Refused):
            self.check(volume={'Labels': {m.LABEL: 'foreign'}})

    def test_url_requires_exact_owned_credential_host_database_and_options(self):
        for raw in ['', 'postgres://neon.invalid/production',
                    self.state['url'].replace(m.DATABASE, 'development'),
                    self.state['url'].replace('SECRET_MARKER', 'other'),
                    self.state['url'] + '&host=neon.invalid']:
            with self.subTest(raw=raw), self.assertRaises(m.Refused) as error:
                m.validate_url(raw, self.state['url'])
            self.assertNotIn('SECRET_MARKER', str(error.exception))

    def test_engine_change_rejected_before_inspect(self):
        with patch.object(m, 'engine', return_value='other'), patch.object(m, 'docker') as docker:
            with self.assertRaises(m.Refused):
                m.checked(self.state)
            docker.assert_not_called()

    def test_no_environment_or_missing_state_fallback(self):
        with patch.dict(os.environ, {'DATABASE_URL': 'SECRET_MARKER'}, clear=True), patch.object(m, 'STATE', Path('/does-not-exist/state.json')):
            with self.assertRaises(m.Refused):
                m.load()

    def test_external_test_url_rejected(self):
        with patch.dict(os.environ, {'TEST_DATABASE_URL': 'SECRET_MARKER'}, clear=True), patch.object(m, 'engine', return_value='engine-owned'), patch.object(m, 'docker') as docker:
            with self.assertRaises(m.Refused):
                m.checked(self.state)
            docker.assert_not_called()

    def test_docker_failure_redacts_stderr(self):
        class Result:
            returncode = 1
            stdout = ''
            stderr = 'SECRET_MARKER'
        with patch.object(m.subprocess, 'run', return_value=Result()):
            with self.assertRaises(m.Refused) as error:
                m.docker('inspect', 'fake')
        self.assertNotIn('SECRET_MARKER', str(error.exception))

    def test_remote_engine_override_rejected(self):
        for name in ['DOCKER_HOST', 'DOCKER_CONTEXT']:
            with patch.dict(os.environ, {name: 'foreign'}, clear=True), patch.object(m, 'docker') as docker:
                with self.assertRaises(m.Refused):
                    m.engine()
                docker.assert_not_called()


if __name__ == '__main__':
    unittest.main()
