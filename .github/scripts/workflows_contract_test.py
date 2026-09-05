#!/usr/bin/env python3
"""Contract tests for .github/workflows.

One consolidated test instead of kandev's per-workflow contract files: it
encodes the invariants that have actually cost this repo time (merge-queue
drift, required checks that never report, auto-deploy creeping back onto
production, queue entries ejected by a cancelled run). Run:

    python3 .github/scripts/workflows_contract_test.py
"""

import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path

import yaml

WF_DIR = Path(__file__).resolve().parents[1] / "workflows"
MERGE_GROUP_GUARD = "${{ github.event_name != 'merge_group' }}"

# Workflows deliberately exempt from "must also run on merge_group": advisory
# gates and event-scoped automation that are not required status checks.
NON_REQUIRED = {
    "accessibility.yml",
    "auto-merge.yml",
    "claude-code-review.yml",
    "claude-review.yml",
    "ci-base-image.yml",
    "deploy-production.yml",
    "deploy-staging.yml",
    "docker-smoke.yml",
    "error-monitoring.yml",
    "lighthouse.yml",
    "merge-queue-watchdog.yml",
    "operational-failure-monitoring.yml",
    "pr-title.yml",
    "pr-walkthrough.yml",
    "scheduled-synthetics.yml",
    "sync-monitoring.yml",
}


def load(path: Path) -> dict:
    data = yaml.safe_load(path.read_text())
    # PyYAML parses the bare `on:` key as the boolean True.
    if True in data and "on" not in data:
        data["on"] = data.pop(True)
    return data


def triggers(wf: dict) -> dict:
    on = wf.get("on", {})
    if isinstance(on, str):
        return {on: None}
    if isinstance(on, list):
        return {k: None for k in on}
    return on


def real_jobs(wf: dict) -> dict:
    return {name: j for name, j in wf.get("jobs", {}).items() if isinstance(j, dict) and "steps" in j}


WORKFLOWS = {p.name: load(p) for p in sorted(WF_DIR.glob("*.yml"))}


def auto_merge_shell() -> str:
    """Return the production inline shell, rather than a copied test fixture."""
    steps = real_jobs(WORKFLOWS["auto-merge.yml"])["auto-merge"]["steps"]
    return next(step["run"] for step in steps if step.get("name") == "Enable or disable auto-merge, and enqueue, per target PR")


class AutoMergeWorkflowShellTest(unittest.TestCase):
    """Exercise the checked-in workflow shell with a deterministic fake gh."""

    def run_shell(self, info: dict, *, queued: bool, fail_dequeue: bool = False, race_dequeued: bool = False):
        with tempfile.TemporaryDirectory() as directory:
            temp = Path(directory)
            fake_gh = temp / "gh"
            log = temp / "gh.log"
            fake_gh.write_text(
                """#!/usr/bin/env python3
import os
import sys

args = sys.argv[1:]
joined = " ".join(args)
log = os.environ["GH_LOG"]
with open(log, "a") as output:
    if args[:2] == ["pr", "view"]:
        output.write("VIEW\\n")
        print(os.environ["PR_INFO"])
    elif args[:2] == ["pr", "merge"]:
        output.write("MERGE " + " ".join(args[2:]) + "\\n")
    elif "dequeuePullRequest" in joined:
        output.write("DEQUEUE\\n")
        if os.environ.get("FAIL_DEQUEUE") == "1":
            print("simulated dequeue failure", file=sys.stderr)
            sys.exit(1)
    elif "enqueuePullRequest" in joined:
        output.write("ENQUEUE\\n")
    elif "mergeQueueEntry" in joined:
        count_path = os.environ["GH_QUERY_COUNT"]
        count = int(open(count_path).read()) if os.path.exists(count_path) else 0
        with open(count_path, "w") as count_file:
            count_file.write(str(count + 1))
        output.write("QUEUE_QUERY\\n")
        value = os.environ["QUEUED"] == "1"
        if os.environ.get("RACE_DEQUEUED") == "1" and count > 0:
            value = False
        print(str(value).lower())
    else:
        print("unexpected gh invocation: " + joined, file=sys.stderr)
        sys.exit(2)
"""
            )
            fake_gh.chmod(0o755)
            environment = {
                **os.environ,
                "PATH": f"{temp}:{os.environ['PATH']}",
                "NUMBERS": "[42]",
                "GH_TOKEN": "test-token",
                "GH_REPO": "KARSIFT/vocanova-platform-sandbox",
                "PR_INFO": json.dumps(info),
                "GH_LOG": str(log),
                "GH_QUERY_COUNT": str(temp / "query-count"),
                "QUEUED": "1" if queued else "0",
                "FAIL_DEQUEUE": "1" if fail_dequeue else "0",
                "RACE_DEQUEUED": "1" if race_dequeued else "0",
            }
            result = subprocess.run(
                ["bash", "-c", auto_merge_shell()],
                cwd=WF_DIR.parents[1],
                env=environment,
                text=True,
                capture_output=True,
            )
            return result, log.read_text().splitlines() if log.exists() else []

    @staticmethod
    def pr_info(*, draft: bool = False, held: bool = False, auto_merge: bool = True) -> dict:
        return {
            "isDraft": draft,
            "labels": [{"name": "hold"}] if held else [],
            "id": "PR_node_42",
            "state": "OPEN",
            "autoMergeRequest": {"enabledAt": "2026-09-05T00:00:00Z"} if auto_merge else None,
        }

    def test_draft_disables_auto_merge_and_dequeues(self) -> None:
        result, log = self.run_shell(self.pr_info(draft=True), queued=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(log, ["VIEW", "MERGE 42 --disable-auto", "QUEUE_QUERY", "DEQUEUE"])

    def test_hold_disables_auto_merge_and_dequeues(self) -> None:
        result, log = self.run_shell(self.pr_info(held=True), queued=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(log, ["VIEW", "MERGE 42 --disable-auto", "QUEUE_QUERY", "DEQUEUE"])

    def test_absent_queue_entry_is_not_dequeued(self) -> None:
        result, log = self.run_shell(self.pr_info(draft=True, auto_merge=False), queued=False)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(log, ["VIEW", "QUEUE_QUERY"])

    def test_dequeue_failure_is_visible_when_entry_remains(self) -> None:
        result, log = self.run_shell(self.pr_info(draft=True), queued=True, fail_dequeue=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("remains in the merge queue", result.stderr)
        self.assertEqual(log, ["VIEW", "MERGE 42 --disable-auto", "QUEUE_QUERY", "DEQUEUE", "QUEUE_QUERY"])

    def test_concurrent_dequeue_is_safe(self) -> None:
        result, log = self.run_shell(
            self.pr_info(draft=True), queued=True, fail_dequeue=True, race_dequeued=True
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("already removed", result.stderr)
        self.assertEqual(log, ["VIEW", "MERGE 42 --disable-auto", "QUEUE_QUERY", "DEQUEUE", "QUEUE_QUERY"])

    def test_ready_unheld_pr_still_enables_and_enqueues(self) -> None:
        result, log = self.run_shell(self.pr_info(auto_merge=False), queued=False)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(log, ["VIEW", "MERGE 42 --auto", "ENQUEUE"])


class WorkflowContractTest(unittest.TestCase):
    def test_all_workflows_parse(self) -> None:
        self.assertTrue(WORKFLOWS)
        for name, wf in WORKFLOWS.items():
            self.assertIn("jobs", wf, f"{name}: no jobs")

    def test_every_job_has_a_timeout(self) -> None:
        for name, wf in WORKFLOWS.items():
            for job_name, job in real_jobs(wf).items():
                self.assertIn(
                    "timeout-minutes", job,
                    f"{name}: job '{job_name}' has no timeout-minutes (a hung job burns 6h of runner time)",
                )

    def test_every_job_runs_least_privilege(self) -> None:
        for name, wf in WORKFLOWS.items():
            if "permissions" in wf:
                continue
            for job_name, job in real_jobs(wf).items():
                self.assertIn(
                    "permissions", job,
                    f"{name}: no top-level permissions and job '{job_name}' sets none "
                    "(defaults to the token's full scope)",
                )

    def test_production_deploy_is_manual_only(self) -> None:
        # #1142 made production manual-dispatch only; a push trigger would fire
        # staging and production on the same commit. Do not let it back in.
        t = triggers(WORKFLOWS["deploy-production.yml"])
        self.assertEqual(set(t), {"workflow_dispatch"}, f"deploy-production triggers changed: {sorted(t)}")

    def test_staging_deploys_on_push_to_main(self) -> None:
        push = triggers(WORKFLOWS["deploy-staging.yml"]).get("push") or {}
        self.assertEqual(push.get("branches"), ["main"])

    def test_merge_group_workflows_never_cancel_a_queue_run(self) -> None:
        for name, wf in WORKFLOWS.items():
            if "merge_group" not in triggers(wf):
                continue
            blocks = [wf.get("concurrency")] + [j.get("concurrency") for j in real_jobs(wf).values()]
            guards = [b.get("cancel-in-progress") for b in blocks if isinstance(b, dict)]
            self.assertIn(
                MERGE_GROUP_GUARD, [str(g) for g in guards],
                f"{name}: runs on merge_group but no concurrency block guards cancel-in-progress with "
                f"{MERGE_GROUP_GUARD!r} — a cancelled queue run is read as a failure and ejects the PR",
            )

    def test_merge_group_workflows_have_no_paths_filter(self) -> None:
        for name, wf in WORKFLOWS.items():
            t = triggers(wf)
            if "merge_group" not in t:
                continue
            for ev in ("pull_request", "push"):
                cfg = t.get(ev)
                if isinstance(cfg, dict):
                    self.assertNotIn(
                        "paths", cfg,
                        f"{name}: '{ev}' has a paths filter but the workflow is a merge-queue check — "
                        "a skipped required check strands the queue entry",
                    )

    def test_required_checks_also_report_on_the_pull_request(self) -> None:
        for name, wf in WORKFLOWS.items():
            t = triggers(wf)
            if "merge_group" not in t:
                continue
            self.assertIn(
                "pull_request", t,
                f"{name}: runs on merge_group but not pull_request — it would not report on the PR itself",
            )
            if name not in NON_REQUIRED:
                self.assertNotIn("paths", t.get("pull_request") or {}, name)

    def test_ci_api_rejects_openapi_drift(self) -> None:
        workflow = WORKFLOWS["ci-api.yml"]
        steps = real_jobs(workflow)["ci-api"]["steps"]
        step = next(
            (item for item in steps if item.get("name") == "Verify committed OpenAPI spec"),
            None,
        )

        self.assertIsNotNone(step, "ci-api.yml: missing the OpenAPI drift check")
        self.assertEqual(step.get("working-directory"), "apps/api")
        command = step.get("run", "")
        self.assertIn("go run ./cmd/openapi", command)
        self.assertIn("diff -u openapi/vocanova.openapi.json", command)
        self.assertIn(
            "OpenAPI spec is out of date",
            command,
            "ci-api.yml: drift failure must explain how to regenerate the spec",
        )

    def test_auto_merge_stays_squash_and_hold_guarded(self) -> None:
        # Auto-merge must never widen past a squash landing, and the `hold`
        # label has to remain a real brake — a green PR merges on its own
        # otherwise, so these two strings are load-bearing.
        text = (WF_DIR / "auto-merge.yml").read_text()
        self.assertIn("--auto", text, "auto-merge.yml no longer arms auto-merge")
        self.assertIn("--disable-auto", text, "auto-merge.yml no longer turns auto-merge back off")
        self.assertIn('"hold"', text, "auto-merge.yml no longer honours the `hold` label")
        self.assertIn("draft", text.lower(), "auto-merge.yml no longer skips drafts")
        self.assertIn(
            "enqueuePullRequest", text,
            "auto-merge.yml no longer explicitly enqueues - arming auto-merge alone "
            "does not reliably re-enqueue a PR into a required merge queue (found live, PR #1171/#1172)",
        )


if __name__ == "__main__":
    unittest.main()
