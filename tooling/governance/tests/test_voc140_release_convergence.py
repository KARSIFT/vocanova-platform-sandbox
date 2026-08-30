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
REPOSITORY_SETTINGS_PATH = REPO_ROOT / "docs/governance/repository-settings.md"
RECONCILIATION_NOTES_PATH = (
    REPO_ROOT / "docs/operations/19-governance-reconciliation-notes.md"
)

AUTHORITATIVE_PIN = "599436835371f27fac52ec6b47a18b36257366ac"
CURRENT_PIN = "9fdff24cd387cc2cdc468c84a3012b0c34b6c8e8"
IMPLEMENTATION_PR_BASE = "c59548375764d938265910cd07f2c2a73e337c01"

MIRRORED_FILE_HASHES = {
    ".github/workflows/merge-gate.yml": (
        "0762609f8903ed9d4bfdd3c2e22bb3e64994eeb60baa50d2ce3de6f17b4cd40e"
    ),
    ".github/workflows/release.yml": (
        "8198c14b3ff9ebb4d047ced6eaf2d5c37398c8da9f5709aaa0dbbf5252eef56f"
    ),
    "config/actions-check-recovery-runner.py": (
        "ff3e79cd31aa684e4ed591471f14d13afae58ce068a87abdd71eb47f0d571733"
    ),
    "config/actions_check_recovery.py": (
        "fa7b8052c6b11801fe9446e9589adfe4a4c5d2272afc61e6283d02e5893a9cfb"
    ),
    "config/authoritative-checks-runner.py": (
        "09963eaa7dc517ba2c52e2f2faaf4d3e119da2dea1eb480c41dfac36f74e8e84"
    ),
    "config/production_merge_guard.py": (
        "a12720fdf6d67c533f5d478ff8396526900a5109f96e3fc63f2068ae820c724f"
    ),
    "config/promotion-status-attestation-runner.py": (
        "d69a71d97e5ea9130b25fdbe63c922c18183f050535e747ee587d6917c57c216"
    ),
    "config/promotion_ci_attestation.py": (
        "06cfa79df461c1920ddcbddb533807b0b2d4c9f5bcf41e6c4117964735690290"
    ),
    "config/promotion_status_attestation.py": (
        "18329e515df88dda113c23fea3dd32275d51635370b9c8a35aff39f6763eb15a"
    ),
    "config/verify-production-merge-guard.sh": (
        "bb5587724c1ff38995c7b78b3bf17eff1de596ea404ad0559cf9a3fe401000f3"
    ),
    "tests/test_adoption_handoff.py": (
        "fda591948b4e0e540ba69ae2e42dc6b4267d4a005d1d7655a646972f1c21990c"
    ),
    "tests/test_production_merge_guard.py": (
        "e42ebe9536ceb55192f105b2e617c0dc031a673e3e009819bcd228bcbbd58190"
    ),
    "tests/test_promotion_status_attestation.py": (
        "6c55ff93fe464d0ab29be9ea68452ff437672a456ba39d22d613356eae15f393"
    ),
    "tests/test_voc114_actions_check_recovery.py": (
        "87eb36a0c8f5bf57fc474c29b0eca9642ab6a0195b7648ead3d1deec09544486"
    ),
    "tests/test_voc121_actions_check_recovery.py": (
        "44346ff783d8e56d289c94d8aec4914feee35c4942e9773c5dc0e25a377e5c31"
    ),
    "tests/test_voc122_actions_check_recovery.py": (
        "748e15ee723f002e93bd4451ace894c863509c9cc58ade398159b0f4573d1242"
    ),
    "tests/test_voc140_release_carrier_attestation.py": (
        "8382d8999c6559164488c02424fca6acc3e41dae8b9199b64d0bc973aa1671b0"
    ),
    "tests/test_voc140_production_merge_guard.py": (
        "da0352223067e5f781db3950bdd81facb2749d22755ca0510c07af774864ae05"
    ),
}
EXECUTABLE_MIRRORS = frozenset(
    {
        "config/production_merge_guard.py",
        "config/verify-production-merge-guard.sh",
    }
)

CURRENT_AUTHORITY_DOCUMENTS = {
    ".github/CODEOWNERS": ("active A-004", "at any risk class"),
    ".github/README.md": ("A-004 is the active", "RL1/RL2"),
    "CLAUDE.md": ("predecessor to active A-004", "R3 and R4"),
    "CONTRIBUTING.md": ("Under active A-004", "production deployment are enabled"),
    "docs/README.md": (
        "Current A-004-backed",
        "RL1/RL2 technical activation remains disabled",
    ),
    "docs/decisions/README.md": (
        "Under active A-004",
        "R4 retains stronger evidence",
    ),
    "docs/engineering/04-technical-architecture.md": (
        "A-004-remove-founder-approval-gates",
        "a004-transition-state.yaml",
    ),
    "docs/governance/16-autonomous-development-operating-model.md": (
        "A-004 now supersedes",
        "R0-R4 production releases may proceed automatically",
    ),
    "docs/governance/README.md": (
        "A-004 is the effective",
        "frozen historical audit evidence",
    ),
    "docs/governance/change-risk-classification.md": (
        "Under active A-004",
        "Historical initial-governance bootstrap classification",
    ),
    "docs/governance/protected-areas.md": ("Historical bootstrap", "active A-004"),
    "docs/operations/10-development-workflow.md": (
        "push-to-`main` production deployment are enabled",
        "RL1/RL2 technical activation remains disabled",
    ),
    "docs/operations/15-ai-native-product-and-engineering-operating-model.md": (
        "production deployment are\nimplemented and enabled",
        "Under active A-004, a founder `approved` comment cannot waive",
    ),
    "docs/templates/change-specification.md": ("active-A-004", "R0-R4"),
    "docs/templates/release-record.md": (
        "Active-A-004",
        "no founder-comment workflow gate",
    ),
    "docs/templates/technical-approval-request.md": (
        "under active A-004",
        "founder product/legal/strategy clarification",
    ),
}

STALE_LIVE_AUTHORITY_LITERALS = {
    ".github/CODEOWNERS": ("active A-003 it must not recreate",),
    ".github/README.md": ("A-003 governance authority is active",),
    "CLAUDE.md": ("A-003 has been effectively active since",),
    "CONTRIBUTING.md": (
        "Under active A-003, routine R3",
        "R4 founder authority remains unchanged",
        "R3 production remains blocked",
    ),
    "docs/README.md": ("that narrower capability (A-003 §10)",),
    "docs/decisions/README.md": (
        "Under active A-003 they do not",
        "consequential R4 decisions require the founder",
    ),
    "docs/engineering/04-technical-architecture.md": (
        "A-003](../governance/amendments/A-003-governed-autonomous-engineering-authority.md), and the\n[approval matrix",
    ),
    "docs/governance/16-autonomous-development-operating-model.md": (
        "Active A-003 now supersedes",
        "Under active\nUnder A-003",
        "Low-risk, reversible R0-R1 production releases may proceed automatically",
    ),
    "docs/governance/README.md": ("A-003 has been effectively active since",),
    "docs/governance/change-risk-classification.md": (
        "Under active A-003, an R3 path",
        "R3 production remains blocked until a qualified human steward",
    ),
    "docs/governance/protected-areas.md": (
        "ordinary R3 steward requirements apply and R3 production remains blocked",
    ),
    "docs/operations/10-development-workflow.md": (
        "production deployment are not technically active",
        "R4 consequences require founder approval",
        "future\nstaging source",
    ),
    "docs/operations/15-ai-native-product-and-engineering-operating-model.md": (
        "no staging or production deployment stage exists in the live pipeline",
        "High-risk waivers require founder approval before production",
    ),
    "docs/templates/change-specification.md": (
        "active-A-003 /",
        "Under active A-003",
    ),
    "docs/templates/release-record.md": (
        "Active-A-003 strengthened",
        "Founder approval when R4",
    ),
    "docs/templates/technical-approval-request.md": (
        "under active A-003",
        "R4 founder escalation",
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
        cls.repository_settings = REPOSITORY_SETTINGS_PATH.read_text(encoding="utf-8")
        cls.reconciliation_notes = RECONCILIATION_NOTES_PATH.read_text(encoding="utf-8")

    def test_pin_advances_to_reviewed_voc140_infra_merge(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_mirrored_fixture_files_match_recorded_sha256_hashes(self):
        for relative, expected in MIRRORED_FILE_HASHES.items():
            with self.subTest(relative=relative):
                path = FIXTURE_INFRA_ROOT / relative
                self.assertEqual(sha256_file(path), expected, relative)

    def test_mirrored_fixture_files_preserve_authoritative_modes(self):
        for relative in MIRRORED_FILE_HASHES:
            with self.subTest(relative=relative):
                expected = 0o755 if relative in EXECUTABLE_MIRRORS else 0o644
                actual = (FIXTURE_INFRA_ROOT / relative).stat().st_mode & 0o777
                self.assertEqual(actual, expected, relative)

    def test_readme_records_voc140_recovery_and_guard_contract(self):
        self.assertIn(CURRENT_PIN, self.readme)
        self.assertIn("promotion-pr-validation", self.readme)
        self.assertIn("Administration", self.readme)
        self.assertIn("never attestable", self.readme)
        self.assertIn("skipped release definition remains eligible", self.readme)
        self.assertNotIn("the App token remains mutation-only", self.readme)

    def test_current_docs_record_release_identity_and_two_token_contract(self):
        for document in (self.repository_settings, self.reconciliation_notes):
            with self.subTest(document=document[:40]):
                self.assertIn("not completed successfully", document)
                self.assertIn("non-skipped `release / converge`", document)
                self.assertIn("`promotion-pr-validation PR #<n>`", document)
                self.assertIn("skipped", document)
                self.assertIn("mutation", document)
                self.assertIn("token grants", document)
                self.assertIn("exactly Contents, Issues, and Pull", document)
                self.assertIn("Administration-write-only", document)
                self.assertIn("used only", document)

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

    def test_live_authority_docs_reconcile_a004_and_dispose_every_path(self):
        self.assertEqual(
            set(CURRENT_AUTHORITY_DOCUMENTS), set(STALE_LIVE_AUTHORITY_LITERALS)
        )
        for relative, required_literals in CURRENT_AUTHORITY_DOCUMENTS.items():
            with self.subTest(relative=relative):
                document = (REPO_ROOT / relative).read_text(encoding="utf-8")
                for required in required_literals:
                    self.assertIn(required, document, relative)
                for stale in STALE_LIVE_AUTHORITY_LITERALS[relative]:
                    self.assertNotIn(stale, document, relative)
                self.assertIn(f"`{relative}`", self.evidence, relative)

        self.assertIn("VOC-138 / VOC-139", self.evidence)
        self.assertIn("DOC-17 / DOC-18", self.evidence)
        self.assertIn("validator/test historical strings", self.evidence)


if __name__ == "__main__":
    unittest.main()
