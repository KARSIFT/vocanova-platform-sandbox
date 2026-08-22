#!/usr/bin/env python3
"""Sanitized operator escalation marker for remediation ownership (VOC-106)."""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys

from remediate_ownership import OPERATOR_ESCALATION_MARKER_PREFIX


SAFE_REASON_RE = re.compile(r"^[a-z][a-z0-9_]{0,79}$")
TRUSTED_BOT_LOGINS = {
    "app/karsift-ai-infra-bot",
    "karsift-ai-infra-bot",
    "github-actions[bot]",
}


def pr_has_marker(repo: str, pr_number: int, token: str, prefix: str) -> bool:
    env = os.environ.copy()
    env["GH_TOKEN"] = token
    env["GH_REPO"] = repo
    completed = subprocess.run(
        [
            "gh",
            "pr",
            "view",
            str(pr_number),
            "--json",
            "comments",
            "--repo",
            repo,
        ],
        check=True,
        capture_output=True,
        text=True,
        env=env,
    )
    comments = json.loads(completed.stdout)["comments"]
    matches = [comment for comment in comments if prefix in comment.get("body", "")]
    if any(
        (comment.get("author") or {}).get("login") not in TRUSTED_BOT_LOGINS
        for comment in matches
    ):
        raise ValueError("untrusted_operator_escalation_marker")
    if len(matches) > 1:
        raise ValueError("duplicate_operator_escalation_marker")
    return bool(matches)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repository", default=os.environ.get("GITHUB_REPOSITORY", ""))
    parser.add_argument("--token", default=os.environ.get("GITHUB_TOKEN", ""))
    parser.add_argument("--task-id", required=True)
    parser.add_argument("--package-path", required=True)
    parser.add_argument("--pr-number", type=int, required=True)
    parser.add_argument("--run-id", required=True)
    parser.add_argument("--head-sha", required=True)
    parser.add_argument("--base-sha", required=True)
    parser.add_argument("--reason", required=True)
    parser.add_argument("--ownership", required=True)
    args = parser.parse_args()

    if not args.repository or not args.token:
        print("repository and token are required", file=sys.stderr)
        return 2
    if args.ownership not in {"operator", "live-actions"}:
        print("invalid operator ownership", file=sys.stderr)
        return 2
    if not SAFE_REASON_RE.fullmatch(args.reason):
        print("invalid escalation reason", file=sys.stderr)
        return 2

    prefix = f"{OPERATOR_ESCALATION_MARKER_PREFIX} `{args.task_id}`"
    try:
        if pr_has_marker(args.repository, args.pr_number, args.token, prefix):
            print("operator escalation marker already present", file=sys.stderr)
            return 0
    except (subprocess.SubprocessError, ValueError, json.JSONDecodeError):
        print("operator escalation marker validation failed", file=sys.stderr)
        return 1

    body = "\n".join(
        [
            prefix,
            "",
            "Operator-owned / live-evidence-only task. Remediation did not dispatch",
            "the general implementer for this review FAIL or CI failure.",
            "",
            f"should_retry: `false`",
            f"reason_code: `{args.reason}`",
            f"ownership: `{args.ownership}`",
            f"task_id: `{args.task_id}`",
            f"package_path: `{args.package_path}`",
            f"pr_number: `{args.pr_number}`",
            f"run_id: `{args.run_id}`",
            f"head_sha: `{args.head_sha}`",
            f"base_sha: `{args.base_sha}`",
            "",
            "No raw logs, artifacts, secrets, OAuth/session material, user identifiers,",
            "or evidence payloads were copied into this comment.",
        ]
    )
    env = os.environ.copy()
    env["GH_TOKEN"] = args.token
    env["GH_REPO"] = args.repository
    subprocess.run(
        [
            "gh",
            "pr",
            "comment",
            str(args.pr_number),
            "--body",
            body,
            "--repo",
            args.repository,
        ],
        check=True,
        env=env,
    )
    print(f"operator escalation marker posted for {args.task_id}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
