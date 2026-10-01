#!/usr/bin/env python3
"""Focused contracts for release-triggered Hakopod deployments."""

from pathlib import Path
import re
import unittest


ROOT = Path(__file__).resolve().parents[1]


def workflow(name: str) -> str:
    return (ROOT / ".github" / "workflows" / name).read_text()


class HakopodWorkflowTests(unittest.TestCase):
    def test_shared_lock_preserves_every_pending_deployment(self):
        deploy = workflow("hakopod-deploy.yml")
        self.assertRegex(
            deploy,
            re.compile(
                r"concurrency:\n\s+group: hakopod-\$\{\{ vars\.HAKOPOD_APPLICATION_ID \}\}\n"
                r"\s+queue: max\n\s+cancel-in-progress: false"
            ),
        )
        self.assertIn("permissions: {}", deploy)
        self.assertIn("uses: hakopod/deploy@v1", deploy)
        self.assertIn("wait: 'true'", deploy)
        self.assertIn("timeout: '600'", deploy)

        for filename in (
            "backend-release.yml",
            "frontend-release.yml",
            "component-release.yml",
            "mcp-container.yml",
            "payments-release.yml",
        ):
            release = workflow(filename)
            self.assertIn("queue: max", release, filename)
            self.assertNotRegex(release, r"queue: max\n\s+cancel-in-progress: (?!false)")

    def test_every_service_deploys_an_immutable_published_image(self):
        expected = {
            "backend-release.yml": "backend",
            "frontend-release.yml": "frontend",
            "mcp-container.yml": "mcp",
            "payments-release.yml": "payments",
        }
        for filename, service in expected.items():
            release = workflow(filename)
            self.assertIn("uses: ./.github/workflows/hakopod-deploy.yml", release)
            self.assertIn(f"service: {service}", release)
            self.assertRegex(release, r"image: \$\{\{ needs\.(release|publish)\.outputs\.image \}\}")

    def test_frontend_waits_outside_deployment_lock_and_mcp_tags_do_not_deploy(self):
        frontend = workflow("frontend-release.yml")
        self.assertLess(frontend.index("wait-for-backend:"), frontend.index("\n  deploy:"))
        self.assertIn("needs: [release, wait-for-backend]", frontend)
        self.assertIn("!cancelled()", frontend)

        mcp = workflow("mcp-container.yml")
        self.assertIn("tags: ['mcp-v*']", mcp)
        self.assertIn("if: github.ref == format('refs/heads/{0}', github.event.repository.default_branch)", mcp)


if __name__ == "__main__":
    unittest.main()
