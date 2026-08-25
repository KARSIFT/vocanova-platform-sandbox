"""VOC-121 caller fixture regressions for infra contract changes."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


class Voc121ImplementFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.recovery_runner = read_fixture("config/actions-check-recovery-runner.py")
        cls.required_checks = read_fixture("config/required_check_satisfaction.py")
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.readme = read_fixture("README.md")

    def test_fixture_implement_preserves_self_correction_helpers(self):
        self.assertIn("/tmp/karsift-implement-helpers/prepare_cursor_model.py", self.implement)
        self.assertIn("publish-source:", self.implement)
        self.assertIn("repositories: karsift-ai-infra", self.implement)

    def test_fixture_release_checks_required_pr_view(self):
        self.assertIn("gh pr checks \"$PR_NUMBER\" --required", self.release)

    def test_fixture_reruns_ruleset_selected_failed_run(self):
        self.assertIn('f"repos/{repository}/actions/runs/{run_id}/rerun"', self.recovery_runner)
        self.assertIn("payload.get(\"run_attempt\") != 1", self.recovery_runner)
        self.assertIn("plan_required_check_recovery", self.required_checks)
        self.assertIn("ambiguous_required_check_run", self.required_checks)

    def test_fixture_readme_no_false_status_override_claim(self):
        self.assertIn("gh pr checks --required", self.readme)
        self.assertNotIn(
            "a successful exact-head run alone cannot satisfy the PR's required contexts",
            self.readme,
        )

    def test_fixture_pin_matches_reviewed_infrastructure_merge(self):
        expected = "99476c2a1018e42d4bd442657b5257885ac9f1c9"
        self.assertEqual(self.pin, expected)
        self.assertIn(expected, self.readme)
        self.assertNotIn("VOC-121-D10 bootstrap", self.readme)


if __name__ == "__main__":
    unittest.main()
