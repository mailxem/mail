import unittest
import importlib.util
from pathlib import Path
spec = importlib.util.spec_from_file_location('component_release', Path(__file__).with_name('component-release.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
next_version = module.next_version

class ReleaseTests(unittest.TestCase):
    def test_components_cannot_bump_each_other(self):
        tags = ['frontend-v0.2.4', 'backend-v5.6.7', 'v99.99.99', 'frontend-v0.2.4-beta']
        self.assertEqual(next_version('frontend', tags), ('0.2.5', 'frontend-v0.2.4'))
        self.assertEqual(next_version('backend', tags), ('5.6.8', 'backend-v5.6.7'))

    def test_numeric_order(self):
        self.assertEqual(next_version('backend', ['backend-v0.1.9', 'backend-v0.1.10']), ('0.1.11', 'backend-v0.1.10'))

    def test_first_release(self):
        self.assertEqual(next_version('frontend', ['v1.0.0']), ('0.1.0', ''))

    def test_unknown_component(self):
        with self.assertRaises(ValueError): next_version('other', [])
