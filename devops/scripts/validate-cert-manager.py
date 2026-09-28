#!/usr/bin/env python3
"""Validate CRs against pinned upstream cert-manager OpenAPI schemas."""
import jsonschema
from pathlib import Path
import sys
import urllib.request
import yaml

for resource in yaml.safe_load_all(Path(sys.argv[1]).read_text()):
    if not resource or resource.get('kind') not in ('Certificate', 'Issuer'):
        continue
    kind = resource['kind'].lower() + 's'
    url = f'https://raw.githubusercontent.com/cert-manager/cert-manager/v1.18.2/deploy/crds/crd-{kind}.yaml'
    with urllib.request.urlopen(url, timeout=30) as response:
        crd = yaml.safe_load(response)
    schema = next(v['schema']['openAPIV3Schema'] for v in crd['spec']['versions'] if v['name'] == 'v1')
    jsonschema.validate(resource, schema)
    print(f"PASS cert-manager v1.18.2 schema: {resource['kind']}")
