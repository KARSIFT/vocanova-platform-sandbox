"""VOC-080 adopt/reconcile idempotency policy regressions (TEST-04)."""

from __future__ import annotations

import unittest

from voc080_fixtures import REPOSITORY_ROOT, read_fixture


class Voc080AdoptionReconcilePolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.adopt = read_fixture(".github/workflows/adopt.yml")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.pipeline = (
            REPOSITORY_ROOT / ".github/workflows/pipeline.yml"
        ).read_text(encoding="utf-8")

    def test_adopt_header_documents_idempotent_reconcile_contract(self):
        self.assertIn(
            "Re-running this workflow with the same merged PR safely reconciles it",
            self.adopt,
        )
        self.assertIn(
            "existing task issues are reused and an unchanged roster is a no-op",
            self.adopt,
        )
        self.assertIn(
            "workflow_dispatch action=reconcile with",
            self.adopt,
        )
        self.assertIn("plan_pr_number=", self.adopt)

    def test_adopt_reuses_existing_task_issues_instead_of_duplicating(self):
        self.assertIn("gh issue list --state all", self.adopt)
        self.assertIn("already exists", self.adopt)
        self.assertIn("reusing it, not creating a duplicate", self.adopt)
        self.assertIn(
            "Reconciliation is already complete; no roster or adoption changes remain",
            self.adopt,
        )

    def test_caller_exposes_reconcile_without_replaying_old_events(self):
        self.assertIn(
            "options: [implement, plan, reconcile, reconcile-release, reconcile-live-evidence, verify-auto-advance-live-evidence, verify-ready-for-review-reuse, verify-remediate-operator-ownership, recover-promotion-pr-checks, verify-promotion-check-recovery, verify-post-promotion-workflow]",
            self.pipeline,
        )
        self.assertIn("inputs.action == 'reconcile'", self.pipeline)
        self.assertIn("plan_pr_number:", self.pipeline)
        self.assertIn(
            "safely repairs the handoff if that",
            self.pipeline,
        )

    def test_release_reconcile_converges_without_founder_comment(self):
        self.assertIn("  converge:", self.release)
        self.assertIn("group: release-converge-", self.release)
        self.assertEqual(self.release.count("gh pr merge"), 1)
        self.assertNotIn("COMMENT_AUTHOR", self.release)
        self.assertNotIn("COMMENT_BODY", self.release)
        self.assertIn("inputs.release_issue_number != ''", self.release)
        self.assertIn("reconcile-release", self.pipeline)


if __name__ == "__main__":
    unittest.main()
