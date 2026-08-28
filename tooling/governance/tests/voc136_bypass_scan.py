"""VOC-136 exhaustive caller-diff bypass scanner (source-safe literals)."""

from __future__ import annotations

import re
from pathlib import Path

FIXTURE_MIRROR_PREFIX = "tooling/governance/fixtures/karsift-ai-infra/"
SCAN_EXCLUDE_PREFIXES = (FIXTURE_MIRROR_PREFIX,)
SCAN_ALLOW_PROVENANCE_TEST_PATHS = frozenset(
    {"scripts/foundation/voc112-navigation-benchmark.test.mjs"}
)
EXECUTABLE_SUFFIXES = {".mjs", ".js", ".sh", ".py"}

_VOC112 = "VOC" + "112_"
_CAPTURE = "CAPTURE_"
_PROVENANCE = "PROVENANCE_"
_MODE = "MODE"
_PROV_ENV = _VOC112 + _CAPTURE + _PROVENANCE + _MODE

_PR = "PR_"
_BASE = "BASE_"
_HEAD = "HEAD_"
_SHA = "SHA"
_PR_BASE_ENV = _PR + _BASE + _SHA
_PR_HEAD_ENV = _PR + _HEAD + _SHA

_GIT = "git"
_FETCH = "fetch"
_HYDRATE = "hy" + "drate"
_MATERIALIZE = "material" + "ize"
_EVIDENCE = "evid" + "ence"
_ENSURE = "ensure"
_CAPTURE_SUBJECT = "f9d11e23"

CAPTURE_FETCH_PATTERN = re.compile(
    "".join(
        [
            _GIT,
            r"\s+",
            _FETCH,
            r"[^\n]*(?:capture|",
            _EVIDENCE,
            "|",
            _CAPTURE_SUBJECT,
            "|",
            "voc112",
            ")",
            "|",
            r"subprocess\.(?:run|call|Popen)\([^)]*['\"]",
            _GIT,
            r"['\"][^)]*['\"]",
            _FETCH,
            r"['\"]",
        ]
    ),
    re.IGNORECASE,
)
HYDRATE_PATTERN = re.compile(
    "".join(
        [
            _HYDRATE,
            r"[-_ ]?",
            "voc112|",
            _MATERIALIZE,
            r"[-_ ]?",
            _EVIDENCE,
            "|",
            _ENSURE,
            "-voc112-capture",
        ]
    ),
    re.IGNORECASE,
)
PROVENANCE_MODE_SET_PATTERN = re.compile(
    "".join(
        [
            r"(?:export\s+",
            _PROV_ENV,
            r"\s*=|",
            r"(?:^|[;\s&|])",
            _PROV_ENV,
            r"\s*=|",
            r"process\.env\.",
            _PROV_ENV,
            r"\s*=|",
            r"os\.environ\[['\"]",
            _PROV_ENV,
            r"['\"]\]\s*=|",
            r"os\.putenv\s*\(\s*['\"]",
            _PROV_ENV,
            r"['\"])",
        ]
    ),
    re.MULTILINE,
)
PR_SHA_SET_PATTERN = re.compile(
    "".join(
        [
            r"(?:export\s+",
            _PR_BASE_ENV,
            r"\s*=|",
            r"export\s+",
            _PR_HEAD_ENV,
            r"\s*=|",
            r"(?:^|[;\s&|])",
            _PR_BASE_ENV,
            r"\s*=|",
            r"(?:^|[;\s&|])",
            _PR_HEAD_ENV,
            r"\s*=|",
            r"process\.env\.",
            _PR_BASE_ENV,
            r"\s*=|",
            r"process\.env\.",
            _PR_HEAD_ENV,
            r"\s*=|",
            r"os\.environ\[['\"]",
            _PR_BASE_ENV,
            r"['\"]\]\s*=|",
            r"os\.environ\[['\"]",
            _PR_HEAD_ENV,
            r"['\"]\]\s*=|",
            r"os\.putenv\s*\(\s*['\"]",
            _PR_BASE_ENV,
            r"['\"]|",
            r"os\.putenv\s*\(\s*['\"]",
            _PR_HEAD_ENV,
            r"['\"])",
        ]
    ),
    re.MULTILINE,
)
LOCAL_FAIL_CLOSED_BYPASS_PATTERN = re.compile(
    "".join(
        [
            _PROV_ENV,
            r"\s*=\s*['\"]pr-(?:validation|ancestry)['\"]",
        ]
    ),
    re.IGNORECASE,
)
def is_scan_excluded(relative: str) -> bool:
    return any(relative.startswith(prefix) for prefix in SCAN_EXCLUDE_PREFIXES)


def should_scan_path(relative: str) -> bool:
    if is_scan_excluded(relative):
        return False
    if relative == "package.json" or relative.startswith("scripts/"):
        return True
    return Path(relative).suffix in EXECUTABLE_SUFFIXES


def scan_changed_path_for_bypasses(relative: str, text: str) -> None:
    if relative in SCAN_ALLOW_PROVENANCE_TEST_PATHS:
        return
    lowered = text.lower()
    if CAPTURE_FETCH_PATTERN.search(text):
        raise AssertionError(f"{relative} fetches capture/evidence commits")
    if HYDRATE_PATTERN.search(lowered):
        raise AssertionError(f"{relative} hydrates or materializes evidence")
    is_script_or_package = relative == "package.json" or relative.startswith("scripts/")
    is_executable = Path(relative).suffix in EXECUTABLE_SUFFIXES
    if is_script_or_package or is_executable:
        if PROVENANCE_MODE_SET_PATTERN.search(text):
            raise AssertionError(f"{relative} sets provenance override mode")
        if PR_SHA_SET_PATTERN.search(text):
            raise AssertionError(f"{relative} sets PR base/head SHA overrides")
        if LOCAL_FAIL_CLOSED_BYPASS_PATTERN.search(text):
            raise AssertionError(f"{relative} bypasses local fail-closed provenance")


def scan_exclude_prefixes_do_not_skip_caller_tests() -> None:
    for prefix in SCAN_EXCLUDE_PREFIXES:
        if prefix == "tooling/governance/tests/":
            raise AssertionError("caller tests must not be wholesale scan exclusions")
