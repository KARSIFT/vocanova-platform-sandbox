#!/usr/bin/env python3
"""Pure ownership classification for auto-advance dispatch decisions (VOC-102)."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import re
from typing import Literal

from live_evidence_reconcile import ContractError, parse_contract_yaml


TASK_ID_RE = re.compile(r"^[A-Za-z][A-Za-z0-9]*-\d+-T\d+[a-z]?$")
EVIDENCE_SUFFIX_RE = re.compile(r"-T(\d+[a-z]?)$")
MARKER_LINE_RE = re.compile(
    r"^- Automation ownership: (operator|live-actions)\s*$"
)
WAITING_MARKER_PREFIX = (
    "VOC-102: Live-evidence carrier prepared; no implementer run was started."
)
FAIL_CLOSED_MARKER_PREFIX = (
    "VOC-102: Auto-advance fail-closed on ownership metadata for"
)

Decision = Literal["none", "implement", "prepare-live-evidence", "fail-closed"]


@dataclass(frozen=True)
class Classification:
    decision: Decision
    reason: str = ""
    ownership: str | None = None
    evidence_relative_path: str | None = None


def derive_evidence_relative_path(task_id: str) -> str:
    if not TASK_ID_RE.fullmatch(task_id):
        raise ValueError("invalid_task_id_for_evidence")
    match = EVIDENCE_SUFFIX_RE.search(task_id)
    if match is None:
        raise ValueError("invalid_task_id_for_evidence")
    return f"t{match.group(1).lower()}-evidence.md"


def parse_automation_ownership_markers(tasks_md: str, task_id: str) -> tuple[str, ...]:
    heading = f"## {task_id}"
    start = tasks_md.find(heading)
    if start < 0:
        return ()
    section_start = start + len(heading)
    next_heading = tasks_md.find("\n## ", section_start)
    section = tasks_md[section_start:] if next_heading < 0 else tasks_md[section_start:next_heading]
    markers: list[str] = []
    for line in section.splitlines():
        match = MARKER_LINE_RE.fullmatch(line.rstrip())
        if match:
            markers.append(match.group(1))
    return tuple(markers)


def _load_contract_ownership(contract_path: Path, expected_task_id: str) -> str:
    try:
        text = contract_path.read_text(encoding="utf-8")
    except OSError as exc:
        raise ContractError("unreadable_contract") from exc
    data = parse_contract_yaml(text)
    task_id = data.get("task_id")
    if not isinstance(task_id, str) or task_id != expected_task_id:
        raise ContractError("task_id_mismatch")
    ownership = data.get("ownership")
    if ownership not in {"operator", "live-actions"}:
        raise ContractError("invalid_ownership")
    return ownership


def classify_next_task(
    package_path: str,
    next_task_id: str,
    tasks_md: str | None = None,
) -> Classification:
    if not TASK_ID_RE.fullmatch(next_task_id):
        return Classification("fail-closed", "invalid_task_id")

    markers = parse_automation_ownership_markers(tasks_md or "", next_task_id)
    if len(markers) > 1:
        return Classification("fail-closed", "duplicate_automation_marker")
    marker = markers[0] if markers else None

    contract_path = Path(package_path) / ".karsift/live-evidence" / f"{next_task_id}.yaml"
    contract_exists = contract_path.is_file()

    if contract_exists:
        try:
            ownership = _load_contract_ownership(contract_path, next_task_id)
        except ContractError as exc:
            return Classification("fail-closed", exc.code)
        if marker is not None and marker != ownership:
            return Classification("fail-closed", "marker_contract_conflict")
        try:
            evidence_relative_path = derive_evidence_relative_path(next_task_id)
        except ValueError:
            return Classification("fail-closed", "invalid_task_id_for_evidence")
        return Classification(
            "prepare-live-evidence",
            ownership=ownership,
            evidence_relative_path=evidence_relative_path,
        )

    if marker is not None:
        return Classification("fail-closed", "marker_without_contract")

    return Classification("implement")


def pending_evidence_body(task_id: str, change_id: str, package_path: str) -> str:
    return "\n".join(
        [
            f"# {task_id} — Evidence (pending operator live evidence)",
            "",
            "Deterministic evidence carrier created by auto-advance (VOC-102).",
            "No implementer run was started for this operator-owned task.",
            "",
            f"Package: `{package_path}`",
            f"Change: `{change_id}`",
            "",
            "Record allowlisted metadata only when operator evidence is available.",
            "See docs/operations/live-evidence.md.",
            "",
        ]
    )


def carrier_pr_body(
    *,
    change_id: str,
    task_id: str,
    package_path: str,
    issue_number: int,
    evidence_relative_path: str,
) -> str:
    return "\n".join(
        [
            f"Operator-owned live-evidence carrier for `{task_id}` from `{change_id}`.",
            "",
            f"Pending evidence path: `{package_path}/{evidence_relative_path}`",
            "",
            "No implementer run was started. Repository-controlled reconcile observes",
            "the declared live-evidence contract after operator proof arrives.",
            "",
            f"Package path: `{package_path}`",
            "",
            f"Tracking issue: #{issue_number}",
            "",
        ]
    )


def branch_name(change_id: str, task_id: str) -> str:
    slug = re.sub(r"-+", "-", re.sub(r"[^a-z0-9]+", "-", task_id.lower())).strip("-")
    change_lower = change_id.lower()
    return f"agent/{change_lower}-{slug}"


def is_valid_carrier_pr(
    *,
    pr_title: str,
    pr_body: str,
    change_id: str,
    task_id: str,
    package_path: str,
) -> bool:
    expected_title = f"{change_id}: {task_id}"
    if pr_title.strip() != expected_title:
        return False
    if f"Package path: `{package_path}`" not in pr_body:
        return False
    if f"Pending evidence path: `{package_path}/" not in pr_body:
        return False
    return True
