"""VOC-122 caller fixture regressions for promotion-recovery replan."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


class Voc122ImplementFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.recovery_runner = read_fixture("config/actions-check-recovery-runner.py")
        cls.recover_workflow = read_fixture(".github/workflows/recover-actions-checks.yml")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.readme = read_fixture("README.md")
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.voc122_tests = read_fixture("tests/test_voc122_actions_check_recovery.py")

    def test_fixture_runner_replans_during_promotion_polling(self):
        self.assertIn("apply_promotion_pr_recovery_plan", self.recovery_runner)
        self.assertIn("dispatched_contexts: set[str]", self.recovery_runner)
        loop_body = self.recovery_runner.split("while time.time() < deadline:", 1)[1]
        self.assertIn("apply_promotion_pr_recovery_plan(", loop_body)
        self.assertNotIn(
            "apply_integration_push_recovery_plan(",
            loop_body,
        )

    def test_fixture_recovery_workflow_documents_replan_contract(self):
        self.assertIn("VOC-122", self.recover_workflow)
        self.assertIn("replans against the", self.recover_workflow)
        self.assertIn("not only from the first", self.recover_workflow)

    def test_fixture_release_recovery_step_documents_replan_contract(self):
        recovery_step = self.release.split("Recover missing exact-head promotion checks", 1)[1]
        self.assertIn("replans against GitHub's current required PR view", recovery_step)
        self.assertIn("after the first snapshot", recovery_step)

    def test_fixture_includes_time_evolving_voc122_tests(self):
        self.assertIn("test_absent_then_cancelled_selected_row_is_rerun_once_and_succeeds", self.voc122_tests)
        self.assertIn("test_ambiguous_required_check_run_fails_closed_on_replan", self.voc122_tests)
        self.assertIn("32912850066", self.voc122_tests)

    def test_fixture_pin_matches_reviewed_infrastructure_merge(self):
        expected = "863fc1f35b1d35e4981a59166b0e939be1a2b681"
        self.assertEqual(self.pin, expected)
        self.assertIn(expected, self.readme)


if __name__ == "__main__":
    unittest.main()
