#!/usr/bin/env python3
"""Reject submodules and component-local GitHub configuration in the monorepo."""
from pathlib import Path
import subprocess

violations = []
if Path('.gitmodules').exists():
    violations.append('.gitmodules is not supported')
for entry in subprocess.check_output(['git', 'ls-files', '--stage', '-z']).decode().split('\0'):
    if not entry:
        continue
    metadata, path = entry.split('\t', 1)
    if metadata.startswith('160000 '):
        violations.append('Submodule: ' + path)
    if '/.github/' in path:
        violations.append('Move GitHub configuration to the root: ' + path)
if violations:
    raise SystemExit('\n'.join(violations))
print('All components are ordinary directories; GitHub configuration is centralized.')
