from pathlib import Path
import os
import subprocess
import sys
import unittest


REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_INFRA_ROOT = REPOSITORY_ROOT / "tooling/governance/fixtures/karsift-ai-infra"


class ReadyForReviewReuseFixtureSuiteTests(unittest.TestCase):
    def test_pinned_shared_infra_policy_suite(self):
        env = os.environ.copy()
        env["PYTHONPATH"] = str(FIXTURE_INFRA_ROOT / "config")
        result = subprocess.run(
            [
                sys.executable,
                "-m",
                "unittest",
                "discover",
                "-s",
                str(FIXTURE_INFRA_ROOT / "tests"),
                "-p",
                "test_*.py",
                "-v",
            ],
            cwd=FIXTURE_INFRA_ROOT,
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(
            result.returncode,
            0,
            result.stderr or result.stdout or "pinned reuse policy suite failed",
        )

    def test_green_reused_evidence_cannot_start_auto_merge_while_draft(self):
        workflow = (
            FIXTURE_INFRA_ROOT / ".github/workflows/merge-gate.yml"
        ).read_text(encoding="utf-8")
        auto_merge = workflow.split("  auto-merge:", 1)[1]
        condition = auto_merge.split("    uses:", 1)[0]

        # Model the exact dangerous combination explicitly: checks and trusted
        # review are green, automatic merge is enabled, but the PR is a draft.
        # The workflow condition must retain a separate false-only draft gate.
        self.assertIn("inputs.auto_merge_enabled == 'true'", condition)
        self.assertIn("needs.report-status.outputs.checks_ok == 'true'", condition)
        self.assertIn("needs.report-status.outputs.verdict != 'FAIL'", condition)
        self.assertIn("needs.report-status.outputs.verdict != 'WAITING'", condition)
        self.assertIn("needs.report-status.outputs.verdict != 'PENDING'", condition)
        self.assertIn(
            "needs.report-status.outputs.is_draft == 'false'",
            condition,
        )
        self.assertNotIn(
            "needs.report-status.outputs.is_draft == 'true'",
            condition,
        )


if __name__ == "__main__":
    unittest.main()
