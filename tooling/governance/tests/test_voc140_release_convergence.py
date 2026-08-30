"""VOC-140 caller regressions for pin advance and mirrored infra contract."""

from __future__ import annotations

import hashlib
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-140-release-convergence-cannot-trust-its-own-ci-run/t00-evidence.md"
)

AUTHORITATIVE_PIN = "599436835371f27fac52ec6b47a18b36257366ac"
CURRENT_PIN = "0ee1daf1aecdb5039ecc0fc74f5c64b24cdd5f5d"
IMPLEMENTATION_PR_BASE = "c59548375764d938265910cd07f2c2a73e337c01"

MIRRORED_FILE_HASHES = {
    "config/promotion_ci_attestation.py": (
        "84b07e3c2dc0a8ef2faed81310bd9c692e8a9203f85d49d67e84888732888c47"
    ),
    "config/production_merge_guard.py": (
        "a12720fdf6d67c533f5d478ff8396526900a5109f96e3fc63f2068ae820c724f"
    ),
    "config/verify-production-merge-guard.sh": (
        "bb5587724c1ff38995c7b78b3bf17eff1de596ea404ad0559cf9a3fe401000f3"
    ),
    "config/actions_check_recovery.py": (
        "c13a2da50773757a628ef3a40a48e024d8bbbfb84cbcc14adaa81f6746b023d8"
    ),
    "config/actions-check-recovery-runner.py": (
        "04e03532a086a81eb53ffe8ac0d4b8cdf086f2902b32dc4e7c8ed643060662ea"
    ),
    "config/authoritative-checks-runner.py": (
        "7b6e31c1fae952d12a7671eb6aa20ee03de9ab5340d48c4afe30cc60cb70c288"
    ),
    "config/promotion_status_attestation.py": (
        "18329e515df88dda113c23fea3dd32275d51635370b9c8a35aff39f6763eb15a"
    ),
    ".github/workflows/release.yml": (
        "52b70fd84bcaf08614bf3ed0aa27526c0ac74843937eff6d9671cf885a064eee"
    ),
    ".github/workflows/merge-gate.yml": (
        "3de99346f463fee640d284b592d103b983bdd20a4e8478474cf67b5b96030162"
    ),
    "tests/test_voc140_release_carrier_attestation.py": (
        "4dda251fecbf87cebb82ff73e11954b3bd71f383bff71737f9631dddff68afe7"
    ),
    "tests/test_voc140_production_merge_guard.py": (
        "e5bd7b36b7df2d323211042181de77b8e138c060beba02974d944325dfee267a"
    ),
    "README.md": (
        "e3f37664447ed954e7a220868daaab09a1d1352b1ad0ad88e7b43b47e5306a38"
    ),
}


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class Voc140ReleaseConvergenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.readme = read_fixture("README.md")
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")

    def test_pin_advances_to_reviewed_voc140_infra_merge(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_mirrored_fixture_files_match_recorded_sha256_hashes(self):
        for relative, expected in MIRRORED_FILE_HASHES.items():
            with self.subTest(relative=relative):
                path = FIXTURE_INFRA_ROOT / relative
                self.assertEqual(sha256_file(path), expected, relative)

    def test_readme_records_voc140_recovery_and_guard_contract(self):
        self.assertIn(CURRENT_PIN, self.readme)
        self.assertIn("promotion-pr-validation", self.readme)
        self.assertIn("Administration", self.readme)
        self.assertNotIn("the App token remains mutation-only", self.readme)

    def test_release_workflow_has_isolated_guard_mint(self):
        release = read_fixture(".github/workflows/release.yml")
        self.assertIn("Mint App installation token for production merge guard", release)
        self.assertIn("permission-administration: write", release)
        self.assertIn(
            'GH_TOKEN="$GUARD_TOKEN" bash karsift-ai-infra/config/verify-production-merge-guard.sh',
            release,
        )

    def test_fixture_promotion_ci_attestation_module_exists(self):
        text = read_fixture("config/promotion_ci_attestation.py")
        self.assertIn("is_release_carrier_run", text)
        self.assertIn("parent_run_is_attestable", text)

    def test_evidence_records_implementation_pr_base(self):
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(CURRENT_PIN, self.evidence)


if __name__ == "__main__":
    unittest.main()
