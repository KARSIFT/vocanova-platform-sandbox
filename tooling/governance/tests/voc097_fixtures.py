"""Shared fixture root and policy module loader for VOC-097 regressions."""

from __future__ import annotations

from importlib.util import module_from_spec, spec_from_file_location
from pathlib import Path
import sys

REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_INFRA_ROOT = (
    REPOSITORY_ROOT / "tooling" / "governance" / "fixtures" / "karsift-ai-infra"
)
CALLER_PIPELINE = REPOSITORY_ROOT / ".github/workflows/pipeline.yml"


def read_fixture(relative: str) -> str:
    path = FIXTURE_INFRA_ROOT / relative
    return path.read_text(encoding="utf-8")


def load_policy_module(name: str, relative: str):
    path = FIXTURE_INFRA_ROOT / relative
    spec = spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise AssertionError(f"cannot load {path}")
    module = module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module
