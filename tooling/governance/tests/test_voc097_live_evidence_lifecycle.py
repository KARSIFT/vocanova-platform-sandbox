"""VOC-097 waiting lifecycle regressions (TEST-02 through TEST-05)."""

from __future__ import annotations

from pathlib import Path
import os
import subprocess
import tempfile
import textwrap
import unittest

from voc097_fixtures import (
    CALLER_PIPELINE,
    FIXTURE_INFRA_ROOT,
    load_policy_module,
    read_fixture,
)

classifier = load_policy_module(
    "voc097_classify_review_verdict",
    "config/classify-review-verdict.py",
)
decider = load_policy_module(
    "voc097_decide_remediation",
    "config/decide-remediation.py",
)
head_guard = load_policy_module(
    "voc097_verify_expected_head",
    "config/verify-expected-head.py",
)


class Voc097LiveEvidenceLifecycleTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.review_workflow = read_fixture(".github/workflows/review.yml")
        cls.remediate_workflow = read_fixture(".github/workflows/remediate.yml")
        cls.implement_workflow = read_fixture(".github/workflows/implement.yml")
        cls.merge_workflow = read_fixture(".github/workflows/merge-gate.yml")
        cls.review_prompt = read_fixture("prompts/review.md")
        cls.pipeline = CALLER_PIPELINE.read_text(encoding="utf-8")

    @staticmethod
    def _run_block(workflow: str, step_name: str) -> str:
        lines = workflow.splitlines()
        marker = f"- name: {step_name}"
        step_index = next(
            index for index, line in enumerate(lines) if line.strip() == marker
        )
        run_index = next(
            index
            for index in range(step_index + 1, len(lines))
            if lines[index].strip() == "run: |"
        )
        run_indent = len(lines[run_index]) - len(lines[run_index].lstrip())
        block = []
        for line in lines[run_index + 1 :]:
            if line.strip() and len(line) - len(line.lstrip()) <= run_indent:
                break
            block.append(line)
        return textwrap.dedent("\n".join(block))

    def _execute_missing_sha_path(self, workflow: str, step_name: str):
        script = self._run_block(workflow, step_name)
        script = script.replace("${{ inputs.pr_number }}", "1")
        script = script.replace("${{ github.event.pull_request.number }}", "1")
        script = script.replace("${{ inputs.expected_head_sha }}", "")
        script = script.replace("${{ inputs.expected_base_sha }}", "")
        script = script.replace("${{ inputs.reuse_outcome }}", "")
        script = script.replace("${{ inputs.reuse_prior_run_id }}", "")
        gh_stub = """
        gh() {
          if [ "$1 $2 $3" = "pr view 1" ]; then
            printf '%s\\n' '{"body":"Risk classification: R1","title":"fixture","author":{"login":"fixture"},"headRefOid":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","baseRefOid":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","isDraft":false}'
            return 0
          fi
          printf 'unexpected gh invocation: %s\\n' "$*" >&2
          return 97
        }
        """
        with tempfile.TemporaryDirectory() as scratch:
            script = script.replace("/tmp/pr.json", f"{scratch}/pr.json")
            script = script.replace("/tmp/pr.diff", f"{scratch}/pr.diff")
            with tempfile.NamedTemporaryFile(dir=scratch) as output:
                env = os.environ.copy()
                env["GITHUB_OUTPUT"] = output.name
                env["EXPECTED_HEAD_SHA"] = ""
                env["EXPECTED_BASE_SHA"] = ""
                completed = subprocess.run(
                    ["bash", "-c", textwrap.dedent(gh_stub) + script],
                    cwd=FIXTURE_INFRA_ROOT,
                    env=env,
                    text=True,
                    capture_output=True,
                    check=False,
                )
                output.seek(0)
                return completed, output.read().decode()

    def _execute_remediation_path(self, verdict: str, *, ci_failed: bool = False):
        head = "a" * 40
        base = "b" * 40
        package = "specs/changes/VOC-097-example"
        script = self._run_block(
            self.remediate_workflow,
            "Parse verdict, attempt, and package identity from the PR",
        )
        script = script.replace("${{ github.repository }}", "KARSIFT/example")
        script = script.replace("karsift-ai-infra/config/", "config/")
        gh_stub = f"""
        gh() {{
          if [ "$1 $2 $3" = "pr view 12" ]; then
            printf '%s\\n' '{{"body":"Implements task `VOC-097-T02` from `VOC-097` (`{package}`).\\n\\nCloses #34.\\n\\nPackage path: `{package}`\\n\\nImplemented by the implementer role (attempt 1 of 2).","headRefOid":"{head}","baseRefOid":"{base}"}}'
            return 0
          fi
          if [ "$1" = "api" ]; then
            printf '%s\\n' '[[{{"id":1,"created_at":"2026-08-21T00:00:00Z","user":{{"login":"karsift-ai-infra-bot[bot]","type":"Bot"}},"body":"**Independent verification - bound to commit `{head}`**\\ntask_id: `VOC-097-T02`\\npackage_path: `{package}`\\nauthority_issue: `34`\\nbase_sha: `{base}`\\nVERDICT: {verdict}"}}]]'
            return 0
          fi
          printf 'unexpected gh invocation: %s\\n' "$*" >&2
          return 97
        }}
        """
        with tempfile.TemporaryDirectory() as scratch:
            verdict_path = Path(scratch) / "review-verdict.md"
            script = script.replace("/tmp/review-verdict.md", str(verdict_path))
            output_path = Path(scratch) / "github-output"
            env = {
                **os.environ,
                "GITHUB_OUTPUT": str(output_path),
                "PR_NUMBER": "12",
                "GH_REPO": "KARSIFT/example",
                "CI_FAILED": str(ci_failed).lower(),
                "REVIEW_JOB_FAILED": "false",
                "EXPECTED_HEAD_SHA": head,
                "EXPECTED_BASE_SHA": base,
                "PACKAGE_PATH": package,
            }
            completed = subprocess.run(
                ["bash", "-c", textwrap.dedent(gh_stub) + script],
                cwd=FIXTURE_INFRA_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )
            return completed, output_path.read_text(encoding="utf-8")

    def test_voc097_test_02_waiting_marker_is_machine_detectable_and_fail_dominant(self):
        waiting = "VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE"
        failure = "VERDICT: FAIL"
        self.assertIn(waiting, self.review_prompt)
        self.assertIn("config/classify-review-verdict.py", self.remediate_workflow)
        self.assertIn('decision" = "WAITING"', self.remediate_workflow)
        self.assertIn(
            "waiting_for_operator_live_evidence=true",
            self.remediate_workflow,
        )
        self.assertEqual(classifier.classify(waiting), "WAITING")
        self.assertEqual(classifier.classify(failure), "FAIL")
        self.assertEqual(classifier.classify(f"{waiting}\n{failure}"), "FAIL")
        self.assertEqual(classifier.classify(f"{failure}\n{waiting}"), "FAIL")
        self.assertEqual(classifier.classify("no machine verdict"), "PENDING")

    def test_voc097_test_03_waiting_does_not_set_remediation_retry(self):
        self.assertEqual(
            decider.decide(
                expected_sha="a" * 40,
                current_sha="a" * 40,
                review_state="WAITING",
                ci_failed=False,
                review_job_failed=False,
            ),
            "WAITING",
        )
        waiting_guard = self.remediate_workflow.index('decision" = "WAITING"')
        retry_output = self.remediate_workflow.index('echo "should_retry=true"')
        self.assertLess(waiting_guard, retry_output)
        self.assertIn('echo "should_retry=false"', self.remediate_workflow)
        result, output = self._execute_remediation_path(
            "WAITING FOR OPERATOR LIVE EVIDENCE"
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("should_retry=false", output)
        self.assertIn("next_attempt=", output)
        self.assertIn("waiting_for_operator_live_evidence=true", output)
        self.assertNotIn("should_retry=true", output)

    def test_voc097_test_04_genuine_fail_and_ci_failure_still_retry(self):
        common = {"expected_sha": "a" * 40, "current_sha": "a" * 40}
        self.assertEqual(
            decider.decide(
                **common,
                review_state="FAIL",
                ci_failed=False,
                review_job_failed=False,
            ),
            "RETRY",
        )
        self.assertEqual(
            decider.decide(
                **common,
                review_state="WAITING",
                ci_failed=True,
                review_job_failed=False,
            ),
            "RETRY",
        )
        self.assertIn('decision" != "RETRY"', self.remediate_workflow)
        self.assertIn('echo "should_retry=true"', self.remediate_workflow)
        for verdict, ci_failed in (("FAIL", False), ("PASS", True)):
            with self.subTest(verdict=verdict, ci_failed=ci_failed):
                result, output = self._execute_remediation_path(
                    verdict,
                    ci_failed=ci_failed,
                )
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn("should_retry=true", output)
                self.assertIn("next_attempt=2", output)
                self.assertNotIn("waiting_for_operator_live_evidence=true", output)

    def test_voc097_test_05_implementer_has_no_general_actions_permission(self):
        permissions = self.implement_workflow.split("    permissions:\n", 1)[1].split(
            "    steps:\n", 1
        )[0]
        self.assertNotIn("actions:", permissions)
        reconcile = read_fixture(".github/workflows/live-evidence-reconcile.yml")
        operator_permissions = reconcile.split("    permissions:", 1)[1].split(
            "    steps:", 1
        )[0]
        self.assertIn("actions: write", operator_permissions)
        app_token = reconcile.split("      - name: Mint separate operator token", 1)[
            1
        ].split("      - name: Reconcile declared live evidence", 1)[0]
        self.assertNotIn("permission-actions:", app_token)
        self.assertIn("permission-pull-requests: write", app_token)

    def test_caller_binds_exact_head_and_cancels_superseded_runs(self):
        self.assertIn(
            "types: [opened, synchronize, reopened, ready_for_review, closed]",
            self.pipeline,
        )
        self.assertIn(
            "group: ${{ github.workflow }}-${{ github.event.pull_request.number || github.run_id }}",
            self.pipeline,
        )
        self.assertIn(
            "cancel-in-progress: ${{ github.event_name == 'pull_request' && github.event.action != 'closed' }}",
            self.pipeline,
        )
        self.assertEqual(
            self.pipeline.count(
                "expected_head_sha: ${{ github.event.pull_request.head.sha }}"
            ),
            5,
        )

    def test_stale_run_never_retries_even_when_failed(self):
        self.assertEqual(
            decider.decide(
                expected_sha="a" * 40,
                current_sha="b" * 40,
                review_state="FAIL",
                ci_failed=True,
                review_job_failed=True,
            ),
            "STALE",
        )

    def test_omitted_sha_is_transition_compatible_but_runtime_fail_closed(self):
        review_result, review_output = self._execute_missing_sha_path(
            self.review_workflow,
            "Fetch PR diff, metadata, and exact SHA",
        )
        self.assertEqual(review_result.returncode, 0, review_result.stderr)
        self.assertIn("stale=true", review_output)
        self.assertIn("Skipping reviewer model invocation", review_result.stdout)

        merge_result, merge_output = self._execute_missing_sha_path(
            self.merge_workflow,
            "Determine risk class, checks, and verification status",
        )
        self.assertEqual(merge_result.returncode, 0, merge_result.stderr)
        self.assertIn("checks_ok=false", merge_output)
        self.assertIn("verdict=PENDING", merge_output)


if __name__ == "__main__":
    unittest.main()
