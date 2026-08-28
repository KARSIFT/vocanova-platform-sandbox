"""VOC-138 caller regressions for promotion PR pr-validation and recovery repair."""

from __future__ import annotations

import hashlib
import os
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

AUTHORITATIVE_PIN = "123735c80fec813a5b46a004f3e1122bd425cde2"
STALE_PIN_167 = "b263c0c110591cc798b89277dfc35542abb1597b"
PROTECTED_COMPARISON_ANCHOR = "b9e74fc2db4691c48c637639b265d527de9f4505"
IMPLEMENTATION_PR_BASE = "e89a02723cfbcaed952a868f2ab3f1442fd04fae"
VOC112_SUBJECT_REVISION = "f9d11e232a07c7d7a9c433d02c9267912543ba10"

MIRRORED_FILE_HASHES = {
    ".github/workflows/ci.yml": (
        "54dd080ece5e9dd6564788810025b0c0bf8b3bfe49d509b9771fd2ac88f3828a"
    ),
    ".github/workflows/self-ci.yml": (
        "3c2d074afd694da31cfbefc38cf42de74132d5fc07e769e2aeb4d4a59d9761be"
    ),
    "config/run-app-checks.sh": (
        "90c9f94db19825c30168f03d13ea1de21e72e1bb1c7a5fb41c93118d62e0c4b7"
    ),
    "config/actions-check-recovery-runner.py": (
        "e3f3504e0e6104ea5ff7f540ac591ded12e59d49c11cc810502ba5a6b84468e9"
    ),
    "config/promotion_status_attestation.py": (
        "53a50e2c70d31f38750ad4134abeaf30e31c63ac9e3ac74197c638ce2a8cc1ca"
    ),
    "config/promotion-status-attestation-runner.py": (
        "354cb65e434b158983f37440aee5bc14c2d60ba4db2ad9b5feb4446bade4ee2f"
    ),
    "templates/project-repo/.github/workflows/pipeline.yml": (
        "7a1532acec1354b5f2b9ce5096f031d9c6060fb9f2d769d47dd1e4f770c24616"
    ),
    "tests/test_app_check_context.py": (
        "d572b91eeb5c8270082e52e9650194df7ca644dccbbeebaeb8de02fa2c3a6e35"
    ),
    "tests/test_promotion_status_attestation.py": (
        "5f924fa4857931a03ac7a69197314f9a4d517461e3de47ab778fa7129e02ffe4"
    ),
    "tests/test_voc122_actions_check_recovery.py": (
        "8ca6b211aab6bca8cf64f2dfa9169855685b52a7ba6e34ac77bf7ea4f28f29b0"
    ),
    "tests/test_voc138_promotion_pr_provenance.py": (
        "253870621ab895a42258d9b0b5a8285b7dd05e29d97eea603291f3eb75dc51ff"
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

    def test_actual_voc112_assertion_accepts_pr_validation_and_rejects_ancestry(self):
        subject_lookup = subprocess.run(
            ["git", "cat-file", "-e", f"{VOC112_SUBJECT_REVISION}^{{commit}}"],
            cwd=REPO_ROOT,
            capture_output=True,
        )
        self.assertNotEqual(subject_lookup.returncode, 0)
        head = subprocess.check_output(
            ["git", "rev-parse", "HEAD"], cwd=REPO_ROOT, text=True
        ).strip()
        environment = os.environ.copy()
        environment.update(
            {
                "PR_BASE_SHA": IMPLEMENTATION_PR_BASE,
                "PR_HEAD_SHA": head,
                "VOC112_CAPTURE_PROVENANCE_MODE": "pr-validation",
            }
        )
        command = [
            "node",
            "--test",
            "scripts/foundation/voc112-navigation-benchmark.test.mjs",
        ]
        promotion = subprocess.run(
            command,
            cwd=REPO_ROOT,
            env=environment,
            text=True,
            capture_output=True,
        )
        self.assertEqual(promotion.returncode, 0, promotion.stderr)

        environment["VOC112_CAPTURE_PROVENANCE_MODE"] = "pr-ancestry"
        ordinary = subprocess.run(
            command,
            cwd=REPO_ROOT,
            env=environment,
            text=True,
            capture_output=True,
        )
        self.assertNotEqual(ordinary.returncode, 0)
        self.assertIn(
            "PR ancestry mode requires every captured commit object",
            ordinary.stdout + ordinary.stderr,
        )

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
