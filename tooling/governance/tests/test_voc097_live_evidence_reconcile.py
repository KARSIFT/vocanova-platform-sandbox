"""VOC-097 reconcile policy regressions (TEST-06 through TEST-14)."""

from __future__ import annotations

from datetime import datetime, timedelta, timezone
import json
from types import SimpleNamespace
import unittest
from unittest.mock import patch

from voc097_fixtures import (
    CALLER_PIPELINE,
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
        evidence = policy.qualify_run(
            parsed_contract(),
            run_fixture(),
            jobs_fixture(),
            pr_head_sha=HEAD,
            now=NOW,
        )
        serialized = json.loads(policy.evidence_json(evidence))
        self.assertEqual(serialized["state"], "qualified")
        self.assertEqual(serialized["job_ids"], [7001, 7002])
        self.assertNotIn("logs", serialized)
        self.assertNotIn("artifacts", serialized)
        self.assertNotIn("actor", serialized)
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
        self.assertIn("karsift-live-evidence-timeout", self.runner_source)
        self.assertIn("comment_exists", self.runner_source)

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


if __name__ == "__main__":
    unittest.main()
