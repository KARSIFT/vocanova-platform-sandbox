"""VOC-080 plan vs task PR path policy regressions (TEST-01)."""

from __future__ import annotations

import unittest

from voc080_fixtures import FIXTURE_INFRA_ROOT, REPOSITORY_ROOT, read_fixture


class Voc080PlanPathPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.plan_review = read_fixture(".github/workflows/plan-review.yml")
        cls.review = read_fixture(".github/workflows/review.yml")
        cls.merge_gate = read_fixture(".github/workflows/merge-gate.yml")
        cls.pipeline = (
            REPOSITORY_ROOT / ".github/workflows/pipeline.yml"
        ).read_text(encoding="utf-8")

    def test_plan_review_uses_plan_reviewer_role(self):
        self.assertIn("plan_reviewer", self.plan_review)
        self.assertIn("Resolve plan_reviewer model", self.plan_review)
        self.assertIn("prompts/plan-review.md", self.plan_review)
        self.assertIn("**Independent verification - bound to commit", self.plan_review)

    def test_task_review_uses_reviewer_role(self):
        self.assertIn("Resolve reviewer model", self.review)
        self.assertIn("prompts/review.md", self.review)
        self.assertIn("**Independent verification - bound to commit", self.review)

    def test_caller_routes_plan_and_agent_branches_separately(self):
        self.assertIn("startsWith(github.head_ref, 'plan/')", self.pipeline)
        self.assertIn("startsWith(github.head_ref, 'agent/')", self.pipeline)
        self.assertIn("plan-review:", self.pipeline)
        self.assertIn("  review:", self.pipeline)

    def test_merge_gate_requires_independent_verification_verdict(self):
        self.assertIn("Independent verification", self.merge_gate)
        self.assertIn("verdict != 'PENDING'", self.merge_gate)
        self.assertIn("verdict != 'FAIL'", self.merge_gate)

    def test_fixture_pin_is_recorded(self):
        pin = (FIXTURE_INFRA_ROOT / "PINNED_SHA.txt").read_text(encoding="utf-8").strip()
        self.assertRegex(pin, r"^[0-9a-f]{40}$")


if __name__ == "__main__":
    unittest.main()
