"""VOC-127 caller fixture regressions for exact-merge-SHA develop sync."""

from __future__ import annotations

import re
import unittest
from pathlib import Path

from voc080_fixtures import read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_PIPELINE = (REPO_ROOT / ".github/workflows/pipeline.yml").read_text(
    encoding="utf-8"
)
INFRA_PIN = "a9df74a63976d5239b84151fd01310835c999e7c"
DRAFTING_PIN = "60afda3a44fd06b8c00b219771de7112f1aded6e"
MAX_DISPATCH_INPUTS = 25


def count_workflow_dispatch_inputs(text: str) -> int:
    in_inputs = False
    count = 0
    for line in text.splitlines():
        if re.match(r"^  workflow_dispatch:\s*$", line):
            in_inputs = False
            continue
        if re.match(r"^    inputs:\s*$", line):
            in_inputs = True
            continue
        if in_inputs and re.match(r"^    [A-Za-z]", line) and not line.startswith("      "):
            in_inputs = False
        if in_inputs and re.match(r"^      (\w+):", line):
            count += 1
    return count


class Voc127ImplementPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.release = read_fixture(".github/workflows/release.yml")
        cls.reconcile = read_fixture(
            ".github/workflows/reconcile-production-change.yml"
        )
        cls.branch_sync = read_fixture("config/branch_sync.py")
        cls.runner = read_fixture("config/branch-sync-runner.py")
        cls.template = read_fixture(
            "templates/project-repo/.github/workflows/pipeline.yml"
        )
        cls.readme = read_fixture("README.md")
        cls.pin = read_fixture("PINNED_SHA.txt").strip()

    def test_fixture_pin_matches_reviewed_infrastructure_merge(self):
        self.assertEqual(self.pin, INFRA_PIN)
        self.assertNotEqual(self.pin, DRAFTING_PIN)
        self.assertIn(INFRA_PIN, self.readme)

    def test_release_converges_integration_to_exact_merge_before_close(self):
        merge = self.release.index("Perform the single exact-head merge decision")
        sync = self.release.index(
            "Synchronize integration to the exact promotion merge", merge
        )
        close = self.release.index(
            "Close the release audit after exact branch convergence", sync
        )
        self.assertLess(merge, sync)
        self.assertLess(sync, close)
        self.assertIn("branch-sync-runner.py", self.release[sync:close])
        self.assertNotIn('-f sha="$CHECKED_HEAD_SHA"', self.release)
        self.assertEqual(self.release.count("gh pr merge"), 1)

    def test_release_does_not_treat_ahead_by_zero_with_unequal_shas_as_promoted(self):
        promotion = self.release.split("Open or reuse the single promotion PR", 1)[1]
        self.assertIn('!= "$production_sha"', promotion)
        self.assertIn('!= "$integration_sha"', promotion)
        self.assertIn("sync_needed=true", promotion)
        self.assertNotIn('Already promoted" without', promotion)

    def test_governed_main_only_reconcile_workflow_is_exact_sha_bound(self):
        self.assertIn("--mode governed-main-only", self.reconcile)
        self.assertIn("branch-sync-runner.py", self.reconcile)
        self.assertNotIn("workflow_dispatch:", self.reconcile)

    def test_caller_pipeline_exposes_exceptional_main_only_reconciliation(self):
        self.assertIn("reconcile-production-change:", CALLER_PIPELINE)
        self.assertIn(
            "reconcile-production-change.yml@main",
            CALLER_PIPELINE,
        )
        self.assertIn("needs: [merge-gate, reconcile-production-change]", CALLER_PIPELINE)
        self.assertNotIn("reconcile-production-change:", (
            REPO_ROOT / ".github/workflows/pipeline-verify.yml"
        ).read_text(encoding="utf-8"))
        dispatch = CALLER_PIPELINE.split("workflow_dispatch:", 1)[1].split(
            "\n# Read-only verifier", 1
        )[0]
        self.assertIn("reconcile-production-change", dispatch)
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("expected_base_sha:", dispatch)

    def test_template_and_caller_dispatch_blocks_stay_within_github_limit(self):
        for text in (CALLER_PIPELINE, self.template):
            count = count_workflow_dispatch_inputs(text)
            self.assertLessEqual(count, MAX_DISPATCH_INPUTS)

    def test_branch_sync_helper_refuses_unique_integration_commits(self):
        self.assertIn("integration_has_unique_commits", self.branch_sync)
        self.assertIn("production_ref_moved", self.branch_sync)
        self.assertIn("promotion_merge_parents_mismatch", self.branch_sync)

    def test_branch_sync_runner_uses_lease_and_tree_equivalence(self):
        self.assertIn('f"--force-with-lease={lease}"', self.runner)
        self.assertIn("tree_equivalent", self.runner)
        self.assertIn("branch_sync_postcondition_failed", self.runner)


if __name__ == "__main__":
    unittest.main()
