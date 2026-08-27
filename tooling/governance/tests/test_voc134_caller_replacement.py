"""VOC-134 caller replacement regressions for infra #166 pin and contracts."""

from __future__ import annotations

import hashlib
import re
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, REPOSITORY_ROOT, read_fixture


AUTHORITATIVE_PIN = "f3d79177bf8a9abe0dae550f39502165d494c576"
STALE_PIN_164 = "863fc1f35b1d35e4981a59166b0e939be1a2b681"
STALE_PIN_165 = "8ce2b77a09a729e458a9f4cbea1ca26eb114d398"
IMMUTABLE_CARRIER_BASE = "b9e74fc2db4691c48c637639b265d527de9f4505"
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"
MAX_DISPATCH_INPUTS = 25

FIXTURE_ROOT = FIXTURE_INFRA_ROOT
CALLER_WORKFLOWS = REPOSITORY_ROOT / ".github/workflows"
EVIDENCE = (
    REPOSITORY_ROOT
    / "specs/changes/VOC-134-replace-exhausted-voc-133-with-feasible-exact-sha/t00-evidence.md"
)

VOC112_NO_CHANGE_PATHS = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
    "scripts/foundation/voc112-navigation-benchmark.test.mjs",
    "AGENTS.md",
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
)

FIXTURE_SHA256 = {
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
    digest = hashlib.sha256()
    digest.update(path.read_bytes())
    return digest.hexdigest()


def git_show_bytes(sha: str, relative: str) -> bytes:
    completed = subprocess.run(
        ["git", "show", f"{sha}:{relative}"],
        cwd=REPOSITORY_ROOT,
        capture_output=True,
        check=False,
    )
    if completed.returncode != 0:
        raise AssertionError(
            f"cannot resolve {sha}:{relative}: {completed.stderr.decode()}"
        )
    return completed.stdout


def assert_commit_exists(sha: str) -> None:
    completed = subprocess.run(
        ["git", "cat-file", "-e", f"{sha}^{{commit}}"],
        cwd=REPOSITORY_ROOT,
        capture_output=True,
        check=False,
    )
    if completed.returncode != 0:
        raise AssertionError(f"commit object does not resolve: {sha}")


def paths_in_diff(base_sha: str, *paths: str) -> list[str]:
    completed = subprocess.run(
        ["git", "diff", "--name-only", base_sha],
        cwd=REPOSITORY_ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    changed = set(completed.stdout.splitlines())
    return [path for path in paths if path in changed]


class Voc134CallerReplacementTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.pipeline = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.fixture_readme = read_fixture("README.md")
        cls.evidence = EVIDENCE.read_text(encoding="utf-8")

    def test_pin_equals_authoritative_infra_merge_and_not_stale_pins(self):
        self.assertEqual(self.pin, AUTHORITATIVE_PIN)
        self.assertNotEqual(self.pin, STALE_PIN_164)
        self.assertNotEqual(self.pin, STALE_PIN_165)

    def test_foundation_pin_literals_match_authoritative_merge(self):
        for script in (
            "voc097-fixture-matrix.test.mjs",
            "voc104-ready-for-review-reuse.test.mjs",
            "voc108-authoritative-lifecycle.test.mjs",
        ):
            text = (REPOSITORY_ROOT / "scripts/foundation" / script).read_text(
                encoding="utf-8"
            )
            self.assertIn(AUTHORITATIVE_PIN, text, msg=script)
            self.assertNotIn(STALE_PIN_164, text, msg=script)

    def test_fixture_mirrors_necessary_166_files_by_recorded_sha256(self):
        for relative, expected in FIXTURE_SHA256.items():
            path = FIXTURE_ROOT / relative
            self.assertTrue(path.is_file(), f"missing fixture file: {relative}")
            self.assertEqual(sha256_file(path), expected, msg=relative)
        self.assertEqual(
            self.release.count("Restore shared lifecycle policy after caller checkout"),
            2,
        )
        release_policy = (FIXTURE_ROOT / "tests/test_release_policy.py").read_text(
            encoding="utf-8"
        )
        self.assertIn(
            "test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers",
            release_policy,
        )
        implement_policy = (
            FIXTURE_ROOT / "tests/test_voc121_implement_policy.py"
        ).read_text(encoding="utf-8")
        self.assertIn("test_nested_checkout_classifier_rejects_symlink", implement_policy)
        self.assertIn(
            "test_nested_checkout_classifier_rejects_parent_git_inheritance",
            implement_policy,
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

    def test_restore_uses_immutable_workflow_revision_without_persisting_credentials(self):
        for job_name, end_marker, helper in (
            ("  identify:", "  converge:", "task-completion-runner.py validate-task"),
            ("  converge:", None, "task-completion-runner.py validate-roster"),
        ):
            job = self.release.split(job_name, 1)[1]
            if end_marker:
                job = job.split(end_marker, 1)[0]
            restore = job.index("Restore shared lifecycle policy after caller checkout")
            helper_use = job.index(helper)
            restored = job[restore:helper_use].split("\n", 8)[:8]
            restored_block = "\n".join(restored)
            self.assertIn("repository: ${{ job.workflow_repository }}", restored_block)
            self.assertIn("ref: ${{ job.workflow_sha }}", restored_block)
            self.assertIn("path: karsift-ai-infra", restored_block)
            self.assertIn("persist-credentials: false", restored_block)
            self.assertNotIn("ref: ${{ inputs.integration_branch }}", restored_block)

    def test_164_contracts_remain_in_pinned_fixture(self):
        release_policy = (FIXTURE_ROOT / "tests/test_release_policy.py").read_text(
            encoding="utf-8"
        )
        self.assertIn("Synchronize integration to the exact promotion merge", self.release)
        self.assertIn("branch-sync-runner.py", self.release)
        self.assertNotIn('-f sha="$CHECKED_HEAD_SHA"', self.release)
        self.assertIn(
            "test_promotion_converges_integration_to_exact_merge_before_close",
            release_policy,
        )
        self.assertIn("reconcile-production-change", self.pipeline)
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split(
            "\n# Read-only verifier", 1
        )[0]
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("target_sha:", dispatch)

    def test_166_helper_lifetime_and_nested_checkout_contracts(self):
        preserve = self.implement.index("- name: Preserve post-implementer lifecycle helpers")
        implement = self.implement.index("- name: Run implementer (cursor-agent)")
        commit = self.implement.index("- name: Commit implementer's work")
        self.assertLess(preserve, implement)
        self.assertLess(implement, commit)
        for helper in (
            "run-app-checks.sh",
            "prepare_cursor_model.py",
            "implementer_source_carrier.py",
            "cross_repo_reference.py",
            "implementer_nested_checkout.py",
        ):
            self.assertIn(helper, self.implement[preserve:commit])
        commit_block = self.implement[
            commit : self.implement.index("- name: Pre-push validation")
        ]
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
        sys.path.insert(0, str(FIXTURE_ROOT / "config"))
        try:
            from implementer_nested_checkout import (  # noqa: E402
                NestedCheckoutError,
                classify_nested_checkout,
            )
        finally:
            sys.path.pop(0)

        with tempfile.TemporaryDirectory() as scratch:
            nested = Path(scratch) / "karsift-ai-infra"
            nested.write_text("not a directory\n", encoding="utf-8")
            with self.assertRaisesRegex(
                NestedCheckoutError,
                "nested_checkout_not_directory",
            ):
                classify_nested_checkout(nested)

    def test_live_pipeline_dispatch_limit_and_roles_unchanged(self):
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
        roles = read_fixture("config/roles.yml")
        self.assertIn("implementer:", roles)
        self.assertNotIn("openai", roles.lower())

    def test_fixture_readme_names_current_pin_restore_and_helper_contract(self):
        self.assertIn(AUTHORITATIVE_PIN, self.fixture_readme)
        self.assertIn("VOC-134-T00", self.fixture_readme)
        self.assertIn("Restore shared lifecycle policy after caller checkout", self.release)
        self.assertIn("Preserve post-implementer lifecycle helpers", self.implement)

    def test_complete_voc112_no_change_boundary_against_immutable_carrier_base(self):
        self.assertRegex(IMMUTABLE_CARRIER_BASE, r"^[0-9a-f]{40}$")
        assert_commit_exists(IMMUTABLE_CARRIER_BASE)
        for relative in VOC112_NO_CHANGE_PATHS:
            carrier_bytes = git_show_bytes(IMMUTABLE_CARRIER_BASE, relative)
            working_bytes = (REPOSITORY_ROOT / relative).read_bytes()
            self.assertEqual(
                working_bytes,
                carrier_bytes,
                msg=f"{relative} differs from immutable carrier base",
            )
        changed = paths_in_diff(IMMUTABLE_CARRIER_BASE, *VOC112_NO_CHANGE_PATHS)
        self.assertEqual(changed, [], msg=f"protected paths in diff: {changed}")
        for json_path in VOC112_NO_CHANGE_PATHS[:2]:
            text = (REPOSITORY_ROOT / json_path).read_text(encoding="utf-8")
            self.assertIn(VOC112_SUBJECT_REVISION, text)
        provenance = (
            REPOSITORY_ROOT / "scripts/foundation/voc112-navigation-benchmark.test.mjs"
        ).read_text(encoding="utf-8")
        self.assertIn("local", provenance)
        self.assertIn(
            "a full local checkout must already contain the captured commit",
            provenance,
        )

    def test_package_json_is_carrier_base_identical_and_provenance_not_bypassed(self):
        assert_commit_exists(IMMUTABLE_CARRIER_BASE)
        carrier_bytes = git_show_bytes(IMMUTABLE_CARRIER_BASE, "package.json")
        working_bytes = (REPOSITORY_ROOT / "package.json").read_bytes()
        self.assertEqual(working_bytes, carrier_bytes)
        changed = paths_in_diff(IMMUTABLE_CARRIER_BASE, "package.json")
        self.assertEqual(changed, [])
        package_json = (REPOSITORY_ROOT / "package.json").read_text(encoding="utf-8")
        self.assertNotIn("VOC112_CAPTURE_PROVENANCE_MODE", package_json)
        self.assertNotIn("ensure-voc112-capture-commits", package_json)
        self.assertFalse(
            (REPOSITORY_ROOT / "scripts/foundation/ensure-voc112-capture-commits.mjs").exists()
        )

    def test_feasible_exact_revision_evidence_contract(self):
        self.assertIn(IMMUTABLE_CARRIER_BASE, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)
        self.assertIn("App-authored independent-review comment/check", self.evidence)
        self.assertIn("must not be required to contain its own SHA", self.evidence)
        for digest in FIXTURE_SHA256.values():
            self.assertIn(digest, self.evidence)

    def test_replacement_carrier_is_voc134_not_exhausted_carriers(self):
        self.assertIn("#1065", self.evidence)
        self.assertIn("#1051", self.evidence)
        self.assertIn("#1056", self.evidence)
        self.assertIn("not #1051", self.evidence)
        self.assertIn("not #1056", self.evidence)
        self.assertIn("not #1065", self.evidence)
        self.assertIn("VOC-134-T00 attempt `1`", self.evidence)
        self.assertIn("not redispatched #1059", self.evidence)
        self.assertIn("not redispatched #1063", self.evidence)


if __name__ == "__main__":
    unittest.main()
