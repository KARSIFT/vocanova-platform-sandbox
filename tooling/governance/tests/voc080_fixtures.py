"""Shared fixture root for VOC-080 infra contract policy tests."""

from __future__ import annotations

from pathlib import Path

REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_INFRA_ROOT = (
    REPOSITORY_ROOT / "tooling" / "governance" / "fixtures" / "karsift-ai-infra"
)


def read_fixture(relative: str) -> str:
    path = FIXTURE_INFRA_ROOT / relative
    return path.read_text(encoding="utf-8")
