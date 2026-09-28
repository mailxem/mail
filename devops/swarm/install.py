#!/usr/bin/env python3
"""Single-node Xem bootstrap. Standard library only; never executes config as code."""
import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
import secrets
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
STACK = 'xem'
# Verified multi-architecture image index (linux/amd64 and linux/arm64).
STORAGE_IMAGE = 'rustfs/rustfs@sha256:8cc9801755448b71a786705ce76692c77e14936cccd87cf2fc31842e58f4d1ff'

def run(*args, capture=False):
    return subprocess.run(args, check=True, text=True,
                          stdout=subprocess.PIPE if capture else None).stdout

def save(path, data):
    tmp = path.with_suffix('.tmp')
    with tmp.open('w', encoding='utf-8') as stream:
        os.chmod(tmp, 0o600)
        json.dump(data, stream, indent=2)
        stream.write('\n')
    tmp.replace(path)

def validate(config):
    for key in ('app_url', 'api_url', 'storage_url'):
        url = urllib.parse.urlsplit(config[key])
        if (url.scheme not in ('http', 'https') or not url.hostname or
                url.username or url.password or url.query or url.fragment or
                url.path not in ('', '/') or url.hostname in ('localhost', '127.0.0.1', '::1')):
            raise ValueError(f'{key} must be a public/LAN origin reachable from browsers AND containers; no localhost or URL paths.')
        config[key] = config[key].rstrip('/')
    for key in ('app_port', 'api_port', 'storage_port'):
        if type(config.get(key)) is not int or not 1 <= config[key] <= 65535:
            raise ValueError(f'{key} must be an integer from 1 to 65535')
    if len({config[k] for k in ('app_port', 'api_port', 'storage_port')}) != 3:
        raise ValueError('App, API, and storage ports must differ')

def prompt_config():
    # curl | bash consumes stdin; interactive questions use the controlling terminal.
    with open('/dev/tty', 'r+') as tty:
        def ask(label, default=None):
            tty.write(label + (f' [{default}]' if default is not None else '') + ': ')
            tty.flush()
            value = tty.readline()
            if not value:
                raise ValueError('No terminal input; use --config for unattended setup')
            return value.strip() or default
        print('Use a DNS name or LAN/public IP reachable from this host and its containers.\n'
              'The stack includes PostgreSQL, Redis, and RustFS object storage.')
        return {
            'app_url': ask('App origin (e.g. http://192.168.1.10:3000)'),
            'api_url': ask('API origin (e.g. http://192.168.1.10:9001)'),
            'app_port': int(ask('Published app port', '3000')),
            'api_port': int(ask('Published API port', '9001')),
            'storage_url': ask('Storage origin (e.g. http://192.168.1.10:9000)'),
            'storage_port': int(ask('Published storage port', '9000')),
        }

def stack_spec(c, node, version):
    deploy = {
        'replicas': 1,
        'placement': {'constraints': [f'node.id == {node}']},
        'restart_policy': {'condition': 'any', 'delay': '10s'},
        'update_config': {'parallelism': 1, 'order': 'stop-first', 'failure_action': 'pause'},
    }
    def service(image, **kwargs):
        return {'image': image, 'networks': ['private'], 'deploy': deploy,
                'logging': {'driver': 'json-file', 'options': {'max-size': '10m', 'max-file': '3'}}, **kwargs}
    api_env = {
        'SERVER_HOST': '0.0.0.0', 'SERVER_PORT': '9001', 'PUBLIC_URL': c['api_url'],
        'POSTGRES_HOST': 'postgres', 'POSTGRES_PORT': '5432', 'POSTGRES_USER': 'xem',
        'POSTGRES_DB': 'xem', 'POSTGRES_PASSWORD': c['db_password'], 'POSTGRES_SSLMODE': 'disable',
        'REDIS_HOST': 'redis', 'REDIS_PORT': '6379', 'REDIS_DB': '0',
        'JWT_SECRET': c['jwt_secret'], 'PRIVATE_KEY': c['private_key'],
        'STORAGE_PROVIDER': 's3', 'S3_BUCKET_NAME': 'xem',
        'S3_REGION': 'us-east-1', 'S3_ENDPOINT_URL': c['storage_url'],
        'S3_ACCESS_KEY': c['storage_user'], 'S3_SECRET_KEY': c['storage_password'],
        'CORS_ALLOWED_ORIGINS': c['app_url'], 'APP_ENV': 'production',
        'AI_ENABLED': 'false', 'MANAGED_SENDING_ENABLED': 'false', 'MANAGED_SMTP_ENABLED': 'false',
    }
    # Dollar signs in user credentials must survive Compose interpolation.
    def escaped(env):
        return {k: str(v).replace('$', '$$') for k, v in env.items()}
    return {
        'version': '3.8',
        'services': {
            'postgres': service('postgres:16-alpine', environment=escaped({
                'POSTGRES_USER': 'xem', 'POSTGRES_DB': 'xem', 'POSTGRES_PASSWORD': c['db_password']}),
                volumes=['postgres:/var/lib/postgresql/data'],
                healthcheck={'test': ['CMD-SHELL', 'pg_isready -U xem -d xem'], 'interval': '10s', 'retries': 12}),
            'redis': service('redis:7-alpine', command=['redis-server', '--appendonly', 'yes'],
                volumes=['redis:/data'], healthcheck={'test': ['CMD', 'redis-cli', 'ping'], 'interval': '10s', 'retries': 12}),
            'storage': service(STORAGE_IMAGE,
                environment=escaped({'RUSTFS_ACCESS_KEY': c['storage_user'], 'RUSTFS_SECRET_KEY': c['storage_password'], 'RUSTFS_VOLUMES': '/data', 'RUSTFS_CONSOLE_ENABLE': 'false'}),
                healthcheck={'test': ['CMD', 'curl', '-fsS', 'http://localhost:9000/health'], 'interval': '10s', 'retries': 12},
                volumes=['storage:/data'],
                ports=[{'target': 9000, 'published': c['storage_port'], 'protocol': 'tcp', 'mode': 'host'}]),
            'storage-init': {
                **service(STORAGE_IMAGE,
                    environment=escaped({'STORAGE_USER': c['storage_user'], 'STORAGE_PASSWORD': c['storage_password']}),
                    entrypoint=['/bin/sh', '-c'],
                    command=['for i in $$(seq 1 60); do if curl --max-time 10 -fsS --aws-sigv4 aws:amz:us-east-1:s3 --user "$$STORAGE_USER:$$STORAGE_PASSWORD" --head http://storage:9000/xem >/dev/null || curl --max-time 10 -fsS --aws-sigv4 aws:amz:us-east-1:s3 --user "$$STORAGE_USER:$$STORAGE_PASSWORD" -X PUT http://storage:9000/xem; then exit 0; fi; sleep 5; done; exit 1']),
                'deploy': {**deploy, 'restart_policy': {'condition': 'on-failure', 'delay': '10s', 'max_attempts': 3}},
            },
            'api': service(f'xem-api:{version}', environment=escaped(api_env),
                ports=[{'target': 9001, 'published': c['api_port'], 'protocol': 'tcp', 'mode': 'host'}]),
            'app': service(f'xem-app:{version}', environment=escaped({
                'NEXT_PUBLIC_API_URL': c['api_url'] + '/api/v1',
                'INTERNAL_API_URL': 'http://api:9001/api/v1',
                'NEXTAUTH_URL': c['app_url'], 'AUTH_URL': c['app_url'],
                'NEXTAUTH_SECRET': c['auth_secret'], 'AUTH_SECRET': c['auth_secret'],
                'AUTH_TRUST_HOST': 'true', 'XEM_ASSISTANT_ENABLED': 'false',
            }), ports=[{'target': 3000, 'published': c['app_port'], 'protocol': 'tcp', 'mode': 'host'}]),
        },
        'networks': {'private': {'driver': 'overlay'}},
        'volumes': {'postgres': {}, 'redis': {}, 'storage': {}},
    }

def wait_ready(c, timeout=600):
    deadline = time.monotonic() + timeout
    urls = [c['api_url'] + '/health', c['app_url'] + '/auth/login']
    while time.monotonic() < deadline:
        ready = True
        for url in urls:
            try:
                with urllib.request.urlopen(url, timeout=5) as response:
                    if response.status != 200:
                        ready = False
            except (OSError, urllib.error.URLError):
                ready = False
        if ready:
            return
        time.sleep(5)
    raise RuntimeError('Readiness timed out. Services remain for diagnosis: docker stack ps xem --no-trunc; docker service logs xem_api')

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--config', type=Path, help='JSON configuration; existing state is reused unless supplied')
    parser.add_argument('--state', type=Path, default=Path.home() / '.local/share/xem')
    args = parser.parse_args()
    os.umask(0o077)
    state = args.state.resolve()
    state.mkdir(parents=True, exist_ok=True, mode=0o700)
    # Atomic directory lock prevents concurrent installs/credential replacement.
    lock = state / '.install-lock'
    lock.mkdir()
    try:
        install(args, state)
    finally:
        lock.rmdir()

def install(args, state):
    state_file = state / 'config.json'
    c = json.loads(state_file.read_text()) if state_file.exists() else {}
    if args.config:
        supplied = json.loads(args.config.read_text())
        allowed = {'app_url', 'api_url', 'app_port', 'api_port', 'storage_url', 'storage_port'}
        if set(supplied) - allowed:
            raise ValueError('Unknown configuration keys; use config.example.json')
        c.update(supplied)
    elif not c:
        c = prompt_config()
    validate(c)
    info = json.loads(run('docker', 'info', '--format', '{{json .}}', capture=True))
    if info['OSType'] != 'linux':
        raise ValueError('Linux containers are required')
    swarm = info['Swarm']
    if swarm['LocalNodeState'] not in ('inactive', 'active'):
        raise ValueError('Resolve the existing Swarm state first')
    if swarm['LocalNodeState'] == 'active' and not swarm['ControlAvailable']:
        raise ValueError('Run on a Swarm manager, not a worker')
    existing = run('docker', 'stack', 'ls', '--format', '{{.Name}}', capture=True).split() if swarm['LocalNodeState'] == 'active' else []
    if STACK in existing and not c.get('node_id'):
        raise ValueError('An unowned xem stack exists. Restore its state directory or use another host.')
    if c.get('node_id') and c['node_id'] != swarm.get('NodeID'):
        raise ValueError('Saved state belongs to another Swarm node. Follow the backup/restore guide; refusing to create empty volumes.')
    for key in ('db_password', 'jwt_secret', 'auth_secret', 'storage_password'):
        c.setdefault(key, secrets.token_hex(32))
    c.setdefault('storage_user', 'xemadmin')
    if 'private_key' not in c:
        pem = run('openssl', 'genrsa', '2048', capture=True)
        c['private_key'] = base64.b64encode(pem.encode()).decode()
    # Persist credentials before any potentially failing build/deployment.
    save(state_file, c)
    revision = run('git', '-C', str(ROOT), 'rev-parse', 'HEAD', capture=True).strip()
    build_id = revision[:12] + '-' + hashlib.sha256(c['api_url'].encode()).hexdigest()[:8]
    print('Building frontend and backend from this checkout. The first build may take several minutes.', flush=True)
    run('docker', 'build', '-t', f'xem-api:{build_id}', str(ROOT / 'server'))
    run('docker', 'build', '--build-arg', 'NEXT_PUBLIC_API_URL=' + c['api_url'] + '/api/v1',
        '--build-arg', 'NEXT_PUBLIC_PAYWALL_URL=' + c['app_url'] + '/billing-unavailable',
        '-t', f'xem-app:{build_id}', str(ROOT / 'client'))
    if swarm['LocalNodeState'] == 'inactive':
        run('docker', 'swarm', 'init')
        swarm = json.loads(run('docker', 'info', '--format', '{{json .Swarm}}', capture=True))
    c['node_id'] = swarm['NodeID']
    save(state_file, c)
    spec = stack_spec(c, swarm['NodeID'], build_id)
    stack_file = state / 'stack.json'
    save(stack_file, spec)
    # Validate without printing the resolved service credentials.
    run('docker', 'stack', 'config', '-c', str(stack_file), capture=True)
    run('docker', 'stack', 'deploy', '--resolve-image', 'never', '-c', str(stack_file), STACK)
    print('Waiting for the API and login page (up to 10 minutes)...', flush=True)
    wait_ready(c)
    save(state / 'release.json', {'revision': revision, 'image_tag': build_id, 'node_id': swarm['NodeID']})
    print(f"Ready: {c['app_url']}/auth/register\nState: {state}\nUse your own SMTP provider; managed sending and hosted billing are not included.")

if __name__ == '__main__':
    try:
        main()
    except (ValueError, RuntimeError, OSError, subprocess.CalledProcessError) as error:
        # subprocess errors include commands but not secret-bearing config values.
        print(f'Installation failed: {error}', file=sys.stderr)
        sys.exit(1)
