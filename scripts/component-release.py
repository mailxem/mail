#!/usr/bin/env python3
"""Independent, collision-free release versions and scoped notes for the monorepo."""
import re
import subprocess
import sys


def git(*args):
    return subprocess.check_output(['git', *args], text=True).strip()


def next_version(component, tags):
    if component not in ('frontend', 'backend', 'payments'):
        raise ValueError('Unknown component')
    matches = []
    for tag in tags:
        match = re.fullmatch(re.escape(component) + r'-v(\d+)\.(\d+)\.(\d+)', tag)
        if match:
            matches.append((tuple(map(int, match.groups())), tag))
    if not matches:
        return '0.1.0', ''
    (major, minor, patch), previous = max(matches)
    return f'{major}.{minor}.{patch + 1}', previous


if __name__ == '__main__':
    if sys.argv[1] == 'version':
        component = sys.argv[2]
        version, previous = next_version(component, git('tag', '--list').splitlines())
        print(f'version={version}\ntag={component}-v{version}\nprevious={previous}')
    elif sys.argv[1] == 'notes':
        directory, previous = sys.argv[2:4]
        if directory not in ('client', 'server', 'payments.go', 'sdk-go'):
            raise ValueError('Unknown component directory')
        revision = f'{previous}..HEAD' if previous else 'HEAD'
        print(f'Changes in `{directory}/`\n')
        print(git('log', '--no-merges', '--format=- %s (%h)', revision, '--', directory) or 'Build and release workflow update.')
    else:
        raise SystemExit('Use version <component> or notes <directory> <previous-tag>')
