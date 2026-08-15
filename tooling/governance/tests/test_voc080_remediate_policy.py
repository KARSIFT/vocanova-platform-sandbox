"""VOC-080 remediate policy regressions (TEST-02, TEST-05 fail-closed)."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


class Voc080RemediatePolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.remediate = read_fixture(".github/workflows/remediate.yml")

    def test_no_founder_override_path(self):
        self.assertNotIn("founder_username", self.remediate)
        self.assertNotIn("COMMENT_AUTHOR", self.remediate)
        self.assertNotIn("issue_comment:", self.remediate)
        self.assertNotIn("approve-and-merge", self.remediate)
        # Literal approved-comment merge authority must not appear.
        self.assertNotRegex(
            self.remediate,
            r"(?i)reply\s+[`']?approved[`']?",
        )

    def test_retry_is_bounded_and_fail_closed(self):
        self.assertIn('next_attempt=$((attempt + 1))', self.remediate)
        self.assertIn('if [ "$next_attempt" -gt 2 ]; then', self.remediate)
        self.assertIn("not retrying automatically", self.remediate)
        self.assertIn("should_retry=false", self.remediate)

    def test_retry_dispatches_implement_not_human_approval(self):
        self.assertIn("uses: KARSIFT/karsift-ai-infra/.github/workflows/implement.yml@main", self.remediate)
        self.assertIn("needs.decide.outputs.should_retry == 'true'", self.remediate)
        self.assertIn("attempt: ${{ needs.decide.outputs.next_attempt }}", self.remediate)

    def test_missing_identity_fields_refuse_guess(self):
        self.assertIn("refusing to guess", self.remediate)
        self.assertIn("should_retry=false", self.remediate)


if __name__ == "__main__":
    unittest.main()
