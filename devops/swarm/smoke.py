"""Disposable CI signup/login/API check. Never sends email."""
import json
from pathlib import Path
import secrets
import sys
import urllib.request

config = json.loads(Path(sys.argv[1]).read_text())
base = config['api_url'] + '/api/v1'
account = {'email': 'smoke-' + secrets.token_hex(6) + '@example.invalid',
           'password': secrets.token_urlsafe(24), 'first_name': 'Smoke', 'last_name': 'Test'}

def request(path, payload=None, token=None):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    req = urllib.request.Request(base + path, data=json.dumps(payload).encode() if payload else None, headers=headers)
    with urllib.request.urlopen(req, timeout=30) as response:
        return json.load(response)

request('/auth/register', account)
login = request('/auth/login', {'email': account['email'], 'password': account['password']})
user = request('/users/me', token=login['token'])
assert user['email'] == account['email']
boundary = 'xem-' + secrets.token_hex(16)
payload = (f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="smoke.txt"\r\n'
           f'Content-Type: text/plain\r\n\r\nXem storage smoke test\r\n--{boundary}--\r\n').encode()
upload = urllib.request.Request(base + '/files/upload', data=payload, headers={
    'Content-Type': 'multipart/form-data; boundary=' + boundary,
    'Authorization': 'Bearer ' + login['token'],
})
with urllib.request.urlopen(upload, timeout=30) as response:
    assert json.load(response)['file']
print('Signup, API login, authenticated lookup, and upload passed. No email sent.')
