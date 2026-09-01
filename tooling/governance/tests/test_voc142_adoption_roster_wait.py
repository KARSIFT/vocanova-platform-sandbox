"""VOC-142 caller regressions for pin advance and roster wait/reuse contract."""

from __future__ import annotations

import hashlib
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-142-adoption-roster-wait-ignores-pending-required-ci/t00-evidence.md"
)
AGENTS_PATH = REPO_ROOT / "AGENTS.md"

AUTHORITATIVE_PIN = "67bdfd13ef875dead23ce4be01d7d0e8b976e289"
VOC142_INFRA_MERGE = "8993e867640dfb604dec0466c4e0787e68d8e258"
CURRENT_PIN = "ad2b27784e6fc33b3ac7e9dab48245dd6d08ac7f"
IMPLEMENTATION_PR_BASE = "1fc3473576bef96cffd861f4304168ad147296ef"

FINAL_INFRA_CHANGED_MIRRORS = frozenset(
    {
        ".github/workflows/adopt.yml",
        "config/authoritative-checks-runner.py",
        "config/roster_pr_wait.py",
        "config/roster_carrier.py",
        "config/roster-pr-wait-runner.py",
        "config/roster-carrier-runner.py",
        "tests/test_adoption_handoff.py",
        "tests/test_voc142_roster_wait_and_carrier.py",
    }
)

MIRRORED_FILE_HASHES = {
    ".github/workflows/adopt.yml": (
        "1cef92ded8504a53752b504477949013feec421f9e99fbd9a3224bb7adce1c49"
    ),
    "config/authoritative-checks-runner.py": (
        "cb0fe7867a2fd95a99935451057a288a01fedbc779d2f68da133e1291b1042d3"
    ),
    "config/roster_pr_wait.py": (
        "0615d9ed6fe253a96884980155d934f5c7f66938f863b1bfae4adb788170d8f3"
    ),
    "config/roster_carrier.py": (
        "9613922bd8593ba4f1d9ccd5cfee70a357024da8489531ce9e8dffc9bbe666fa"
    ),
    "config/roster-pr-wait-runner.py": (
        "abe71c87b53388f4861bfedde925b3b380d99d135b5586a47880fd5f19d1effc"
    ),
    "config/roster-carrier-runner.py": (
        "c06a3a7c1314d80edf53831fde399db424d48a8a04154b55e965708ced4d5c8a"
    ),
    "tests/test_adoption_handoff.py": (
        "b16e84853283d04c9e1d297175c6e23ee32f85f1e56c9983bdd1a82f3a125d8e"
    ),
    "tests/test_voc142_roster_wait_and_carrier.py": (
        "f445854bd40f350dd47c355d6ad9bcc27a5cea2d773967cff3f69437a68d1dcc"
    ),
}

CURRENT_PIN_ASSERTION_PATHS = (
    "scripts/foundation/voc097-fixture-matrix.test.mjs",
    "scripts/foundation/voc104-ready-for-review-reuse.test.mjs",
    "scripts/foundation/voc108-authoritative-lifecycle.test.mjs",
    "tooling/governance/tests/test_voc121_implement_policy.py",
    "tooling/governance/tests/test_voc122_implement_policy.py",
    "tooling/governance/tests/test_voc124_implement_policy.py",
    "tooling/governance/tests/test_voc125_implement_fixture.py",
    "tooling/governance/tests/test_voc125_implement_policy.py",
    "tooling/governance/tests/test_voc126_workflow_dispatch_input_limit.py",
    "tooling/governance/tests/test_voc129_caller_replacement.py",
    "tooling/governance/tests/test_voc136_caller_replacement.py",
    "tooling/governance/tests/test_voc137_pr_sha_scan.py",
    "tooling/governance/tests/test_voc138_promotion_pr_provenance.py",
    "tooling/governance/tests/test_voc139_promotion_recovery_metadata.py",
    "tooling/governance/tests/test_voc140_release_convergence.py",
    "tooling/governance/tests/test_voc142_adoption_roster_wait.py",
    "tooling/governance/tests/test_voc145_caller_replacement.py",
)


class Voc142AdoptionRosterWaitCallerTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.readme = read_fixture("README.md")
        cls.adopt = read_fixture(".github/workflows/adopt.yml")
        cls.agents = AGENTS_PATH.read_text(encoding="utf-8")
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")

    def test_current_pin_matches_independently_reviewed_infra_merge(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_changed_mirrors_match_declared_hashes(self):
        for relative in sorted(FINAL_INFRA_CHANGED_MIRRORS):
            path = FIXTURE_INFRA_ROOT / relative
            digest = hashlib.sha256(path.read_bytes()).hexdigest()
            self.assertEqual(
                digest,
                MIRRORED_FILE_HASHES[relative],
                relative,
            )

    def test_live_current_pin_assertions_reference_new_merge(self):
        self.assertEqual(len(CURRENT_PIN_ASSERTION_PATHS), 17)
        for relative in CURRENT_PIN_ASSERTION_PATHS:
            text = (REPO_ROOT / relative).read_text(encoding="utf-8")
            self.assertIn(CURRENT_PIN, text, relative)

    def test_current_docs_state_complete_required_set_and_carrier_reuse(self):
        self.assertIn("ruleset-required check set", self.agents.lower())
        self.assertIn("reuse", self.agents.lower())
        self.assertIn("roster-pr-wait-runner.py", self.adopt)
        self.assertIn("roster-carrier-runner.py", self.adopt)
        self.assertIn("required_complete", self.adopt)
        self.assertIn("reuse_merged", self.adopt)
        self.assertIn(CURRENT_PIN, self.readme)
        self.assertIn("two stable", self.readme.lower())

    def test_evidence_records_implementation_base_and_infra_merge(self):
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(VOC142_INFRA_MERGE, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)
        self.assertIn("roster-pr-wait-runner.py", self.evidence)
        self.assertIn("roster-carrier-runner.py", self.evidence)

    def test_voc141_package_records_remain_unmodified(self):
        package_dir = (
            REPO_ROOT / "specs/changes/VOC-141-promotion-recovery-waits-30-minutes-when-green-ci"
        )
        change_yaml = (package_dir / "change.yaml").read_text(encoding="utf-8")
        self.assertIn(AUTHORITATIVE_PIN, change_yaml)


if __name__ == "__main__":
    unittest.main()
