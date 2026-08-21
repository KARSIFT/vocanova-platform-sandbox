"""VOC-097 reconcile policy regressions (TEST-06 through TEST-14)."""

from __future__ import annotations

from datetime import datetime, timedelta, timezone
import json
import os
import subprocess
import tempfile
import textwrap
from types import SimpleNamespace
import unittest
from unittest.mock import patch

from voc097_fixtures import (
    CALLER_PIPELINE,
    FIXTURE_INFRA_ROOT,
    load_policy_module,
    read_fixture,
)

policy = load_policy_module(
    "live_evidence_reconcile",
    "config/live_evidence_reconcile.py",
)
runner = load_policy_module(
    "live_evidence_reconcile_runner",
    "config/live-evidence-reconcile-runner.py",
)
normalizer = load_policy_module(
    "normalize_review_narrative",
    "config/normalize-review-narrative.py",
)

NOW = datetime(2026, 8, 21, 0, 0, tzinfo=timezone.utc)
HEAD = "a" * 40
RUN_SHA = "b" * 40
BASE = "c" * 40
PACKAGE = "specs/changes/VOC-097-example"


def contract_text(**overrides):
    values = {
        "workflow_file": "deploy-production.yml",
        "job_names": "  - deploy-production\n  - verify-production",
        "events": "  - push",
        "branch": "main",
        "lineage": "  mode: exact_pr_head",
        "max_age": "72h",
        "dispatch": "",
    }
    values.update(overrides)
    return f"""schema_version: 1
task_id: VOC-097-T02
ownership: operator
workflow_file: {values['workflow_file']}
job_names:
{values['job_names']}
events:
{values['events']}
branch: {values['branch']}
sha_lineage:
{values['lineage']}
conclusion: success
max_age: {values['max_age']}
{values['dispatch']}"""


def parsed_contract(**overrides):
    return policy.validate_contract(
        policy.parse_contract_yaml(contract_text(**overrides)),
        "VOC-097-T02",
    )


def run_fixture(**overrides):
    run = {
        "id": 12345,
        "workflow_id": 91,
        "name": "deploy-production",
        "path": ".github/workflows/deploy-production.yml",
        "event": "push",
        "head_branch": "main",
        "head_sha": HEAD,
        "conclusion": "success",
        "run_started_at": "2026-08-20T23:55:00Z",
        "updated_at": "2026-08-20T23:59:00Z",
    }
    run.update(overrides)
    return run


def jobs_fixture():
    return [
        {"id": 7001, "name": "deploy-production", "conclusion": "success"},
        {"id": 7002, "name": "verify-production", "conclusion": "success"},
    ]


class Voc097LiveEvidenceReconcileTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.workflow = read_fixture(".github/workflows/live-evidence-reconcile.yml")
        cls.runner_source = read_fixture("config/live-evidence-reconcile-runner.py")
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.review = read_fixture(".github/workflows/review.yml")
        cls.remediate = read_fixture(".github/workflows/remediate.yml")
        cls.pipeline = CALLER_PIPELINE.read_text(encoding="utf-8")

    def assert_rejected(self, contract, run=None, jobs=None, **kwargs):
        with self.assertRaises(policy.ContractError):
            policy.qualify_run(
                contract,
                run or run_fixture(),
                jobs or jobs_fixture(),
                pr_head_sha=HEAD,
                now=NOW,
                **kwargs,
            )

    def test_voc097_test_06_wrong_workflow_identity_is_rejected(self):
        contract = parsed_contract()
        self.assert_rejected(contract, run_fixture(path=".github/workflows/other.yml"))

    def test_voc097_test_07_wrong_or_missing_required_job_is_rejected(self):
        contract = parsed_contract()
        self.assert_rejected(
            contract,
            jobs=[{"id": 7001, "name": "deploy-production", "conclusion": "success"}],
        )
        failed = jobs_fixture()
        failed[1]["conclusion"] = "failure"
        self.assert_rejected(contract, jobs=failed)

    def test_voc097_test_08_wrong_event_branch_and_sha_lineage_are_rejected(self):
        contract = parsed_contract()
        self.assert_rejected(contract, run_fixture(event="schedule"))
        self.assert_rejected(contract, run_fixture(head_branch="develop"))
        self.assert_rejected(contract, run_fixture(head_sha=RUN_SHA))

        integration = parsed_contract(lineage="  mode: integration_contains_pr_head")
        self.assert_rejected(
            integration,
            run_fixture(head_sha=RUN_SHA),
            integration_contains_pr=True,
            integration_contains_run=False,
        )

    def test_voc097_test_09_qualifying_output_contains_allowlisted_metadata_only(self):
        forbidden = "forbidden-sensitive-fixture"
        evidence = policy.qualify_run(
            parsed_contract(),
            run_fixture(
                logs=forbidden,
                artifacts=forbidden,
                token=forbidden,
                cookie=forbidden,
                credential=forbidden,
                actor={"login": forbidden},
            ),
            [
                {**job, "output": forbidden, "user": forbidden}
                for job in jobs_fixture()
            ],
            pr_head_sha=HEAD,
            now=NOW,
        )
        evidence_json = policy.evidence_json(evidence)
        serialized = json.loads(evidence_json)
        self.assertEqual(serialized["state"], "qualified")
        self.assertEqual(serialized["job_ids"], [7001, 7002])
        self.assertNotIn(forbidden, evidence_json)

        class ReadApi:
            repository = "KARSIFT/example"

            def get(self, endpoint):
                if endpoint != "repos/KARSIFT/example/pulls/12":
                    raise AssertionError(f"unexpected read endpoint: {endpoint}")
                return {
                    "state": "open",
                    "merged_at": None,
                    "head": {
                        "sha": HEAD,
                        "ref": "agent/voc-097-t02",
                        "repo": {"full_name": self.repository},
                    },
                    "base": {"sha": BASE},
                }

        class WriteApi:
            repository = "KARSIFT/example"
            body = None

            def mutate(self, method, endpoint, payload):
                if method != "POST" or endpoint != "repos/KARSIFT/example/issues/12/comments":
                    raise AssertionError(f"unexpected write: {method} {endpoint}")
                self.body = payload["body"]
                return {"id": 1}

        task = SimpleNamespace(
            task_id="VOC-097-T02",
            pr_number=12,
            head_sha=HEAD,
            head_ref="agent/voc-097-t02",
            base_sha=BASE,
        )
        write_api = WriteApi()
        runner.post_qualified_comment(
            ReadApi(),
            write_api,
            task,
            {**evidence, "logs": forbidden, "token": forbidden},
            "d" * 40,
        )
        self.assertIsNotNone(write_api.body)
        self.assertNotIn(forbidden, write_api.body)
        with self.assertRaises(policy.ContractError):
            policy.evidence_json({**evidence, "arbitrary_output": "forbidden"})

    def test_voc097_test_10_workflow_never_calls_log_or_artifact_apis(self):
        combined = self.workflow + self.runner_source
        self.assertNotIn("/logs", combined)
        self.assertNotIn("/artifacts", combined)
        self.assertNotIn("download-artifact", combined)
        self.assertNotIn("steps_url", combined)

    def test_voc097_test_11_qualification_is_one_commit_then_fresh_pr_review(self):
        self.assertIn("append_result_commit", self.runner_source)
        self.assertIn("result_already_present", self.runner_source)
        self.assertIn("fresh exact-SHA independent review", self.runner_source)
        self.assertIn("pull_request:", self.pipeline)
        self.assertEqual(
            self.pipeline.count(
                "expected_head_sha: ${{ github.event.pull_request.head.sha }}"
            ),
            4,
        )

        task = SimpleNamespace(task_id="VOC-097-T02", pr_number=12)
        evidence = {"run_id": 12345}
        events = []

        def record(name, result=None):
            def callback(*_args, **_kwargs):
                events.append(name)
                return result

            return callback

        with (
            patch.object(
                runner,
                "result_already_present",
                side_effect=[False, True],
            ),
            patch.object(runner, "candidate_runs", return_value=[{"id": 12345}]),
            patch.object(runner, "qualify", side_effect=record("qualify", evidence)),
            patch.object(
                runner,
                "append_result_commit",
                side_effect=record("append", "d" * 40),
            ),
            patch.object(
                runner,
                "post_qualified_comment",
                side_effect=record("comment"),
            ),
            patch.object(
                runner,
                "advance_result_ref",
                side_effect=record("advance"),
            ),
        ):
            self.assertTrue(
                runner.reconcile_task(object(), object(), task, NOW, None)
            )
            self.assertFalse(
                runner.reconcile_task(object(), object(), task, NOW, None)
            )
        self.assertEqual(events, ["qualify", "append", "comment", "advance"])

        narrative = normalizer.normalize_narrative(
            b"""Reviewer preamble.
**Independent verification - bound to commit `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`**
task_id: `VOC-097-T02`
package_path: `specs/changes/VOC-097-example`
authority_issue: `12`
base_sha: `cccccccccccccccccccccccccccccccccccccccc`
No blocking findings.
VERDICT: PASS
"""
        )
        self.assertNotIn("bound to commit", narrative)
        self.assertEqual(narrative.splitlines()[-1], "VERDICT: PASS")
        for workflow in (self.review, read_fixture(".github/workflows/plan-review.yml")):
            self.assertIn("normalize-review-narrative.py", workflow)
            self.assertIn("extract-cursor-result.py", workflow)
            self.assertLess(
                workflow.index("normalize-review-narrative.py"),
                workflow.index("Build verification record for isolated publisher"),
            )
        self.assertTrue(
            (FIXTURE_INFRA_ROOT / "config/extract-cursor-result.py").is_file(),
            "every review helper referenced by the pinned workflows must be vendored",
        )

        merge_gate = read_fixture(".github/workflows/merge-gate.yml")
        script = self._workflow_run_block(
            merge_gate,
            "Determine risk class, checks, and verification status",
        )
        script = script.replace("${{ github.event.pull_request.number }}", "12")
        script = script.replace("${{ inputs.expected_head_sha }}", HEAD)
        script = script.replace("${{ inputs.expected_base_sha }}", BASE)
        script = script.replace("${{ github.repository }}", "KARSIFT/example")
        script = script.replace("karsift-ai-infra/config/", "config/")
        old_head = "e" * 40
        gh_stub = f"""
        gh() {{
          if [ "$1 $2 $3" = "pr view 12" ]; then
            printf '%s\\n' '{{"body":"Risk classification: R4\\nImplements task `VOC-097-T02`\\nPackage path: `specs/changes/VOC-097-example`\\nCloses #34","author":{{"login":"fixture"}},"headRefName":"agent/voc-097-t02","headRefOid":"{HEAD}","baseRefOid":"{BASE}","isDraft":false}}'
            return 0
          fi
          if [ "$1 $2 $3" = "pr checks 12" ]; then
            printf '%s\\n' '[{{"name":"ci / ci","state":"SUCCESS"}},{{"name":"review / publish-review","state":"SUCCESS"}}]'
            return 0
          fi
          if [ "$1" = "api" ]; then
            printf '%s\\n' '[[{{"id":1,"created_at":"2026-08-21T00:00:00Z","user":{{"login":"karsift-ai-infra-bot[bot]","type":"Bot"}},"body":"**Independent verification - bound to commit `{old_head}`**\\ntask_id: `VOC-097-T02`\\npackage_path: `specs/changes/VOC-097-example`\\nauthority_issue: `34`\\nbase_sha: `{BASE}`\\nVERDICT: PASS"}}]]'
            return 0
          fi
          printf 'unexpected gh invocation: %s\\n' "$*" >&2
          return 97
        }}
        """
        with tempfile.TemporaryDirectory() as scratch:
            checks_path = os.path.join(scratch, "pr-checks.json")
            verdict_path = os.path.join(scratch, "review-verdict.md")
            script = script.replace("/tmp/pr-checks.json", checks_path)
            script = script.replace("/tmp/review-verdict.md", verdict_path)
            output_path = os.path.join(scratch, "github-output")
            env = {**os.environ, "GITHUB_OUTPUT": output_path}
            result = subprocess.run(
                ["bash", "-c", textwrap.dedent(gh_stub) + script],
                cwd=FIXTURE_INFRA_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            with open(output_path, encoding="utf-8") as output_file:
                outputs = output_file.read()
            self.assertIn("checks_ok=true", outputs)
            self.assertIn("verdict=PENDING", outputs)

    def test_voc097_test_12_stale_and_non_success_runs_are_rejected(self):
        contract = parsed_contract(max_age="1h")
        self.assert_rejected(
            contract,
            run_fixture(updated_at="2026-08-20T22:00:00Z"),
        )
        self.assert_rejected(contract, run_fixture(conclusion="failure"))
        self.assert_rejected(
            parsed_contract(),
            run_fixture(updated_at="2026-08-20T23:59:00Z"),
            completed_by=datetime(2026, 8, 20, 23, 58, tzinfo=timezone.utc),
        )

    def test_voc097_test_13_timeout_is_bounded_and_marker_is_single_use(self):
        self.assertFalse(policy.timed_out(NOW - timedelta(hours=71), NOW))
        self.assertTrue(policy.timed_out(NOW - timedelta(hours=72), NOW))

        comments = []
        mutations = []
        task = SimpleNamespace(
            task_id="VOC-097-T02",
            issue_number=34,
            pr_number=12,
            head_sha=HEAD,
            waiting_since=NOW - timedelta(hours=72),
        )

        class ReadApi:
            repository = "KARSIFT/example"

            def get_all(self, endpoint, key=None):
                self.assert_comments_endpoint(endpoint)
                return comments

            @staticmethod
            def assert_comments_endpoint(endpoint):
                if endpoint != "repos/KARSIFT/example/issues/34/comments":
                    raise AssertionError(f"unexpected read endpoint: {endpoint}")

        class WriteApi:
            repository = "KARSIFT/example"

            def mutate(self, method, endpoint, payload):
                mutations.append((method, endpoint, payload))
                comments.append({
                    "user": {
                        "login": "karsift-ai-infra-bot[bot]",
                        "type": "Bot",
                    },
                    "body": payload["body"],
                })
                return {"id": 1}

        with (
            patch.object(runner, "result_already_present", return_value=False),
            patch.object(runner, "candidate_runs", return_value=[]),
            patch.object(runner, "append_result_commit") as append_result,
            patch.object(runner, "post_qualified_comment") as qualified_comment,
            patch.object(runner, "advance_result_ref") as advance_result,
            patch.object(runner, "dispatch_once") as dispatch,
        ):
            self.assertFalse(
                runner.reconcile_task(ReadApi(), WriteApi(), task, NOW, None)
            )
            self.assertFalse(
                runner.reconcile_task(
                    ReadApi(),
                    WriteApi(),
                    task,
                    NOW + timedelta(hours=1),
                    None,
                )
            )

        self.assertEqual(len(mutations), 1)
        method, endpoint, payload = mutations[0]
        self.assertEqual(method, "POST")
        self.assertEqual(endpoint, "repos/KARSIFT/example/issues/34/comments")
        self.assertIn("karsift-live-evidence-timeout", payload["body"])
        append_result.assert_not_called()
        qualified_comment.assert_not_called()
        advance_result.assert_not_called()
        dispatch.assert_not_called()

    def test_voc097_test_14_duplicate_result_short_circuits_reconciliation(self):
        task = SimpleNamespace(result_path="result.json", head_sha=HEAD, base_sha=BASE)
        task.pr_number = 12

        class AttestedApi:
            repository = "KARSIFT/example"

            def get_all(self, endpoint, key=None):
                return [{
                    "user": {"login": "karsift-ai-infra-bot[bot]", "type": "Bot"},
                    "body": (
                        "**Live-evidence reconcile — qualified**\n"
                        f"result_head_sha: `{HEAD}`\n"
                        f"base_sha: `{BASE}`"
                    ),
                }]

        with patch.object(
            runner,
            "read_repository_file",
            return_value='{"schema_version":1,"state":"qualified","run_id":12345}',
        ):
            self.assertTrue(runner.result_already_present(AttestedApi(), task))
        self.assertIn(
            "group: live-evidence-reconcile-${{ github.repository }}",
            self.workflow,
        )
        self.assertIn("cancel-in-progress: false", self.workflow)

    def test_caller_polls_without_workflow_run_recursion(self):
        self.assertIn('cron: "0 * * * *"', self.pipeline)
        self.assertNotIn("  workflow_run:", self.pipeline)
        self.assertIn("reconcile-live-evidence", self.pipeline)
        self.assertIn("live_evidence_run_id", self.pipeline)

    def test_live_evidence_stays_separate_from_operational_failure_observer(self):
        observer = (
            CALLER_PIPELINE.parent / "operational-failure-monitoring.yml"
        ).read_text(encoding="utf-8")
        combined = self.workflow + self.runner_source
        self.assertNotIn("operational-failure-monitoring", combined)
        self.assertNotIn("open-failure-issue.sh", combined)
        self.assertNotIn("karsift-live-evidence", observer)

    @staticmethod
    def _workflow_run_block(workflow, step_name):
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


if __name__ == "__main__":
    unittest.main()
