#!/usr/bin/env python3
"""Local safety checks: plan guard, actual certificate validation/rotation, chart behavior."""
import base64
import contextlib
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import yaml

ROOT = Path(__file__).resolve().parent.parent


def load(name, file):
    spec = importlib.util.spec_from_file_location(name, ROOT / "scripts" / file)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def run(*args):
    return subprocess.run(args, check=True, capture_output=True, text=True).stdout


guard = load("guard", "check-adoption-plan.py")
exporter = load("exporter", "export-smtp-certificate.py")


class ManagedReceivingTemplate(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        path = ROOT / "managed-receiving" / "cloudformation.yaml"
        cls.raw = path.read_text()
        # BaseLoader preserves CloudFormation's short-form intrinsic tags for
        # offline structural assertions without contacting AWS.
        cls.template = yaml.load(cls.raw, Loader=yaml.BaseLoader)
        cls.resources = cls.template["Resources"]

    def test_is_additive_and_never_activates_receipt_rules(self):
        types = [resource["Type"] for resource in self.resources.values()]
        self.assertIn("AWS::SES::ReceiptRuleSet", types)
        self.assertNotIn("AWS::SES::ReceiptRule", types)
        self.assertNotIn("SetActiveReceiptRuleSet", self.raw)
        self.assertNotIn("AWS::S3::BucketName", self.raw)
        for name in ["RawMailBucket", "ReceiptTopic", "ReceiptQueue", "DeadLetterQueue", "InactiveReceiptRuleSet"]:
            self.assertEqual(self.resources[name]["DeletionPolicy"], "Retain")
            self.assertEqual(self.resources[name]["UpdateReplacePolicy"], "Retain")

    def test_private_storage_and_retention_boundaries(self):
        bucket = self.resources["RawMailBucket"]["Properties"]
        self.assertEqual(set(bucket["PublicAccessBlockConfiguration"].values()), {"true"})
        rules = bucket["LifecycleConfiguration"]["Rules"]
        self.assertEqual(rules, [{"Id": "ExpireUncommittedIncomingMIME", "Status": "Enabled",
                                 "Prefix": "incoming/", "ExpirationInDays": "30"}])
        policy = self.resources["RawMailBucketPolicy"]["Properties"]["PolicyDocument"]["Statement"]
        self.assertTrue(any(s["Sid"] == "DenyInsecureTransport" for s in policy))
        ses = next(s for s in policy if s["Sid"] == "AllowSESReceiptWritesForConfiguredRuleSet")
        self.assertEqual(ses["Action"], "s3:PutObject")
        self.assertIn("incoming/*", ses["Resource"])
        self.assertIn("AWS:SourceAccount", ses["Condition"]["StringEquals"])
        self.assertIn("receipt-rule-set/${RuleSetName}:receipt-rule/*", ses["Condition"]["ArnLike"]["AWS:SourceArn"])

    def test_sns_sqs_path_is_restricted_and_bounded(self):
        subscription = self.resources["ReceiptSubscription"]["Properties"]
        self.assertEqual(subscription["RawMessageDelivery"], "false")
        queue = self.resources["ReceiptQueue"]["Properties"]
        self.assertEqual(queue["ReceiveMessageWaitTimeSeconds"], "20")
        self.assertEqual(queue["SqsManagedSseEnabled"], "true")
        self.assertEqual(self.resources["DeadLetterQueue"]["Properties"]["MessageRetentionPeriod"], "1209600")
        statement = self.resources["ReceiptQueuePolicy"]["Properties"]["PolicyDocument"]["Statement"][0]
        self.assertEqual(statement["Sid"], "DenyInsecureTransport")
        statements = self.resources["ReceiptQueuePolicy"]["Properties"]["PolicyDocument"]["Statement"]
        self.assertEqual(self.resources["ReceiptQueuePolicy"]["Properties"]["Queues"], ["ReceiptQueue"])
        self.assertTrue(all(not isinstance(s["Resource"], list) for s in statements))
        allowed = next(s for s in statements if s["Sid"] == "AllowOnlyReceiptTopic")
        self.assertEqual(allowed["Principal"], {"Service": "sns.amazonaws.com"})
        self.assertEqual(allowed["Condition"]["StringEquals"]["aws:SourceArn"], "ReceiptTopic")
        dlq_policy = self.resources["DeadLetterQueuePolicy"]["Properties"]
        self.assertEqual(dlq_policy["Queues"], ["DeadLetterQueue"])
        self.assertEqual(len(dlq_policy["PolicyDocument"]["Statement"]), 1)
        dlq_tls = dlq_policy["PolicyDocument"]["Statement"][0]
        self.assertEqual(dlq_tls["Sid"], "DenyInsecureTransport")
        self.assertEqual(dlq_tls["Resource"], "DeadLetterQueue.Arn")
        self.assertNotIn("AllowOnlyReceiptTopic", str(dlq_policy))
        topic_statements = self.resources["ReceiptTopicPolicy"]["Properties"]["PolicyDocument"]["Statement"]
        topic_tls = next(s for s in topic_statements if s["Sid"] == "DenyInsecureTransport")
        self.assertEqual(topic_tls["Action"], "sns:Publish")
        self.assertFalse(any(
            action == "sns:*"
            for statement in topic_statements
            for action in ([statement["Action"]] if isinstance(statement["Action"], str) else statement["Action"])
        ))
        publish = next(s for s in topic_statements if s["Sid"] == "AllowSESReceiptPublishForConfiguredRuleSet")
        self.assertEqual(publish["Principal"], {"Service": "ses.amazonaws.com"})
        self.assertIn("receipt-rule-set/${RuleSetName}:receipt-rule/*",
                      publish["Condition"]["ArnLike"]["AWS:SourceArn"])

    def test_runtime_policy_has_no_unrelated_storage_or_activation_access(self):
        statements = self.resources["RuntimePolicy"]["Properties"]["PolicyDocument"]["Statement"]
        actions = {action for statement in statements for action in
                   ([statement["Action"]] if isinstance(statement["Action"], str) else statement["Action"])}
        self.assertNotIn("s3:DeleteObject", actions)
        self.assertNotIn("ses:SetActiveReceiptRuleSet", actions)
        self.assertNotIn("ses:DeleteReceiptRuleSet", actions)
        self.assertIn("s3:PutObject", actions)
        self.assertIn("sqs:ReceiveMessage", actions)
        resources = str([statement["Resource"] for statement in statements])
        self.assertIn("incoming/*", resources)
        self.assertIn("mail/*", resources)

    def test_backlog_and_poison_alarms_exist(self):
        for name in ["QueueBacklogAgeAlarm", "DeadLetterQueueAlarm"]:
            alarm = self.resources[name]
            self.assertEqual(alarm["Type"], "AWS::CloudWatch::Alarm")
            self.assertIn("AlarmActions", alarm["Properties"])
            self.assertIn("OKActions", alarm["Properties"])


class AdoptionGuard(unittest.TestCase):
    def plan(self, actions):
        return {"resource_changes": [{"address": "example", "change": {"actions": actions}}]}

    def test_noop(self):
        self.assertEqual(list(guard.violations(self.plan(["no-op"]))), [])

    def test_create_delete_replace_rejected(self):
        for actions in [["create"], ["delete"], ["create", "delete"], ["delete", "create"]]:
            self.assertTrue(list(guard.violations(self.plan(actions), True)))

    def test_update_requires_review(self):
        self.assertTrue(list(guard.violations(self.plan(["update"]))))
        self.assertEqual(list(guard.violations(self.plan(["update"]), True)), [])

    def test_incomplete_input_rejected(self):
        for value in [{}, {"complete": False}, {"errored": True}, {"deferred_changes": [{}]}]:
            self.assertTrue(list(guard.violations(value)))


class Certificates(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.ca = self.root / "ca.pem"
        self.ca_key = self.root / "ca.key"
        run("openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes", "-keyout", str(self.ca_key),
            "-out", str(self.ca), "-days", "3", "-subj", "/CN=Xem Test CA")
        self.ext = self.root / "extensions.txt"
        self.ext.write_text("subjectAltName=DNS:smtp.xem.email\nextendedKeyUsage=serverAuth\n")
        self.source = self.root / "acme.json"
        self.dest = self.root / "smtp"
        for key, value in dict(SOURCE=self.source, DEST=self.dest, LOCK=self.root / "lock", TRUST_STORE=str(self.ca)).items():
            p = patch.object(exporter, key, value); p.start(); self.addCleanup(p.stop)
        p = patch.object(exporter.os, "chown"); p.start(); self.addCleanup(p.stop)
        # The production exporter changes umask; restore the caller's mask after each test.
        old_mask = os.umask(0o077); self.addCleanup(os.umask, old_mask)

    def certificate(self, suffix):
        key = self.root / f"{suffix}.key"; cert = self.root / f"{suffix}.pem"; csr = self.root / f"{suffix}.csr"
        run("openssl", "req", "-newkey", "rsa:2048", "-nodes", "-keyout", str(key), "-out", str(csr), "-subj", "/CN=smtp.xem.email")
        run("openssl", "x509", "-req", "-in", str(csr), "-CA", str(self.ca), "-CAkey", str(self.ca_key),
            "-CAcreateserial", "-out", str(cert), "-days", "3", "-extfile", str(self.ext))
        return cert.read_bytes() + self.ca.read_bytes(), key.read_bytes()

    def write_acme(self, cert, key):
        self.source.write_text(json.dumps({"letsencrypt": {"Certificates": [{"domain": {"main": "smtp.xem.email"},
            "certificate": base64.b64encode(cert).decode(), "key": base64.b64encode(key).decode()}]}}))

    def export(self):
        with contextlib.redirect_stdout(io.StringIO()): exporter.main()

    def test_rotation_permissions_and_invalid_key_preserve_last_good(self):
        cert1, key1 = self.certificate("one"); self.write_acme(cert1, key1); self.export()
        current1 = os.readlink(self.dest / "current")
        self.assertEqual((self.dest / "privkey.pem").stat().st_mode & 0o777, 0o640)
        self.assertEqual((self.dest / "fullchain.pem").read_bytes(), cert1)
        self.export(); self.assertEqual(os.readlink(self.dest / "current"), current1)
        cert2, key2 = self.certificate("two"); self.write_acme(cert2, key2); self.export()
        current2 = os.readlink(self.dest / "current"); self.assertNotEqual(current1, current2)
        self.write_acme(cert2, key1)
        with self.assertRaises(RuntimeError): self.export()
        self.assertEqual(os.readlink(self.dest / "current"), current2)
        self.assertEqual((self.dest / "privkey.pem").read_bytes(), key2)

    def test_wrong_hostname_rejected(self):
        self.ext.write_text("subjectAltName=DNS:other.example.com\nextendedKeyUsage=serverAuth\n")
        cert, key = self.certificate("wrong"); self.write_acme(cert, key)
        with self.assertRaises(subprocess.CalledProcessError): self.export()
        self.assertFalse((self.dest / "current").exists())

    def test_missing_dedicated_certificate_fails_closed(self):
        self.source.write_text('{"letsencrypt":{"Certificates":[]}}')
        with self.assertRaises(RuntimeError): self.export()


class Chart(unittest.TestCase):
    def render(self, *args):
        return list(filter(None, yaml.safe_load_all(run("helm", "template", "test", str(ROOT / "charts/xem"), *args))))

    def test_default_no_public_smtp_or_tokens(self):
        docs = self.render()
        self.assertFalse(any(d['kind'] == 'Service' and d['spec']['type'] == 'LoadBalancer' for d in docs))
        for d in docs:
            if d['kind'] == 'Deployment':
                spec = d['spec']['template']['spec']
                self.assertFalse(spec['automountServiceAccountToken'])
                self.assertFalse(any(v['name'] == 'aws-token' for v in spec['volumes']))

    def test_managed_wiring(self):
        docs = self.render('-f', str(ROOT / 'charts/xem/ci/managed.yaml'))
        backend = next(d for d in docs if d['kind'] == 'Deployment' and d['metadata']['name'].endswith('-backend'))
        spec = backend['spec']['template']['spec']; container = spec['containers'][0]
        self.assertEqual(backend['spec']['replicas'], 1)
        self.assertEqual(backend['spec']['strategy']['type'], 'Recreate')
        self.assertEqual(container['readinessProbe']['tcpSocket']['port'], 'smtp')
        mount = next(v for v in container['volumeMounts'] if v['name'] == 'smtp-tls')
        self.assertTrue(mount['readOnly']); self.assertNotIn('subPath', mount)
        env = {e['name']: e['value'] for e in container['env']}
        self.assertFalse(any(k.startswith('MANAGED_') for k in env))
        self.assertEqual(env['AWS_EC2_METADATA_DISABLED'], 'true')
        token = next(v for v in spec['volumes'] if v['name'] == 'aws-token')
        self.assertEqual(token['projected']['sources'][0]['serviceAccountToken']['audience'], 'sts.amazonaws.com')
        smtp = next(d for d in docs if d['kind'] == 'Service' and d['metadata']['name'].endswith('-smtp'))
        self.assertEqual(smtp['spec']['ports'][0], dict(name='smtp', port=587, targetPort='smtp', protocol='TCP'))
        self.assertEqual(smtp['spec']['externalTrafficPolicy'], 'Local')
        self.assertEqual(len([d for d in docs if d['kind'] == 'Deployment']), 4)
        for d in docs:
            if d['kind'] == 'Deployment':
                c = d['spec']['template']['spec']['containers'][0]
                self.assertTrue(c['securityContext']['readOnlyRootFilesystem'])
                self.assertEqual(c['securityContext']['capabilities']['drop'], ['ALL'])
                self.assertFalse(c['securityContext']['allowPrivilegeEscalation'])

    def test_invalid_settings_rejected(self):
        for setting in ['components.backend.replicas=2', 'workloadIdentity.enabled=false',
                        'managedSending.hostname=', 'components.backend.port=8000',
                        'components.backend.env.MANAGED_SMTP_ENABLED=true', 'components.backend.uid=0',
                        'certificate.issuerRef.kind=ClusterIssuer', 'components.backend.enabled=false',
                        'managedSending.service.annotations.ssl-cert=bad']:
            with self.subTest(setting=setting):
                result = subprocess.run(['helm', 'template', 'test', str(ROOT / 'charts/xem'), '-f',
                    str(ROOT / 'charts/xem/ci/managed.yaml'), '--set', setting], capture_output=True)
                self.assertNotEqual(result.returncode, 0)


if __name__ == '__main__': unittest.main(verbosity=2)
