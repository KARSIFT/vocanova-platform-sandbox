"""VOC-124 caller regressions against the nested infra checkout when present."""

from __future__ import annotations

import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, REPOSITORY_ROOT, read_fixture


NESTED_INFRA_ROOT = REPOSITORY_ROOT / "karsift-ai-infra"
NESTED_IMPLEMENT = NESTED_INFRA_ROOT / ".github" / "workflows" / "implement.yml"


def caller_publish_job(workflow: str) -> str:
    _, remainder = workflow.split("\n  publish:", 1)
    publish_job, _ = remainder.split("\n  publish-source:", 1)
    return publish_job


def publish_source_job(workflow: str) -> str:
    return workflow[workflow.index("\n  publish-source:") :]


class Voc124ImplementPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.fixture_implement = read_fixture(".github/workflows/implement.yml")
        cls.fixture_pin = read_fixture("PINNED_SHA.txt").strip()
        cls.nested_implement = (
            NESTED_IMPLEMENT.read_text(encoding="utf-8")
            if NESTED_IMPLEMENT.is_file()
            else None
        )

    def test_fixture_pin_unchanged_until_bootstrap_merge(self):
        self.assertEqual(
            self.fixture_pin,
            "7500a4171d96a8e0d38889a9c92ad5dc092ad8dd",
        )

    def test_fixture_implement_still_matches_prior_reviewed_merge(self):
        self.assertNotIn("permission-workflows: write", self.fixture_implement)
        self.assertIn("required human approval are still pending", self.fixture_implement)

    def test_nested_publish_source_mint_requests_workflows_write(self):
        if self.nested_implement is None:
            self.skipTest("nested karsift-ai-infra checkout not present")
        mint = publish_source_job(self.nested_implement)
        self.assertIn("permission-workflows: write", mint)
        self.assertIn("repositories: karsift-ai-infra", mint)

    def test_nested_caller_publish_mint_omits_workflows_write(self):
        if self.nested_implement is None:
            self.skipTest("nested karsift-ai-infra checkout not present")
        publish_job = caller_publish_job(self.nested_implement)
        mint = publish_job.split(
            "- name: Mint least-privilege App token on the clean runner", 1
        )[1]
        self.assertNotIn("permission-workflows: write", mint)
        self.assertIn("cannot publish workflow-file changes", publish_job)

    def test_nested_caller_publish_pr_body_matches_active_a004(self):
        if self.nested_implement is None:
            self.skipTest("nested karsift-ai-infra checkout not present")
        publish_job = caller_publish_job(self.nested_implement)
        self.assertNotIn("required human approval are still pending", publish_job)
        self.assertIn("Independent", publish_job)
        self.assertIn("exact-revision review is still pending", publish_job)

    def test_nested_carrier_files_exist_for_post_merge_fixture_sync(self):
        if self.nested_implement is None:
            self.skipTest("nested karsift-ai-infra checkout not present")
        for relative in (
            ".github/workflows/implement.yml",
            "README.md",
            "tests/test_voc124_workflow_permissions.py",
            "tests/test_voc121_source_carrier_publisher.py",
            "tests/test_voc121_implement_policy.py",
            "tests/test_live_evidence_reconcile.py",
        ):
            self.assertTrue(
                (NESTED_INFRA_ROOT / relative).is_file(),
                f"missing nested path {relative}",
            )


if __name__ == "__main__":
    unittest.main()
