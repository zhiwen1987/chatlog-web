#!/usr/bin/env python3
"""Optional JSON Schema tests of synthetic examples, not MCP/ACL integration."""
from __future__ import annotations

import copy
import json
from pathlib import Path
import sys

try:
    import jsonschema
except ImportError:
    print('Optional check unavailable: install requirements-doc-check.txt in an isolated environment.', file=sys.stderr)
    sys.exit(2)

from verify_delivery import context_input_semantics, strict_json

ROOT = Path(__file__).resolve().parents[1]


def load(name: str) -> object:
    return strict_json((ROOT / 'mcp/examples' / name).read_text(encoding='utf-8'))


def main() -> int:
    input_schema = load('messages-context.input.schema.json')
    output_schema = load('messages-context.output.schema.json')
    input_value = load('messages-context.input.json')
    output_value = load('messages-context.output.json')
    records = []
    for name, schema in (('input_schema_valid', input_schema), ('output_schema_valid', output_schema)):
        try:
            jsonschema.Draft202012Validator.check_schema(schema)
            records.append({'case': name, 'passed': True})
        except jsonschema.SchemaError:
            records.append({'case': name, 'passed': False})
    iv = jsonschema.Draft202012Validator(input_schema, format_checker=jsonschema.FormatChecker())
    ov = jsonschema.Draft202012Validator(output_schema, format_checker=jsonschema.FormatChecker())
    cases = [('input_example', iv, input_value, True), ('output_example', ov, output_value, True)]
    for name, patch in (
        ('tenant_field_rejected', {'tenantId': 'another-tenant'}),
        ('fraction_before_rejected', {'before': 1.5}),
        ('negative_before_rejected', {'before': -1}),
        ('large_before_rejected', {'before': 201}),
        ('boolean_before_rejected', {'before': True}),
    ):
        value = copy.deepcopy(input_value)
        value.update(patch)
        cases.append((name, iv, value, False))
    cases.append(('missing_message_id_rejected', iv, {'before': 0}, False))
    value = copy.deepcopy(output_value); value['secret'] = 'synthetic'
    cases.append(('extra_output_field_rejected', ov, value, False))
    value = copy.deepcopy(output_value); value['messages'][0]['revision'] = 1
    cases.append(('numeric_revision_rejected', ov, value, False))
    value = copy.deepcopy(output_value); value['coverage']['asOf'] = 'not-a-date'
    cases.append(('invalid_asof_rejected', ov, value, False))
    value = copy.deepcopy(output_value); value['messages'] *= 202
    cases.append(('too_many_messages_rejected', ov, value, False))
    value = copy.deepcopy(output_value); value['messages'][0]['decodeStatus'] = 'made_up'
    cases.append(('unknown_decode_status_rejected', ov, value, False))
    for name, validator, value, expected in cases:
        records.append({'case': name, 'passed': validator.is_valid(value) == expected})
    # Show explicitly what shape-only validation cannot prove.
    over_total = {'messageId': 'fixture-message-001', 'before': 200, 'after': 200}
    records.append({'case': 'schema_alone_does_not_check_sum', 'passed': iv.is_valid(over_total)})
    records.append({'case': 'service_semantics_reject_sum', 'passed': bool(context_input_semantics(over_total))})
    failed = sum(not row['passed'] for row in records)
    print(json.dumps({'scope': 'synthetic_contract_examples_only', 'count': len(records), 'failed': failed, 'cases': records}, ensure_ascii=False, indent=2))
    return 1 if failed else 0


if __name__ == '__main__':
    sys.exit(main())
