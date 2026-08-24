"""Regression coverage for package task roster policy (VOC-115)."""

from __future__ import annotations

from pathlib import Path
import sys
import unittest

ROOT = Path(__file__).resolve().parents[1]
CONFIG = ROOT / "config"
if str(CONFIG) not in sys.path:
    sys.path.insert(0, str(CONFIG))

from package_task_policy import (  # noqa: E402
    PackageTaskPolicyError,
    parse_task_sections,
    validate_package_tasks,
)


def _one_task_package() -> str:
    return """# VOC-900 — Tasks

## VOC-900-T00 — Implement the feature end to end

- Requirement source: issue #1
- Acceptance criteria: VOC-900-AC-00
- Tests: VOC-900-TEST-00
- Evidence: VOC-900-EV-00
"""


def _multi_skill_one_task() -> str:
    return """# VOC-901 — Tasks

## VOC-901-T00 — Add related agent skills with configuration, docs, and tests

Deliver several related skills, their adapters, configuration, documentation, and
regression tests in one end-to-end implementation carrier.
"""


def _justified_multi_task() -> str:
    return """# VOC-902 — Tasks

## VOC-902-T00 — Land schema migration first

Schema-only migration that must merge before runtime code.

## VOC-902-T01 — Implement runtime behavior

- Split reason: merge-order-dependency — runtime code depends on the merged schema from T00.
"""


def _invalid_missing_reason() -> str:
    return """# VOC-903 — Tasks

## VOC-903-T00 — First task

## VOC-903-T01 — Second task without a split reason
"""


def _invalid_small_reason() -> str:
    return """# VOC-903 — Tasks

## VOC-903-T00 — First task

## VOC-903-T01 — Second task

- Split reason: small — keep tasks easy to review.
"""


def _invalid_docs_vs_code() -> str:
    return """# VOC-903 — Tasks

## VOC-903-T00 — First task

## VOC-903-T01 — Docs only

- Split reason: docs-vs-code
"""


def _four_tasks_without_justification() -> str:
    return """# VOC-904 — Tasks

## VOC-904-T00 — First

## VOC-904-T01 — Second

- Split reason: merge-order-dependency — T01 follows T00.

## VOC-904-T02 — Third

- Split reason: independently-releasable-unit — can roll back T02 alone.

## VOC-904-T03 — Fourth

- Split reason: mutually-exclusive-execution-environment — runs only in staging.
"""


def _four_tasks_with_justification() -> str:
    return """# VOC-904 — Tasks

## Package-level multi-task justification

Four tasks remain necessary because each boundary isolates a distinct merge-order,
rollback, and execution-environment concern that cannot honestly share one carrier.

## VOC-904-T00 — First

## VOC-904-T01 — Second

- Split reason: merge-order-dependency — T01 follows T00.

## VOC-904-T02 — Third

- Split reason: independently-releasable-unit — can roll back T02 alone.

## VOC-904-T03 — Fourth

- Split reason: mutually-exclusive-execution-environment — runs only in staging.
"""


class PackageTaskPolicyTests(unittest.TestCase):
    def test_one_coherent_request_defaults_to_one_task(self):
        sections = validate_package_tasks(_one_task_package(), "VOC-900")
        self.assertEqual(len(sections), 1)
        self.assertEqual(sections[0].task_id, "VOC-900-T00")

    def test_related_skills_remain_one_task(self):
        sections = validate_package_tasks(_multi_skill_one_task(), "VOC-901")
        self.assertEqual(len(sections), 1)

    def test_code_tests_docs_stay_together_by_default(self):
        text = _one_task_package().replace(
            "Implement the feature end to end",
            "Implement API, tests, docs, and evidence together",
        )
        sections = validate_package_tasks(text, "VOC-900")
        self.assertEqual(len(sections), 1)

    def test_missing_split_reason_fails_closed(self):
        with self.assertRaises(PackageTaskPolicyError) as ctx:
            validate_package_tasks(_invalid_missing_reason(), "VOC-903")
        self.assertEqual(ctx.exception.code, "missing_split_reason")

    def test_invalid_split_reason_slug_fails_closed(self):
        with self.assertRaises(PackageTaskPolicyError) as ctx:
            validate_package_tasks(_invalid_small_reason(), "VOC-903")
        self.assertEqual(ctx.exception.code, "invalid_split_reason_slug")

        with self.assertRaises(PackageTaskPolicyError) as ctx:
            validate_package_tasks(_invalid_docs_vs_code(), "VOC-903")
        self.assertEqual(ctx.exception.code, "invalid_split_reason_slug")

    def test_more_than_three_tasks_requires_package_justification(self):
        with self.assertRaises(PackageTaskPolicyError) as ctx:
            validate_package_tasks(_four_tasks_without_justification(), "VOC-904")
        self.assertEqual(ctx.exception.code, "missing_multi_task_justification")

        sections = validate_package_tasks(_four_tasks_with_justification(), "VOC-904")
        self.assertEqual(len(sections), 4)

    def test_justified_multi_task_preserves_order(self):
        sections = validate_package_tasks(_justified_multi_task(), "VOC-902")
        self.assertEqual(
            [section.task_id for section in sections],
            ["VOC-902-T00", "VOC-902-T01"],
        )

    def test_cross_reference_headings_are_not_tasks(self):
        text = """# VOC-905 — Tasks

## VOC-905-T00 — Primary task

References VOC-034-T01 in prose only.

## VOC-905-T01 — Follow-up

- Split reason: merge-order-dependency — must follow T00 merge.
"""
        sections = parse_task_sections(text, "VOC-905")
        self.assertEqual([section.task_id for section in sections], ["VOC-905-T00", "VOC-905-T01"])


if __name__ == "__main__":
    unittest.main()
