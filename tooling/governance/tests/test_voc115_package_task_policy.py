"""VOC-115 package/task default and split-reason policy regressions."""

from __future__ import annotations

from pathlib import Path
import sys
import unittest

import yaml

from voc080_fixtures import FIXTURE_INFRA_ROOT, REPOSITORY_ROOT, read_fixture

FIXTURE_CONFIG = FIXTURE_INFRA_ROOT / "config"
if str(FIXTURE_CONFIG) not in sys.path:
    sys.path.insert(0, str(FIXTURE_CONFIG))

from package_task_policy import (  # noqa: E402
    PackageTaskPolicyError,
    validate_package_tasks,
)
from auto_advance_ownership import next_roster_task  # noqa: E402


class Voc115PackageTaskPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.plan_prompt = read_fixture("prompts/plan.md")
        cls.plan_review_prompt = read_fixture("prompts/plan-review.md")
        cls.implement_prompt = read_fixture("prompts/implement.md")
        cls.review_prompt = read_fixture("prompts/review.md")
        cls.plan_workflow = read_fixture(".github/workflows/plan.yml")
        cls.adopt_workflow = read_fixture(".github/workflows/adopt.yml")
        cls.auto_advance_workflow = read_fixture(".github/workflows/auto-advance.yml")
        cls.voc115_tasks = (
            REPOSITORY_ROOT
            / "specs/changes/VOC-115-reduce-plan-and-task-fragmentation-in-governed/tasks.md"
        ).read_text(encoding="utf-8")

    def test_planner_prompt_defaults_to_one_task(self):
        self.assertIn("largest safe, coherent change unit", self.plan_prompt)
        self.assertIn("broad or massive", self.plan_prompt)
        self.assertIn("minimum sufficient number of tasks", self.plan_prompt)
        self.assertIn("one end-to-end", self.plan_prompt)
        self.assertIn("implementation task", self.plan_prompt)
        self.assertNotIn("small,\nordered", self.plan_prompt)

    def test_plan_review_requires_split_reasons_for_extra_tasks(self):
        normalized = " ".join(self.plan_review_prompt.split())
        self.assertIn("minimum sufficient number of maximal tasks", normalized)
        self.assertIn("hard dependency", normalized)
        self.assertIn("More than three tasks is exceptional", normalized)

    def test_plan_workflow_enforces_task_policy_in_retry_loop(self):
        self.assertIn("package-task-policy-runner.py validate", self.plan_workflow)
        self.assertIn("minimum sufficient number of maximal tasks", self.plan_workflow)

    def test_implement_and_review_prompts_keep_causal_work_together(self):
        self.assertIn("whole named task", self.implement_prompt)
        self.assertIn("causally in scope", self.implement_prompt)
        normalized = " ".join(self.review_prompt.split())
        self.assertIn("largest safe coherent unit", normalized)
        self.assertIn("Do not request a split", normalized)

    def test_fixture_is_pinned_to_final_shared_source_merge(self):
        pin = (FIXTURE_INFRA_ROOT / "PINNED_SHA.txt").read_text(encoding="utf-8").strip()
        self.assertEqual(pin, "3fd40f52aba602fab8399482bc5b772731675d1a")

    def test_voc115_package_is_one_task_by_default(self):
        sections = validate_package_tasks(self.voc115_tasks, "VOC-115")
        self.assertEqual(len(sections), 1)
        self.assertEqual(sections[0].task_id, "VOC-115-T00")

    def test_agents_md_distinguishes_in_scope_remediation(self):
        agents = (REPOSITORY_ROOT / "AGENTS.md").read_text(encoding="utf-8")
        self.assertIn("In-scope causal remediation under an active package", agents)
        self.assertIn("unrelated bug fixes", agents)

    def test_doc16_states_one_task_default_and_bounded_remediation(self):
        doc16 = (
            REPOSITORY_ROOT / "docs/governance/16-autonomous-development-operating-model.md"
        ).read_text(encoding="utf-8")
        self.assertIn("one end-to-end implementation task", doc16.lower())
        self.assertIn("In-scope causal remediation", doc16)

    def test_doc15_replaced_seven_task_example(self):
        doc15 = (
            REPOSITORY_ROOT
            / "docs/operations/15-ai-native-product-and-engineering-operating-model.md"
        ).read_text(encoding="utf-8")
        self.assertIn("VOC-023-T00 Implement approved authentication end to end", doc15)
        self.assertNotIn("VOC-023-T07 Provide acceptance-criteria evidence", doc15)

    def test_development_workflow_no_longer_commands_800_line_split(self):
        workflow = (
            REPOSITORY_ROOT / "docs/operations/10-development-workflow.md"
        ).read_text(encoding="utf-8")
        self.assertNotIn("over 800 normally split", workflow)
        self.assertIn("review signals", workflow)

    def test_invalid_split_reason_fixture_fails_closed(self):
        text = """# VOC-999 — Tasks

## VOC-999-T00 — First

## VOC-999-T01 — Second

- Split reason: small
"""
        with self.assertRaises(PackageTaskPolicyError) as ctx:
            validate_package_tasks(text, "VOC-999")
        self.assertEqual(ctx.exception.code, "invalid_split_reason_slug")

    def test_allowed_slug_without_concrete_explanation_fails_closed(self):
        text = """# VOC-999 — Tasks

## VOC-999-T00 — First

## VOC-999-T01 — Second

- Split reason: merge-order-dependency
"""
        with self.assertRaises(PackageTaskPolicyError) as ctx:
            validate_package_tasks(text, "VOC-999")
        self.assertEqual(ctx.exception.code, "split_reason_requires_explanation")

    def test_justified_multi_task_package_preserves_sequential_advancement(self):
        text = """# VOC-999 — Tasks

## VOC-999-T00 — Merge the independently owned contract

## VOC-999-T01 — Implement the dependent runtime

- Split reason: merge-order-dependency — runtime cannot land until the independently merged contract exists.
"""
        sections = validate_package_tasks(text, "VOC-999")
        self.assertEqual(
            [section.task_id for section in sections],
            ["VOC-999-T00", "VOC-999-T01"],
        )

        roster = [
            {"task_id": "VOC-999-T00", "issue": 100, "depends_on": []},
            {
                "task_id": "VOC-999-T01",
                "issue": 101,
                "depends_on": ["VOC-999-T00"],
            },
        ]
        self.assertEqual(
            next_roster_task(roster, "VOC-999-T00"),
            ("VOC-999-T01", 101),
        )
        self.assertIsNone(next_roster_task(roster, "VOC-999-T01"))
        self.assertIn(
            'depends_on=$(jq -cn --arg dependency "$previous_task_id"',
            self.adopt_workflow,
        )
        self.assertIn('previous_task_id="$task_id"', self.adopt_workflow)
        self.assertIn("task-completion-runner.py validate-task", self.auto_advance_workflow)
        self.assertIn("next_roster_task(roster, closed)", self.auto_advance_workflow)

    def test_plan_gate_uses_adoption_yaml_loader_and_rejects_bad_apostrophe(self):
        plan_loader = 'yaml.safe_load(open(sys.argv[1]))'
        adoption_loader = 'data = yaml.safe_load(open(path))'
        self.assertIn(plan_loader, self.plan_workflow)
        self.assertIn(adoption_loader, self.adopt_workflow)
        self.assertLess(
            self.plan_workflow.index(plan_loader),
            self.plan_workflow.index("- name: Open draft PR from clean runner"),
        )

        change_yaml = (
            REPOSITORY_ROOT
            / "specs/changes/VOC-115-reduce-plan-and-task-fragmentation-in-governed/change.yaml"
        ).read_text(encoding="utf-8")
        self.assertEqual(yaml.safe_load(change_yaml)["id"], "VOC-115")
        invalid = "id: VOC-999\nrequirement_source: 'founder's request'\n"
        with self.assertRaises(yaml.YAMLError):
            yaml.safe_load(invalid)


if __name__ == "__main__":
    unittest.main()
