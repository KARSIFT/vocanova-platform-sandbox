"""VOC-080 merge-gate policy regressions (TEST-01, TEST-02, TEST-03)."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


class Voc080MergeGatePolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.merge_gate = read_fixture(".github/workflows/merge-gate.yml")

    def test_r0_r4_auto_merge_eligible_when_gates_pass(self):
        auto_merge = self.merge_gate.split("  auto-merge:", 1)[1]
        self.assertIn("inputs.auto_merge_enabled == 'true'", auto_merge)
        self.assertIn("needs.report-status.outputs.risk != 'unknown'", auto_merge)
        self.assertIn("needs.report-status.outputs.checks_ok == 'true'", auto_merge)
        self.assertIn("needs.report-status.outputs.verdict != 'FAIL'", auto_merge)
        self.assertIn("needs.report-status.outputs.verdict != 'PENDING'", auto_merge)
        # R4 is not excluded from the autonomous path.
        self.assertNotIn("risk != 'R4'", auto_merge)
        self.assertNotIn("risk == 'R4'", auto_merge)
        self.assertIn(
            "Master switch for automatic merge at R0-R4",
            self.merge_gate,
        )

    def test_automatic_merge_allowed_false_is_not_a_founder_gate(self):
        self.assertNotIn("automatic_merge_allowed", self.merge_gate)

    def test_unparseable_risk_fails_closed(self):
        self.assertIn('risk="unknown"', self.merge_gate)
        self.assertIn("never auto-mergeable", self.merge_gate)
        self.assertIn(
            "BLOCKED - no parseable 'Risk classification: R#' line found",
            self.merge_gate,
        )
        auto_merge = self.merge_gate.split("  auto-merge:", 1)[1]
        self.assertIn("risk != 'unknown'", auto_merge)

    def test_founder_comment_is_not_a_merge_authority_path(self):
        self.assertNotIn("approve-and-merge:", self.merge_gate)
        self.assertNotIn("COMMENT_AUTHOR", self.merge_gate)
        self.assertNotIn("issue_comment:", self.merge_gate)
        self.assertIn(
            "Founder comments are not merge authority",
            self.merge_gate,
        )
        self.assertIn(
            "Founder approval comments are not part of the merge path",
            self.merge_gate,
        )

    def test_failed_or_missing_verdict_blocks_auto_merge(self):
        auto_merge = self.merge_gate.split("  auto-merge:", 1)[1]
        self.assertIn("verdict != 'FAIL'", auto_merge)
        self.assertIn("verdict != 'PENDING'", auto_merge)
        self.assertIn("checks_ok == 'true'", auto_merge)


if __name__ == "__main__":
    unittest.main()
