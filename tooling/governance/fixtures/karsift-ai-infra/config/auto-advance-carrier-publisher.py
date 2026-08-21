#!/usr/bin/env python3
"""Deterministic live-evidence carrier publisher for auto-advance (VOC-102)."""

from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import textwrap

from auto_advance_ownership import (
    WAITING_MARKER_PREFIX,
    branch_name,
    carrier_pr_body,
    is_valid_carrier_pr,
    pending_evidence_body,
)


class PublisherError(RuntimeError):
    pass


def run_gh(args: list[str], *, token: str, repo: str) -> str:
    env = os.environ.copy()
    env["GH_TOKEN"] = token
    env["GH_REPO"] = repo
    completed = subprocess.run(
        ["gh", *args, "--repo", repo],
        check=False,
        capture_output=True,
        text=True,
        env=env,
    )
    if completed.returncode != 0:
        raise PublisherError(
            f"gh {' '.join(args)} failed: {completed.stderr.strip() or completed.stdout.strip()}"
        )
    return completed.stdout.strip()


def issue_has_marker(repo: str, issue_number: int, token: str, prefix: str) -> bool:
    comments = json.loads(
        run_gh(
            [
                "issue",
                "view",
                str(issue_number),
                "--json",
                "comments",
            ],
            token=token,
            repo=repo,
        )
    )["comments"]
    return any(prefix in comment.get("body", "") for comment in comments)


def post_deduplicated_comment(
    repo: str,
    issue_number: int,
    token: str,
    body: str,
    prefix: str,
) -> None:
    if issue_has_marker(repo, issue_number, token, prefix):
        return
    run_gh(
        ["issue", "comment", str(issue_number), "--body", body],
        token=token,
        repo=repo,
    )


def find_pr_for_branch(repo: str, branch: str, token: str) -> dict | None:
    raw = run_gh(
        [
            "pr",
            "list",
            "--head",
            branch,
            "--state",
            "all",
            "--json",
            "number,title,body,isDraft,headRefName",
        ],
        token=token,
        repo=repo,
    )
    items = json.loads(raw)
    return items[0] if items else None


def ensure_carrier(
    *,
    repo: str,
    token: str,
    integration_branch: str,
    change_id: str,
    task_id: str,
    package_path: str,
    issue_number: int,
    evidence_relative_path: str,
) -> None:
    branch = branch_name(change_id, task_id)
    evidence_path = f"{package_path}/{evidence_relative_path}"
    existing = find_pr_for_branch(repo, branch, token)
    if existing is not None and not is_valid_carrier_pr(
        pr_title=existing.get("title", ""),
        pr_body=existing.get("body", ""),
        change_id=change_id,
        task_id=task_id,
        package_path=package_path,
    ):
        raise PublisherError("conflicting_existing_pr")

    workdir = Path(os.environ.get("RUNNER_TEMP", "/tmp")) / "auto-advance-carrier"
    if workdir.exists():
        for child in workdir.iterdir():
            if child.is_dir():
                subprocess.run(["rm", "-rf", str(child)], check=False)
    workdir.mkdir(parents=True, exist_ok=True)

    clone_dir = workdir / "repo"
    env = os.environ.copy()
    env["GH_TOKEN"] = token
    subprocess.run(
        [
            "gh",
            "repo",
            "clone",
            repo,
            str(clone_dir),
            "--",
            "--branch",
            integration_branch,
            "--single-branch",
        ],
        check=True,
        env=env,
    )
    subprocess.run(["git", "-C", str(clone_dir), "config", "user.name", "karsift-bot"], check=True)
    subprocess.run(
        ["git", "-C", str(clone_dir), "config", "user.email", "41898282+github-actions[bot]@users.noreply.github.com"],
        check=True,
    )

    remote_branch_exists = subprocess.run(
        ["git", "-C", str(clone_dir), "ls-remote", "--heads", "origin", branch],
        capture_output=True,
        text=True,
        check=False,
    ).stdout.strip()

    if remote_branch_exists:
        subprocess.run(
            ["git", "-C", str(clone_dir), "fetch", "origin", branch],
            check=True,
        )
        subprocess.run(
            ["git", "-C", str(clone_dir), "checkout", "-B", branch, f"origin/{branch}"],
            check=True,
        )
    else:
        subprocess.run(
            ["git", "-C", str(clone_dir), "checkout", "-B", branch],
            check=True,
        )

    target = clone_dir / evidence_relative_path
    target.parent.mkdir(parents=True, exist_ok=True)
    body = pending_evidence_body(task_id, change_id, package_path)
    if not target.exists() or target.read_text(encoding="utf-8") != body:
        target.write_text(body, encoding="utf-8")
        subprocess.run(["git", "-C", str(clone_dir), "add", evidence_relative_path], check=True)
        subprocess.run(
            ["git", "-C", str(clone_dir), "commit", "-m", f"{task_id}: pending operator live-evidence carrier"],
            check=False,
        )

    subprocess.run(
        ["git", "-C", str(clone_dir), "push", "-u", "origin", branch],
        check=True,
        env=env,
    )

    pr_body = carrier_pr_body(
        change_id=change_id,
        task_id=task_id,
        package_path=package_path,
        issue_number=issue_number,
        evidence_relative_path=evidence_relative_path,
    )
    if existing is None:
        run_gh(
            [
                "pr",
                "create",
                "--base",
                integration_branch,
                "--head",
                branch,
                "--title",
                f"{change_id}: {task_id}",
                "--body",
                pr_body,
                "--draft",
            ],
            token=token,
            repo=repo,
        )
    else:
        run_gh(
            ["pr", "edit", str(existing["number"]), "--body", pr_body],
            token=token,
            repo=repo,
        )

    waiting_body = textwrap.dedent(
        f"""\
        {WAITING_MARKER_PREFIX}

        Task `{task_id}` is operator-owned live evidence. Auto-advance created or reused
        the deterministic evidence-carrier PR on `{branch}` and did **not** start
        `implement.yml`.

        Pending evidence path: `{evidence_path}`
        """
    )
    post_deduplicated_comment(
        repo,
        issue_number,
        token,
        waiting_body,
        WAITING_MARKER_PREFIX,
    )


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repository", default=os.environ.get("GITHUB_REPOSITORY", ""))
    parser.add_argument("--token", default=os.environ.get("KARSIFT_APP_TOKEN", ""))
    parser.add_argument("--integration-branch", required=True)
    parser.add_argument("--change-id", required=True)
    parser.add_argument("--package-path", required=True)
    parser.add_argument("--task-id", required=True)
    parser.add_argument("--issue-number", type=int, required=True)
    parser.add_argument("--evidence-relative-path", required=True)
    args = parser.parse_args()

    if not args.repository or not args.token:
        print("repository and token are required", file=sys.stderr)
        return 2

    try:
        ensure_carrier(
            repo=args.repository,
            token=args.token,
            integration_branch=args.integration_branch,
            change_id=args.change_id,
            task_id=args.task_id,
            package_path=args.package_path,
            issue_number=args.issue_number,
            evidence_relative_path=args.evidence_relative_path,
        )
    except PublisherError as exc:
        print(f"carrier publisher failed: {exc}", file=sys.stderr)
        return 1
    print(f"carrier publisher succeeded for {args.task_id}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
