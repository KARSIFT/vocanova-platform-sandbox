#!/usr/bin/env python3
"""Read-only verifier for auto-advance operator-owned carrier proof (VOC-102)."""

from __future__ import annotations

from dataclasses import dataclass
import re

from auto_advance_ownership import (
    WAITING_MARKER_PREFIX,
    branch_name,
    derive_evidence_relative_path,
    is_valid_carrier_pr,
)


IMPLEMENT_JOB_RE = re.compile(r"(^| / )implement( / |$)")


@dataclass(frozen=True)
class VerificationResult:
    ok: bool
    reason: str = ""


def verify_source_run(
    *,
    run: dict,
    repository: str,
    integration_branch: str,
) -> VerificationResult:
    if run.get("repository", {}).get("full_name") != repository:
        return VerificationResult(False, "wrong_repository")
    if run.get("event") != "issues":
        return VerificationResult(False, "wrong_event")
    if run.get("head_branch") != integration_branch:
        return VerificationResult(False, "wrong_branch")
    path = run.get("path") or run.get("workflow_path") or ""
    if not path.endswith("pipeline.yml"):
        return VerificationResult(False, "wrong_workflow")
    if run.get("conclusion") not in {None, "success", "failure"}:
        return VerificationResult(False, "unexpected_run_conclusion")
    return VerificationResult(True)


def verify_no_implement_job(jobs: list[dict], task_id: str) -> VerificationResult:
    for job in jobs:
        name = job.get("name") or ""
        if IMPLEMENT_JOB_RE.search(name):
            return VerificationResult(False, "implement_job_executed")
        if task_id in name and "implement" in name.lower():
            return VerificationResult(False, "implement_job_executed")
    return VerificationResult(True)


def verify_issue_open(state: str) -> VerificationResult:
    if state != "OPEN":
        return VerificationResult(False, "task_issue_not_open")
    return VerificationResult(True)


def count_waiting_markers(comments: list[dict]) -> int:
    return sum(
        1 for comment in comments if WAITING_MARKER_PREFIX in comment.get("body", "")
    )


def verify_carrier_state(
    *,
    pr: dict,
    prs_on_branch: list[dict],
    comments: list[dict],
    change_id: str,
    task_id: str,
    package_path: str,
    evidence_exists: bool,
    current_ref: str,
) -> VerificationResult:
    branch = branch_name(change_id, task_id)
    if pr.get("headRefName") != branch:
        return VerificationResult(False, "wrong_carrier_branch")
    if pr.get("headRefOid") != current_ref:
        return VerificationResult(False, "stale_pr_head")
    if len(prs_on_branch) != 1:
        return VerificationResult(False, "duplicate_carrier")
    if count_waiting_markers(comments) != 1:
        return VerificationResult(False, "duplicate_or_missing_marker")
    if not evidence_exists:
        return VerificationResult(False, "missing_evidence_file")
    if not is_valid_carrier_pr(
        pr_title=pr.get("title", ""),
        pr_body=pr.get("body", ""),
        change_id=change_id,
        task_id=task_id,
        package_path=package_path,
    ):
        return VerificationResult(False, "untrusted_carrier")
    expected_relative = derive_evidence_relative_path(task_id)
    if f"Pending evidence path: `{package_path}/{expected_relative}`" not in pr.get("body", ""):
        return VerificationResult(False, "wrong_evidence_path")
    return VerificationResult(True)


def verify_dispatch_run(
    *,
    run: dict,
    jobs: list[dict],
    expected_job_name: str,
    current_ref: str,
) -> VerificationResult:
    if run.get("event") != "workflow_dispatch":
        return VerificationResult(False, "verifier_wrong_event")
    if run.get("head_sha") != current_ref:
        return VerificationResult(False, "verifier_not_exact_head")
    matching = [job for job in jobs if job.get("name") == expected_job_name]
    if len(matching) != 1:
        return VerificationResult(False, "verifier_job_missing")
    if matching[0].get("conclusion") not in {None, "success"}:
        return VerificationResult(False, "verifier_job_not_success")
    return VerificationResult(True)
