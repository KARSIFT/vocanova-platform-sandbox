"""VOC-133 caller regressions for infra #166 pin, restore, and nested-checkout."""

from __future__ import annotations

import hashlib
import re
import subprocess
import tempfile
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
AUTHORITATIVE_PIN = "f3d79177bf8a9abe0dae550f39502165d494c576"
PIN_164 = "863fc1f35b1d35e4981a59166b0e939be1a2b681"
PIN_165 = "8ce2b77a09a729e458a9f4cbea1ca26eb114d398"
MAX_DISPATCH_INPUTS = 25

MIRRORED_SHA256 = {
    ".github/workflows/implement.yml": (
        "5e44f6a82cdb127f9716faea56cd226965ab3cf86566bde009af375c205ff03c"
    ),
    ".github/workflows/release.yml": (
        "fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08"
    ),
    "config/implementer_nested_checkout.py": (
        "e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9"
    ),
    "tests/test_release_policy.py": (
        "082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07"
    ),
    "tests/test_voc121_implement_policy.py": (
        "78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975"
    ),
    "tests/test_voc123_source_bundle.py": (
        "d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936"
    ),
    "CHANGELOG.md": (
        "7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea"
    ),
}


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


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class Voc133Infra166PinTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.pipeline = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.readme = read_fixture("README.md")
        cls.evidence = (
            REPO_ROOT
            / "specs/changes/VOC-133-consume-exact-infra-166-and-complete-release/t00-evidence.md"
        ).read_text(encoding="utf-8")

    def test_pin_equals_authoritative_infra_merge_and_not_stale_pins(self):
        self.assertEqual(self.pin, AUTHORITATIVE_PIN)
        self.assertNotEqual(self.pin, PIN_164)
        self.assertNotEqual(self.pin, PIN_165)

    def test_mirrored_fixture_files_match_recorded_sha256_hashes(self):
        for relative, expected in MIRRORED_SHA256.items():
            path = FIXTURE_INFRA_ROOT / relative
            self.assertTrue(path.is_file(), f"missing fixture file: {relative}")
            self.assertEqual(
                sha256_file(path),
                expected,
                f"SHA-256 mismatch for {relative}",
            )

    def test_fixture_contains_implementer_nested_checkout_helper(self):
        helper = FIXTURE_INFRA_ROOT / "config/implementer_nested_checkout.py"
        self.assertTrue(helper.is_file())
        self.assertEqual(sha256_file(helper), MIRRORED_SHA256[helper.relative_to(FIXTURE_INFRA_ROOT).as_posix()])

    def test_release_restores_shared_policy_in_identify_before_validate_task(self):
        identify = self.release.split("  identify:", 1)[1].split("  converge:", 1)[0]
        caller_checkout = identify.index("Checkout caller release state")
        restore = identify.index("Restore shared lifecycle policy after caller checkout")
        validate_task = identify.index("task-completion-runner.py validate-task")
        self.assertLess(caller_checkout, restore)
        self.assertLess(restore, validate_task)
        restored = identify[restore:validate_task]
        self.assertIn("repository: ${{ job.workflow_repository }}", restored)
        self.assertIn("ref: ${{ job.workflow_sha }}", restored)
        self.assertIn("path: karsift-ai-infra", restored)
        self.assertIn("persist-credentials: false", restored)

    def test_release_restores_shared_policy_in_converge_before_validate_roster(self):
        converge = self.release.split("  converge:", 1)[1]
        caller_checkout = converge.index("Checkout caller release state")
        restore = converge.index("Restore shared lifecycle policy after caller checkout")
        validate_roster = converge.index("task-completion-runner.py validate-roster")
        self.assertLess(caller_checkout, restore)
        self.assertLess(restore, validate_roster)
        restored = converge[restore:validate_roster]
        self.assertIn("repository: ${{ job.workflow_repository }}", restored)
        self.assertIn("ref: ${{ job.workflow_sha }}", restored)
        self.assertIn("path: karsift-ai-infra", restored)
        self.assertIn("persist-credentials: false", restored)

    def test_release_has_exactly_two_restore_steps(self):
        self.assertEqual(
            self.release.count("Restore shared lifecycle policy after caller checkout"),
            2,
        )

    def test_implement_preserves_helpers_before_model_and_classifies_nested_checkout(self):
        preserve = self.implement.index("- name: Preserve post-implementer lifecycle helpers")
        implement = self.implement.index("- name: Run implementer (cursor-agent)")
        commit = self.implement.index("- name: Commit implementer's work")
        self.assertLess(preserve, implement)
        self.assertLess(implement, commit)
        preserve_block = self.implement[preserve:implement]
        for helper in (
            "run-app-checks.sh",
            "prepare_cursor_model.py",
            "implementer_source_carrier.py",
            "cross_repo_reference.py",
            "implementer_nested_checkout.py",
        ):
            self.assertIn(f"config/{helper}", preserve_block)
        commit_block = self.implement[commit : self.implement.index("- name: Pre-push validation")]
        self.assertIn(
            'python3 "$HELPER_DIR/implementer_nested_checkout.py" karsift-ai-infra',
            commit_block,
        )
        self.assertIn("no nested source changes to publish", commit_block)
        self.assertNotIn(
            "python3 karsift-ai-infra/config/implementer_nested_checkout.py",
            commit_block,
        )

    def test_nested_checkout_classifier_rejects_non_directory_path(self):
        sys_path = str(FIXTURE_INFRA_ROOT / "config")
        import sys

        if sys_path not in sys.path:
            sys.path.insert(0, sys_path)
        from implementer_nested_checkout import (  # noqa: E402
            NestedCheckoutError,
            classify_nested_checkout,
        )

        with tempfile.TemporaryDirectory() as scratch:
            nested = Path(scratch) / "karsift-ai-infra"
            nested.write_text("not a directory", encoding="utf-8")
            with self.assertRaises(NestedCheckoutError) as ctx:
                classify_nested_checkout(nested)
            self.assertEqual(str(ctx.exception), "nested_checkout_not_directory")

    def test_fixture_release_policy_tests_execute_in_fixture(self):
        result = subprocess.run(
            [
                "python3",
                "-m",
                "unittest",
                "discover",
                "-s",
                str(FIXTURE_INFRA_ROOT / "tests"),
                "-p",
                "test_release_policy.py",
            ],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(result.returncode, 0, msg=result.stdout + result.stderr)
        self.assertIn(
            "test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers",
            (FIXTURE_INFRA_ROOT / "tests/test_release_policy.py").read_text(encoding="utf-8"),
        )

    def test_fixture_implement_policy_tests_execute_in_fixture(self):
        result = subprocess.run(
            [
                "python3",
                "-m",
                "unittest",
                "discover",
                "-s",
                str(FIXTURE_INFRA_ROOT / "tests"),
                "-p",
                "test_voc121_implement_policy.py",
            ],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(result.returncode, 0, msg=result.stdout + result.stderr)

    def test_live_pipeline_exposes_reconcile_production_change_within_input_limit(self):
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

    def test_roles_yml_unchanged_and_no_openai_route_added(self):
        roles = read_fixture("config/roles.yml")
        self.assertIn("implementer:", roles)
        self.assertNotIn("openai", roles.lower())

    def test_fixture_readme_names_current_pin_restore_and_helper_lifetime(self):
        self.assertIn(AUTHORITATIVE_PIN, self.readme)
        self.assertIn("Restore shared lifecycle policy after caller checkout", self.readme)
        self.assertIn("Preserve post-implementer lifecycle helpers", self.readme)

    def test_replacement_carrier_is_voc133_not_reused_carriers(self):
        self.assertIn("#1051", self.evidence)
        self.assertIn("#1056", self.evidence)
        self.assertIn("#1059", self.evidence)
        self.assertIn("not #1051", self.evidence)
        self.assertIn("not redispatched", self.evidence)
        self.assertIn("VOC-133-T00 attempt `2`", self.evidence)


if __name__ == "__main__":
    unittest.main()
