from __future__ import annotations

import importlib.util
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
VALIDATOR = REPOSITORY_ROOT / "tooling/governance/validate_repository_foundation.py"


class RepositoryFoundationValidatorTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary.name) / "synthetic-repository"
        shutil.copytree(
            REPOSITORY_ROOT,
            self.root,
            # The validator exercises tracked repository policy. Copying an
            # installed pnpm tree into every synthetic fixture adds hundreds
            # of megabytes and can turn this suite from seconds into many
            # minutes without changing any assertion surface.
            ignore=shutil.ignore_patterns(
                ".git", "node_modules", "__pycache__", "*.pyc"
            ),
        )

    def tearDown(self) -> None:
        self.temporary.cleanup()

    def run_validator(self) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [sys.executable, str(VALIDATOR), "--repository-root", str(self.root)],
            check=False,
            capture_output=True,
            text=True,
            timeout=20,
        )

    def run_classifier(self, declared_risk: str) -> subprocess.CompletedProcess[str]:
        files = self.root / "changed-files.txt"
        body = self.root / "pr-body.md"
        files.write_text(".github/approved-policy/protected-paths.yaml\n", encoding="utf-8")
        body.write_text(f"Risk classification: {declared_risk}\n", encoding="utf-8")
        return subprocess.run(
            [
                "bash",
                "scripts/governance/classify-change-risk.sh",
                "--files-from",
                str(files),
                "--pr-body-file",
                str(body),
                "--require-declaration",
            ],
            cwd=self.root,
            check=False,
            capture_output=True,
            text=True,
            timeout=20,
        )

    def run_classifier_for_path(self, path: str, declared_risk: str) -> subprocess.CompletedProcess[str]:
        files = self.root / "changed-files.txt"
        body = self.root / "pr-body.md"
        files.write_text(f"{path}\n", encoding="utf-8")
        body.write_text(f"Risk classification: {declared_risk}\n", encoding="utf-8")
        return subprocess.run(
            [
                "bash",
                "scripts/governance/classify-change-risk.sh",
                "--files-from",
                str(files),
                "--pr-body-file",
                str(body),
                "--require-declaration",
            ],
            cwd=self.root,
            check=False,
            capture_output=True,
            text=True,
            timeout=20,
        )

    def assert_failure(self, marker: str) -> None:
        result = self.run_validator()
        self.assertEqual(1, result.returncode, result.stdout + result.stderr)
        self.assertIn(marker, result.stderr)

    def replace(self, relative: str, old: str, new: str) -> None:
        path = self.root / relative
        text = path.read_text(encoding="utf-8")
        self.assertIn(old, text)
        path.write_text(text.replace(old, new, 1), encoding="utf-8")

    def test_valid_repository_passes(self) -> None:
        result = self.run_validator()
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assertIn("validation passed", result.stdout)

    def test_valid_evidence_backed_a003_active_state_passes(self) -> None:
        result = self.run_validator()
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)

    def test_classifier_accepts_r4_for_protected_policy(self) -> None:
        result = self.run_classifier("R4")
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assertIn("Detected path-based risk floor: R4", result.stdout)

    def test_classifier_rejects_declaration_below_r4_floor(self) -> None:
        result = self.run_classifier("R3")
        self.assertEqual(1, result.returncode, result.stdout + result.stderr)
        self.assertIn("below the detected floor R4", result.stderr)

    def test_missing_required_root_file_fails(self) -> None:
        (self.root / "AGENTS.md").unlink()
        self.assert_failure("AGENTS.md")

    def test_incomplete_template_package_fails(self) -> None:
        (self.root / "specs/templates/change-package/tasks.md").unlink()
        self.assert_failure("must contain exactly nine files")

    def test_incomplete_voc_001_package_fails(self) -> None:
        (self.root / "specs/changes/VOC-001-repository-foundation/tasks.md").unlink()
        self.assert_failure("must contain exactly nine files")

    def test_incomplete_voc_002_package_fails(self) -> None:
        (self.root / "specs/changes/VOC-002-a003-governance-transition/tasks.md").unlink()
        self.assert_failure("must contain exactly nine files")

    def test_incomplete_voc_003_package_fails(self) -> None:
        (self.root / "specs/changes/VOC-003-a003-lifecycle-sync/tasks.md").unlink()
        self.assert_failure("must contain exactly nine files")

    def test_incomplete_voc_004_package_fails(self) -> None:
        (self.root / "specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/tasks.md").unlink()
        self.assert_failure("must contain exactly nine files")

    def test_voc_002_classifier_floor_is_r4(self) -> None:
        result = self.run_classifier_for_path(
            "specs/changes/VOC-002-a003-governance-transition/README.md", "R4"
        )
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assertIn("Detected path-based risk floor: R4", result.stdout)

    def test_voc_002_classifier_rejects_r3(self) -> None:
        result = self.run_classifier_for_path(
            "specs/changes/VOC-002-a003-governance-transition/README.md", "R3"
        )
        self.assertEqual(1, result.returncode, result.stdout + result.stderr)
        self.assertIn("below the detected floor R4", result.stderr)

    def test_voc_003_classifier_floor_is_r4(self) -> None:
        result = self.run_classifier_for_path(
            "specs/changes/VOC-003-a003-lifecycle-sync/README.md", "R4"
        )
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assertIn("Detected path-based risk floor: R4", result.stdout)

    def test_voc_004_classifier_floor_is_r4(self) -> None:
        result = self.run_classifier_for_path(
            "specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/README.md", "R4"
        )
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assertIn("Detected path-based risk floor: R4", result.stdout)

    def test_voc_004_classifier_rejects_r3(self) -> None:
        result = self.run_classifier_for_path(
            "docs/architecture/17-autonomous-development-architecture.md", "R3"
        )
        self.assertEqual(1, result.returncode, result.stdout + result.stderr)
        self.assertIn("below the detected floor R4", result.stderr)

    def test_a003_frozen_body_change_fails(self) -> None:
        self.replace(
            "docs/governance/amendments/A-003-governed-autonomous-engineering-authority.md",
            "AI performs the work",
            "AI sometimes performs the work",
        )
        self.assert_failure("frozen A-003 substantive body checksum mismatch")

    def test_a003_authority_rollback_without_governed_record_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "authority_model: a003-active",
            "authority_model: pre-a003",
        )
        self.assert_failure("active A-003 requires")

    def test_a003_missing_approved_pr_head_sha_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "approved_pr_head_sha: c858ebff3d97da88fea830bc32a74f69f59a9ad2",
            "approved_pr_head_sha: null",
        )
        self.assert_failure("full approved_pr_head_sha")

    def test_a003_missing_adopted_develop_sha_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "adopted_develop_sha: 9d5b4bc1d4a72e313b013047601265ee837c34f2",
            "adopted_develop_sha: null",
        )
        self.assert_failure("full adopted_develop_sha")

    def test_a003_conflated_revision_shas_fail(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "adopted_develop_sha: 9d5b4bc1d4a72e313b013047601265ee837c34f2",
            "adopted_develop_sha: c858ebff3d97da88fea830bc32a74f69f59a9ad2",
        )
        self.assert_failure("must be distinct records")

    def test_a003_missing_activation_evidence_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            'activation_evidence: "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005456622"',
            "activation_evidence: null",
        )
        self.assert_failure("exact activation_evidence")

    def test_a003_incomplete_post_merge_validation_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "post_merge_validation_status: passed",
            "post_merge_validation_status: incomplete",
        )
        self.assert_failure("post_merge_validation_status: passed")

    def test_a003_migration_approval_reuse_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "migration_approval_status: exhausted-non-reusable",
            "migration_approval_status: reusable",
        )
        self.assert_failure("exhausted-non-reusable")

    def test_a003_permanent_ehr_layer_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "exceptional_human_review_mode: exceptional-only",
            "exceptional_human_review_mode: permanent-routine-approval",
        )
        self.assert_failure("exceptional-only")

    def test_a003_historical_steward_falsification_fails(self) -> None:
        self.replace(
            "docs/governance/technical-steward-appointment.md",
            "Appointed qualified human technical steward: `@m-e-h-r-d-a-a-d`",
            "Appointed qualified human technical steward: `@someone-else`",
        )
        self.assert_failure("historical evidence marker")

    def test_a003_routine_r3_human_approval_marker_removal_fails(self) -> None:
        self.replace(
            "docs/governance/approval-matrix.md",
            "No standing technical-steward approval; no founder approval merely because work is R3",
            "Standing technical-steward and founder approval required for every R3",
        )
        self.assert_failure("missing A-003 authority marker")

    def test_a003_rl2_false_activation_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "rl2_technical_activation: false",
            "rl2_technical_activation: true",
        )
        self.assert_failure("rl2_technical_activation must remain false")

    def test_a003_rl1_false_activation_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "rl1_technical_activation: false",
            "rl1_technical_activation: true",
        )
        self.assert_failure("rl1_technical_activation must remain false")

    def test_a003_automatic_merge_enablement_fails(self) -> None:
        # 2026-08-08: the repository is now authorized (see the
        # AUTONOMOUS-RELEASE-AUTHORIZED-2026-08-08 marker in
        # a003-transition-state.yaml) and this field's required value is
        # "true", not "false" - the tripwire this test exercises is that
        # DEVIATING from the currently-required value (in either direction)
        # still fails closed, not that the field is frozen at one constant
        # forever regardless of authorization.
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "automatic_merge_allowed: true",
            "automatic_merge_allowed: false",
        )
        self.assert_failure("automatic_merge_allowed must equal 'true' once authorized")

    def test_a003_automatic_merge_enablement_without_marker_fails(self) -> None:
        # The other half of the same tripwire: the marker and the four
        # merge/release/deployment fields must move together. Flipping a
        # field to the authorized value while REMOVING the marker (as if
        # someone tried to sneak the capability in without the recorded
        # authorization) must fail just as hard as the reverse.
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "AUTONOMOUS-RELEASE-AUTHORIZED-2026-08-08",
            "MARKER-REMOVED-FOR-TEST",
        )
        self.assert_failure("automatic_merge_allowed must remain 'false' without an authorization marker")

    def test_a003_autonomous_merge_enablement_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "autonomous_merge_allowed: true",
            "autonomous_merge_allowed: false",
        )
        self.assert_failure("autonomous_merge_allowed must equal 'true' once authorized")

    def test_a003_autonomous_production_enablement_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "autonomous_production_release: enabled",
            "autonomous_production_release: disabled",
        )
        self.assert_failure("autonomous_production_release must equal 'enabled' once authorized")

    def test_a003_doc_17_adoption_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "doc_17_repository_adoption: true",
            "doc_17_repository_adoption: false",
        )
        self.assert_failure("doc_17_repository_adoption must be true")

    def test_a003_doc_18_adoption_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "doc_18_repository_adoption: true",
            "doc_18_repository_adoption: false",
        )
        self.assert_failure("doc_18_repository_adoption must be true")

    def test_doc_17_frozen_body_change_fails(self) -> None:
        self.replace(
            "docs/architecture/17-autonomous-development-architecture.md",
            "AI workers are replaceable.",
            "AI workers are permanent.",
        )
        self.assert_failure("frozen substantive body checksum mismatch")

    def test_doc_18_frozen_body_change_fails(self) -> None:
        self.replace(
            "docs/planning/18-autonomous-development-implementation-roadmap.md",
            "Production autonomy is not activated early.",
            "Production autonomy is activated early.",
        )
        self.assert_failure("frozen substantive body checksum mismatch")

    def test_doc_17_false_technical_activation_fails(self) -> None:
        self.replace(
            "docs/architecture/17-autonomous-development-architecture.md",
            "technical_activation_status: inactive",
            "technical_activation_status: active",
        )
        self.assert_failure("technical_activation_status: inactive")

    def test_doc_17_pre_merge_lifecycle_fails(self) -> None:
        self.replace(
            "docs/architecture/17-autonomous-development-architecture.md",
            "repository_adoption_status: adopted",
            "repository_adoption_status: candidate-pending-merge",
        )
        self.assert_failure("repository_adoption_status: adopted")

    def test_doc_18_missing_adopted_develop_sha_fails(self) -> None:
        self.replace(
            "docs/planning/18-autonomous-development-implementation-roadmap.md",
            "adopted_develop_sha: 2b5ecb19b532a9b23250e1255ff1e7fb9a78ef77",
            "adopted_develop_sha: null",
        )
        self.assert_failure("adopted_develop_sha: 2b5ecb19b532a9b23250e1255ff1e7fb9a78ef77")

    def test_voc_004_incomplete_lifecycle_sync_fails(self) -> None:
        self.replace(
            "specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/change.yaml",
            "canonical_lifecycle_sync_status: complete",
            "canonical_lifecycle_sync_status: pending",
        )
        self.assert_failure("canonical_lifecycle_sync_status must equal 'complete'")

    def test_control_plane_false_implementation_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "control_plane_implementation: false",
            "control_plane_implementation: true",
        )
        self.assert_failure("control_plane_implementation must remain false")

    def test_production_deployment_enablement_fails(self) -> None:
        self.replace(
            "docs/governance/a003-transition-state.yaml",
            "production_deployment: enabled",
            "production_deployment: disabled",
        )
        self.assert_failure("production_deployment must equal 'enabled' once authorized")

    def test_protected_policy_partial_adoption_fails(self) -> None:
        self.replace(
            ".github/approved-policy/protected-paths.yaml",
            "doc_18_repository_adoption: true",
            "doc_18_repository_adoption: false",
        )
        self.assert_failure("canonical protected policy requires doc_18_repository_adoption")

    def test_package_path_id_mismatch_fails(self) -> None:
        self.replace(
            "specs/changes/VOC-001-repository-foundation/change.yaml",
            "canonical_path: specs/changes/VOC-001-repository-foundation",
            "canonical_path: specs/changes/VOC-999-wrong",
        )
        self.assert_failure("canonical_path must equal")

    def test_duplicate_stable_identifier_fails(self) -> None:
        path = self.root / "specs/changes/VOC-001-repository-foundation/tasks.md"
        path.write_text(path.read_text(encoding="utf-8") + "\n## VOC-001-T01 — Duplicate\n", encoding="utf-8")
        self.assert_failure("duplicate stable identifier")

    def test_invalid_lifecycle_fails(self) -> None:
        self.replace(
            "specs/changes/VOC-001-repository-foundation/change.yaml",
            "status: implementing",
            "status: impossible",
        )
        self.assert_failure("status must equal")

    def test_invalid_risk_fails(self) -> None:
        self.replace(
            "specs/changes/VOC-001-repository-foundation/change.yaml",
            "risk: R4",
            "risk: R9",
        )
        self.assert_failure("risk must equal")

    def test_unsupported_yaml_construct_fails(self) -> None:
        path = self.root / "specs/changes/VOC-001-repository-foundation/change.yaml"
        path.write_text(path.read_text(encoding="utf-8") + "anchor: &unsafe value\n", encoding="utf-8")
        self.assert_failure("unsupported YAML construct")

    def test_duplicate_yaml_key_fails(self) -> None:
        path = self.root / "specs/changes/VOC-001-repository-foundation/change.yaml"
        path.write_text(path.read_text(encoding="utf-8") + "risk: R4\n", encoding="utf-8")
        self.assert_failure("duplicate YAML key 'risk'")

    def test_root_decisions_directory_fails(self) -> None:
        (self.root / "decisions").mkdir()
        self.assert_failure("root decision directory is prohibited")

    def test_uppercase_duplicate_pr_template_fails(self) -> None:
        (self.root / ".github/PULL_REQUEST_TEMPLATE.md").write_text("duplicate", encoding="utf-8")
        self.assert_failure("uppercase duplicate PR template")

    def test_missing_codeowners_protected_path_fails(self) -> None:
        self.replace(
            ".github/CODEOWNERS",
            "/tooling/governance/                       @m-e-h-r-d-a-a-d",
            "# removed tooling owner",
        )
        self.assert_failure("missing exact protected path owner")

    def test_invented_governance_team_fails(self) -> None:
        path = self.root / ".github/CODEOWNERS"
        path.write_text(path.read_text(encoding="utf-8") + "\n/example/ @KARSIFT/vocanova-governance\n", encoding="utf-8")
        self.assert_failure("invented or unverified governance team")

    def test_ai_or_bot_owner_fails(self) -> None:
        path = self.root / ".github/CODEOWNERS"
        path.write_text(path.read_text(encoding="utf-8") + "\n/example/ @claude-bot\n", encoding="utf-8")
        self.assert_failure("AI or bot identity")

    def test_workflow_write_permission_fails(self) -> None:
        self.replace(".github/workflows/repository-governance.yml", "contents: read", "contents: write")
        self.assert_failure("contents: read")

    def test_pull_request_target_fails(self) -> None:
        self.replace(".github/workflows/repository-governance.yml", "pull_request:", "pull_request_target:")
        self.assert_failure("pull_request_target")

    def test_path_filtered_workflow_fails(self) -> None:
        self.replace(
            ".github/workflows/repository-governance.yml",
            "  pull_request:\n    branches:",
            "  pull_request:\n    paths:\n      - docs/**\n    branches:",
        )
        self.assert_failure("paths:")

    def test_unpinned_external_action_fails(self) -> None:
        self.replace(
            ".github/workflows/repository-governance.yml",
            "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1",
            "actions/checkout@v7",
        )
        self.assert_failure("not pinned")

    def test_false_autonomous_activation_claim_fails(self) -> None:
        # Targets post-merge-activation-checklist.md, not
        # protected-paths.yaml: as of 2026-08-08 the latter is a legitimately
        # authorized exception to this pattern (see validate_false_activation's
        # own comment) - this test still needs to prove the tripwire is real
        # everywhere ELSE a false activation claim could appear.
        path = self.root / "docs/governance/post-merge-activation-checklist.md"
        path.write_text(path.read_text(encoding="utf-8") + "automatic_merge: true\n", encoding="utf-8")
        self.assert_failure("false claim")

    def test_non_r4_automatic_merge_false_fails(self) -> None:
        self.replace(
            "specs/changes/VOC-077-voc-072-t00-evidence-file-regressed-to-pending/change.yaml",
            "automatic_merge_allowed: true",
            "automatic_merge_allowed: false",
        )
        self.assert_failure("automatic_merge_allowed: false requires risk: R4")

    def test_historical_non_r4_automatic_merge_false_exemption_passes(self) -> None:
        # VOC-051 is grandfathered under VOC-075-T04 until a follow-up backfill
        # flips its automatic_merge_allowed field.
        text = (self.root / "specs/changes/VOC-051-add-hourly-sentry-based-log-error-monitoring/change.yaml").read_text(
            encoding="utf-8"
        )
        self.assertIn("automatic_merge_allowed: false", text)
        self.assertIn("risk: R3", text)
        result = self.run_validator()
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)

    def test_internal_error_returns_two_and_fails_closed(self) -> None:
        spec = importlib.util.spec_from_file_location("foundation_validator", VALIDATOR)
        self.assertIsNotNone(spec)
        self.assertIsNotNone(spec.loader)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        with mock.patch.object(module, "validate_repository", side_effect=RuntimeError("synthetic parser defect")):
            self.assertEqual(2, module.main(["--repository-root", str(self.root)]))


if __name__ == "__main__":
    unittest.main()
