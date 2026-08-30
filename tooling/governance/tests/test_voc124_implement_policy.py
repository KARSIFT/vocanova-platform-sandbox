"""VOC-124 caller fixture regressions for publish-source workflow-write permission."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


def caller_publish_job(workflow: str) -> str:
    _, remainder = workflow.split("\n  publish:", 1)
    publish_job, _ = remainder.split("\n  publish-source:", 1)
    return publish_job


def publish_source_job(workflow: str) -> str:
    return workflow[workflow.index("\n  publish-source:") :]


class Voc124ImplementFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.readme = read_fixture("README.md")

    def test_fixture_pin_matches_voc124_bootstrap_merge(self):
        expected = "9fdff24cd387cc2cdc468c84a3012b0c34b6c8e8"
        self.assertEqual(self.pin, expected)
        self.assertIn("123735c80fec813a5b46a004f3e1122bd425cde2", self.readme)
        self.assertIn(expected, self.readme)
        self.assertIn("VOC-124-T00", self.readme)

    def test_fixture_publish_source_mint_requests_workflows_write(self):
        mint = publish_source_job(self.implement)
        self.assertIn("permission-workflows: write", mint)
        self.assertIn("repositories: karsift-ai-infra", mint)
        self.assertIn("permission-contents: write", mint)
        self.assertIn("permission-issues: write", mint)
        self.assertIn("permission-pull-requests: write", mint)

    def test_fixture_caller_publish_mint_omits_workflows_write(self):
        publish_job = caller_publish_job(self.implement)
        mint = publish_job.split(
            "- name: Mint least-privilege App token on the clean runner", 1
        )[1]
        self.assertNotIn("permission-workflows: write", mint)
        self.assertIn("cannot publish workflow-file changes", publish_job)

    def test_fixture_caller_publish_pr_body_matches_active_a004(self):
        publish_job = caller_publish_job(self.implement)
        self.assertNotIn("required human approval are still pending", publish_job)
        self.assertIn("Independent", publish_job)
        self.assertIn("exact-revision review is still pending", publish_job)
        self.assertIn("not authorized to", publish_job)
        self.assertIn("merge on its own", publish_job)

    def test_fixture_implement_records_voc124_current_state(self):
        self.assertIn("Current state (VOC-124, 2026-08-26)", self.implement)
        self.assertIn("requests `workflows: write`", self.implement)

    def test_fixture_includes_voc124_workflow_permission_tests(self):
        voc124 = read_fixture("tests/test_voc124_workflow_permissions.py")
        self.assertIn("permission-workflows: write", voc124)
        self.assertIn("test_publish_source_mint_requests_workflows_write", voc124)


if __name__ == "__main__":
    unittest.main()
