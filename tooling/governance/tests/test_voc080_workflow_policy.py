from __future__ import annotations

from pathlib import Path
import re
import unittest


REPOSITORY_ROOT = Path(__file__).resolve().parents[3]

LIVE_DOC_PATHS = [
    "AGENTS.md",
    "CLAUDE.md",
    "docs/operations/15-ai-native-product-and-engineering-operating-model.md",
    "docs/governance/approval-matrix.md",
    "docs/governance/change-risk-classification.md",
    "docs/governance/repository-settings.md",
    "docs/governance/protected-areas.md",
    "docs/governance/post-merge-activation-checklist.md",
    "docs/governance/16-autonomous-development-operating-model.md",
    ".github/workflows/pipeline.yml",
    "specs/templates/change-package/change.yaml",
    "specs/templates/change-package/README.md",
]

LIVE_FOUNDER_GATE_PHRASES = [
    "Founder approval is required for",
    "Requires founder approval",
    "Founder approves develop",
    "Publication to production requires founder",
    "does not replace founder approval",
    "cannot reach `main` or production without founder",
    "reply `approved`",
    "A-003 remains effective until `VOC-080-T07`",
    "Under active A-003 until A-004 activation",
]


class Voc080WorkflowPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pipeline = (
            REPOSITORY_ROOT / ".github/workflows/pipeline.yml"
        ).read_text(encoding="utf-8")
        cls.agents = (REPOSITORY_ROOT / "AGENTS.md").read_text(encoding="utf-8")
        cls.deploy_production = (
            REPOSITORY_ROOT / ".github/workflows/deploy-production.yml"
        ).read_text(encoding="utf-8")
        cls.template_change_yaml = (
            REPOSITORY_ROOT / "specs/templates/change-package/change.yaml"
        ).read_text(encoding="utf-8")

    def test_pipeline_has_no_founder_username_wiring(self):
        self.assertNotIn("founder_username", self.pipeline)

    def test_pipeline_exposes_reconcile_dispatches(self):
        self.assertIn(
            "options: [implement, plan, reconcile, reconcile-release, reconcile-live-evidence, verify-auto-advance-live-evidence, verify-ready-for-review-reuse, recover-promotion-pr-checks, verify-promotion-check-recovery, verify-post-promotion-workflow, verify-remediate-operator-ownership]",
            self.pipeline,
        )
        self.assertIn("inputs.action == 'reconcile'", self.pipeline)
        self.assertIn("inputs.action == 'reconcile-release'", self.pipeline)
        self.assertIn("plan_pr_number:", self.pipeline)
        self.assertIn("release_issue_number:", self.pipeline)

    def test_pipeline_enables_autonomous_merge_for_this_repo(self):
        merge_gate = self.pipeline.split("  merge-gate:", 1)[1]
        self.assertIn('auto_merge_enabled: "true"', merge_gate)

    def test_pipeline_routes_plan_and_task_prs_to_distinct_review_jobs(self):
        self.assertIn("startsWith(github.head_ref, 'plan/')", self.pipeline)
        self.assertIn("startsWith(github.head_ref, 'agent/')", self.pipeline)
        self.assertIn("plan-review:", self.pipeline)
        self.assertIn("review:", self.pipeline)

    def test_agents_md_requires_automatic_merge_allowed_true_for_all_risk_classes(self):
        self.assertIn("**R0–R4:** set `automatic_merge_allowed: true`", self.agents)
        self.assertIn(
            "merge-gate no longer treats `false` as a founder-attention gate",
            self.agents,
        )

    def test_change_package_template_defaults_automatic_merge_allowed_true(self):
        self.assertIn("automatic_merge_allowed: true", self.template_change_yaml)
        self.assertIn("including R4", self.template_change_yaml)

    def test_deploy_production_is_push_driven_without_founder_approval_job(self):
        self.assertIn("on:\n  push:\n    branches: [main]", self.deploy_production)
        self.assertNotIn("issue_comment:", self.deploy_production)
        self.assertNotIn("founder_username", self.deploy_production)

    def test_pipeline_remediate_has_no_founder_override_inputs(self):
        remediate = self.pipeline.split("  remediate:", 1)[1].split(
            "  merge-gate:", 1
        )[0]
        self.assertIn("remediate.yml@main", remediate)
        self.assertNotIn("founder_username", remediate)
        self.assertNotIn("approved", remediate)

    def test_live_docs_do_not_claim_founder_comment_engineering_gates(self):
        for relative in LIVE_DOC_PATHS:
            text = (REPOSITORY_ROOT / relative).read_text(encoding="utf-8")
            for phrase in LIVE_FOUNDER_GATE_PHRASES:
                self.assertNotIn(
                    phrase,
                    text,
                    msg=f"{relative} still contains live founder-gate phrase: {phrase}",
                )

    def test_claude_md_describes_a004_active_without_standing_founder_merge_gates(self):
        claude = (REPOSITORY_ROOT / "CLAUDE.md").read_text(encoding="utf-8")
        self.assertIn("A-004 is the effective authority model", claude)
        self.assertIn("No autonomous engineering workflow waits on a founder", claude)
        self.assertNotIn("A-003 remains effective until A-004 activation", claude)
        self.assertNotRegex(
            claude,
            re.compile(
                r"R4.*requires founder approval.*merge",
                re.IGNORECASE | re.DOTALL,
            ),
        )

    def test_a004_activation_does_not_fabricate_or_require_approval(self):
        state = (
            REPOSITORY_ROOT / "docs/governance/a004-transition-state.yaml"
        ).read_text(encoding="utf-8")
        amendment = (
            REPOSITORY_ROOT
            / "docs/governance/amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md"
        ).read_text(encoding="utf-8")
        for text in (state, amendment):
            self.assertIn("not-required-explicitly-revoked", text)
            self.assertNotIn("2026-08-15T08:30:00Z", text)
            self.assertNotIn("approved-exact-revision-github-evidence", text)
        self.assertIn("required-external-exact-revision-pass", state)
        self.assertIn("issuecomment-5301333790", state)


if __name__ == "__main__":
    unittest.main()
