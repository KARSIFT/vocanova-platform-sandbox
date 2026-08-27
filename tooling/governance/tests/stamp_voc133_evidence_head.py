"""Refresh VOC-133 implementation-head binding in t00-evidence.md."""

from __future__ import annotations

import re
import subprocess
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-133-consume-exact-infra-166-and-complete-release/t00-evidence.md"
)
IMPLEMENTATION_HEAD_PATTERN = re.compile(
    r"(\|\s*Implementation head \(attempt 2 carrier\)\s*\|\s*`)[0-9a-f]{40}(`\s*\|)"
)


def git_head() -> str:
    return subprocess.check_output(
        ["git", "rev-parse", "HEAD"],
        cwd=REPO_ROOT,
        text=True,
    ).strip()


def bind_voc133_evidence_head(content: str, head: str | None = None) -> str:
    bound_head = head or git_head()
    updated, count = IMPLEMENTATION_HEAD_PATTERN.subn(
        rf"\g<1>{bound_head}\g<2>",
        content,
        count=1,
    )
    if count != 1:
        raise ValueError(
            "VOC-133 evidence is missing the Implementation head (attempt 2 carrier) row"
        )
    return updated


def stamp_voc133_evidence_head() -> str:
    head = git_head()
    content = EVIDENCE_PATH.read_text(encoding="utf-8")
    updated = bind_voc133_evidence_head(content, head)
    if updated != content:
        EVIDENCE_PATH.write_text(updated, encoding="utf-8")
    return head


if __name__ == "__main__":
    print(stamp_voc133_evidence_head())
