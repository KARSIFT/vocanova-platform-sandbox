"""VOC-121 caller fixture regressions for infra contract changes."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


class Voc121ImplementFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.readme = read_fixture("README.md")

    def test_fixture_implement_preserves_self_correction_helpers(self):
        self.assertIn("/tmp/karsift-implement-helpers/prepare_cursor_model.py", self.implement)
        self.assertIn("publish-source:", self.implement)
        self.assertIn("repositories: karsift-ai-infra", self.implement)

    def test_fixture_release_checks_required_pr_view(self):
        self.assertIn("gh pr checks \"$PR_NUMBER\" --required", self.release)

    def test_fixture_readme_no_false_status_override_claim(self):
        self.assertIn("gh pr checks --required", self.readme)
        self.assertNotIn(
            "a successful exact-head run alone cannot satisfy the PR's required contexts",
            self.readme,
        )

    def test_fixture_readme_does_not_claim_false_pin_sync(self):
        self.assertIn("VOC-121-D10", self.readme)
        self.assertIn("PINNED_SHA.txt` remains", self.readme)
        self.assertNotIn(
            "fixture file content in this directory is synchronized to that merge",
            self.readme,
        )


if __name__ == "__main__":
    unittest.main()
