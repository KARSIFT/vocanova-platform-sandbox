"""VOC-080 builder/verifier separation policy regressions (TEST-07)."""

from __future__ import annotations

import re
import unittest

from voc080_fixtures import read_fixture


class Voc080RoleSeparationPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.roles = read_fixture("config/roles.yml")
        cls.implement = read_fixture("prompts/implement.md")
        cls.review = read_fixture("prompts/review.md")
        cls.plan_review = read_fixture("prompts/plan-review.md")

    def _role_value(self, role: str) -> str:
        match = re.search(rf"^{role}:\s*(\S+)\s*$", self.roles, re.MULTILINE)
        self.assertIsNotNone(match, msg=f"missing role binding for {role}")
        return match.group(1)

    def test_implementer_and_reviewer_are_distinct_bindings(self):
        implementer = self._role_value("implementer")
        reviewer = self._role_value("reviewer")
        self.assertNotEqual(
            implementer,
            reviewer,
            msg="implementer and reviewer must not share the same model binding",
        )

    def test_implementer_prompt_forbids_self_review_and_self_merge(self):
        self.assertIn("You cannot merge your own work", self.implement)
        self.assertIn(
            "You cannot mark yourself as the independent reviewer",
            self.implement,
        )

    def test_reviewer_prompt_is_verification_not_approval_authority(self):
        self.assertIn("independent verifier", self.review)
        self.assertIn("cannot grant founder or steward approval", self.review)
        self.assertIn(
            "whether the implementer attempted to approve or merge its own work",
            self.review,
        )

    def test_plan_reviewer_cannot_grant_adoption_approval(self):
        self.assertIn("cannot grant founder or adoption approval", self.plan_review)
        self.assertIn("no repository-write, merge", self.plan_review)


if __name__ == "__main__":
    unittest.main()
