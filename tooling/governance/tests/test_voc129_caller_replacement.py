"""VOC-129 caller replacement regressions for infra #164 pin and dispatch."""

from __future__ import annotations

import re
import subprocess
import unittest
from pathlib import Path

from voc080_fixtures import read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
FIXTURE_ROOT = REPO_ROOT / "tooling/governance/fixtures/karsift-ai-infra"
AUTHORITATIVE_PIN = "123735c80fec813a5b46a004f3e1122bd425cde2"
CURRENT_PIN = "1edd60b98e1785057f63b7686ee2822706574a97"
STALE_PIN_164 = "863fc1f35b1d35e4981a59166b0e939be1a2b681"
PREVIOUS_DEVELOP_PIN = "60afda3a44fd06b8c00b219771de7112f1aded6e"
MAX_DISPATCH_INPUTS = 25


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


class Voc129CallerReplacementTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.pipeline = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.production_reconcile = read_fixture(
            ".github/workflows/reconcile-production-change.yml"
        )
        cls.template = read_fixture(
            "templates/project-repo/.github/workflows/pipeline.yml"
        )
        cls.evidence = (
            REPO_ROOT
            / "specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the/t00-evidence.md"
        ).read_text(encoding="utf-8")

    def test_pin_equals_current_infra_merge_and_not_stale_pins(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, STALE_PIN_164)
        self.assertNotEqual(self.pin, PREVIOUS_DEVELOP_PIN)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_fixture_contains_in_scope_infra_164_files(self):
        required = [
            ".github/workflows/release.yml",
            ".github/workflows/reconcile-production-change.yml",
            "config/release-checkout-ref-runner.py",
            "config/branch_sync.py",
            "config/branch-sync-runner.py",
            "templates/project-repo/.github/workflows/pipeline.yml",
            "tests/test_release_checkout_ref_runner.py",
            "tests/test_branch_sync.py",
            "tests/test_branch_sync_runner.py",
        ]
        for relative in required:
            path = FIXTURE_ROOT / relative
            self.assertTrue(path.is_file(), f"missing fixture file: {relative}")
        self.assertIn("Synchronize integration to the exact promotion merge", self.release)
        self.assertIn("branch-sync-runner.py", self.release)
        self.assertNotIn('-f sha="$CHECKED_HEAD_SHA"', self.release)

    def test_checkout_ref_runner_tests_execute_in_fixture(self):
        result = subprocess.run(
            [
                "python3",
                "-m",
                "unittest",
                "discover",
                "-s",
                str(FIXTURE_ROOT / "tests"),
                "-p",
                "test_release_checkout_ref_runner.py",
            ],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(
            result.returncode,
            0,
            msg=result.stdout + result.stderr,
        )

    def test_live_pipeline_exposes_reconcile_production_change_within_input_limit(self):
        self.assertIn("reconcile-production-change", self.pipeline)
        self.assertIn(
            "reconcile-production-change.yml@main",
            self.pipeline,
        )
        self.assertIn("authority_issue_number:", self.pipeline)
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split(
            "\n# Read-only verifier", 1
        )[0]
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("target_sha:", dispatch)
        self.assertNotIn("sha:", dispatch.lower())
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

    def test_release_and_auto_advance_wait_on_production_change_reconcile(self):
        release = self.pipeline.split("  release:", 1)[1].split("  auto-advance:", 1)[0]
        auto_advance = self.pipeline.split("  auto-advance:", 1)[1].split(
            "  live-evidence-reconcile:", 1
        )[0]
        self.assertIn("needs: [merge-gate, reconcile-production-change]", release)
        self.assertIn(
            "needs.reconcile-production-change.result == 'success'",
            release,
        )
        self.assertIn("needs: [reconcile-production-change]", auto_advance)
        self.assertIn(
            "needs.reconcile-production-change.result == 'success'",
            auto_advance,
        )
        self.assertNotIn("schedule:", self.production_reconcile)
        implement = self.pipeline.split("  implement:", 1)[1].split("  plan:", 1)[0]
        self.assertIn("existing_pr_number: ${{ inputs.existing_pr_number }}", implement)
        self.assertNotIn(
            "existing_pr_number",
            self.production_reconcile,
        )

    def test_fixture_release_policy_no_longer_restores_checked_head_sha(self):
        self.assertNotIn(
            "test_promotion_preserves_long_lived_integration_branch",
            (FIXTURE_ROOT / "tests/test_release_policy.py").read_text(encoding="utf-8"),
        )
        self.assertIn("test_promotion_converges_integration_to_exact_merge_before_close", (
            FIXTURE_ROOT / "tests/test_release_policy.py"
        ).read_text(encoding="utf-8"))

    def test_replacement_carrier_is_voc129_not_voc127_attempt_three(self):
        self.assertIn("#1041", self.evidence)
        self.assertIn("attempt `3` is forbidden", self.evidence)
        self.assertIn("not #1041", self.evidence)
        self.assertIn("VOC-129-T00 attempt `1`", self.evidence)

    def test_roles_yml_unchanged_and_no_openai_route_added(self):
        roles = read_fixture("config/roles.yml")
        self.assertIn("implementer:", roles)
        self.assertNotIn("openai", roles.lower())


if __name__ == "__main__":
    unittest.main()
