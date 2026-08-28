"""VOC-139 caller regressions for promotion head-hash binding and metadata."""

from __future__ import annotations

import hashlib
import json
import os
import stat
import subprocess
import tempfile
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-139-promotion-recovery-cannot-validate-an-accumulated/t00-evidence.md"
)

AUTHORITATIVE_PIN = "123735c80fec813a5b46a004f3e1122bd425cde2"
CURRENT_PIN = "1edd60b98e1785057f63b7686ee2822706574a97"
PROTECTED_COMPARISON_ANCHOR = "b9e74fc2db4691c48c637639b265d527de9f4505"
IMPLEMENTATION_PR_BASE = "ecb3d6d8e30628a9691928ea4594523f7193b961"
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"

NO_CHANGE_PATHS = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
    "scripts/foundation/voc112-navigation-benchmark-run.mjs",
    "scripts/foundation/validate-workspace.mjs",
    "AGENTS.md",
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
    "package.json",
)

METADATA_SCRIPT = """
set -euo pipefail
payload=$(gh api "repos/$GITHUB_REPOSITORY/pulls/$PROMOTION_PR_NUMBER")
base_ref=$(jq -r .base.ref <<<"$payload")
head_ref=$(jq -r .head.ref <<<"$payload")
base_repo=$(jq -r .base.repo.full_name <<<"$payload")
head_repo=$(jq -r .head.repo.full_name <<<"$payload")
state=$(jq -r .state <<<"$payload")
base_sha=$(jq -r .base.sha <<<"$payload")
head_sha=$(jq -r .head.sha <<<"$payload")
if [ "$state" != "open" ]; then
  echo "promotion PR is not open" >&2
  exit 1
fi
if [ "$base_ref" != "main" ] || [ "$head_ref" != "develop" ] ||
   [ "$base_repo" != "$GITHUB_REPOSITORY" ] ||
   [ "$head_repo" != "$GITHUB_REPOSITORY" ]; then
  echo "promotion pair mismatch" >&2
  exit 1
fi
if ! [[ "$base_sha" =~ ^[0-9a-f]{40}$ ]] ||
   ! [[ "$head_sha" =~ ^[0-9a-f]{40}$ ]]; then
  echo "promotion PR metadata missing immutable SHAs" >&2
  exit 1
fi
echo "pr_base_sha=$base_sha"
echo "pr_head_sha=$head_sha"
"""

VALID_PAYLOAD = {
    "state": "open",
    "base": {
        "ref": "main",
        "sha": "0" * 40,
        "repo": {"full_name": "KARSIFT/vocanova-platform-sandbox"},
    },
    "head": {
        "ref": "develop",
        "sha": "1" * 40,
        "repo": {"full_name": "KARSIFT/vocanova-platform-sandbox"},
    },
    "headRepository": {"id": 1, "name": "vocanova-platform-sandbox"},
}


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class Voc139PromotionRecoveryMetadataTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.pipeline_text = (CALLER_WORKFLOWS / "pipeline.yml").read_text(encoding="utf-8")
        cls.fixture_pipeline = read_fixture(
            "templates/project-repo/.github/workflows/pipeline.yml"
        )
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")
        cls.readme = (FIXTURE_INFRA_ROOT / "README.md").read_text(encoding="utf-8")

    def test_pin_advances_to_voc139_infra_merge(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_seven_voc112_paths_remain_frozen_against_protected_anchor(self):
        for relative in NO_CHANGE_PATHS:
            working = REPO_ROOT / relative
            anchor = subprocess.check_output(
                ["git", "show", f"{PROTECTED_COMPARISON_ANCHOR}:{relative}"],
                cwd=REPO_ROOT,
                text=True,
            )
            self.assertEqual(working.read_text(encoding="utf-8"), anchor, relative)

    def test_provenance_test_allows_promotion_head_hash_binding(self):
        head = subprocess.check_output(
            ["git", "rev-parse", "HEAD"], cwd=REPO_ROOT, text=True
        ).strip()
        base = subprocess.check_output(
            ["git", "rev-parse", "HEAD~1"], cwd=REPO_ROOT, text=True
        ).strip()
        environment = os.environ.copy()
        environment.update(
            {
                "PR_BASE_SHA": base,
                "PR_HEAD_SHA": head,
                "VOC112_CAPTURE_PROVENANCE_MODE": "pr-validation",
                "VOC112_PROMOTION_PR": "true",
            }
        )
        result = subprocess.run(
            ["node", "--test", "scripts/foundation/voc112-navigation-benchmark.test.mjs"],
            cwd=REPO_ROOT,
            env=environment,
            text=True,
            capture_output=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_caller_and_fixture_metadata_are_repository_explicit(self):
        for text in (self.pipeline_text, self.fixture_pipeline):
            self.assertIn(
                'gh api "repos/$GITHUB_REPOSITORY/pulls/$PROMOTION_PR_NUMBER"',
                text,
            )
            self.assertNotIn("headRepository.nameWithOwner", text)
            metadata_section = text.split("promotion-pr-metadata:", 1)[1].split("\n  ci:", 1)[0]
            self.assertNotIn("actions/checkout", metadata_section)

    def test_metadata_command_succeeds_without_git_repository(self):
        with tempfile.TemporaryDirectory() as directory:
            workspace = Path(directory)
            gh_path = workspace / "gh"
            gh_path.write_text(
                "#!/usr/bin/env bash\n"
                'if [ "$1" != "api" ]; then exit 99; fi\n'
                f"cat <<'EOF'\n{json.dumps(VALID_PAYLOAD)}\nEOF\n"
            )
            gh_path.chmod(gh_path.stat().st_mode | stat.S_IEXEC)
            environment = os.environ.copy()
            environment.update(
                {
                    "PATH": f"{workspace}:{environment['PATH']}",
                    "GITHUB_REPOSITORY": "KARSIFT/vocanova-platform-sandbox",
                    "PROMOTION_PR_NUMBER": "1090",
                    "GH_TOKEN": "fixture-token",
                }
            )
            result = subprocess.run(
                ["bash", "-c", METADATA_SCRIPT],
                cwd=workspace,
                text=True,
                capture_output=True,
                env=environment,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn(f"pr_base_sha={'0' * 40}", result.stdout)

    def test_fixture_readme_records_voc139_pin_and_hash_contract(self):
        self.assertIn(CURRENT_PIN, self.readme)
        self.assertIn("head/source-revision", self.readme)
        self.assertIn("merge-base-anchored", self.readme)

    def test_evidence_records_implementation_pr_base_and_infra_merge(self):
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(CURRENT_PIN, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)


if __name__ == "__main__":
    unittest.main()
