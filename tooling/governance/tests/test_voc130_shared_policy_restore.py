"""VOC-130 caller regressions for infra #165 pin and shared-policy restore."""

from __future__ import annotations

import re
import subprocess
import unittest
from pathlib import Path

from voc080_fixtures import read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
FIXTURE_ROOT = REPO_ROOT / "tooling/governance/fixtures/karsift-ai-infra"
AUTHORITATIVE_PIN = "8ce2b77a09a729e458a9f4cbea1ca26eb114d398"
PREVIOUS_PIN_164 = "863fc1f35b1d35e4981a59166b0e939be1a2b681"
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


class Voc130SharedPolicyRestoreTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.pipeline = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.release_policy_tests = read_fixture("tests/test_release_policy.py")
        cls.readme = read_fixture("README.md")
        cls.roles = read_fixture("config/roles.yml")
        cls.evidence = (
            REPO_ROOT
            / "specs/changes/VOC-130-release-blocker-caller-checkout-deletes-shared/t00-evidence.md"
        ).read_text(encoding="utf-8")
        cls.voc129_evidence = (
            REPO_ROOT
            / "specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the/t00-evidence.md"
        ).read_text(encoding="utf-8")

    def test_pin_equals_authoritative_infra_merge_and_not_previous_pin(self):
        self.assertEqual(self.pin, AUTHORITATIVE_PIN)
        self.assertNotEqual(self.pin, PREVIOUS_PIN_164)

    def test_foundation_pin_literals_match_authoritative_merge(self):
        foundation_dir = REPO_ROOT / "scripts/foundation"
        for name in (
            "voc097-fixture-matrix.test.mjs",
            "voc104-ready-for-review-reuse.test.mjs",
            "voc108-authoritative-lifecycle.test.mjs",
        ):
            text = (foundation_dir / name).read_text(encoding="utf-8")
            self.assertIn(AUTHORITATIVE_PIN, text)
            self.assertNotIn(PREVIOUS_PIN_164, text, msg=name)

    def test_fixture_mirrors_in_scope_infra_165_restore_files(self):
        required = [
            ".github/workflows/release.yml",
            "tests/test_release_policy.py",
        ]
        for relative in required:
            path = FIXTURE_ROOT / relative
            self.assertTrue(path.is_file(), f"missing fixture file: {relative}")
        self.assertEqual(
            self.release.count(
                "Restore shared lifecycle policy after caller checkout"
            ),
            2,
        )
        self.assertIn(
            "test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers",
            self.release_policy_tests,
        )

    def test_identify_restores_shared_policy_before_validate_task(self):
        job = self.release.split("  identify:", 1)[1].split("  converge:", 1)[0]
        policy_checkout = job.index("Checkout shared lifecycle policy")
        resolver = job.index("release-checkout-ref-runner.py")
        caller_checkout = job.index("Checkout caller release state")
        restore = job.index("Restore shared lifecycle policy after caller checkout")
        helper = job.index("task-completion-runner.py validate-task")
        self.assertLess(policy_checkout, resolver)
        self.assertLess(resolver, caller_checkout)
        self.assertLess(caller_checkout, restore)
        self.assertLess(restore, helper)

    def test_converge_restores_shared_policy_before_validate_roster(self):
        job = self.release.split("  converge:", 1)[1]
        caller_checkout = job.index("Checkout caller release state")
        restore = job.index("Restore shared lifecycle policy after caller checkout")
        helper = job.index("task-completion-runner.py validate-roster")
        self.assertLess(caller_checkout, restore)
        self.assertLess(restore, helper)

    def test_restore_uses_immutable_reusable_workflow_revision(self):
        for job_name, end_marker, helper in (
            ("  identify:", "  converge:", "task-completion-runner.py validate-task"),
            ("  converge:", None, "task-completion-runner.py validate-roster"),
        ):
            job = self.release.split(job_name, 1)[1]
            if end_marker:
                job = job.split(end_marker, 1)[0]
            restore = job.index(
                "Restore shared lifecycle policy after caller checkout"
            )
            helper_use = job.index(helper)
            restored = job[restore:helper_use]
            restore_step, _, remainder = restored.partition("\n      - name:")
            self.assertIn("repository: ${{ job.workflow_repository }}", restore_step)
            self.assertIn("ref: ${{ job.workflow_sha }}", restore_step)
            self.assertIn("path: karsift-ai-infra", restore_step)
            self.assertNotIn("inputs.integration_branch", restore_step)
            self.assertTrue(remainder, "restore step should end before the next step")

    def test_164_contracts_remain_after_165_pin(self):
        self.assertIn("Synchronize integration to the exact promotion merge", self.release)
        self.assertIn("branch-sync-runner.py", self.release)
        self.assertNotIn('-f sha="$CHECKED_HEAD_SHA"', self.release)
        self.assertIn("reconcile-production-change", self.pipeline)
        self.assertIn("reconcile-production-change.yml@main", self.pipeline)
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split(
            "\n# Read-only verifier", 1
        )[0]
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("target_sha:", dispatch)
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

    def test_existing_controls_roles_and_docs_remain(self):
        self.assertIn("implementer:", self.roles)
        self.assertNotIn("openai", self.roles.lower())
        self.assertIn(AUTHORITATIVE_PIN, self.readme)
        self.assertIn(
            "Restore shared lifecycle policy after caller checkout",
            self.readme,
        )
        self.assertIn("VOC-130-T00", self.readme)

    def test_replacement_carrier_is_voc130_not_voc129_rewrite(self):
        self.assertIn("#1046", self.evidence)
        self.assertIn("33066533397", self.evidence)
        self.assertIn("Why VOC-129 is not retried", self.evidence)
        self.assertIn("snapshot-the-gap", self.evidence.lower())
        self.assertIn("do not re-implement", self.evidence.lower())
        self.assertIn("#1041", self.voc129_evidence)


if __name__ == "__main__":
    unittest.main()
