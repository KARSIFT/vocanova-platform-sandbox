"""VOC-137 filename-independent PR base/head SHA override scanner regressions."""

from __future__ import annotations

import hashlib
import inspect
import subprocess
import unittest
from pathlib import Path

import voc136_bypass_scan
from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture
from voc136_bypass_scan import (
    _PR_BASE_ENV,
    _PR_HEAD_ENV,
    FIXTURE_MIRROR_PREFIX,
    PR_SHA_SET_PATTERN,
    SCAN_EXCLUDE_PREFIXES,
    scan_changed_path_for_bypasses,
    scan_exclude_prefixes_do_not_skip_caller_tests,
    should_scan_path,
)


REPO_ROOT = Path(__file__).resolve().parents[3]
AUTHORITATIVE_PIN = "b263c0c110591cc798b89277dfc35542abb1597b"
CURRENT_PIN = "1edd60b98e1785057f63b7686ee2822706574a97"
PROTECTED_COMPARISON_ANCHOR = "b9e74fc2db4691c48c637639b265d527de9f4505"
IMPLEMENTATION_PR_BASE = "ebe4c460d892b87b6de38915f9fbd5e30d3c051b"
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-137-fail-closed-on-pr-base-head-sha-overrides-in/t00-evidence.md"
)

MIRRORED_FILE_HASHES = {
    ".github/workflows/ci.yml": (
        "54dd080ece5e9dd6564788810025b0c0bf8b3bfe49d509b9771fd2ac88f3828a"
    ),
    ".github/workflows/implement.yml": (
        "e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a"
    ),
    ".github/workflows/release.yml": (
        "fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08"
    ),
    "config/run-app-checks.sh": (
        "4adab35c3a5ec91ee09c8917edd3f02e6ae861e22c9d375b78b7c2cb39fe09ed"
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
        "7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea"
    ),
}

NO_CHANGE_PATHS = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
    "scripts/foundation/voc112-navigation-benchmark-run.mjs",
    "scripts/foundation/validate-workspace.mjs",
    "AGENTS.md",
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
    "package.json",
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


class Voc137PrShaScanTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")
        cls.roles = read_fixture("config/roles.yml")
        cls.run_app_checks = read_fixture("config/run-app-checks.sh")

    def test_scan_exclude_prefixes_fixture_mirror_only(self):
        self.assertEqual(SCAN_EXCLUDE_PREFIXES, (FIXTURE_MIRROR_PREFIX,))
        scan_exclude_prefixes_do_not_skip_caller_tests()

    def test_filename_gate_removed_from_scanner_source(self):
        source = inspect.getsource(voc136_bypass_scan.scan_changed_path_for_bypasses)
        self.assertNotIn("validate-workspace", source)
        self.assertNotIn(".test.mjs", source)

    def test_pr_sha_pattern_uses_source_safe_env_parts(self):
        pattern_source = inspect.getsource(voc136_bypass_scan)
        self.assertIn("_PR_BASE_ENV", pattern_source)
        self.assertIn("_PR_HEAD_ENV", pattern_source)
        self.assertNotRegex(pattern_source, r"PR_(?:BASE|HEAD)_SHA")

    def test_voc136_fixture_assertion_literal_is_source_safe(self):
        voc136_source = (
            REPO_ROOT
            / "tooling/governance/tests/test_voc136_caller_replacement.py"
        ).read_text(encoding="utf-8")
        export_pr_base_literal = "export " + "PR_" + "BASE_SHA="
        self.assertNotIn(export_pr_base_literal, voc136_source)

    def test_shell_arbitrary_wrapper_rejects_issue_payload(self):
        export_kw = "ex" + "port"
        pr_base = _PR_BASE_ENV
        payload = f"{export_kw} {pr_base}=deadbeef\npnpm test\n"
        relative = "scripts/arbitrary-wrapper.sh"
        self.assertNotIn("validate-workspace", relative)
        self.assertFalse(relative.endswith(".test.mjs"))
        with self.assertRaises(AssertionError):
            scan_changed_path_for_bypasses(relative, payload)

    def test_node_arbitrary_wrapper_rejects_pr_head_assignment(self):
        relative = "scripts/arbitrary-head-wrapper.mjs"
        pr_head = _PR_HEAD_ENV
        payload = f"process.env.{pr_head}='deadbeef';\nimport 'node:test';\n"
        self.assertNotIn("validate-workspace", relative)
        self.assertFalse(relative.endswith(".test.mjs"))
        with self.assertRaises(AssertionError):
            scan_changed_path_for_bypasses(relative, payload)

    def test_python_arbitrary_wrapper_rejects_pr_base_assignment(self):
        relative = "scripts/arbitrary_wrapper.py"
        pr_base = _PR_BASE_ENV
        payload = (
            f"import os\nos.environ['{pr_base}']='deadbeef'\n"
            "import subprocess\nsubprocess.run(['pnpm', 'test'])\n"
        )
        with self.assertRaises(AssertionError):
            scan_changed_path_for_bypasses(relative, payload)

    def test_python_outside_fixture_mirror_is_scanned(self):
        relative = "tooling/governance/tests/synthetic_pr_sha.py"
        pr_base = _PR_BASE_ENV
        payload = f"os.environ['{pr_base}']='deadbeef'\n"
        self.assertTrue(should_scan_path(relative))
        self.assertFalse(should_scan_path(f"{FIXTURE_MIRROR_PREFIX}tests/test_app_check_context.py"))
        with self.assertRaises(AssertionError):
            scan_changed_path_for_bypasses(relative, payload)

    def test_benign_mentions_and_scanner_module_do_not_false_positive(self):
        prov_label = "VOC112_" + "CAPTURE_" + "PROVENANCE_" + "MODE"
        pr_label = _PR_BASE_ENV
        benign = (
            f"# regression documents that {prov_label} must not be set by wrappers\n"
            f"# and that {pr_label} overrides around validation are forbidden\n"
            "def test_example():\n"
            "    assert True\n"
        )
        scan_changed_path_for_bypasses(
            "tooling/governance/tests/synthetic_benign_voc137.py", benign
        )
        scanner_path = REPO_ROOT / "tooling/governance/tests/voc136_bypass_scan.py"
        scan_changed_path_for_bypasses(
            "tooling/governance/tests/voc136_bypass_scan.py",
            scanner_path.read_text(encoding="utf-8"),
        )
        self.assertIsNotNone(PR_SHA_SET_PATTERN)

    def test_fixture_run_app_checks_excluded_from_scan(self):
        fixture_relative = f"{FIXTURE_MIRROR_PREFIX}config/run-app-checks.sh"
        self.assertFalse(should_scan_path(fixture_relative))
        export_pr_base = "export " + "PR_" + "BASE_SHA="
        self.assertIn(export_pr_base, self.run_app_checks)

    def test_pin_and_mirrored_fixture_bytes_unchanged(self):
        self.assertEqual(read_fixture("PINNED_SHA.txt").strip(), CURRENT_PIN)
        for relative, expected in MIRRORED_FILE_HASHES.items():
            path = FIXTURE_INFRA_ROOT / relative
            self.assertTrue(path.is_file(), f"missing fixture file: {relative}")
            self.assertEqual(sha256_file(path), expected, relative)
        if CURRENT_PIN == AUTHORITATIVE_PIN:
            diff_names = git_diff_names(IMPLEMENTATION_PR_BASE)
            for name in diff_names:
                self.assertFalse(
                    name.startswith("tooling/governance/fixtures/karsift-ai-infra/"),
                    f"fixture path in diff: {name}",
                )

    def test_seven_no_change_paths_and_roles_unchanged(self):
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
        self.assertIn("implementer: cursor/composer-2.5", self.roles)
        self.assertIn(
            "reviewer: cursor/grok-4.6[effort=high,fast=false]", self.roles
        )
        self.assertNotIn("openai", self.roles.lower())

    def test_tracked_modules_are_scan_clean(self):
        modules = (
            "tooling/governance/tests/voc136_bypass_scan.py",
            "tooling/governance/tests/test_voc136_caller_replacement.py",
            "tooling/governance/tests/test_voc137_pr_sha_scan.py",
        )
        for relative in modules:
            path = REPO_ROOT / relative
            self.assertTrue(should_scan_path(relative), relative)
            scan_changed_path_for_bypasses(relative, path.read_text(encoding="utf-8"))

    def test_evidence_records_base_and_binding_contract(self):
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(PROTECTED_COMPARISON_ANCHOR, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)
        self.assertIn("App-authored independent-review comment/check", self.evidence)
        self.assertIn("does **not** require", self.evidence)
        self.assertNotRegex(
            self.evidence,
            r"implementation head SHA:\s*[0-9a-f]{40}",
        )


if __name__ == "__main__":
    unittest.main()
