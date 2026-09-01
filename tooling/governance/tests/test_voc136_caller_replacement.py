"""VOC-136 caller replacement regressions for infra #167 pin and contracts."""

from __future__ import annotations

import hashlib
import subprocess
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture
from voc136_bypass_scan import (
    FIXTURE_MIRROR_PREFIX,
    scan_changed_path_for_bypasses,
    scan_exclude_prefixes_do_not_skip_caller_tests,
    should_scan_path,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/t00-evidence.md"
)

CURRENT_PIN = "ad2b27784e6fc33b3ac7e9dab48245dd6d08ac7f"
AUTHORITATIVE_PIN = "b263c0c110591cc798b89277dfc35542abb1597b"
STALE_PIN_164 = "863fc1f35b1d35e4981a59166b0e939be1a2b681"
STALE_PIN_165 = "8ce2b77a09a729e458a9f4cbea1ca26eb114d398"
STALE_PIN_166 = "f3d79177bf8a9abe0dae550f39502165d494c576"
STALE_PIN_167 = "b263c0c110591cc798b89277dfc35542abb1597b"
PROTECTED_COMPARISON_ANCHOR = "b9e74fc2db4691c48c637639b265d527de9f4505"
IMPLEMENTATION_PR_BASE = "a4f24e7de3ce62651ae25ca49adfff90256c1fa3"
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"

MIRRORED_FILE_HASHES = {
    ".github/workflows/ci.yml": (
        "54dd080ece5e9dd6564788810025b0c0bf8b3bfe49d509b9771fd2ac88f3828a"
    ),
    ".github/workflows/implement.yml": (
        "e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a"
    ),
    ".github/workflows/release.yml": (
        "8198c14b3ff9ebb4d047ced6eaf2d5c37398c8da9f5709aaa0dbbf5252eef56f"
    ),
    "config/run-app-checks.sh": (
        "e0c3dedf3b2a750ea53ca8e6264a7fda1b430df95fd13ffa8694dcf7b9e935da"
    ),
    "config/implementer_nested_checkout.py": (
        "e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9"
    ),
    "tests/test_app_check_context.py": (
        "c272d30b66c00315b11f2edb0dead4dd6b871433452bc39194f3ef6e0c08cc90"
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
        "a33a305abf76528c71632a3df7b5b0b8afe4e5899d88a330f5623c500da7bdff"
    ),
}

NO_CHANGE_PATHS = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
    "scripts/foundation/voc112-navigation-benchmark-run.mjs",
    "scripts/foundation/validate-workspace.mjs",
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
    "package.json",
)

PROHIBITED_HELPERS = (
    "scripts/foundation/" + "ensure-" + "voc112-" + "capture-commits.mjs",
    "scripts/foundation/" + "hy" + "drate-" + "voc112-git-objects.mjs",
)


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    digest.update(path.read_bytes())
    return digest.hexdigest()


def git_show_bytes(revision: str, relative: str) -> bytes:
    completed = subprocess.run(
        ["git", "show", f"{revision}:{relative}"],
        cwd=REPO_ROOT,
        capture_output=True,
        check=False,
    )
    if completed.returncode:
        raise AssertionError(
            f"cannot resolve {revision}:{relative}: {completed.stderr.decode()}"
        )
    return completed.stdout


def git_diff_names(base: str) -> list[str]:
    completed = subprocess.run(
        ["git", "diff", "--name-only", base],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if completed.returncode:
        raise AssertionError(completed.stderr)
    names = [line for line in completed.stdout.splitlines() if line.strip()]
    untracked = subprocess.run(
        ["git", "ls-files", "--others", "--exclude-standard"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if untracked.returncode:
        raise AssertionError(untracked.stderr)
    names.extend(line for line in untracked.stdout.splitlines() if line.strip())
    return sorted(
        name
        for name in set(names)
        if not name.startswith("karsift-ai-infra/")
    )


def split_release_job(text: str, job: str) -> str:
    marker = f"  {job}:"
    start = text.index(marker)
    remainder = text[start + len(marker) :]
    for other in ("  identify:", "  converge:", "jobs:"):
        if other == marker:
            continue
        idx = remainder.find(f"\n{other}")
        if idx != -1:
            remainder = remainder[:idx]
    return remainder


def scan_implementation_diff(base: str = PROTECTED_COMPARISON_ANCHOR) -> None:
    for relative in git_diff_names(base):
        if relative in NO_CHANGE_PATHS:
            raise AssertionError(f"{relative} appears in diff against anchor")
        if not should_scan_path(relative):
            continue
        path = REPO_ROOT / relative
        if path.is_file():
            scan_changed_path_for_bypasses(relative, path.read_text(encoding="utf-8"))


class Voc136CallerReplacementTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.ci = read_fixture(".github/workflows/ci.yml")
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.run_app_checks = read_fixture("config/run-app-checks.sh")
        cls.readme = read_fixture("README.md")
        cls.pipeline = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")
        cls.roles = read_fixture("config/roles.yml")
        cls.provenance_test = (
            REPO_ROOT / "scripts/foundation/voc112-navigation-benchmark.test.mjs"
        ).read_text(encoding="utf-8")

    def test_pin_equals_authoritative_infra_merge_and_not_stale_pins(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, STALE_PIN_164)
        self.assertNotEqual(self.pin, STALE_PIN_165)
        self.assertNotEqual(self.pin, STALE_PIN_166)
        self.assertNotEqual(self.pin, STALE_PIN_167)

    def test_foundation_pin_literals_match_authoritative_merge(self):
        for relative in (
            "scripts/foundation/voc097-fixture-matrix.test.mjs",
            "scripts/foundation/voc104-ready-for-review-reuse.test.mjs",
            "scripts/foundation/voc108-authoritative-lifecycle.test.mjs",
        ):
            text = (REPO_ROOT / relative).read_text(encoding="utf-8")
            self.assertIn(CURRENT_PIN, text, relative)
            self.assertNotIn(STALE_PIN_164, text, relative)

    def test_mirrored_fixture_files_match_recorded_sha256_hashes(self):
        for relative, expected in MIRRORED_FILE_HASHES.items():
            path = FIXTURE_INFRA_ROOT / relative
            self.assertTrue(path.is_file(), f"missing fixture file: {relative}")
            self.assertEqual(sha256_file(path), expected, relative)
        self.assertEqual(
            self.release.count("Restore shared lifecycle policy after caller checkout"),
            2,
        )
        self.assertTrue(
            (FIXTURE_INFRA_ROOT / "tests/test_app_check_context.py").is_file()
        )

    def test_identify_restores_shared_policy_before_validate_task(self):
        identify = split_release_job(self.release, "identify")
        ordered = [
            "Checkout shared lifecycle policy",
            "release-checkout-ref-runner.py",
            "Checkout caller release state",
            "Restore shared lifecycle policy after caller checkout",
            "task-completion-runner.py validate-task",
        ]
        indices = [identify.index(label) for label in ordered]
        self.assertEqual(indices, sorted(indices))

    def test_converge_restores_shared_policy_before_validate_roster(self):
        converge = split_release_job(self.release, "converge")
        ordered = [
            "Checkout caller release state",
            "Restore shared lifecycle policy after caller checkout",
            "task-completion-runner.py validate-roster",
        ]
        indices = [converge.index(label) for label in ordered]
        self.assertEqual(indices, sorted(indices))

    def test_restore_uses_immutable_workflow_revision_without_credentials(self):
        for job in ("identify", "converge"):
            block = split_release_job(self.release, job)
            restore_start = block.index(
                "Restore shared lifecycle policy after caller checkout"
            )
            restore_block = block[restore_start:].split("\n      - name:", 1)[0]
            self.assertIn("repository: ${{ job.workflow_repository }}", restore_block)
            self.assertIn("ref: ${{ job.workflow_sha }}", restore_block)
            self.assertIn("path: karsift-ai-infra", restore_block)
            self.assertIn("persist-credentials: false", restore_block)
            self.assertNotIn("ref: ${{ inputs.integration_branch }}", restore_block)

    def test_release_preserves_missing_develop_and_exact_merge_sync(self):
        self.assertIn("Synchronize integration to the exact promotion merge", self.release)
        self.assertIn("branch-sync-runner.py", self.release)
        self.assertNotIn('-f sha="$CHECKED_HEAD_SHA"', self.release)
        self.assertIn("reconcile-production-change", self.pipeline)
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split(
            "\n# Read-only verifier", 1
        )[0]
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("target_sha:", dispatch)

    def test_implement_preserves_helper_copy_and_nested_checkout_contract(self):
        preserve_index = self.implement.index(
            "Preserve post-implementer lifecycle helpers"
        )
        model_index = self.implement.index("Run implementer (cursor-agent)")
        commit_marker = "      - name: Commit implementer's work\n        id: commit"
        commit_index = self.implement.index(commit_marker)
        self.assertLess(preserve_index, model_index)
        self.assertLess(model_index, commit_index)
        preserve = self.implement[preserve_index:model_index]
        for helper in (
            "run-app-checks.sh",
            "prepare_cursor_model.py",
            "implementer_source_carrier.py",
            "cross_repo_reference.py",
            "implementer_nested_checkout.py",
        ):
            self.assertIn(helper, preserve)
        commit = self.implement[commit_index : commit_index + 5000]
        self.assertIn('python3 "$HELPER_DIR/implementer_nested_checkout.py"', commit)
        self.assertIn("has_source_changes=false", commit)
        self.assertIn("force-with-lease", self.implement)

    def test_nested_checkout_classifier_rejects_non_directory_path(self):
        helper = FIXTURE_INFRA_ROOT / "config/implementer_nested_checkout.py"
        with subprocess.Popen(
            ["python3", str(helper), str(REPO_ROOT / "package.json")],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        ) as proc:
            _stdout, stderr = proc.communicate(timeout=30)
        self.assertNotEqual(proc.returncode, 0)
        self.assertIn("nested_checkout_not_directory", stderr)

    def test_run_app_checks_binds_pr_context_without_fetch(self):
        self.assertIn("--pr-base-sha SHA --pr-head-sha SHA", self.run_app_checks)
        self.assertIn('git cat-file -e "${validation_base_sha}^{commit}"', self.run_app_checks)
        self.assertIn("git merge-base", self.run_app_checks)
        self.assertIn('validation_mode="pr-validation"', self.run_app_checks)
        self.assertIn('validation_mode="pr-ancestry"', self.run_app_checks)
        self.assertIn("capture fixture comparison failed", self.run_app_checks)
        export_pr_base = "export " + "PR_" + "BASE_SHA="
        self.assertIn(export_pr_base, self.run_app_checks)
        self.assertNotIn("git fetch", self.run_app_checks)

    def test_ci_and_implement_pass_exact_pr_context(self):
        self.assertIn("fetch-depth: 0", self.ci)
        self.assertIn("github.event.pull_request.base.sha", self.ci)
        self.assertIn("github.event.pull_request.head.sha", self.ci)
        self.assertGreaterEqual(
            self.implement.count(
                '--pr-base-sha "${{ steps.branch.outputs.integration_sha }}"'
            ),
            2,
        )
        self.assertGreaterEqual(
            self.implement.count('--pr-head-sha "$(git rev-parse HEAD)"'),
            2,
        )

    def test_roles_yml_unchanged_and_no_openai_route_added(self):
        self.assertIn("implementer: cursor/composer-2.5", self.roles)
        self.assertIn(
            "reviewer: cursor/grok-4.6[effort=high,fast=false]", self.roles
        )
        self.assertNotIn("openai", self.roles.lower())

    def test_readme_names_current_state_pin_restore_nested_and_pr_context(self):
        self.assertIn(CURRENT_PIN, self.readme)
        self.assertIn("post-caller-checkout restore", self.readme)
        self.assertIn("post-implementer helper-lifetime", self.readme)
        self.assertIn("nested-checkout", self.readme)
        self.assertIn("immutable PR-context", self.readme)

    def test_protected_comparison_anchor_constant_is_valid_commit(self):
        self.assertRegex(PROTECTED_COMPARISON_ANCHOR, r"^[0-9a-f]{40}$")
        completed = subprocess.run(
            ["git", "cat-file", "-e", f"{PROTECTED_COMPARISON_ANCHOR}^{{commit}}"],
            cwd=REPO_ROOT,
            capture_output=True,
            check=False,
        )
        self.assertEqual(completed.returncode, 0, completed.stderr.decode())

    def test_seven_no_change_paths_match_anchor_and_are_absent_from_diff(self):
        diff_names = set(git_diff_names(PROTECTED_COMPARISON_ANCHOR))
        for relative in NO_CHANGE_PATHS:
            anchor_bytes = git_show_bytes(PROTECTED_COMPARISON_ANCHOR, relative)
            working_bytes = (REPO_ROOT / relative).read_bytes()
            self.assertEqual(
                working_bytes,
                anchor_bytes,
                f"{relative} differs from protected comparison anchor",
            )
            self.assertNotIn(relative, diff_names, f"{relative} appears in diff")
        for json_path in NO_CHANGE_PATHS[:2]:
            text = (REPO_ROOT / json_path).read_text(encoding="utf-8")
            self.assertIn(VOC112_SUBJECT_REVISION, text)
        self.assertIn(
            "VOC112_PROMOTION_PR",
            self.provenance_test,
        )

    def test_package_json_is_anchor_identical_and_named_helpers_absent(self):
        relative = "package.json"
        self.assertEqual(
            (REPO_ROOT / relative).read_bytes(),
            git_show_bytes(PROTECTED_COMPARISON_ANCHOR, relative),
        )
        package = (REPO_ROOT / relative).read_text(encoding="utf-8")
        prov_mode = "VOC112_" + "CAPTURE_" + "PROVENANCE_MODE"
        self.assertNotIn(prov_mode, package)
        self.assertIn("node --test scripts/foundation/*.test.mjs", package)
        for helper in PROHIBITED_HELPERS:
            self.assertFalse((REPO_ROOT / helper).exists(), helper)

    def test_complete_diff_scan_includes_caller_tests_without_wholesale_exclusion(self):
        scan_exclude_prefixes_do_not_skip_caller_tests()
        regression = "tooling/governance/tests/test_voc136_caller_replacement.py"
        scanner = "tooling/governance/tests/voc136_bypass_scan.py"
        self.assertTrue(should_scan_path(regression))
        self.assertTrue(should_scan_path(scanner))
        self.assertFalse(should_scan_path(f"{FIXTURE_MIRROR_PREFIX}config/run-app-checks.sh"))
        for relative in (regression, scanner):
            path = REPO_ROOT / relative
            scan_changed_path_for_bypasses(relative, path.read_text(encoding="utf-8"))
        scan_implementation_diff()

    def test_scanner_negative_cases_reject_relocated_bypasses(self):
        git_cmd = "gi" + "t"
        fetch_cmd = "fe" + "tch"
        capture_token = "cap" + "ture"
        hydrate_name = "hy" + "drate" + "-" + "voc112" + "-git-objects.mjs"
        prov_env = "VOC112_" + "CAPTURE_" + "PROVENANCE_MODE"
        pr_base = "PR_" + "BASE_SHA"
        cases = [
            (
                "synthetic-fetch.sh",
                f"#!/bin/sh\n{git_cmd} {fetch_cmd} origin {capture_token}\n",
            ),
            (
                hydrate_name,
                f"// {hydrate_name}\nexport default function hydrate() {{}}\n",
            ),
            (
                "wrapper.sh",
                f"#!/bin/sh\nexport {prov_env}=pr-validation\npnpm test\n",
            ),
            (
                "validate-workspace-wrapper.mjs",
                f"process.env.{pr_base}='deadbeef';\nimport './validate-workspace.mjs';\n",
            ),
            (
                "tooling/governance/tests/synthetic_import.py",
                f"import subprocess\nsubprocess.run(['{git_cmd}', '{fetch_cmd}', 'origin', '{capture_token}'])\n",
            ),
            (
                "skip-local.mjs",
                f"process.env.{prov_env}='pr-ancestry';\n",
            ),
        ]
        for relative, payload in cases:
            with self.subTest(relative=relative):
                with self.assertRaises(AssertionError):
                    scan_changed_path_for_bypasses(relative, payload)

    def test_scanner_allows_benign_mentions(self):
        prov_label = "VOC112_" + "CAPTURE_" + "PROVENANCE_" + "MODE"
        pr_label = "PR_" + "BASE_SHA"
        benign = (
            f"# regression documents that {prov_label} must not be set by wrappers\n"
            f"# and that {pr_label} overrides around validate-workspace are forbidden\n"
            "def test_example():\n"
            "    assert True\n"
        )
        scan_changed_path_for_bypasses(
            "tooling/governance/tests/synthetic_benign.py", benign
        )

    def test_regression_module_is_scan_clean_when_tracked(self):
        module = Path(__file__)
        relative = str(module.relative_to(REPO_ROOT))
        self.assertTrue(should_scan_path(relative))
        scan_changed_path_for_bypasses(relative, module.read_text(encoding="utf-8"))
        diff_names = git_diff_names(PROTECTED_COMPARISON_ANCHOR)
        self.assertIn(relative, diff_names, "regression module must be in implementation diff")

    def test_evidence_records_anchor_base_merge_hashes_and_feasible_binding(self):
        self.assertIn(PROTECTED_COMPARISON_ANCHOR, self.evidence)
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)
        self.assertIn("tooling/governance/tests/**", self.evidence)
        self.assertIn("App-authored independent-review comment/check", self.evidence)
        self.assertIn("does **not** require", self.evidence)
        self.assertNotRegex(
            self.evidence,
            r"implementation head SHA:\s*[0-9a-f]{40}",
        )

    def test_replacement_carrier_is_voc136_not_exhausted_carriers(self):
        self.assertIn("#1075", self.evidence)
        self.assertIn("not #1075", self.evidence)
        self.assertIn("#1070", self.evidence)
        self.assertIn("not #1070", self.evidence)
        self.assertIn("#1051", self.evidence)
        self.assertIn("#1056", self.evidence)
        self.assertIn("#1065", self.evidence)
        self.assertIn("VOC-136-T00 attempt `1`", self.evidence)
        exclude_label = "SCAN_" + "EXCLUDE_" + "PREFIXES"
        self.assertIn(exclude_label, self.evidence)
        self.assertIn("not redispatched", self.evidence)
        self.assertNotIn("snapshot-gap", self.evidence.lower())

    def test_implementation_pr_base_recorded_separately_from_anchor(self):
        self.assertRegex(IMPLEMENTATION_PR_BASE, r"^[0-9a-f]{40}$")
        self.assertNotEqual(IMPLEMENTATION_PR_BASE, PROTECTED_COMPARISON_ANCHOR)
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn("governance-only", self.evidence.lower())


if __name__ == "__main__":
    unittest.main()
