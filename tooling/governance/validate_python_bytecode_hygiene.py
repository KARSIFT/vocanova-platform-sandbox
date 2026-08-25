#!/usr/bin/env python3
"""Fail-closed validation for tracked Python bytecode/cache hygiene."""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path

BYTECODE_SUFFIXES = (".pyc", ".pyo", ".pyd")
REQUIRED_PYCACHE_PATTERN = "__pycache__/"
ACCEPTED_PYC_PATTERNS = ("*.py[cod]", "*.pyc", "*.pyo", "*.pyd")


def list_tracked_paths(repository_root: Path) -> list[str]:
    completed = subprocess.run(
        ["git", "ls-files", "-z"],
        cwd=repository_root,
        check=False,
        capture_output=True,
    )
    if completed.returncode != 0:
        stderr = completed.stderr.decode("utf-8", errors="replace").strip()
        raise RuntimeError(f"git ls-files failed in {repository_root}: {stderr}")
    if not completed.stdout:
        return []
    return completed.stdout.decode("utf-8", errors="replace").split("\0")[:-1]


def find_tracked_bytecode_artifacts(tracked_paths: list[str]) -> list[str]:
    violations: list[str] = []
    for path in tracked_paths:
        normalized = path.replace("\\", "/")
        if "/__pycache__/" in normalized or normalized.startswith("__pycache__/"):
            violations.append(path)
            continue
        lower = normalized.lower()
        if any(lower.endswith(suffix) for suffix in BYTECODE_SUFFIXES):
            violations.append(path)
    return sorted(violations)


def parse_gitignore_patterns(gitignore_text: str) -> list[str]:
    patterns: list[str] = []
    for line in gitignore_text.splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        patterns.append(stripped)
    return patterns


def missing_required_gitignore_patterns(patterns: list[str]) -> list[str]:
    errors: list[str] = []
    has_pycache = any(
        pattern == "__pycache__/" or pattern.rstrip("/") == "__pycache__"
        for pattern in patterns
    )
    if not has_pycache:
        errors.append("repository .gitignore is missing required pattern: __pycache__/")

    has_pyc_coverage = any(
        pattern in ACCEPTED_PYC_PATTERNS for pattern in patterns
    )
    if not has_pyc_coverage:
        errors.append(
            "repository .gitignore is missing required Python bytecode coverage "
            "(expected *.py[cod] or equivalent *.pyc / *.pyo / *.pyd rules)"
        )
    return errors


def validate_gitignore_file(gitignore_path: Path) -> list[str]:
    if not gitignore_path.is_file():
        return [f"missing repository .gitignore: {gitignore_path}"]
    return missing_required_gitignore_patterns(
        parse_gitignore_patterns(gitignore_path.read_text(encoding="utf-8"))
    )


def validate_repository(repository_root: Path) -> list[str]:
    errors: list[str] = []
    tracked_paths = list_tracked_paths(repository_root)
    tracked_bytecode = find_tracked_bytecode_artifacts(tracked_paths)
    if tracked_bytecode:
        preview = ", ".join(tracked_bytecode[:5])
        suffix = "..." if len(tracked_bytecode) > 5 else ""
        errors.append(
            "tracked Python bytecode/cache artifacts remain in the index: "
            f"{preview}{suffix}"
        )
    errors.extend(validate_gitignore_file(repository_root / ".gitignore"))
    return errors


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Validate tracked Python bytecode/cache hygiene."
    )
    parser.add_argument(
        "--repository-root",
        type=Path,
        default=Path("."),
        help="Repository root to validate (default: current directory).",
    )
    return parser


def main(argv: list[str] | None = None) -> int:
    try:
        args = build_parser().parse_args(argv)
        errors = validate_repository(args.repository_root.resolve())
    except Exception as exc:  # fail closed on validator defects
        print(f"python bytecode hygiene validator internal failure: {exc}", file=sys.stderr)
        return 2
    if errors:
        for error in errors:
            print(f"ERROR: {error}", file=sys.stderr)
        print(
            f"Python bytecode hygiene validation failed with {len(errors)} error(s).",
            file=sys.stderr,
        )
        return 1
    print("Python bytecode hygiene validation passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
