#!/usr/bin/env python3
"""Offline documentation-bundle checks. No network, commands, or product access.

This is not a full Markdown/YAML/JSON-Schema validator or a product security test.
See --help and reports/文档校验报告.md for the explicit verification boundary.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import sys
import unicodedata
from urllib.parse import unquote, urlsplit

EXCLUDED_DIRS = {'.git', '.delivery', '__pycache__', 'node_modules', 'target', 'dist', '.venv'}
REQUIRED = (
    'AGENTS.md', '开发规则.md', '开发手册.md', 'README.md', '00-开始使用.md',
    '项目介绍.md', '功能说明.md', '安装指南.md', '项目交接.md', 'SECURITY.md',
    'docs/install/服务端安装.md', 'docs/install/桌面客户端安装.md',
    'docs/install/授权中心安装.md', 'docs/install/DSH与Agent-Skills-MCP接入.md',
    'docs/handoff/维护交接.md', 'docs/handoff/安全与密钥交接.md',
    'docs/operations/备份恢复与应急.md', 'docs/agent/Skills目录.md',
    'docs/mcp/MCP接入与权限.md', 'docs/provenance/来源与版本.md',
    'templates/HANDOFF.md', 'templates/IMPLEMENTATION_STATUS.example.json',
    'templates/RELEASE_MANIFEST.example.json', 'mcp/tool-inventory.json',
)
LINK = re.compile(r'\[[^\]\n]*\]\(([^\)\n]+)\)')
NAME = re.compile(r'^[a-z0-9]+(?:-[a-z0-9]+)*$')
TOOL = re.compile(r'^[a-z][a-z0-9_]*$')


def strict_json(text: str) -> object:
    def pairs(items: list[tuple[str, object]]) -> dict[str, object]:
        result: dict[str, object] = {}
        for key, value in items:
            if key in result:
                raise ValueError(f'duplicate JSON key: {key}')
            result[key] = value
        return result

    def invalid_constant(value: str) -> None:
        raise ValueError(f'invalid JSON constant: {value}')

    return json.loads(text, object_pairs_hook=pairs, parse_constant=invalid_constant)


def prose_only(text: str) -> tuple[str, bool]:
    """Remove fenced and inline code; return whether a code fence is unclosed."""
    lines: list[str] = []
    current: tuple[str, int] | None = None
    for line in text.splitlines():
        match = re.match(r'^\s{0,3}(`{3,}|~{3,})(.*)$', line)
        if match:
            marker, rest = match.groups()
            if current is None:
                current = (marker[0], len(marker))
                continue
            if marker[0] == current[0] and len(marker) >= current[1] and not rest.strip():
                current = None
                continue
        if current is None:
            lines.append(re.sub(r'(`+).*?\1', '', line))
    return '\n'.join(lines), current is not None


def anchors(text: str) -> set[str]:
    prose, _ = prose_only(text)
    found = set(re.findall(r'<a\s+[^>]*id=["\']([^"\']+)["\']', prose))
    seen: dict[str, int] = {}
    for line in prose.splitlines():
        m = re.match(r'^#{1,6}\s+(.+?)\s*#*$', line)
        if not m:
            continue
        heading = re.sub(r'<[^>]+>', '', m.group(1)).lower()
        slug = ''.join(c for c in heading if c.isalnum() or c in ' _-')
        slug = slug.replace(' ', '-')
        number = seen.get(slug, 0)
        seen[slug] = number + 1
        found.add(slug if number == 0 else f'{slug}-{number}')
    return found


def safe_target(root: Path, source: Path, reference: str) -> tuple[Path | None, str]:
    value = reference.strip()
    # This bundle uses ordinary link targets, not Markdown title annotations.
    if value.startswith('<') and value.endswith('>'):
        value = value[1:-1]
    split = urlsplit(value)
    if split.scheme:
        if split.scheme.lower() not in {'https', 'http', 'mailto'}:
            raise ValueError(f'unsafe or unsupported link scheme: {split.scheme}')
        return None, ''  # External targets are not fetched.
    if split.netloc or value.startswith('/') or '\\' in value:
        raise ValueError('absolute or network link target is not allowed')
    decoded = unquote(split.path)
    if '\\' in decoded or '\x00' in decoded or Path(decoded).is_absolute():
        raise ValueError('invalid relative link target')
    target = source if not decoded else source.parent / decoded
    normalized = target.resolve()
    try:
        normalized.relative_to(root.resolve())
    except ValueError as error:
        raise ValueError('link escapes bundle root') from error
    # Never follow a symlink within a link path, even if it points inside the bundle.
    cursor = root.resolve()
    for part in normalized.relative_to(root.resolve()).parts:
        cursor = cursor / part
        if cursor.is_symlink():
            raise ValueError('symlink target is not allowed')
    # Detect lexical symlinks before resolution, too.
    cursor = target
    while cursor != root and cursor != cursor.parent:
        if cursor.is_symlink():
            raise ValueError('symlink target is not allowed')
        cursor = cursor.parent
    return normalized, unquote(split.fragment)


def check_links(root: Path, source: Path, text: str) -> list[str]:
    errors: list[str] = []
    prose, unclosed = prose_only(text)
    if unclosed:
        errors.append(f'{source.name}: unclosed fenced code block')
    for match in LINK.finditer(prose):
        reference = match.group(1)
        try:
            target, fragment = safe_target(root, source, reference)
            if target is None:
                continue
            if not target.exists():
                errors.append(f'{source.relative_to(root)}: missing link {reference}')
            elif fragment:
                if not target.is_file() or target.suffix.lower() != '.md':
                    errors.append(f'{source.relative_to(root)}: cannot validate anchor {reference}')
                elif fragment not in anchors(target.read_text(encoding='utf-8')):
                    errors.append(f'{source.relative_to(root)}: missing anchor {reference}')
        except (OSError, ValueError, UnicodeError) as error:
            errors.append(f'{source.relative_to(root)}: bad link {reference}: {error}')
    return errors


def check_skill(path: Path, expected_audience: str) -> list[str]:
    """Validate this bundle's constrained frontmatter; not arbitrary YAML."""
    errors: list[str] = []
    text = path.read_text(encoding='utf-8')
    lines = text.splitlines()
    if not lines or lines[0] != '---':
        return [f'{path.parent.name}: missing frontmatter']
    try:
        end = lines.index('---', 1)
    except ValueError:
        return [f'{path.parent.name}: unclosed frontmatter']
    frontmatter = '\n'.join(lines[1:end])
    values: dict[str, str] = {}
    for line in lines[1:end]:
        m = re.match(r'^([a-z][a-z0-9-]*):\s*(.*?)\s*$', line)
        if m:
            if m.group(1) in values:
                errors.append(f'{path.parent.name}: duplicate frontmatter key')
            values[m.group(1)] = m.group(2).strip('"\'')
    name = values.get('name', '')
    description = values.get('description', '')
    if not NAME.fullmatch(name) or len(name) > 64 or name != path.parent.name:
        errors.append(f'{path.parent.name}: invalid/mismatched skill name')
    if not 1 <= len(description) <= 1024:
        errors.append(f'{path.parent.name}: missing/oversized description')
    if not re.search(rf'^\s+audience:\s*["\']?{re.escape(expected_audience)}["\']?\s*$', frontmatter, re.M):
        errors.append(f'{path.parent.name}: unexpected audience')
    if len(lines) > 500:
        errors.append(f'{path.parent.name}: skill exceeds this bundle 500-line design limit')
    if not '\n'.join(lines[end + 1:]).strip():
        errors.append(f'{path.parent.name}: empty skill body')
    return errors


def check_inventory(value: object) -> list[str]:
    if not isinstance(value, dict):
        return ['tool inventory must be an object']
    errors = []
    if value.get('runtimeImportAllowed') is not False:
        errors.append('design inventory must not be runtime importable')
    tools = value.get('tools', [])
    if not isinstance(tools, list):
        return errors + ['tools must be a list']
    names = []
    for row in tools:
        if not isinstance(row, dict) or not TOOL.fullmatch(str(row.get('name', ''))):
            errors.append('invalid tool entry/name')
            continue
        names.append(row['name'])
        if row.get('registrationAllowed') is not False:
            errors.append(f'{row["name"]}: unimplemented design tool cannot be registered')
    if len(names) != len(set(names)):
        errors.append('duplicate tool name')
    if value.get('toolCount') != len(tools):
        errors.append('toolCount differs from tools length')
    return errors


def context_input_semantics(value: object) -> list[str]:
    """Only the supplied messages_context input example; not a Schema engine."""
    if not isinstance(value, dict):
        return ['context input must be an object']
    errors = []
    if set(value) - {'messageId', 'before', 'after'}:
        errors.append('context input has unknown fields')
    mid = value.get('messageId')
    if not isinstance(mid, str) or not 1 <= len(mid) <= 128:
        errors.append('invalid messageId')
    nums = []
    for name in ('before', 'after'):
        n = value.get(name, 20)
        if type(n) is not int or not 0 <= n <= 200:
            errors.append(f'invalid {name}')
        else:
            nums.append(n)
    if len(nums) == 2 and sum(nums) > 200:
        errors.append('before + after exceeds 200')
    return errors


def iter_files(root: Path) -> list[Path]:
    result = []
    for current, dirs, files in os.walk(root, followlinks=False):
        dirs[:] = sorted(d for d in dirs if d not in EXCLUDED_DIRS)
        for d in list(dirs):
            p = Path(current) / d
            if p.is_symlink():
                result.append(p)
                dirs.remove(d)
        result.extend(Path(current) / name for name in sorted(files))
    return sorted(result)


def check_manifest(root: Path) -> list[str]:
    manifest = root / 'MANIFEST.sha256'
    if not manifest.exists():
        return []  # Draft bundle before final checksum generation.
    errors = []
    covered = set()
    for line in manifest.read_text(encoding='utf-8').splitlines():
        if not line.strip():
            continue
        m = re.fullmatch(r'([0-9a-f]{64})  (.+)', line)
        if not m:
            errors.append('malformed checksum line')
            continue
        digest, relative = m.groups()
        try:
            if relative in covered:
                raise ValueError('duplicate checksum path')
            covered.add(relative)
            target, fragment = safe_target(root, root / 'MANIFEST.sha256', relative)
            if target is None or fragment or not target.is_file() or target.name == 'MANIFEST.sha256':
                raise ValueError('invalid checksum target')
            if hashlib.sha256(target.read_bytes()).hexdigest() != digest:
                raise ValueError('checksum mismatch')
        except (OSError, ValueError) as error:
            errors.append(f'manifest {relative}: {error}')
    expected = {p.relative_to(root).as_posix() for p in iter_files(root) if p.is_file() and p.name != 'MANIFEST.sha256'}
    if expected != covered:
        errors.append(f'checksum coverage mismatch: missing={len(expected-covered)} extra={len(covered-expected)}')
    return errors


def validate(root: Path) -> tuple[list[str], dict[str, int]]:
    root = root.resolve()
    errors: list[str] = []
    stats = {'files': 0, 'markdown': 0, 'json': 0, 'developmentSkills': 0, 'runtimeSkills': 0, 'designTools': 0}
    if not root.is_dir():
        return ['bundle root does not exist'], stats
    for relative in REQUIRED:
        if not (root / relative).is_file():
            errors.append(f'missing required file: {relative}')
    names: dict[str, str] = {}
    for path in iter_files(root):
        stats['files'] += 1
        relative = path.relative_to(root).as_posix()
        collision_key = unicodedata.normalize('NFC', relative).casefold()
        if collision_key in names:
            errors.append(f'case/Unicode path collision: {names[collision_key]}, {relative}')
        names[collision_key] = relative
        if path.is_symlink():
            errors.append(f'symlink forbidden in delivery: {relative}')
            continue
        try:
            if path.suffix.lower() in {'.md', '.json', '.yaml', '.py', '.txt', '.sha256', '.log'}:
                if path.stat().st_size > 16 * 1024 * 1024:
                    errors.append(f'oversized documentation file: {relative}')
                    continue
                text = path.read_text(encoding='utf-8')
                if '\x00' in text:
                    errors.append(f'NUL in text file: {relative}')
                if path.suffix == '.md':
                    stats['markdown'] += 1
                    errors.extend(check_links(root, path, text))
                elif path.suffix == '.json':
                    stats['json'] += 1
                    strict_json(text)
        except (UnicodeError, OSError, ValueError) as error:
            errors.append(f'{relative}: {error}')
    for folder, audience, count_key in (('.agents/skills','development','developmentSkills'),('runtime-skills','runtime-readonly','runtimeSkills')):
        for skill in sorted((root / folder).glob('*/SKILL.md')):
            stats[count_key] += 1
            errors.extend(check_skill(skill, audience))
    try:
        inventory = strict_json((root / 'mcp/tool-inventory.json').read_text(encoding='utf-8'))
        errors.extend(check_inventory(inventory))
        stats['designTools'] = len(inventory['tools'])
        errors.extend(context_input_semantics(strict_json((root / 'mcp/examples/messages-context.input.json').read_text(encoding='utf-8'))))
        for relative in ('templates/RELEASE_MANIFEST.example.json', 'templates/IMPLEMENTATION_STATUS.example.json'):
            data = strict_json((root / relative).read_text(encoding='utf-8'))
            if data.get('releaseVerified') is not False:
                errors.append(f'{relative}: template must not claim release verified')
        acceptance = strict_json((root / 'templates/PRODUCT_ACCEPTANCE.example.json').read_text(encoding='utf-8'))
        if acceptance['count'] != len(acceptance['items']) or any(row['state'] != 'not_tested' or row['evidence'] for row in acceptance['items']):
            errors.append('product acceptance template contains invalid count or completion claims')
        for transport in ('remote', 'stdio'):
            data = strict_json((root / f'config/examples/dsh-mcp-{transport}.config.fragment.json').read_text(encoding='utf-8'))
            if data.get('headers', {}) or data.get('env', {}):
                errors.append('example configuration must not carry credentials')
            if data.get('failOnStartupError') is not True:
                errors.append('example configuration hides startup failures')
            if transport == 'remote' and not urlsplit(data['url']).hostname.endswith('.invalid'):
                errors.append('remote example must use .invalid placeholder')
            if transport == 'stdio' and not data['command'].startswith('__VERIFIED_'):
                errors.append('stdio example must use explicit unverified placeholder')
    except (OSError, ValueError, KeyError, TypeError, AttributeError) as error:
        errors.append(f'contract/config validation: {error}')
    errors.extend(check_manifest(root))
    return errors, stats


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--root', type=Path, default=Path(__file__).resolve().parents[1], help='documentation bundle root (default: parent of scripts/)')
    args = parser.parse_args()
    errors, stats = validate(args.root)
    print(json.dumps({'status': 'failed' if errors else 'passed', 'scope': 'documentation_bundle_only', 'stats': stats, 'errors': errors}, ensure_ascii=False, indent=2))
    return 1 if errors else 0


if __name__ == '__main__':
    sys.exit(main())
