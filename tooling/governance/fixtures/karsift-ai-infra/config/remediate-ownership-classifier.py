#!/usr/bin/env python3
"""CLI wrapper for remediation ownership classification (VOC-106)."""

from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import sys

from remediate_ownership import classify_task_for_remediation


def write_output(*, ownership: str, reason: str, task_id: str, package_path: str) -> None:
    github_output = os.environ.get("GITHUB_OUTPUT")
    payload = {
        "ownership": ownership,
        "ownership_reason": reason,
        "task_id": task_id,
        "package_path": package_path,
    }
    if github_output:
        with open(github_output, "a", encoding="utf-8") as handle:
            for key, value in payload.items():
                handle.write(f"{key}={value}\n")
    else:
        print(json.dumps(payload, sort_keys=True))


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--package-path", required=True)
    parser.add_argument("--task-id", required=True)
    args = parser.parse_args()

    tasks_md_path = Path(args.package_path) / "tasks.md"
    if not tasks_md_path.is_file():
        ownership = "fail-closed"
        reason = "missing_tasks_file"
    else:
        try:
            tasks_md = tasks_md_path.read_text(encoding="utf-8")
        except (OSError, UnicodeError):
            ownership = "fail-closed"
            reason = "unreadable_tasks_file"
        else:
            ownership, reason = classify_task_for_remediation(
                args.package_path,
                args.task_id,
                tasks_md,
            )
    write_output(
        ownership=ownership,
        reason=reason,
        task_id=args.task_id,
        package_path=args.package_path,
    )
    if ownership == "fail-closed":
        print(
            f"remediation: fail-closed ({reason}) for {args.task_id}",
            file=sys.stderr,
        )
        return 0
    print(
        f"remediation: ownership={ownership} task={args.task_id}",
        file=sys.stderr,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
