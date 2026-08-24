#!/usr/bin/env python3
"""Deterministic package task roster policy (VOC-115)."""

from __future__ import annotations

from dataclasses import dataclass
import re
from typing import Iterable


class PackageTaskPolicyError(ValueError):
    """Fail-closed package task policy rejection with a stable reason code."""

    def __init__(self, code: str, detail: str = ""):
        message = code if not detail else f"{code}: {detail}"
        super().__init__(message)
        self.code = code
        self.detail = detail


ALLOWED_SPLIT_REASONS = frozenset(
    {
        "merge-order-dependency",
        "independently-releasable-unit",
        "distinct-owner-authority-risk-boundary",
        "mutually-exclusive-execution-environment",
        "post-merge-evidence-not-in-carrier",
        "single-pr-review-size-boundary",
    }
)

INVALID_SPLIT_REASON_SLUGS = frozenset(
    {
        "small",
        "smaller",
        "convenience",
        "review-convenience",
        "docs",
        "documentation",
        "tests",
        "test",
        "code",
        "evidence",
        "docs-vs-code",
        "tests-vs-code",
        "code-vs-tests",
        "file-type",
        "file-count",
        "component-count",
        "skill-count",
        "size",
        "l",
        "xs",
        "s",
        "m",
        "800-lines",
        "line-count",
        "changed-lines",
    }
)

TASK_HEADING_RE = re.compile(
    r"^##[ \t]+([A-Z][A-Z0-9]*-\d+-T\d+[a-z]?)(?:[ \t]+(?:—|-).*)?[ \t]*$",
    re.MULTILINE,
)
SPLIT_REASON_LINE_RE = re.compile(
    r"^-\s*Split reason:\s*([a-z0-9-]+)(?:\s*(?:—|-)\s*(.+))?\s*$",
    re.IGNORECASE | re.MULTILINE,
)
MULTI_TASK_JUSTIFICATION_HEADING = re.compile(
    r"^##[ \t]+Package-level multi-task justification[ \t]*$",
    re.MULTILINE | re.IGNORECASE,
)
CHANGE_ID_RE = re.compile(r"^([A-Z][A-Z0-9]*-\d+)-", re.IGNORECASE)
REVIEW_SIZE_EXPLANATION_MIN_LEN = 30
MULTI_TASK_JUSTIFICATION_MIN_LEN = 40
EXCEPTIONAL_TASK_COUNT = 3


@dataclass(frozen=True)
class TaskSection:
    task_id: str
    body: str
    split_reason: str | None = None
    split_explanation: str | None = None


def change_id_from_package_path(package_path: str) -> str:
    name = package_path.rstrip("/").rsplit("/", 1)[-1]
    match = CHANGE_ID_RE.match(name)
    if match is None:
        raise PackageTaskPolicyError("invalid_package_path", package_path)
    return match.group(1).upper()


def parse_task_sections(tasks_md: str, change_id: str) -> list[TaskSection]:
    if not tasks_md.strip():
        raise PackageTaskPolicyError("missing_tasks_md")

    change_prefix = f"{change_id.upper()}-T"
    headings = list(TASK_HEADING_RE.finditer(tasks_md))
    if not headings:
        raise PackageTaskPolicyError("no_task_headings", change_id)

    sections: list[TaskSection] = []
    for index, match in enumerate(headings):
        task_id = match.group(1)
        if not task_id.upper().startswith(change_prefix):
            continue
        start = match.end()
        end = headings[index + 1].start() if index + 1 < len(headings) else len(tasks_md)
        body = tasks_md[start:end]
        split_reason = None
        split_explanation = None
        reason_matches = list(SPLIT_REASON_LINE_RE.finditer(body))
        if reason_matches:
            if len(reason_matches) != 1:
                raise PackageTaskPolicyError(
                    "duplicate_split_reason", task_id
                )
            slug = reason_matches[0].group(1).lower()
            split_reason = slug
            split_explanation = (reason_matches[0].group(2) or "").strip() or None
        sections.append(
            TaskSection(
                task_id=task_id,
                body=body,
                split_reason=split_reason,
                split_explanation=split_explanation,
            )
        )

    if not sections:
        raise PackageTaskPolicyError("no_package_tasks", change_id)
    return sections


def package_level_multi_task_justification(tasks_md: str) -> str | None:
    match = MULTI_TASK_JUSTIFICATION_HEADING.search(tasks_md)
    if match is None:
        return None
    start = match.end()
    next_heading = tasks_md.find("\n## ", start)
    body = tasks_md[start:] if next_heading < 0 else tasks_md[start:next_heading]
    text = "\n".join(
        line.strip()
        for line in body.splitlines()
        if line.strip() and not line.strip().startswith("#")
    ).strip()
    return text or None


def _validate_split_reason(section: TaskSection) -> None:
    if section.split_reason is None:
        raise PackageTaskPolicyError("missing_split_reason", section.task_id)

    slug = section.split_reason.lower()
    if slug in INVALID_SPLIT_REASON_SLUGS:
        raise PackageTaskPolicyError("invalid_split_reason_slug", section.task_id)
    if slug not in ALLOWED_SPLIT_REASONS:
        raise PackageTaskPolicyError("unknown_split_reason", section.task_id)
    if slug == "single-pr-review-size-boundary":
        explanation = (section.split_explanation or "").strip()
        if len(explanation) < REVIEW_SIZE_EXPLANATION_MIN_LEN:
            raise PackageTaskPolicyError(
                "review_size_boundary_requires_explanation",
                section.task_id,
            )


def validate_package_tasks(tasks_md: str, change_id: str) -> list[TaskSection]:
    sections = parse_task_sections(tasks_md, change_id)
    for index, section in enumerate(sections):
        if index == 0:
            if section.split_reason is not None:
                raise PackageTaskPolicyError(
                    "first_task_must_not_split", section.task_id
                )
            continue
        _validate_split_reason(section)

    if len(sections) > EXCEPTIONAL_TASK_COUNT:
        justification = package_level_multi_task_justification(tasks_md)
        if not justification or len(justification) < MULTI_TASK_JUSTIFICATION_MIN_LEN:
            raise PackageTaskPolicyError("missing_multi_task_justification")

    return sections


def ordered_task_ids(sections: Iterable[TaskSection]) -> list[str]:
    return [section.task_id for section in sections]
