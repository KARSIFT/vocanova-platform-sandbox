"""VOC-125 caller fixture regressions for existing-carrier resume identity."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


class Voc125ImplementFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.remediate = read_fixture(".github/workflows/remediate.yml")
        cls.pipeline = read_fixture(
            "templates/project-repo/.github/workflows/pipeline.yml"
        )
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.readme = read_fixture("README.md")

    def test_fixture_pin_records_voc125_content(self):
        self.assertEqual(self.pin, "9fdff24cd387cc2cdc468c84a3012b0c34b6c8e8")
        self.assertIn("existing_pr_number", self.implement)
        self.assertIn("existing_pr_number=<open PR>", self.readme)

    def test_fixture_implement_bind_step_before_branch(self):
        self.assertIn("- name: Bind existing-carrier recovery identity", self.implement)
        bind_index = self.implement.index(
            "- name: Bind existing-carrier recovery identity"
        )
        branch_index = self.implement.index("- name: Create implementation branch")
        self.assertLess(bind_index, branch_index)

    def test_fixture_remediate_forwards_existing_pr_number(self):
        retry = self.remediate.split("  retry:", 1)[1]
        self.assertIn("existing_pr_number: ${{ inputs.pr_number }}", retry)

    def test_fixture_template_pipeline_exposes_existing_pr_number(self):
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split("jobs:", 1)[0]
        self.assertIn("existing_pr_number:", dispatch)
        self.assertNotIn("expected_head_sha:", dispatch)
        implement = self.pipeline.split("  implement:", 1)[1].split("\n  plan:", 1)[0]
        self.assertIn(
            "existing_pr_number: ${{ inputs.existing_pr_number }}", implement
        )


if __name__ == "__main__":
    unittest.main()
