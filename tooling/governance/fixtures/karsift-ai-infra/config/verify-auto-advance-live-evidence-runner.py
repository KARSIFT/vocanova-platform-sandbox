#!/usr/bin/env python3
"""Hosted runner adapter for verify-auto-advance-live-evidence (VOC-102)."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path

from verify_auto_advance_live_evidence import (
    verify_carrier_state,
    verify_issue_open,
    verify_no_implement_job,
    verify_source_run,
)


class GitHubApi:
    def __init__(self, token: str, repository: str):
        self.token = token
        self.repository = repository

    def gh(self, args: list[str]) -> str:
        env = os.environ.copy()
        env["GH_TOKEN"] = self.token
        env["GH_REPO"] = self.repository
        completed = subprocess.run(
            ["gh", *args, "--repo", self.repository],
            check=True,
            capture_output=True,
            text=True,
            env=env,
        )
        return completed.stdout.strip()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repository", default=os.environ.get("GITHUB_REPOSITORY", ""))
    parser.add_argument("--source-run-id", required=True)
    parser.add_argument("--waiting-pr-number", type=int, required=True)
    parser.add_argument("--change-id", required=True)
    parser.add_argument("--task-id", required=True)
    parser.add_argument("--package-path", required=True)
    parser.add_argument("--integration-branch", default="develop")
    parser.add_argument("--current-ref", default=os.environ.get("GITHUB_SHA", ""))
    args = parser.parse_args()

    token = os.environ.get("GITHUB_TOKEN", "")
    if not token or not args.repository or not args.current_ref:
        print("GITHUB_TOKEN, repository, and current ref are required", file=sys.stderr)
        return 2

    api = GitHubApi(token, args.repository)
    run = json.loads(
        api.gh(
            [
                "api",
                f"/repos/{args.repository}/actions/runs/{args.source_run_id}",
            ]
        )
    )
    source_check = verify_source_run(
        run={
            "repository": run.get("repository", {}),
            "event": run.get("event"),
            "head_branch": run.get("head_branch"),
            "path": run.get("path"),
            "conclusion": run.get("conclusion"),
        },
        repository=args.repository,
        integration_branch=args.integration_branch,
    )
    if not source_check.ok:
        print(f"verify failed: {source_check.reason}", file=sys.stderr)
        return 1

    jobs = json.loads(
        api.gh(
            [
                "api",
                f"/repos/{args.repository}/actions/runs/{args.source_run_id}/jobs",
                "-q",
                ".jobs",
            ]
        )
    )
    implement_check = verify_no_implement_job(jobs, args.task_id)
    if not implement_check.ok:
        print(f"verify failed: {implement_check.reason}", file=sys.stderr)
        return 1

    pr = json.loads(
        api.gh(
            [
                "pr",
                "view",
                str(args.waiting_pr_number),
                "--json",
                "number,title,body,headRefName,headRefOid,isDraft",
            ]
        )
    )
    branch = pr.get("headRefName", "")
    prs_on_branch = json.loads(
        api.gh(
            [
                "pr",
                "list",
                "--head",
                branch,
                "--state",
                "all",
                "--json",
                "number",
            ]
        )
    )
    issue_number = None
    for line in pr.get("body", "").splitlines():
        if line.startswith("Tracking issue: #"):
            issue_number = int(line.split("#", 1)[1].strip())
            break
    if issue_number is None:
        print("verify failed: missing_tracking_issue", file=sys.stderr)
        return 1

    issue = json.loads(
        api.gh(
            [
                "issue",
                "view",
                str(issue_number),
                "--json",
                "state,comments",
            ]
        )
    )
    issue_check = verify_issue_open(issue.get("state", ""))
    if not issue_check.ok:
        print(f"verify failed: {issue_check.reason}", file=sys.stderr)
        return 1

    evidence_relative = Path(args.package_path) / __import__(
        "auto_advance_ownership", fromlist=["derive_evidence_relative_path"]
    ).derive_evidence_relative_path(args.task_id)
    evidence_exists = evidence_relative.is_file()

    carrier_check = verify_carrier_state(
        pr=pr,
        prs_on_branch=prs_on_branch,
        comments=issue.get("comments", []),
        change_id=args.change_id,
        task_id=args.task_id,
        package_path=args.package_path,
        evidence_exists=evidence_exists,
        current_ref=args.current_ref,
    )
    if not carrier_check.ok:
        print(f"verify failed: {carrier_check.reason}", file=sys.stderr)
        return 1

    print(
        json.dumps(
            {
                "source_run_id": args.source_run_id,
                "waiting_pr_number": args.waiting_pr_number,
                "task_id": args.task_id,
                "carrier_head_sha": pr.get("headRefOid"),
                "verify_result": "pass",
            },
            sort_keys=True,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
