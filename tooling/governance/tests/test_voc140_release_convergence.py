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
    "config/promotion_ci_attestation.py": None,
    "config/production_merge_guard.py": None,
    "config/verify-production-merge-guard.sh": None,
    "config/actions_check_recovery.py": None,
    "config/actions-check-recovery-runner.py": None,
    "config/authoritative-checks-runner.py": None,
    "config/promotion_status_attestation.py": None,
    ".github/workflows/release.yml": None,
    ".github/workflows/merge-gate.yml": None,
    "tests/test_voc140_release_carrier_attestation.py": None,
    "tests/test_voc140_production_merge_guard.py": None,
}


def sha256_bytes(content: bytes) -> str:
    return hashlib.sha256(content).hexdigest()


class Voc140ReleaseConvergenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.readme = read_fixture("README.md")
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")
        for relative, _ in list(MIRRORED_FILE_HASHES.items()):
            MIRRORED_FILE_HASHES[relative] = sha256_bytes(
                read_fixture(relative).encode("utf-8")
            )

    def test_pin_advances_to_reviewed_voc140_infra_merge(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_mirrored_files_are_byte_stable(self):
        for relative, expected in MIRRORED_FILE_HASHES.items():
            with self.subTest(relative=relative):
                actual = sha256_bytes(read_fixture(relative).encode("utf-8"))
                self.assertEqual(actual, expected)

    def test_readme_records_voc140_recovery_and_guard_contract(self):
        self.assertIn(CURRENT_PIN, self.readme)
        self.assertIn("promotion-pr-validation", self.readme)
        self.assertIn("Administration", self.readme)

    def test_release_workflow_has_isolated_guard_mint(self):
        release = read_fixture(".github/workflows/release.yml")
        self.assertIn("Mint App installation token for production merge guard", release)
        self.assertIn("permission-administration: write", release)
        self.assertIn('GH_TOKEN="$GUARD_TOKEN" bash karsift-ai-infra/config/verify-production-merge-guard.sh', release)

    def test_fixture_promotion_ci_attestation_module_exists(self):
        text = read_fixture("config/promotion_ci_attestation.py")
        self.assertIn("is_release_carrier_run", text)
        self.assertIn("parent_run_is_attestable", text)


if __name__ == "__main__":
    unittest.main()
