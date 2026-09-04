#!/usr/bin/env python3
"""Contract tests for .github/workflows.

One consolidated test instead of kandev's per-workflow contract files: it
encodes the invariants that have actually cost this repo time (merge-queue
drift, required checks that never report, auto-deploy creeping back onto
production, queue entries ejected by a cancelled run). Run:

    python3 .github/scripts/workflows_contract_test.py
"""

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
