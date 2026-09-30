"""Tests of the delivery validator only; never connect to product services."""
from __future__ import annotations

import copy
import hashlib
from pathlib import Path
import sys
import tempfile
import unittest

BUNDLE = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(BUNDLE / 'scripts'))
import verify_delivery as checker


class ValidatorTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name).resolve()
        self.source = self.root / 'README.md'
        self.source.write_text('# Example\n', encoding='utf-8')

    def tearDown(self):
        self.temp.cleanup()

    def make_skill(self, front_extra='', body='## Steps\nCheck the actual repository.\n'):
        p = self.root / 'wcm-example' / 'SKILL.md'
        p.parent.mkdir(exist_ok=True)
        p.write_text('---\nname: wcm-example\ndescription: A valid example.\nmetadata:\n  audience: "development"\n' + front_extra + '---\n' + body, encoding='utf-8')
        return p

    def test_json_valid(self):
        self.assertEqual(checker.strict_json('{"value":1}'), {'value': 1})

    def test_json_duplicate_top_level(self):
        with self.assertRaises(ValueError):
            checker.strict_json('{"a":1,"a":2}')

    def test_json_duplicate_nested(self):
        with self.assertRaises(ValueError):
            checker.strict_json('{"a":{"b":1,"b":2}}')

    def test_json_nonstandard_nan(self):
        with self.assertRaises(ValueError):
            checker.strict_json('{"value":NaN}')

    def test_fences_balanced(self):
        prose, unclosed = checker.prose_only('# A\n```text\n[bad](missing.md)\n```\nTail')
        self.assertFalse(unclosed)
        self.assertNotIn('missing.md', prose)
        self.assertIn('Tail', prose)

    def test_fences_unclosed(self):
        self.assertTrue(checker.prose_only('```text\nx')[1])

    def test_inline_code_ignored(self):
        self.assertEqual(checker.check_links(self.root, self.source, '`[bad](missing.md)`'), [])

    def test_relative_link(self):
        (self.root / '功能.md').write_text('# 功能\n', encoding='utf-8')
        self.assertEqual(checker.check_links(self.root, self.source, '[f](功能.md)'), [])

    def test_missing_link(self):
        self.assertTrue(checker.check_links(self.root, self.source, '[f](missing.md)'))

    def test_explicit_anchor(self):
        (self.root / '功能.md').write_text('<a id="stable"></a>\n# 功能', encoding='utf-8')
        self.assertEqual(checker.check_links(self.root, self.source, '[f](功能.md#stable)'), [])

    def test_missing_anchor(self):
        self.assertTrue(checker.check_links(self.root, self.source, '[f](README.md#missing)'))

    def test_chinese_heading(self):
        self.assertIn('中文-标题', checker.anchors('# 中文 标题'))

    def test_duplicate_heading(self):
        self.assertIn('hello-1', checker.anchors('# Hello\n# Hello'))

    def test_external_not_fetched(self):
        self.assertEqual(checker.check_links(self.root, self.source, '[f](https://not-requested.invalid/file)'), [])

    def test_traversal_rejected(self):
        with self.assertRaises(ValueError):
            checker.safe_target(self.root, self.source, '../outside.md')

    def test_encoded_traversal_rejected(self):
        with self.assertRaises(ValueError):
            checker.safe_target(self.root, self.source, '%2e%2e/outside.md')

    def test_javascript_scheme_rejected(self):
        with self.assertRaises(ValueError):
            checker.safe_target(self.root, self.source, 'javascript:alert')

    def test_absolute_path_rejected(self):
        with self.assertRaises(ValueError):
            checker.safe_target(self.root, self.source, '/tmp/not-read')

    def test_backslash_rejected(self):
        with self.assertRaises(ValueError):
            checker.safe_target(self.root, self.source, '..\\outside.md')

    def test_symlink_rejected(self):
        target = self.root / 'real.md'
        target.write_text('# Real', encoding='utf-8')
        link = self.root / 'linked.md'
        try:
            link.symlink_to(target)
        except OSError:
            self.skipTest('This OS does not permit test symlinks')
        with self.assertRaises(ValueError):
            checker.safe_target(self.root, self.source, 'linked.md')

    def test_skill_valid(self):
        self.assertEqual(checker.check_skill(self.make_skill(), 'development'), [])

    def test_skill_name_mismatch(self):
        p = self.make_skill()
        p.write_text(p.read_text().replace('name: wcm-example', 'name: wrong-name'))
        self.assertTrue(checker.check_skill(p, 'development'))

    def test_skill_duplicate_key(self):
        self.assertTrue(checker.check_skill(self.make_skill(front_extra='name: wcm-example\n'), 'development'))

    def test_skill_wrong_audience(self):
        self.assertTrue(checker.check_skill(self.make_skill(), 'runtime-readonly'))

    def test_skill_missing_body(self):
        self.assertTrue(checker.check_skill(self.make_skill(body=''), 'development'))

    def test_inventory_valid(self):
        data = {'runtimeImportAllowed': False, 'toolCount': 1, 'tools': [{'name': 'messages_get', 'registrationAllowed': False}]}
        self.assertEqual(checker.check_inventory(data), [])

    def test_inventory_duplicate(self):
        row = {'name': 'messages_get', 'registrationAllowed': False}
        self.assertTrue(checker.check_inventory({'runtimeImportAllowed': False, 'toolCount': 2, 'tools': [row, copy.copy(row)]}))

    def test_inventory_count_mismatch(self):
        self.assertTrue(checker.check_inventory({'runtimeImportAllowed': False, 'toolCount': 10, 'tools': []}))

    def test_design_not_registrable(self):
        self.assertTrue(checker.check_inventory({'runtimeImportAllowed': True, 'toolCount': 1, 'tools': [{'name': 'messages_get', 'registrationAllowed': True}]}))

    def test_context_valid(self):
        self.assertEqual(checker.context_input_semantics({'messageId':'a','before':100,'after':100}), [])

    def test_context_defaults(self):
        self.assertEqual(checker.context_input_semantics({'messageId':'a'}), [])

    def test_context_tenant_injection(self):
        self.assertTrue(checker.context_input_semantics({'messageId':'a','tenantId':'b'}))

    def test_context_budget(self):
        self.assertTrue(checker.context_input_semantics({'messageId':'a','before':101,'after':100}))

    def test_context_bool_not_integer(self):
        self.assertTrue(checker.context_input_semantics({'messageId':'a','before':True}))

    def test_context_fraction_rejected(self):
        self.assertTrue(checker.context_input_semantics({'messageId':'a','before':1.5}))

    def test_context_negative_rejected(self):
        self.assertTrue(checker.context_input_semantics({'messageId':'a','before':-1}))

    def test_manifest_valid(self):
        digest = hashlib.sha256(self.source.read_bytes()).hexdigest()
        (self.root/'MANIFEST.sha256').write_text(f'{digest}  README.md\n')
        self.assertEqual(checker.check_manifest(self.root), [])

    def test_manifest_tamper(self):
        (self.root/'MANIFEST.sha256').write_text('0'*64+'  README.md\n')
        self.assertTrue(checker.check_manifest(self.root))

    def test_manifest_missing_file_coverage(self):
        (self.root/'MANIFEST.sha256').write_text('')
        self.assertTrue(checker.check_manifest(self.root))

    def test_missing_required_documents(self):
        errors, _ = checker.validate(self.root)
        self.assertTrue(any('missing required file' in e for e in errors))

    def test_real_delivery_bundle(self):
        errors, stats = checker.validate(BUNDLE)
        self.assertEqual(errors, [])
        self.assertEqual(stats['developmentSkills'], 13)
        self.assertEqual(stats['runtimeSkills'], 4)
        self.assertEqual(stats['designTools'], 141)


if __name__ == '__main__':
    unittest.main()
