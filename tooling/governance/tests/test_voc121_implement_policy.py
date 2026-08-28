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
        expected = "599436835371f27fac52ec6b47a18b36257366ac"
        self.assertEqual(self.pin, expected)
        self.assertIn("123735c80fec813a5b46a004f3e1122bd425cde2", self.readme)
        self.assertIn(expected, self.readme)
        self.assertNotIn("VOC-121-D10 bootstrap", self.readme)

    def test_fixture_implement_records_voc124_publish_source_workflows_write(self):
        source_publisher = self.implement[self.implement.index("\n  publish-source:") :]
        mint = source_publisher[
            source_publisher.index(
                "- name: Mint least-privilege App token for infrastructure repository"
            ) :
        ]
        self.assertIn("permission-workflows: write", mint)
        _, remainder = self.implement.split("\n  publish:", 1)
        publish_job, _ = remainder.split("\n  publish-source:", 1)
        self.assertNotIn("permission-workflows: write", publish_job)

    def test_fixture_implement_uses_named_ref_source_bundle(self):
        self.assertIn('"$HELPER_DIR/implementer_source_carrier.py"', self.implement)
        self.assertIn("create-bundle \\", self.implement)
        self.assertIn("--output /tmp/implementer-source.bundle", self.implement)
        self.assertNotIn(
            '"${{ steps.infra-checkout.outputs.base_sha }}..$SOURCE_HEAD_SHA"',
            self.implement,
        )
        self.assertIn("Current state (VOC-123, 2026-08-26)", self.implement)
        self.assertIn("VOC-123 (named-ref nested source-carrier bundle tips)", self.readme)


if __name__ == "__main__":
    unittest.main()
