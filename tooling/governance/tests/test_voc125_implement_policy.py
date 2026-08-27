"""VOC-125 caller fixture regressions for existing-carrier resume identity."""

from __future__ import annotations

import unittest

from voc080_fixtures import read_fixture


class Voc125ImplementFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.remediate = read_fixture(".github/workflows/remediate.yml")
        cls.pipeline_template = read_fixture(
            "templates/project-repo/.github/workflows/pipeline.yml"
        )
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.readme = read_fixture("README.md")
        cls.voc125_tests = read_fixture("tests/test_voc125_existing_carrier.py")

    def test_fixture_pin_matches_voc126_infra_merge(self):
        expected = "ac0edc4b5b8f6165fa5e23a7b166dc2a0c2ea18f"
        self.assertEqual(self.pin, expected)
        self.assertIn(expected, self.readme)
        self.assertIn("VOC-126-T00", self.readme)

    def test_fixture_implement_declares_existing_pr_number(self):
        self.assertIn("existing_pr_number:", self.implement)
        self.assertIn("Current state (VOC-125, 2026-08-26)", self.implement)
        self.assertIn("Bind existing-carrier recovery identity", self.implement)
        self.assertIn("bind-existing-carrier-runner.py", self.implement)
        bind_index = self.implement.index("- name: Bind existing-carrier recovery identity")
        branch_index = self.implement.index("- name: Create implementation branch")
        model_index = self.implement.index("- name: Resolve implementer model")
        self.assertLess(bind_index, branch_index)
        self.assertLess(bind_index, model_index)

    def test_fixture_remediate_retry_forwards_existing_pr_number(self):
        retry = self.remediate.split("  retry:", 1)[1]
        self.assertIn("existing_pr_number: ${{ inputs.pr_number }}", retry)
        self.assertIn("expected_head_sha: ${{ inputs.expected_head_sha }}", retry)
        self.assertIn("expected_base_sha: ${{ inputs.expected_base_sha }}", retry)

    def test_fixture_template_pipeline_exposes_existing_pr_number_only(self):
        dispatch = self.pipeline_template.split("workflow_dispatch:", 1)[1].split(
            "jobs:", 1
        )[0]
        self.assertIn("existing_pr_number:", dispatch)
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("expected_base_sha:", dispatch)
        implement = self.pipeline_template.split("  implement:", 1)[1].split(
            "\n  plan:", 1
        )[0]
        self.assertIn("existing_pr_number: ${{ inputs.existing_pr_number }}", implement)

    def test_fixture_includes_voc125_existing_carrier_tests(self):
        self.assertIn("test_attempt3_fails_closed", self.voc125_tests)
        self.assertIn("test_attempt1_with_existing_carrier_fails_closed", self.voc125_tests)
        self.assertIn("test_implement_workflow_declares_bind_step_before_branch", self.voc125_tests)
        self.assertIn("from bind_existing_carrier import", self.voc125_tests)

    def test_fixture_readme_documents_operator_resume_on_pipeline(self):
        self.assertIn("existing_pr_number=<open PR>", self.readme)
        self.assertIn("pipeline-verify.yml", self.readme)


if __name__ == "__main__":
    unittest.main()
