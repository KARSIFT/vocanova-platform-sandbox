"""VOC-126 caller workflow_dispatch input-count and relocation regression."""

from __future__ import annotations

import re
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
MAX_DISPATCH_INPUTS = 25
INVALID_VOC125_TEMPLATE_SHA = "1f1705dbad41729563b0ad1e878e4154e5511e93"

PIPELINE_MUTATING_OPTIONS = (
    "implement, plan, reconcile, reconcile-release, reconcile-live-evidence, "
    "recover-integration-push, recover-promotion-pr-checks"
)
PIPELINE_VERIFY_OPTIONS = (
    "verify-auto-advance-live-evidence, verify-ready-for-review-reuse, "
    "verify-remediate-operator-ownership, verify-promotion-check-recovery, "
    "verify-post-promotion-workflow"
)


def count_workflow_dispatch_inputs(text: str) -> int:
    in_inputs = False
    count = 0
    for line in text.splitlines():
        if re.match(r"^  workflow_dispatch:\s*$", line):
            in_inputs = False
            continue
        if re.match(r"^    inputs:\s*$", line):
            in_inputs = True
            continue
        if in_inputs and re.match(r"^    [A-Za-z]", line) and not line.startswith("      "):
            in_inputs = False
        if in_inputs and re.match(r"^      (\w+):", line):
            count += 1
    return count


class Voc126CallerWorkflowDispatchTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pipeline = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.pipeline_verify = (
            CALLER_WORKFLOWS / "pipeline-verify.yml"
        ).read_text(encoding="utf-8")

    def test_live_workflow_dispatch_blocks_stay_within_github_limit(self):
        for path in sorted(CALLER_WORKFLOWS.glob("*.yml")):
            text = path.read_text(encoding="utf-8")
            if "workflow_dispatch:" not in text:
                continue
            count = count_workflow_dispatch_inputs(text)
            self.assertLessEqual(
                count,
                MAX_DISPATCH_INPUTS,
                f"{path.name} declares {count} workflow_dispatch inputs",
            )

    def test_pipeline_keeps_mutating_actions_and_existing_pr_number(self):
        self.assertIn(f"options: [{PIPELINE_MUTATING_OPTIONS}]", self.pipeline)
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split(
            "\n# Read-only verifier", 1
        )[0]
        self.assertIn("existing_pr_number:", dispatch)
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("expected_base_sha:", dispatch)
        implement = self.pipeline.split("  implement:", 1)[1].split("  plan:", 1)[0]
        self.assertIn("existing_pr_number: ${{ inputs.existing_pr_number }}", implement)
        for action in PIPELINE_VERIFY_OPTIONS.split(", "):
            self.assertNotIn(f"inputs.action == '{action}'", self.pipeline)

    def test_pipeline_verify_exposes_relocated_verifiers(self):
        self.assertIn(f"options: [{PIPELINE_VERIFY_OPTIONS}]", self.pipeline_verify)
        for job in (
            "verify-auto-advance-live-evidence",
            "verify-ready-for-review-reuse",
            "verify-remediate-operator-ownership",
            "verify-promotion-check-recovery",
            "verify-post-promotion-workflow",
        ):
            self.assertIn(f"  {job}:", self.pipeline_verify)

    def test_pipeline_verify_stays_read_only(self):
        self.assertNotIn("secrets: inherit", self.pipeline_verify)
        self.assertNotIn("actions: write", self.pipeline_verify)
        self.assertNotIn("create-github-app-token", self.pipeline_verify)
        self.assertNotIn("recover-integration-push", self.pipeline_verify)
        self.assertNotIn("recover-promotion-pr-checks", self.pipeline_verify)
        self.assertIn("actions: read", self.pipeline_verify)

    def test_fixture_pin_must_not_equal_invalid_voc125_template_sha(self):
        pin_path = (
            REPO_ROOT
            / "tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt"
        )
        pin = pin_path.read_text(encoding="utf-8").strip()
        self.assertNotEqual(pin, INVALID_VOC125_TEMPLATE_SHA)
        self.assertEqual(pin, "20dcf340fa73a36ebc6074442fde79530dfa5871")


if __name__ == "__main__":
    unittest.main()
