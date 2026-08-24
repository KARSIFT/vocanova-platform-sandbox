#!/usr/bin/env python3
"""CLI wrapper for package task roster policy validation (VOC-115)."""

from __future__ import annotations

import argparse
from pathlib import Path
import sys

from package_task_policy import (
    PackageTaskPolicyError,
    change_id_from_package_path,
    ordered_task_ids,
    validate_package_tasks,
)


def _validate_package(package_path: Path, change_id: str | None) -> int:
    tasks_file = package_path / "tasks.md"
    if not tasks_file.is_file():
        print(f"missing_tasks_md: {tasks_file}", file=sys.stderr)
        return 1
    resolved_change_id = change_id or change_id_from_package_path(str(package_path))
    try:
        sections = validate_package_tasks(
            tasks_file.read_text(encoding="utf-8"),
            resolved_change_id,
        )
    except PackageTaskPolicyError as exc:
        print(str(exc), file=sys.stderr)
        return 1
    print("\n".join(ordered_task_ids(sections)))
    return 0


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    validate = sub.add_parser("validate")
    validate.add_argument("--package-path", required=True)
    validate.add_argument("--change-id", default="")

    args = parser.parse_args(argv)
    if args.command == "validate":
        return _validate_package(Path(args.package_path), args.change_id or None)
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
