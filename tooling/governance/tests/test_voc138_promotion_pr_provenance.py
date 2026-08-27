"""VOC-138 caller regressions for promotion PR pr-validation and recovery repair."""

from __future__ import annotations

import hashlib
import subprocess
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-138-promotion-pr-ci-fails-voc-112-ancestry-when/t00-evidence.md"
)

AUTHORITATIVE_PIN = "ac0edc4b5b8f6165fa5e23a7b166dc2a0c2ea18f"
STALE_PIN_167 = "b263c0c110591cc798b89277dfc35542abb1597b"
PROTECTED_COMPARISON_ANCHOR = "b9e74fc2db4691c48c637639b265d527de9f4505"
IMPLEMENTATION_PR_BASE = "e89a02723cfbcaed952a868f2ab3f1442fd04fae"
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"

MIRRORED_FILE_HASHES = {
    ".github/workflows/ci.yml": (
        "b5f2e0d82bbe3e98f85fe5da064c144732ab500abe1e6cf3348c938b3deef2c1"
    ),
    "config/run-app-checks.sh": (
        "ce875cd2b2450663e1c1c611fd31533b2d222afc5a3c443f0623dbd5d3eca03c"
    ),
    "config/actions-check-recovery-runner.py": (
        "e3f3504e0e6104ea5ff7f540ac591ded12e59d49c11cc810502ba5a6b84468e9"
    ),
    "config/promotion_status_attestation.py": (
        "e504f5a04fe00965e8115b2cf7f218085d246f792fa6f3aba2b75629ad1b50b8"
    ),
    "config/promotion-status-attestation-runner.py": (
        "fee155650894d571398ce199b306ff30be0ca0437948ca18ba80acd0b2e0cd7c"
    ),
    "templates/project-repo/.github/workflows/pipeline.yml": (
        "10f1a5c44ab69140219eaf4101a430bfc0a03ac7cd65662a3614f81a9b69b61d"
    ),
    "tests/test_app_check_context.py": (
        "c585df66f8a00d4526f5eeb5cb08300f5b9e359f7d0d210b15f8890a4480d18d"
    ),
    "tests/test_voc122_actions_check_recovery.py": (
        "8ca6b211aab6bca8cf64f2dfa9169855685b52a7ba6e34ac77bf7ea4f28f29b0"
    ),
    "tests/test_voc138_promotion_pr_provenance.py": (
        "bd160cc0c64b1cff36957c35ef7e1fe0deaab0a5e209ccf44052eb9e084e0d83"
    ),
}

NO_CHANGE_PATHS = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
    "scripts/foundation/voc112-navigation-benchmark.test.mjs",
    "scripts/foundation/voc112-navigation-benchmark-run.mjs",
    "scripts/foundation/validate-workspace.mjs",
    "AGENTS.md",
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
    "package.json",
)


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class Voc138PromotionPrProvenanceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.ci = read_fixture(".github/workflows/ci.yml")
        cls.run_app_checks = read_fixture("config/run-app-checks.sh")
        cls.pipeline = CALLER_WORKFLOWS / "pipeline.yml"
        cls.pipeline_text = cls.pipeline.read_text(encoding="utf-8")
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")
        cls.ops_doc = (
            REPO_ROOT / "docs/operations/11-devops-and-ci-cd.md"
        ).read_text(encoding="utf-8")
        cls.skills_doc = (
            REPO_ROOT / "docs/development/agent-skills.md"
        ).read_text(encoding="utf-8")

    def test_pin_equals_authoritative_infra_merge_and_not_stale_167(self):
        self.assertEqual(self.pin, AUTHORITATIVE_PIN)
        self.assertNotEqual(self.pin, STALE_PIN_167)

    def test_mirrored_fixture_files_match_recorded_sha256_hashes(self):
        for relative, expected in MIRRORED_FILE_HASHES.items():
            path = FIXTURE_INFRA_ROOT / relative
            self.assertTrue(path.is_file(), f"missing fixture file: {relative}")
            self.assertEqual(sha256_file(path), expected, relative)

    def test_eight_voc112_no_change_paths_match_protected_anchor(self):
        for relative in NO_CHANGE_PATHS:
            working = REPO_ROOT / relative
            anchor = subprocess.check_output(
                [
                    "git",
                    "show",
                    f"{PROTECTED_COMPARISON_ANCHOR}:{relative}",
                ],
                cwd=REPO_ROOT,
                text=True,
            )
            self.assertEqual(working.read_text(encoding="utf-8"), anchor, relative)

    def test_run_app_checks_supports_promotion_pr_validation(self):
        self.assertIn("--promotion-pr", self.run_app_checks)
        self.assertIn('if [ "$promotion_pr" != true ]; then', self.run_app_checks)
        self.assertNotIn("git fetch", self.run_app_checks)

    def test_caller_pipeline_passes_promotion_metadata_to_ci(self):
        self.assertIn("promotion-pr-metadata:", self.pipeline_text)
        self.assertIn("promotion_pr:", self.pipeline_text)
        self.assertIn("pr_base_sha:", self.pipeline_text)
        self.assertIn("pr_head_sha:", self.pipeline_text)

    def test_current_state_docs_describe_promotion_pr_validation(self):
        self.assertIn("pr-validation", self.ops_doc)
        self.assertNotIn(
            "promotion PR validates capture\nprovenance with the same `squash-safe-push`",
            self.ops_doc,
        )
        self.assertIn("promotion pull requests deterministically use", self.skills_doc)
        self.assertIn("pr-validation", self.skills_doc)

    def test_evidence_records_implementation_pr_base(self):
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)


if __name__ == "__main__":
    unittest.main()
