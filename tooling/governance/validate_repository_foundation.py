#!/usr/bin/env python3
"""Deterministically validate the VocaNova repository foundation."""

from __future__ import annotations

import argparse
import hashlib
import re
import sys
from collections import defaultdict
from pathlib import Path

PACKAGE_FILES = (
    "README.md",
    "acceptance-criteria.md",
    "change.yaml",
    "impact-analysis.md",
    "implementation-plan.md",
    "release-plan.md",
    "specification.md",
    "tasks.md",
    "test-plan.md",
)

A003_PATH = "docs/governance/amendments/A-003-governed-autonomous-engineering-authority.md"
A003_STATE_PATH = "docs/governance/a003-transition-state.yaml"
A003_FROZEN_SHA256 = "f2b454653a33e6cb76a0eab37c01d48b0174227450c9ea255474f6aac59b4f83"
A003_FROZEN_BODY_SHA256 = "ad05cc8c92047002288245574bc3b76e1cce6f54d43805039ad53393534af4e7"
A004_PATH = "docs/governance/amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md"
A004_STATE_PATH = "docs/governance/a004-transition-state.yaml"
DOC17_PATH = "docs/architecture/17-autonomous-development-architecture.md"
DOC18_PATH = "docs/planning/18-autonomous-development-implementation-roadmap.md"
DOC17_SOURCE_SHA256 = "8c9fd7b714e84d39f4b5e9d5c8a4cf8f00a3231b269e2d6dadf6e0ff7707693a"
DOC18_SOURCE_SHA256 = "717c33649f49cedca64cc4744d8121f4b6f5a371c9760076bfa8134c050a8664"
DOC17_BODY_SHA256 = "b3a157557210f0afecbb5ed4ff53cd2738f50c451c39ef0d012363a6d8df7a40"
DOC18_BODY_SHA256 = "3d578186804cc2b3b500eec72809b26c03d9f236a4a22d3534daa1e2ba34c451"
VOC002_PATH = "specs/changes/VOC-002-a003-governance-transition"
VOC003_PATH = "specs/changes/VOC-003-a003-lifecycle-sync"
VOC004_PATH = "specs/changes/VOC-004-canonical-adoption-doc-17-doc-18"

# VOC-075-T04 / issue #573 (DEP-02 option a, DEP-01 option b): scan every
# specs/changes/*/change.yaml and fail when automatic_merge_allowed: false is set
# without risk: R4. Historical packages that still carried the pre-VOC-075
# violation at T04 implementation time are grandfathered until a follow-up
# backfill removes their false opt-out (see AGENTS.md approve-only-R4 rule).
VOC075_HISTORICAL_NON_R4_FALSE_PACKAGE_IDS = frozenset({
    "VOC-005",
    "VOC-006",
    "VOC-010",
    "VOC-011",
    "VOC-013",
    "VOC-014",
    "VOC-015",
    "VOC-016",
    "VOC-017",
    "VOC-018",
    "VOC-019",
    "VOC-020",
    "VOC-021",
    "VOC-022",
    "VOC-024",
    "VOC-025",
    "VOC-026",
    "VOC-047",
    "VOC-048",
    "VOC-050",
    "VOC-051",
    "VOC-052",
    "VOC-053",
    "VOC-063",
    "VOC-065",
    "VOC-066",
    "VOC-068",
})

REQUIRED_FILES = (
    "AGENTS.md",
    "CLAUDE.md",
    "README.md",
    ".github/CODEOWNERS",
    ".github/approved-policy/protected-paths.yaml",
    ".github/pull_request_template.md",
    ".github/workflows/governance-policy.yml",
    ".github/workflows/repository-governance.yml",
    "docs/README.md",
    DOC17_PATH,
    DOC18_PATH,
    A003_PATH,
    A003_STATE_PATH,
    "docs/decisions/README.md",
    "scripts/governance/classify-change-risk.sh",
    "scripts/governance/validate-governance.sh",
    "specs/README.md",
    f"{VOC002_PATH}/change.yaml",
    f"{VOC003_PATH}/change.yaml",
    f"{VOC004_PATH}/change.yaml",
    "tooling/governance/validate_repository_foundation.py",
    "tooling/governance/tests/test_validate_repository_foundation.py",
)

PROTECTED_PATHS = (
    "AGENTS.md",
    "CLAUDE.md",
    ".github/CODEOWNERS",
    ".github/pull_request_template.md",
    ".github/workflows/",
    ".github/approved-policy/",
    "scripts/governance/",
    "tooling/governance/",
    "docs/governance/",
    DOC17_PATH,
    DOC18_PATH,
    "docs/operations/15-ai-native-product-and-engineering-operating-model.md",
    "docs/decisions/",
    "specs/README.md",
    "specs/templates/",
    "specs/changes/VOC-001-repository-foundation/",
    "specs/changes/VOC-002-a003-governance-transition/",
    "specs/changes/VOC-003-a003-lifecycle-sync/",
    "specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/",
)

TEMPLATE_MARKERS = {
    "README.md": ("Identity and lifecycle", "Verification, approvals, release, and closure"),
    "specification.md": ("Objective and requirement source", "Data, migrations, analytics, and accessibility"),
    "acceptance-criteria.md": ("VOC-000-AC-00", "Evidence"),
    "impact-analysis.md": ("Security and privacy", "Risks, dependencies, and evidence"),
    "implementation-plan.md": ("File reconciliation and implementation sequence", "Deployment and rollback"),
    "tasks.md": ("VOC-000-T00", "Evidence"),
    "test-plan.md": ("VOC-000-TEST-00", "Expected result"),
    "release-plan.md": ("Release and deployment authorization", "human approvals, and closure"),
}

PR_MARKERS = (
    "VOC-###",
    "Change-package status and canonical path",
    "Requirement source",
    "Stable acceptance-criteria mapping",
    "Existing-file reconciliation",
    "Previous governance control",
    "Proposed governance control",
    "Implementer provenance",
    "Verifier provenance",
    "Exact reviewed head SHA",
    "Hosted activation status",
    "Package closure status",
    "Active authority model",
    "Effective-activation evidence",
    "Automatic-merge status",
    "Lightweight R0",
)

ID_TOKEN = re.compile(
    r"VOC-001-(?:D\d+|T\d+|R\d+|(?:AC|TEST|DEP|EV|CON|AM)-\d+)"
)


class Validation:
    def __init__(self, root: Path) -> None:
        self.root = root
        self.errors: list[str] = []

    def error(self, path: str | Path, message: str) -> None:
        self.errors.append(f"{path}: {message}")

    def read(self, relative: str | Path) -> str:
        path = self.root / relative
        try:
            return path.read_text(encoding="utf-8")
        except (OSError, UnicodeError) as exc:
            self.error(relative, f"cannot read UTF-8 text: {exc}")
            return ""


def strip_restricted_yaml_comment(raw: str) -> str:
    quote = ""
    for index, character in enumerate(raw):
        if character in {"\"", "'"}:
            if not quote:
                quote = character
            elif quote == character:
                quote = ""
        elif character == "#" and not quote:
            return raw[:index].rstrip()
    return raw.rstrip()


def validate_restricted_yaml(validation: Validation, relative: str) -> dict[str, str]:
    text = validation.read(relative)
    top_level: dict[str, str] = {}
    contexts: dict[int, tuple[str, ...]] = {-1: ("root",)}
    sequence_counts: defaultdict[tuple[tuple[str, ...], int], int] = defaultdict(int)
    seen: defaultdict[tuple[str, ...], set[str]] = defaultdict(set)

    for number, raw in enumerate(text.splitlines(), 1):
        if "\t" in raw:
            validation.error(relative, f"line {number}: tabs are not allowed in restricted YAML")
        content = strip_restricted_yaml_comment(raw)
        if not content.strip():
            continue
        if re.search(r"(^|\s)[&*!][A-Za-z0-9_-]+", content) or any(x in content for x in ("{", "}", "[", "]")):
            validation.error(relative, f"line {number}: unsupported YAML construct")
        indent = len(content) - len(content.lstrip(" "))
        if indent % 2:
            validation.error(relative, f"line {number}: indentation must use two-space levels")
        match = re.match(r"^\s*(-\s+)?([A-Za-z_][A-Za-z0-9_-]*):(?:\s*(.*))?$", content)
        if not match:
            if not re.match(r"^\s*-\s+[^:]+$", content):
                validation.error(relative, f"line {number}: unsupported restricted-YAML syntax")
            continue
        is_sequence = bool(match.group(1))
        key = match.group(2)
        value = (match.group(3) or "").strip()
        lower_indents = [level for level in contexts if level < indent]
        parent = contexts[max(lower_indents)] if lower_indents else ("root",)
        if is_sequence:
            counter_key = (parent, indent)
            sequence_counts[counter_key] += 1
            scope = parent + (f"#{sequence_counts[counter_key]}",)
            contexts[indent] = scope
        else:
            scope = parent
        if key in seen[scope]:
            validation.error(relative, f"line {number}: duplicate YAML key '{key}'")
        seen[scope].add(key)
        if indent == 0 and not is_sequence:
            top_level[key] = value.strip('"\'')
        if not value and not is_sequence:
            contexts[indent] = scope + (key,)
        for level in tuple(contexts):
            if level > indent:
                del contexts[level]
    return top_level


def require_complete_directory(validation: Validation, relative: str) -> None:
    directory = validation.root / relative
    if not directory.is_dir():
        validation.error(relative, "required directory is missing")
        return
    actual = tuple(sorted(path.name for path in directory.iterdir() if path.is_file()))
    if actual != PACKAGE_FILES:
        missing = sorted(set(PACKAGE_FILES) - set(actual))
        extra = sorted(set(actual) - set(PACKAGE_FILES))
        validation.error(relative, f"must contain exactly nine files; missing={missing}, extra={extra}")
    for name in PACKAGE_FILES:
        path = directory / name
        if path.is_file() and path.stat().st_size == 0:
            validation.error(f"{relative}/{name}", "file is empty")


def validate_templates(validation: Validation) -> None:
    relative = "specs/templates/change-package"
    require_complete_directory(validation, relative)
    yaml_values = validate_restricted_yaml(validation, f"{relative}/change.yaml")
    expected = {
        "id": "VOC-000",
        "slug": "replace-with-approved-slug",
        "status": "draft",
        "risk": "R0",
    }
    for key, value in expected.items():
        if yaml_values.get(key) != value:
            validation.error(f"{relative}/change.yaml", f"{key} must be safe placeholder {value!r}")
    for name, markers in TEMPLATE_MARKERS.items():
        text = validation.read(f"{relative}/{name}")
        if "VOC-001" in text:
            validation.error(f"{relative}/{name}", "template must not look like the approved VOC-001 package")
        for marker in markers:
            if marker not in text:
                validation.error(f"{relative}/{name}", f"missing required template heading/marker: {marker}")


def definition_ids(package: dict[str, str]) -> list[str]:
    definitions: list[str] = []
    patterns = {
        "specification.md": re.compile(r"^-\s+(?:\*\*|`)(VOC-001-(?:D\d+|AM-\d+))", re.MULTILINE),
        "acceptance-criteria.md": re.compile(r"^##\s+(VOC-001-AC-\d+)", re.MULTILINE),
        "tasks.md": re.compile(r"^##\s+(VOC-001-T\d+)", re.MULTILINE),
        "test-plan.md": re.compile(r"^##\s+(VOC-001-TEST-\d+)", re.MULTILINE),
        "impact-analysis.md": re.compile(r"^##\s+(VOC-001-(?:(?:CON|DEP|EV)-\d+|R\d+))", re.MULTILINE),
    }
    for name, pattern in patterns.items():
        definitions.extend(pattern.findall(package.get(name, "")))
    return definitions


def validate_package(validation: Validation) -> None:
    relative = "specs/changes/VOC-001-repository-foundation"
    require_complete_directory(validation, relative)
    yaml_values = validate_restricted_yaml(validation, f"{relative}/change.yaml")
    expected = {
        "schema_version": "1",
        "id": "VOC-001",
        "slug": "repository-foundation",
        "title": "Repository Foundation",
        "type": "infrastructure",
        "status": "implementing",
        "risk": "R4",
        "canonical_path": relative,
    }
    for key, value in expected.items():
        if yaml_values.get(key) != value:
            validation.error(f"{relative}/change.yaml", f"{key} must equal {value!r}")
    if yaml_values.get("status") not in {"draft", "accepted", "implementation-ready", "implementing", "blocked", "closed", "superseded"}:
        validation.error(f"{relative}/change.yaml", "invalid lifecycle status")
    if yaml_values.get("risk") not in {"R0", "R1", "R2", "R3", "R4"}:
        validation.error(f"{relative}/change.yaml", "risk must be R0 through R4")

    package = {name: validation.read(f"{relative}/{name}") for name in PACKAGE_FILES if name.endswith(".md")}
    definitions = definition_ids(package)
    duplicates = sorted(identifier for identifier in set(definitions) if definitions.count(identifier) > 1)
    if duplicates:
        validation.error(relative, f"duplicate stable identifier definitions: {duplicates}")
    defined = set(definitions)
    references = set()
    for text in package.values():
        references.update(ID_TOKEN.findall(text))
    unresolved = sorted(references - defined)
    if unresolved:
        validation.error(relative, f"unresolved stable identifier references: {unresolved}")
    required_ranges = {
        *(f"VOC-001-D{i:02d}" for i in range(1, 109)),
        *(f"VOC-001-AC-{i:02d}" for i in range(1, 29)),
        *(f"VOC-001-T{i:02d}" for i in range(1, 25)),
        *(f"VOC-001-TEST-{i:02d}" for i in range(1, 26)),
        *(f"VOC-001-R{i:02d}" for i in range(1, 11)),
        *(f"VOC-001-DEP-{i:02d}" for i in range(1, 9)),
        *(f"VOC-001-EV-{i:02d}" for i in range(1, 14)),
        *(f"VOC-001-AM-{i:02d}" for i in range(1, 6)),
    }
    missing = sorted(required_ranges - defined)
    if missing:
        validation.error(relative, f"missing approved stable identifier definitions: {missing}")
    combined = "\n".join(package.values())
    for marker in ("GitHub issue #6", "0211d75f28a4986694555f584dd8b84a3228a2ad", "PASS WITH NON-BLOCKING FINDINGS"):
        if marker not in combined:
            validation.error(relative, f"missing reconciled evidence marker: {marker}")


def validate_voc_002_package(validation: Validation) -> None:
    relative = VOC002_PATH
    require_complete_directory(validation, relative)
    values = validate_restricted_yaml(validation, f"{relative}/change.yaml")
    expected = {
        "schema_version": "1",
        "id": "VOC-002",
        "slug": "a003-governance-transition",
        "title": "A-003 Governance Transition",
        "type": "governance",
        "status": "implementing",
        "risk": "R4",
        "protected_technical_effect": "R3",
        "canonical_path": relative,
    }
    for key, value in expected.items():
        if values.get(key) != value:
            validation.error(f"{relative}/change.yaml", f"{key} must equal {value!r}")

    package = {name: validation.read(f"{relative}/{name}") for name in PACKAGE_FILES if name.endswith(".md")}
    combined = "\n".join(package.values())
    patterns = {
        "specification.md": re.compile(r"^- \*\*(VOC-002-R\d+):?\*\*", re.MULTILINE),
        "acceptance-criteria.md": re.compile(r"^## (VOC-002-AC-\d+)", re.MULTILINE),
        "impact-analysis.md": re.compile(r"^## (VOC-002-IMP-\d+)", re.MULTILINE),
        "tasks.md": re.compile(r"^## (VOC-002-T\d+)", re.MULTILINE),
        "test-plan.md": re.compile(r"^## (VOC-002-TEST-\d+)", re.MULTILINE),
    }
    definitions: list[str] = []
    for name, pattern in patterns.items():
        definitions.extend(pattern.findall(package[name]))
    duplicates = sorted(item for item in set(definitions) if definitions.count(item) > 1)
    if duplicates:
        validation.error(relative, f"duplicate VOC-002 stable identifier definitions: {duplicates}")
    required = {
        *(f"VOC-002-R{i:02d}" for i in range(1, 17)),
        *(f"VOC-002-AC-{i:02d}" for i in range(1, 13)),
        *(f"VOC-002-IMP-{i:02d}" for i in range(1, 9)),
        *(f"VOC-002-T{i:02d}" for i in range(1, 9)),
        *(f"VOC-002-TEST-{i:02d}" for i in range(1, 13)),
    }
    missing = sorted(required - set(definitions))
    if missing:
        validation.error(relative, f"missing VOC-002 stable identifier definitions: {missing}")
    references = set(re.findall(r"VOC-002-(?:R\d+|AC-\d+|IMP-\d+|T\d+|TEST-\d+)", combined))
    unresolved = sorted(references - set(definitions))
    if unresolved:
        validation.error(relative, f"unresolved VOC-002 stable identifier references: {unresolved}")
    for marker in (
        A003_FROZEN_SHA256,
        "pre-A-003",
        "R4",
        "R3",
        "exact-SHA Claude",
        "approved PR head SHA",
        "adopted `develop` SHA",
        "one-time",
        "DOC-17",
        "DOC-18",
        "automatic merge",
        "autonomous production release",
    ):
        if marker not in combined:
            validation.error(relative, f"missing VOC-002 transition marker: {marker}")


def validate_voc_003_package(validation: Validation) -> None:
    relative = VOC003_PATH
    require_complete_directory(validation, relative)
    values = validate_restricted_yaml(validation, f"{relative}/change.yaml")
    expected = {
        "schema_version": "1",
        "id": "VOC-003",
        "slug": "a003-lifecycle-sync",
        "title": "A-003 Lifecycle State Synchronization",
        "type": "governance",
        "status": "implementing",
        "risk": "R4",
        "canonical_path": relative,
        "base_sha": "9d5b4bc1d4a72e313b013047601265ee837c34f2",
        "authority_model": "a003-active",
        "post_activation_sync": "true",
        "new_activation_event": "false",
        "automatic_merge": "false",
        "autonomous_merge": "false",
        "rl1_technical_activation": "false",
        "rl2_technical_activation": "false",
        "production_deployment": "false",
        "autonomous_production_release": "disabled",
        "doc_17_repository_adoption": "false",
        "doc_18_repository_adoption": "false",
        "control_plane_implementation": "false",
    }
    for key, value in expected.items():
        if values.get(key) != value:
            validation.error(f"{relative}/change.yaml", f"{key} must equal {value!r}")
    combined = "\n".join(validation.read(f"{relative}/{name}") for name in PACKAGE_FILES)
    for marker in (
        "post-activation canonical synchronization",
        "c858ebff3d97da88fea830bc32a74f69f59a9ad2",
        "9d5b4bc1d4a72e313b013047601265ee837c34f2",
        "2026-07-17T16:44:34Z",
        "R4",
        "exhausted",
        "DOC-17",
        "DOC-18",
        "Control Plane",
        "autonomous production release",
    ):
        if marker not in combined:
            validation.error(relative, f"missing VOC-003 synchronization marker: {marker}")


def validate_voc_004_package(validation: Validation) -> None:
    relative = VOC004_PATH
    require_complete_directory(validation, relative)
    values = validate_restricted_yaml(validation, f"{relative}/change.yaml")
    expected = {
        "schema_version": "1",
        "id": "VOC-004",
        "slug": "canonical-adoption-doc-17-doc-18",
        "title": "Canonical Adoption of DOC-17 and DOC-18",
        "type": "governance",
        "status": "completed",
        "risk": "R4",
        "canonical_path": relative,
        "base_branch": "develop",
        "base_sha": "873038735aea30b754a8c57b3522e1ff41f6d89c",
        "authority_model": "a003-active",
        "approved_candidate_sha": "89013e6a8fab4cee45935e700d9eb3e49d3d39ed",
        "independent_verification_status": "passed-exact-revision-with-non-blocking-findings",
        "independent_verification_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11#issuecomment-5007950942",
        "founder_r4_approval_status": "approved-exact-revision",
        "founder_r4_approval_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11#issuecomment-5007966020",
        "repository_adoption_status": "adopted",
        "adoption_pr": "11",
        "adopted_develop_sha": "2b5ecb19b532a9b23250e1255ff1e7fb9a78ef77",
        "repository_adoption_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11",
        "canonical_lifecycle_sync_status": "complete",
        "doc_17_source_sha256": DOC17_SOURCE_SHA256,
        "doc_18_source_sha256": DOC18_SOURCE_SHA256,
        "doc_17_repository_adoption": "true",
        "doc_18_repository_adoption": "true",
        "control_plane_implementation": "false",
        "rl1_technical_activation": "false",
        "rl2_technical_activation": "false",
        "automatic_merge": "false",
        "autonomous_merge": "false",
        "production_deployment": "disabled",
        "autonomous_production_release": "disabled",
        "ehr": "not-triggered",
        "standing_technical_steward_approval": "not-applicable",
    }
    for key, value in expected.items():
        if values.get(key) != value:
            validation.error(f"{relative}/change.yaml", f"{key} must equal {value!r}")

    package = {
        name: validation.read(f"{relative}/{name}")
        for name in PACKAGE_FILES
        if name.endswith(".md")
    }
    patterns = {
        "specification.md": re.compile(r"^- \*\*(VOC-004-R\d+):?\*\*", re.MULTILINE),
        "acceptance-criteria.md": re.compile(r"^## (VOC-004-AC-\d+)", re.MULTILINE),
        "impact-analysis.md": re.compile(r"^## (VOC-004-IMP-\d+)", re.MULTILINE),
        "tasks.md": re.compile(r"^## (VOC-004-T\d+)", re.MULTILINE),
        "test-plan.md": re.compile(r"^## (VOC-004-TEST-\d+)", re.MULTILINE),
    }
    definitions: list[str] = []
    for name, pattern in patterns.items():
        definitions.extend(pattern.findall(package[name]))
    duplicates = sorted(item for item in set(definitions) if definitions.count(item) > 1)
    if duplicates:
        validation.error(relative, f"duplicate VOC-004 stable identifier definitions: {duplicates}")
    required = {
        *(f"VOC-004-R{i:02d}" for i in range(1, 13)),
        *(f"VOC-004-AC-{i:02d}" for i in range(1, 11)),
        *(f"VOC-004-IMP-{i:02d}" for i in range(1, 8)),
        *(f"VOC-004-T{i:02d}" for i in range(1, 11)),
        *(f"VOC-004-TEST-{i:02d}" for i in range(1, 11)),
    }
    missing = sorted(required - set(definitions))
    if missing:
        validation.error(relative, f"missing VOC-004 stable identifier definitions: {missing}")
    combined = "\n".join(package.values())
    references = set(
        re.findall(r"VOC-004-(?:R\d+|AC-\d+|IMP-\d+|T\d+|TEST-\d+)", combined)
    )
    unresolved = sorted(references - set(definitions))
    if unresolved:
        validation.error(relative, f"unresolved VOC-004 stable identifier references: {unresolved}")
    for marker in (
        "/home/mehrdad/project/vocanova-source/DOC-17-vocanova-autonomous-development-architecture-v1.md",
        "/home/mehrdad/project/vocanova-source/DOC-18-vocanova-autonomous-development-implementation-roadmap.md",
        DOC17_SOURCE_SHA256,
        DOC18_SOURCE_SHA256,
        DOC17_PATH,
        DOC18_PATH,
        "R4",
        "exact-SHA Claude Code",
        "exact-SHA founder R4 approval",
        "EHR is not triggered",
        "standing technical-steward approval",
        "exhausted",
        "VocaNova MVP",
        "89013e6a8fab4cee45935e700d9eb3e49d3d39ed",
        "2b5ecb19b532a9b23250e1255ff1e7fb9a78ef77",
        "PASS WITH NON-BLOCKING FINDINGS",
        "PR #11",
    ):
        if marker not in combined:
            validation.error(relative, f"missing VOC-004 adoption marker: {marker}")


def frontmatter_values(text: str) -> dict[str, str]:
    if not text.startswith("---\n") or "\n---\n" not in text[4:]:
        return {}
    frontmatter = text[4:].split("\n---\n", 1)[0]
    values: dict[str, str] = {}
    for line in frontmatter.splitlines():
        match = re.match(r"^([A-Za-z_][A-Za-z0-9_-]*):\s*(.*)$", line)
        if match:
            values[match.group(1)] = match.group(2).strip().strip("\"'")
    return values


def validate_doc_17_doc_18_adoption(validation: Validation) -> None:
    expected_documents = {
        DOC17_PATH: {
            "id": "DOC-17",
            "status": "approved",
            "canonical_path": DOC17_PATH,
            "founder_direction_status": "approved",
            "formal_repository_approval_status": "approved-exact-revision",
            "repository_adoption_status": "adopted",
            "technical_activation_status": "inactive",
            "frozen_source_sha256": DOC17_SOURCE_SHA256,
            "frozen_substantive_body_sha256": DOC17_BODY_SHA256,
            "adoption_change": "VOC-004",
            "approved_candidate_sha": "89013e6a8fab4cee45935e700d9eb3e49d3d39ed",
            "independent_verification_status": "passed-exact-revision-with-non-blocking-findings",
            "independent_verification_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11#issuecomment-5007950942",
            "founder_r4_approval_status": "approved-exact-revision",
            "founder_r4_approval_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11#issuecomment-5007966020",
            "adoption_pr": "11",
            "adopted_develop_sha": "2b5ecb19b532a9b23250e1255ff1e7fb9a78ef77",
            "repository_adoption_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11",
        },
        DOC18_PATH: {
            "id": "DOC-18",
            "status": "approved",
            "canonical_path": DOC18_PATH,
            "founder_direction_status": "approved",
            "formal_repository_approval_status": "approved-exact-revision",
            "repository_adoption_status": "adopted",
            "technical_activation_status": "inactive",
            "frozen_source_sha256": DOC18_SOURCE_SHA256,
            "frozen_substantive_body_sha256": DOC18_BODY_SHA256,
            "adoption_change": "VOC-004",
            "approved_candidate_sha": "89013e6a8fab4cee45935e700d9eb3e49d3d39ed",
            "independent_verification_status": "passed-exact-revision-with-non-blocking-findings",
            "independent_verification_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11#issuecomment-5007950942",
            "founder_r4_approval_status": "approved-exact-revision",
            "founder_r4_approval_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11#issuecomment-5007966020",
            "adoption_pr": "11",
            "adopted_develop_sha": "2b5ecb19b532a9b23250e1255ff1e7fb9a78ef77",
            "repository_adoption_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/11",
        },
    }
    body_hashes = {DOC17_PATH: DOC17_BODY_SHA256, DOC18_PATH: DOC18_BODY_SHA256}
    for relative, expected in expected_documents.items():
        text = validation.read(relative)
        metadata = frontmatter_values(text)
        for key, value in expected.items():
            if metadata.get(key) != value:
                validation.error(relative, f"canonical adoption requires {key}: {value}")
        body = text.split("---", 2)[2] if text.count("---") >= 2 else ""
        if hashlib.sha256(body.encode("utf-8")).hexdigest() != body_hashes[relative]:
            validation.error(relative, "frozen substantive body checksum mismatch")

    architecture_index = validation.read("docs/architecture/README.md")
    planning_index = validation.read("docs/planning/README.md")
    root_index = validation.read("docs/README.md")
    specs_index = validation.read("specs/README.md")
    for relative, text, marker in (
        ("docs/architecture/README.md", architecture_index, "17-autonomous-development-architecture.md"),
        ("docs/planning/README.md", planning_index, "18-autonomous-development-implementation-roadmap.md"),
        ("docs/README.md", root_index, "DOC-17 and DOC-18 are adopted together"),
        ("specs/README.md", specs_index, "VOC-004 — Canonical Adoption of DOC-17 and DOC-18"),
    ):
        if marker not in text:
            validation.error(relative, f"missing canonical adoption index marker: {marker}")


def validate_a003_lifecycle(validation: Validation) -> None:
    amendment = validation.read(A003_PATH)
    state = validate_restricted_yaml(validation, A003_STATE_PATH)
    metadata = frontmatter_values(amendment)
    full_sha = hashlib.sha256(amendment.encode("utf-8")).hexdigest()
    body = amendment.split("---", 2)[2] if amendment.count("---") >= 2 else ""
    body_sha = hashlib.sha256(body.encode("utf-8")).hexdigest()

    if body_sha != A003_FROZEN_BODY_SHA256:
        validation.error(A003_PATH, "frozen A-003 substantive body checksum mismatch")
    if state.get("frozen_source_sha256") != A003_FROZEN_SHA256:
        validation.error(A003_STATE_PATH, "frozen A-003 source checksum identifier is missing or changed")

    active = state.get("effective_activation_status") == "active" or metadata.get("effective_activation_status") == "active"
    if not active:
        if full_sha != A003_FROZEN_SHA256:
            validation.error(A003_PATH, "inactive adoption candidate must match the exact frozen A-003 source")
        expected_inactive = {
            "authority_model": "pre-a003",
            "transition_stage": "pre-merge-transition",
            "formal_founder_approval_status": "pending-exact-revision-github-evidence",
            "technical_steward_migration_approval_status": "pending-exact-revision-github-evidence",
            "independent_verification_status": "pending-exact-revision-claude-evidence",
            "repository_adoption_status": "pending",
            "effective_activation_status": "inactive",
            "approved_pr_head_sha": "null",
            "adopted_develop_sha": "null",
            "post_merge_validation_status": "not-run",
            "activation_evidence": "null",
            "migration_approval_status": "pending-one-time-use",
            "migration_approval_exhausted": "false",
            "technical_steward_routine_authority_status": "current-until-valid-activation",
            "exceptional_human_review_mode": "exceptional-only-after-valid-activation",
        }
        for key, value in expected_inactive.items():
            if state.get(key) != value:
                validation.error(A003_STATE_PATH, f"inactive transition requires {key}: {value}")
        for key, value in {
            "status": "proposed",
            "formal_founder_approval_status": "pending-exact-revision-github-evidence",
            "repository_adoption_status": "pending",
            "effective_activation_status": "inactive",
            "approved_at": "null",
            "adopted_at": "null",
            "effective_at": "null",
            "approval_evidence": "null",
        }.items():
            if metadata.get(key) != value:
                validation.error(A003_PATH, f"pre-merge A-003 metadata requires {key}: {value}")
    else:
        required_active = {
            "authority_model": "a003-active",
            "transition_stage": "effectively-active",
            "formal_founder_approval_status": "approved-exact-revision",
            "technical_steward_migration_approval_status": "approved-exact-revision-one-time",
            "independent_verification_status": "passed-exact-revision",
            "repository_adoption_status": "adopted",
            "effective_activation_status": "active",
            "post_merge_validation_status": "passed",
            "migration_approval_status": "exhausted-non-reusable",
            "migration_approval_exhausted": "true",
            "technical_steward_routine_authority_status": "historical-retired",
            "exceptional_human_review_mode": "exceptional-only",
            "canonical_lifecycle_sync_status": "complete",
        }
        for key, value in required_active.items():
            if state.get(key) != value:
                validation.error(A003_STATE_PATH, f"active A-003 requires {key}: {value}")
        for key in ("approved_pr_head_sha", "adopted_develop_sha"):
            if not re.fullmatch(r"[0-9a-f]{40}", state.get(key, "")):
                validation.error(A003_STATE_PATH, f"active A-003 requires a full {key}")
        if state.get("approved_pr_head_sha") == state.get("adopted_develop_sha"):
            validation.error(A003_STATE_PATH, "approved PR head SHA and adopted develop SHA must be distinct records")
        exact_active_state = {
            "approved_pr_head_sha": "c858ebff3d97da88fea830bc32a74f69f59a9ad2",
            "adopted_develop_sha": "9d5b4bc1d4a72e313b013047601265ee837c34f2",
            "approved_adopted_tree_sha": "07ef24cbb8602f540600dcea551306ed51a6215f",
            "formal_founder_approval_at": "2026-07-17T16:37:38Z",
            "formal_founder_approval_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005389067",
            "technical_steward_migration_approval_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005389067",
            "independent_verification_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005293621",
            "repository_adoption_at": "2026-07-17T16:41:32Z",
            "repository_adoption_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005429197",
            "effective_activation_at": "2026-07-17T16:44:34Z",
            "post_merge_validation_evidence": "https://github.com/KARSIFT/vocanova-platform/actions/runs/29597154713",
            "activation_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005456622",
        }
        for key, value in exact_active_state.items():
            if state.get(key) != value:
                validation.error(A003_STATE_PATH, f"active A-003 requires exact {key}: {value}")
        active_metadata = {
            "status": "approved",
            "formal_founder_approval_status": "approved-exact-revision-github-evidence",
            "repository_adoption_status": "adopted",
            "effective_activation_status": "active",
        }
        for key, value in active_metadata.items():
            if metadata.get(key) != value:
                validation.error(A003_PATH, f"active transition requires synchronized {key}: {value}")
        exact_metadata = {
            "approved_at": "2026-07-17T16:37:38Z",
            "adopted_at": "2026-07-17T16:41:32Z",
            "effective_at": "2026-07-17T16:44:34Z",
            "approved_pr_head_sha": "c858ebff3d97da88fea830bc32a74f69f59a9ad2",
            "adopted_develop_sha": "9d5b4bc1d4a72e313b013047601265ee837c34f2",
            "approval_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005389067",
            "independent_verification_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005293621",
            "repository_adoption_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005429197",
            "activation_evidence": "https://github.com/KARSIFT/vocanova-platform/pull/8#issuecomment-5005456622",
        }
        for key, value in exact_metadata.items():
            if metadata.get(key) != value:
                validation.error(A003_PATH, f"active transition requires exact {key}: {value}")

    for key in (
        "rl1_technical_activation",
        "rl2_technical_activation",
        "control_plane_implementation",
    ):
        if state.get(key) != "false":
            validation.error(A003_STATE_PATH, f"{key} must remain false")
    for key in ("doc_17_repository_adoption", "doc_18_repository_adoption"):
        if state.get(key) != "true":
            validation.error(A003_STATE_PATH, f"{key} must be true for atomic VOC-004 adoption")

    # Automatic merge/release/production-deployment authority (A-003 SS10-12) is a
    # hard, unconditional invariant UNLESS the file also carries a specific,
    # dated authorization marker - not just the boolean flip. This preserves the
    # original fail-closed tripwire (a silent/accidental flip of just the
    # boolean still fails validation) while allowing the founder's explicit,
    # twice-confirmed 2026-08-08 decision to actually take effect. See
    # AGENTS.md's "Release and deployment authority" section for the record of
    # that decision - this check requires the same marker text to be present
    # here, not just asserted in a doc elsewhere.
    autonomy_authorized = "AUTONOMOUS-RELEASE-AUTHORIZED-2026-08-08" in validation.read(A003_STATE_PATH)
    merge_release_defaults = {
        "automatic_merge_allowed": "false",
        "autonomous_merge_allowed": "false",
        "production_deployment": "disabled",
        "autonomous_production_release": "disabled",
    }
    merge_release_authorized = {
        "automatic_merge_allowed": "true",
        "autonomous_merge_allowed": "true",
        "production_deployment": "enabled",
        "autonomous_production_release": "enabled",
    }
    for key, default in merge_release_defaults.items():
        current = state.get(key)
        if autonomy_authorized:
            if current != merge_release_authorized[key]:
                validation.error(A003_STATE_PATH, f"{key} must equal {merge_release_authorized[key]!r} once authorized")
        elif current != default:
            validation.error(A003_STATE_PATH, f"{key} must remain {default!r} without an authorization marker")

    appointment = validation.read("docs/governance/technical-steward-appointment.md")
    for marker in (
        "Appointed qualified human technical steward: `@m-e-h-r-d-a-a-d`",
        "same verified human presently serves in two explicitly separate",
        "permanent audit history",
        "one-time VOC-002 approval is exhausted and is not reusable",
    ):
        if marker not in appointment:
            validation.error("docs/governance/technical-steward-appointment.md", f"missing historical evidence marker: {marker}")

    authority = validation.read("docs/governance/approval-matrix.md")
    for marker in (
        "No standing technical-steward approval; no founder approval merely because work is R3",
        "R4 founder authority remains unchanged",
        "EHR",
        "must never be reused",
        "CODEOWNERS remains review routing and is not approval evidence",
    ):
        if marker not in authority:
            validation.error("docs/governance/approval-matrix.md", f"missing A-003 authority marker: {marker}")


def validate_a004_lifecycle(validation: Validation) -> None:
    amendment = validation.read(A004_PATH)
    state = validate_restricted_yaml(validation, A004_STATE_PATH)
    metadata = frontmatter_values(amendment)
    full_sha = hashlib.sha256(amendment.encode("utf-8")).hexdigest()

    active = state.get("effective_activation_status") == "active" or metadata.get("effective_activation_status") == "active"
    if not active:
        expected_inactive = {
            "authority_model": "a003-active",
            "transition_stage": "pre-activation-scaffolding",
            "formal_founder_transition_approval_status": "pending-exact-revision-github-evidence",
            "independent_verification_status": "pending-exact-revision",
            "repository_adoption_status": "pending",
            "effective_activation_status": "inactive",
            "migration_approval_status": "pending-one-time-transition",
            "migration_approval_exhausted": "false",
        }
        for key, value in expected_inactive.items():
            if state.get(key) != value:
                validation.error(A004_STATE_PATH, f"inactive A-004 transition requires {key}: {value}")
        if metadata.get("status") != "proposed":
            validation.error(A004_PATH, "inactive A-004 metadata requires status: proposed")
        return

    frozen = state.get("frozen_source_sha256")
    if frozen and frozen != full_sha:
        validation.error(A004_STATE_PATH, "frozen A-004 source checksum mismatch")

    required_active = {
        "authority_model": "a004-active",
        "transition_stage": "effectively-active",
        "formal_founder_transition_approval_status": "approved-exact-revision",
        "independent_verification_status": "passed-exact-revision",
        "repository_adoption_status": "adopted",
        "effective_activation_status": "active",
        "migration_approval_status": "exhausted-non-reusable",
        "migration_approval_exhausted": "true",
        "rehearsal_evidence_status": "complete",
        "exceptional_human_review_mode": "exceptional-only",
    }
    for key, value in required_active.items():
        if state.get(key) != value:
            validation.error(A004_STATE_PATH, f"active A-004 requires {key}: {value}")

    if not re.fullmatch(r"[0-9a-f]{40}", state.get("adopted_develop_sha", "")):
        validation.error(A004_STATE_PATH, "active A-004 requires adopted_develop_sha")

    approved_pr = state.get("approved_pr_head_sha")
    if approved_pr and approved_pr != "null" and not re.fullmatch(r"[0-9a-f]{40}", approved_pr):
        validation.error(A004_STATE_PATH, "approved_pr_head_sha must be a full SHA or null until bound")

    active_metadata = {
        "status": "approved",
        "formal_founder_approval_status": "approved-exact-revision-github-evidence",
        "repository_adoption_status": "adopted",
        "effective_activation_status": "active",
    }
    for key, value in active_metadata.items():
        if metadata.get(key) != value:
            validation.error(A004_PATH, f"active A-004 metadata requires {key}: {value}")

    if metadata.get("adopted_develop_sha") != state.get("adopted_develop_sha"):
        validation.error(A004_PATH, "A-004 amendment adopted_develop_sha must match transition state")

    for key in ("rl1_technical_activation", "rl2_technical_activation", "control_plane_implementation"):
        if state.get(key) != "false":
            validation.error(A004_STATE_PATH, f"{key} must remain false")
    for key in ("doc_17_repository_adoption", "doc_18_repository_adoption"):
        if state.get(key) != "true":
            validation.error(A004_STATE_PATH, f"{key} must be true")

    if "AUTONOMOUS-RELEASE-AUTHORIZED-2026-08-08" not in validation.read(A003_STATE_PATH):
        validation.error(A004_STATE_PATH, "A-004 activation requires A-003 AUTONOMOUS-RELEASE-AUTHORIZED marker")

    for key in (
        "formal_founder_transition_approval_evidence",
        "independent_verification_evidence",
        "repository_adoption_evidence",
        "activation_evidence",
        "rehearsal_evidence",
        "effective_activation_at",
    ):
        if not state.get(key) or state.get(key) == "null":
            validation.error(A004_STATE_PATH, f"active A-004 requires {key}")

    authority = validation.read("docs/governance/approval-matrix.md")
    for marker in (
        "A-004 is effective",
        "no founder `approved` comment",
        "R4 remains a strengthened evidence class",
    ):
        if marker not in authority:
            validation.error("docs/governance/approval-matrix.md", f"missing A-004 authority marker: {marker}")


def validate_ownership(validation: Validation) -> None:
    policy_path = ".github/approved-policy/protected-paths.yaml"
    policy_values = validate_restricted_yaml(validation, policy_path)
    policy = validation.read(policy_path)
    a004_state = validate_restricted_yaml(validation, A004_STATE_PATH)
    a004_active = a004_state.get("effective_activation_status") == "active"
    expected_policy_state = {
        "status": "approved-a004-active" if a004_active else "approved-a003-active",
        "authority_model": "a004-active" if a004_active else "a003-active",
        "hosted_enforcement_status": "not-activated",
        "rl1_technical_activation": "false",
        "rl2_technical_activation": "false",
        "doc_17_repository_adoption": "true",
        "doc_18_repository_adoption": "true",
        "control_plane_implementation": "false",
    }
    for key, value in expected_policy_state.items():
        if policy_values.get(key) != value:
            validation.error(policy_path, f"canonical protected policy requires {key}: {value}")

    # Same authorization-marker gate as validate_a003 applies here - this file
    # mirrors docs/governance/a003-transition-state.yaml's merge/release/
    # deployment fields and must move in lockstep with it, never drift apart.
    autonomy_authorized = "AUTONOMOUS-RELEASE-AUTHORIZED-2026-08-08" in policy
    merge_release_defaults = {
        "automatic_merge_allowed": "false",
        "autonomous_merge_allowed": "false",
        "production_deployment": "disabled",
        "autonomous_production_release": "disabled",
    }
    merge_release_authorized = {
        "automatic_merge_allowed": "true",
        "autonomous_merge_allowed": "true",
        "production_deployment": "enabled",
        "autonomous_production_release": "enabled",
    }
    for key, default in merge_release_defaults.items():
        current = policy_values.get(key)
        if autonomy_authorized:
            if current != merge_release_authorized[key]:
                validation.error(policy_path, f"{key} must equal {merge_release_authorized[key]!r} once authorized")
        elif current != default:
            validation.error(policy_path, f"{key} must remain {default!r} without an authorization marker")
    owners = validation.read(".github/CODEOWNERS")
    listed = set(re.findall(r"^\s*-\s+path:\s*([^\s#]+)", policy, re.MULTILINE))
    for path in PROTECTED_PATHS:
        if path not in listed:
            validation.error(policy_path, f"missing protected path: {path}")
        owner_path = "/" + path
        if not any(line.split() and line.split()[0] == owner_path for line in owners.splitlines() if not line.lstrip().startswith("#")):
            validation.error(".github/CODEOWNERS", f"missing exact protected path owner: {owner_path}")
    active = "\n".join(line for line in owners.splitlines() if not line.lstrip().startswith("#"))
    if "@KARSIFT/" in active:
        validation.error(".github/CODEOWNERS", "invented or unverified governance team is prohibited")
    if re.search(r"@\S*(?:codex|claude|bot|automation)\S*", active, re.IGNORECASE):
        validation.error(".github/CODEOWNERS", "AI or bot identity cannot be a technical-steward owner")
    for line in active.splitlines():
        if line.strip() and "@m-e-h-r-d-a-a-d" not in line:
            validation.error(".github/CODEOWNERS", f"owner must route to @m-e-h-r-d-a-a-d: {line}")


def validate_workflow(validation: Validation) -> None:
    relative = ".github/workflows/repository-governance.yml"
    text = validation.read(relative)
    if not re.search(r"^name:\s*Repository Governance\s*$", text, re.MULTILINE):
        validation.error(relative, "workflow display name must be Repository Governance")
    for marker in ("pull_request:", "push:", "- develop", "- main", "contents: read", "timeout-minutes:"):
        if marker not in text:
            validation.error(relative, f"missing workflow control: {marker}")
    for prohibited in ("pull_request_target", "paths:", "paths-ignore:", "contents: write", "secrets.", "codex", "claude"):
        if prohibited in text.lower():
            validation.error(relative, f"prohibited workflow construct: {prohibited}")
    for action in re.findall(r"^\s*uses:\s*([^\s#]+)", text, re.MULTILINE):
        if not re.fullmatch(r"[^@]+@[0-9a-f]{40}", action):
            validation.error(relative, f"external action is not pinned to a full immutable SHA: {action}")


def validate_governance_language(validation: Validation) -> None:
    agents = validation.read("AGENTS.md")
    exact_chatgpt = (
        "ChatGPT may receive read-only access to KARSIFT/vocanova-platform for\n"
        "repository-grounded product analysis, architecture analysis, specification\n"
        "drafting, and cross-document impact analysis. ChatGPT must not receive\n"
        "repository write, merge, deployment, secret, or production-data access."
    )
    if exact_chatgpt not in agents:
        validation.error("AGENTS.md", "missing exact approved ChatGPT read-only rule")
    combined = agents + "\n" + validation.read("CLAUDE.md")
    for marker in ("approved `VOC-###`", "approve or merge", "independently verifies", "R3", "technical steward", "R4", "founder", "GitHub is the canonical", "Prompt injection"):
        if marker.lower() not in combined.lower():
            validation.error("AGENTS.md", f"missing current R3/R4 governance language: {marker}")
    pr = validation.read(".github/pull_request_template.md")
    for marker in PR_MARKERS:
        if marker not in pr:
            validation.error(".github/pull_request_template.md", f"missing required field: {marker}")


def parse_change_yaml_top_level_field(text: str, field: str) -> str | None:
    match = re.search(rf"^{re.escape(field)}:\s*(\S+)\s*$", text, re.MULTILINE)
    if not match:
        return None
    return match.group(1).strip("\"'")


def validate_automatic_merge_drafting(validation: Validation) -> None:
    changes_root = validation.root / "specs/changes"
    if not changes_root.is_dir():
        validation.error("specs/changes/", "change-package directory is missing")
        return
    for change_yaml in sorted(changes_root.glob("*/change.yaml")):
        relative = change_yaml.relative_to(validation.root).as_posix()
        text = validation.read(relative)
        if parse_change_yaml_top_level_field(text, "automatic_merge_allowed") != "false":
            continue
        risk = parse_change_yaml_top_level_field(text, "risk") or ""
        if risk == "R4":
            continue
        package_id = parse_change_yaml_top_level_field(text, "id") or ""
        if package_id in VOC075_HISTORICAL_NON_R4_FALSE_PACKAGE_IDS:
            continue
        validation.error(
            relative,
            "automatic_merge_allowed: false requires risk: R4 (approve-only-R4; "
            f"see AGENTS.md and VOC-075 / issue #573); found risk: {risk or 'missing'!r}",
        )


def validate_false_activation(validation: Validation) -> None:
    # AUTONOMOUS-RELEASE-AUTHORIZED-2026-08-08: protected-paths.yaml is now
    # legitimately authorized to say automatic_merge_allowed: true and
    # autonomous_production_release: enabled (see that file's own marker
    # comment and validate_ownership's authorization check above, which
    # already enforces that the marker and the four merge/release/deploy
    # fields move together). Excluding it here would make this a check with
    # no teeth if authorization is ever removed without also fixing this
    # file, so instead: only skip the "true"/"enabled" patterns for this one
    # file, and only when the marker is actually present - "Status: Activated"
    # stays banned everywhere unconditionally, since nothing in this
    # authorization concerns hosted-governance activation.
    authorized = "AUTONOMOUS-RELEASE-AUTHORIZED-2026-08-08" in validation.read(
        ".github/approved-policy/protected-paths.yaml"
    )
    paths = (
        ".github/approved-policy/protected-paths.yaml",
        "docs/governance/post-merge-activation-checklist.md",
        "specs/changes/VOC-001-repository-foundation/change.yaml",
    )
    for relative in paths:
        text = validation.read(relative)
        patterns = [r"(?im)^Status:\s*Activated\s*$"]
        if not (authorized and relative == ".github/approved-policy/protected-paths.yaml"):
            patterns += [r"automatic_merge(?:_allowed)?:\s*true", r"autonomous_production_release:\s*enabled"]
        for pattern in patterns:
            if re.search(pattern, text):
                validation.error(relative, "false claim that hosted governance or autonomous release is activated")


def validate_repository(root: Path) -> list[str]:
    validation = Validation(root)
    if not root.is_dir():
        return [f"{root}: repository root is not a directory"]
    for relative in REQUIRED_FILES:
        path = root / relative
        if not path.is_file():
            validation.error(relative, "required repository-foundation file is missing")
        elif path.stat().st_size == 0:
            validation.error(relative, "required repository-foundation file is empty")
    if (root / "decisions").exists():
        validation.error("decisions/", "root decision directory is prohibited; use docs/decisions/")
    if (root / ".github/PULL_REQUEST_TEMPLATE.md").exists():
        validation.error(".github/PULL_REQUEST_TEMPLATE.md", "uppercase duplicate PR template is prohibited")
    validate_templates(validation)
    validate_package(validation)
    validate_voc_002_package(validation)
    validate_voc_003_package(validation)
    validate_voc_004_package(validation)
    validate_doc_17_doc_18_adoption(validation)
    validate_a003_lifecycle(validation)
    validate_a004_lifecycle(validation)
    validate_ownership(validation)
    validate_workflow(validation)
    validate_governance_language(validation)
    validate_automatic_merge_drafting(validation)
    validate_false_activation(validation)
    return sorted(set(validation.errors))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repository-root", required=True, type=Path)
    return parser


def main(argv: list[str] | None = None) -> int:
    try:
        args = build_parser().parse_args(argv)
        errors = validate_repository(args.repository_root.resolve())
    except SystemExit:
        raise
    except Exception as exc:  # fail closed on validator defects
        print(f"validator internal failure: {exc}", file=sys.stderr)
        return 2
    if errors:
        for error in errors:
            print(f"ERROR: {error}", file=sys.stderr)
        print(f"Repository foundation validation failed with {len(errors)} error(s).", file=sys.stderr)
        return 1
    print("Repository foundation validation passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
