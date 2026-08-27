"""VOC-133 complete VOC-112 no-change boundary regressions."""

from __future__ import annotations

import json
import subprocess
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"

VOC112_NO_CHANGE_PATHS = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
    "scripts/foundation/voc112-navigation-benchmark.test.mjs",
    "AGENTS.md",
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
)


def resolve_develop_base() -> str:
    for ref in ("origin/develop", "develop"):
        result = subprocess.run(
            ["git", "rev-parse", ref],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if result.returncode == 0:
            return result.stdout.strip()
    raise unittest.SkipTest("develop base ref is unavailable in this checkout")


def git_show_text(revision: str, relative: str) -> str:
    result = subprocess.run(
        ["git", "show", f"{revision}:{relative}"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise AssertionError(
            f"git show {revision}:{relative} failed: {result.stderr.strip()}"
        )
    return result.stdout


def git_diff_names(base: str) -> set[str]:
    result = subprocess.run(
        ["git", "diff", "--name-only", base, "HEAD", "--", *VOC112_NO_CHANGE_PATHS],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise AssertionError(result.stderr.strip())
    return {line for line in result.stdout.splitlines() if line}


class Voc133CompleteVoc112BoundaryTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.develop_base = resolve_develop_base()
        cls.evidence = (
            REPO_ROOT
            / "specs/changes/VOC-133-consume-exact-infra-166-and-complete-release/t00-evidence.md"
        ).read_text(encoding="utf-8")
        cls.provenance_test = (
            REPO_ROOT / "scripts/foundation/voc112-navigation-benchmark.test.mjs"
        ).read_text(encoding="utf-8")

    def test_voc112_paths_are_byte_identical_to_develop_base(self):
        for relative in VOC112_NO_CHANGE_PATHS:
            working = (REPO_ROOT / relative).read_bytes()
            base = git_show_text(self.develop_base, relative).encode("utf-8")
            self.assertEqual(
                working,
                base,
                f"{relative} differs from develop base {self.develop_base}",
            )

    def test_voc112_paths_are_absent_from_implementation_diff(self):
        changed = git_diff_names(self.develop_base)
        self.assertEqual(
            changed,
            set(),
            f"VOC-112 no-change paths appear in diff against develop: {sorted(changed)}",
        )

    def test_voc112_json_subject_revision_unchanged(self):
        traces = json.loads(
            (REPO_ROOT / VOC112_NO_CHANGE_PATHS[0]).read_text(encoding="utf-8")
        )
        self.assertEqual(traces["subject_revision"], VOC112_SUBJECT_REVISION)
        discovery = json.loads(
            (REPO_ROOT / VOC112_NO_CHANGE_PATHS[1]).read_text(encoding="utf-8")
        )
        for entry in discovery["discoveries"]:
            self.assertEqual(entry["subject_revision"], VOC112_SUBJECT_REVISION)

    def test_provenance_test_fail_closes_local_mode_on_missing_capture_commit(self):
        self.assertIn(
            "a full local checkout must already contain the captured commit",
            self.provenance_test,
        )
        self.assertNotIn(
            'if (mode === "local") {\n      return "squash-safe-push";',
            self.provenance_test,
        )

    def test_evidence_names_infra_merge_and_does_not_claim_false_revert(self):
        self.assertIn("f3d79177bf8a9abe0dae550f39502165d494c576", self.evidence)
        for relative in VOC112_NO_CHANGE_PATHS:
            self.assertNotIn(
                f"reverted `{relative}`",
                self.evidence,
                "evidence must not claim a protected-path revert while the path is unchanged",
            )


if __name__ == "__main__":
    unittest.main()
