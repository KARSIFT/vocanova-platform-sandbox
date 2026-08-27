"""VOC-131 caller replacement regressions for infra #165 pin and restore."""

from __future__ import annotations

import hashlib
import json
import re
import subprocess
import unittest
from pathlib import Path

from voc080_fixtures import read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
FIXTURE_ROOT = REPO_ROOT / "tooling/governance/fixtures/karsift-ai-infra"
AUTHORITATIVE_PIN = "8ce2b77a09a729e458a9f4cbea1ca26eb114d398"
STALE_PIN_164 = "863fc1f35b1d35e4981a59166b0e939be1a2b681"
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"
VOC112_FIXTURES = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
)
# SHA-256 digests of the two mirrored files at infra merge #165
# (8ce2b77a09a729e458a9f4cbea1ca26eb114d398). In-repo proof avoids depending
# on an external karsift-ai-infra checkout at test time.
INFRA_165_FILE_SHA256 = {
    ".github/workflows/release.yml": (
        "fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08"
    ),
    "tests/test_release_policy.py": (
        "082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07"
    ),
}
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


def develop_base_sha() -> str:
    result = subprocess.run(
        ["git", "rev-parse", "origin/develop"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode == 0:
        return result.stdout.strip()
    result = subprocess.run(
        ["git", "rev-parse", "develop"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    return result.stdout.strip()


def git_show_at_revision(revision: str, path: str) -> bytes:
    result = subprocess.run(
        ["git", "show", f"{revision}:{path}"],
        cwd=REPO_ROOT,
        capture_output=True,
        check=True,
    )
    return result.stdout


class Voc131CallerReplacementTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.pipeline = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.release_policy_tests = (
            FIXTURE_ROOT / "tests/test_release_policy.py"
        ).read_text(encoding="utf-8")
        cls.evidence = (
            REPO_ROOT
            / "specs/changes/VOC-131-replacement-release-blocker-consume-infra-165/t00-evidence.md"
        ).read_text(encoding="utf-8")
        cls.develop_base = develop_base_sha()

    def test_pin_equals_authoritative_infra_merge_and_not_stale_pin(self):
        self.assertEqual(self.pin, AUTHORITATIVE_PIN)
        self.assertNotEqual(self.pin, STALE_PIN_164)

    def test_fixture_release_and_policy_tests_are_byte_identical_to_infra_merge(self):
        for relative, expected_sha256 in INFRA_165_FILE_SHA256.items():
            fixture_bytes = (FIXTURE_ROOT / relative).read_bytes()
            digest = hashlib.sha256(fixture_bytes).hexdigest()
            self.assertEqual(
                digest,
                expected_sha256,
                (
                    f"fixture {relative} sha256 {digest} differs from infra "
                    f"merge {AUTHORITATIVE_PIN} expected {expected_sha256}"
                ),
            )

    def test_fixture_restore_steps_precede_lifecycle_helpers_in_both_jobs(self):
        self.assertEqual(
            self.release.count(
                "Restore shared lifecycle policy after caller checkout"
            ),
            2,
        )
        for job_name, end_marker, helper in (
            ("  identify:", "  converge:", "task-completion-runner.py validate-task"),
            ("  converge:", None, "task-completion-runner.py validate-roster"),
        ):
            job = self.release.split(job_name, 1)[1]
            if end_marker:
                job = job.split(end_marker, 1)[0]
            policy_checkout = job.index("Checkout shared lifecycle policy")
            resolver = job.index("release-checkout-ref-runner.py")
            caller_checkout = job.index("Checkout caller release state")
            restore = job.index(
                "Restore shared lifecycle policy after caller checkout"
            )
            helper_use = job.index(helper)
            self.assertLess(policy_checkout, resolver)
            self.assertLess(resolver, caller_checkout)
            self.assertLess(caller_checkout, restore)
            self.assertLess(restore, helper_use)

    def test_restore_uses_immutable_workflow_revision_without_persisted_credentials(self):
        for job_name, end_marker in (
            ("  identify:", "  converge:"),
            ("  converge:", None),
        ):
            job = self.release.split(job_name, 1)[1]
            if end_marker:
                job = job.split(end_marker, 1)[0]
            restore = job.index(
                "Restore shared lifecycle policy after caller checkout"
            )
            next_step = job.index("\n      - name:", restore + 1)
            restored = job[restore:next_step]
            self.assertIn("repository: ${{ job.workflow_repository }}", restored)
            self.assertIn("ref: ${{ job.workflow_sha }}", restored)
            self.assertIn("path: karsift-ai-infra", restored)
            self.assertIn("persist-credentials: false", restored)
            self.assertNotIn("ref: ${{ inputs.integration_branch }}", restored)

    def test_fixture_policy_tests_cover_caller_checkout_rehydration(self):
        self.assertIn(
            "test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers",
            self.release_policy_tests,
        )

    def test_164_missing_develop_and_exact_merge_sync_contracts_remain(self):
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

    def test_roles_yml_unchanged_and_no_openai_route_added(self):
        roles = read_fixture("config/roles.yml")
        self.assertIn("implementer:", roles)
        self.assertNotIn("openai", roles.lower())

    def test_voc112_fixtures_remain_byte_identical_to_develop_base(self):
        for relative in VOC112_FIXTURES:
            current_bytes = (REPO_ROOT / relative).read_bytes()
            base_bytes = git_show_at_revision(self.develop_base, relative)
            self.assertEqual(
                current_bytes,
                base_bytes,
                f"{relative} differs from develop base {self.develop_base}",
            )
            payload = json.loads(current_bytes.decode("utf-8"))
            if relative.endswith("voc112-navigation-benchmark-traces.json"):
                self.assertEqual(
                    payload["subject_revision"],
                    VOC112_SUBJECT_REVISION,
                )
            else:
                for row in payload["discoveries"]:
                    self.assertEqual(
                        row["subject_revision"],
                        VOC112_SUBJECT_REVISION,
                    )

    def test_voc112_fixtures_absent_from_implementation_diff(self):
        result = subprocess.run(
            [
                "git",
                "diff",
                "--name-only",
                self.develop_base,
                "--",
                *VOC112_FIXTURES,
            ],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=True,
        )
        self.assertEqual(result.stdout.strip(), "")

    def test_replacement_carrier_is_voc131_not_voc130_retry(self):
        self.assertIn("#1051", self.evidence)
        self.assertIn("#1049", self.evidence)
        self.assertIn("attempt `1`", self.evidence)
        self.assertIn("not #1051", self.evidence)
        self.assertIn("VOC-131-T00", self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)
        self.assertIn(STALE_PIN_164, self.evidence)


if __name__ == "__main__":
    unittest.main()
