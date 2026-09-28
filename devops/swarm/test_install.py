import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from types import SimpleNamespace
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('installer', Path(__file__).with_name('install.py'))
installer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(installer)

BASE = json.loads(Path(__file__).with_name('config.example.json').read_text())

class InstallerTests(unittest.TestCase):
    def config(self):
        c = copy.deepcopy(BASE)
        c.update(db_password='db$literal', jwt_secret='jwt', auth_secret='auth',
                 private_key='rsa', storage_user='xemadmin', storage_password='storage')
        return c

    def test_origins_and_ports(self):
        installer.validate(copy.deepcopy(BASE))
        for invalid in ('http://localhost:3000', 'file:///tmp/x', 'https://a/b', 'https://u:p@a', 'https://a?q=x'):
            c = copy.deepcopy(BASE)
            c['api_url'] = invalid
            with self.assertRaises(ValueError): installer.validate(c)
        c = copy.deepcopy(BASE)
        c['storage_port'] = c['api_port']
        with self.assertRaises(ValueError): installer.validate(c)

    def test_service_isolation_and_persistence(self):
        stack = installer.stack_spec(self.config(), 'node1', 'abc')
        self.assertEqual(set(stack['volumes']), {'postgres', 'redis', 'storage'})
        for name, service in stack['services'].items():
            self.assertEqual(service['deploy']['placement']['constraints'], ['node.id == node1'])
            self.assertEqual(service['deploy']['replicas'], 1)
            if name in ('postgres', 'redis', 'storage-init'):
                self.assertNotIn('ports', service)
        self.assertEqual(stack['services']['api']['environment']['POSTGRES_PASSWORD'], 'db$$literal')
        self.assertEqual(stack['services']['api']['environment']['S3_ENDPOINT_URL'], BASE['storage_url'])
        self.assertEqual(stack['services']['app']['environment']['INTERNAL_API_URL'], 'http://api:9001/api/v1')
        self.assertEqual(stack['services']['storage-init']['deploy']['restart_policy']['condition'], 'on-failure')

    def test_state_permissions(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'config.json'
            installer.save(path, self.config())
            self.assertEqual(path.stat().st_mode & 0o777, 0o600)
            self.assertEqual(json.loads(path.read_text()), self.config())

    def test_existing_credentials_and_node_are_preserved(self):
        with tempfile.TemporaryDirectory() as directory:
            state = Path(directory)
            config = self.config()
            config['node_id'] = 'node1'
            installer.save(state / 'config.json', config)
            calls = []
            def run(*args, **kwargs):
                calls.append(args)
                if args[:2] == ('docker', 'info'):
                    return json.dumps({'OSType': 'linux', 'Swarm': {'LocalNodeState': 'active', 'ControlAvailable': True, 'NodeID': 'node1'}})
                if args[:3] == ('docker', 'stack', 'ls'): return 'xem\n'
                if args[0] == 'git': return 'a' * 40
                return ''
            with patch.object(installer, 'run', run), patch.object(installer, 'wait_ready'):
                installer.install(SimpleNamespace(config=None), state)
            self.assertEqual(json.loads((state / 'config.json').read_text()), config)
            self.assertFalse(any(c[:3] == ('docker', 'swarm', 'init') for c in calls))
            self.assertTrue(any(c[:3] == ('docker', 'stack', 'deploy') for c in calls))

    def test_worker_unowned_stack_and_node_mismatch_refused(self):
        for control, existing, saved_node in [(False, '', None), (True, 'xem', None), (True, '', 'other-node')]:
            with self.subTest(control=control, existing=existing, saved_node=saved_node), tempfile.TemporaryDirectory() as directory:
                state = Path(directory)
                config = self.config()
                if saved_node: config['node_id'] = saved_node
                installer.save(state / 'config.json', config)
                def run(*args, **kwargs):
                    if args[:2] == ('docker', 'info'):
                        return json.dumps({'OSType': 'linux', 'Swarm': {'LocalNodeState': 'active', 'ControlAvailable': control, 'NodeID': 'node1'}})
                    if args[:3] == ('docker', 'stack', 'ls'): return existing
                    self.fail('Unexpected mutation: ' + str(args))
                with patch.object(installer, 'run', run), self.assertRaises(ValueError):
                    installer.install(SimpleNamespace(config=None), state)

    def test_readiness_failure_is_not_success(self):
        with patch.object(installer.time, 'monotonic', side_effect=[0, 1, 11]), patch.object(installer.time, 'sleep'), patch.object(installer.urllib.request, 'urlopen', side_effect=OSError('unavailable')):
            with self.assertRaises(RuntimeError): installer.wait_ready(BASE, timeout=10)

if __name__ == '__main__':
    unittest.main()
