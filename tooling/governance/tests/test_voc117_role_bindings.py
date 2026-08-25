"""VOC-117 role bindings and parameterized Cursor routing (caller fixture regressions)."""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture

VOC117_BINDINGS = {
    "implementer": "cursor/composer-2.5",
    "implementer_escalation": "cursor/composer-2.5",
    "planner": "cursor/grok-4.6[fast=false]",
    "reviewer": "cursor/grok-4.6[fast=false]",
    "reviewer_fast_retry": "cursor/grok-4.6[fast=false]",
    "plan_reviewer": "cursor/grok-4.6[effort=high,fast=false]",
}

FIXTURE_CONFIG = FIXTURE_INFRA_ROOT / "config"
if str(FIXTURE_CONFIG) not in sys.path:
    sys.path.insert(0, str(FIXTURE_CONFIG))

from prepare_cursor_model import CursorModelError, prepare_cursor_model  # noqa: E402


def active_role(config: str, role: str) -> str:
    matches = re.findall(rf"^{re.escape(role)}:\s*(\S+)\s*$", config, re.MULTILINE)
    if len(matches) != 1:
        raise AssertionError(f"expected one active {role} binding, found {matches}")
    return matches[0]


class Voc117RoleBindingsFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.roles = read_fixture("config/roles.yml")
        cls.plan = read_fixture(".github/workflows/plan.yml")
        cls.review = read_fixture(".github/workflows/review.yml")
        cls.plan_review = read_fixture(".github/workflows/plan-review.yml")
        cls.implement = read_fixture(".github/workflows/implement.yml")

    def test_voc117_test_00_fixture_has_six_exact_bindings(self):
        for role, expected in VOC117_BINDINGS.items():
            with self.subTest(role=role):
                self.assertEqual(active_role(self.roles, role), expected)

    def test_voc117_test_01_workflows_use_prepare_cursor_model(self):
        for workflow in (self.plan, self.review, self.plan_review, self.implement):
            self.assertIn("prepare_cursor_model.py", workflow)
            self.assertIn("--require-api-key", workflow)

    def test_voc117_test_02_plan_reviewer_retains_high_effort(self):
        cli = prepare_cursor_model(VOC117_BINDINGS["plan_reviewer"])
        self.assertIn("effort=high", cli)
        self.assertIn("fast=false", cli)

    def test_voc117_test_03_missing_api_key_fails_closed(self):
        env = os.environ.copy()
        env.pop("CURSOR_API_KEY", None)
        result = subprocess.run(
            [
                sys.executable,
                str(FIXTURE_CONFIG / "prepare_cursor_model.py"),
                "--require-api-key",
                VOC117_BINDINGS["reviewer"],
            ],
            capture_output=True,
            text=True,
            env=env,
            check=False,
        )
        self.assertEqual(result.returncode, 1)
        self.assertIn("missing_cursor_api_key", result.stderr)

    def test_voc117_test_03_unsupported_prefix_fails_closed(self):
        with self.assertRaises(CursorModelError):
            prepare_cursor_model("opencode-go/minimax-m3")

    def test_voc117_test_04_roles_header_describes_voc117_lineup(self):
        header = "\n".join(self.roles.splitlines()[:20])
        self.assertIn("VOC-117", header)
        self.assertIn("not current routing or fallback behavior", header)

    def test_voc117_test_05_fixture_pin_is_recorded(self):
        pin = (FIXTURE_INFRA_ROOT / "PINNED_SHA.txt").read_text(encoding="utf-8").strip()
        self.assertEqual(pin, "773bf7198aec0f5fcdff0f89d712cf14ef0a770e")

    def test_voc117_test_06_cursor_failures_are_sanitized_and_classified(self):
        extractor = read_fixture("config/extract-cursor-result.py")
        self.assertIn('return "model_parameter_invalid"', extractor)
        self.assertIn('return "model_unavailable_or_invalid"', extractor)
        self.assertIn("Raw provider output is withheld.", extractor)
        for workflow in (self.review, self.plan_review):
            self.assertIn("extract-cursor-result.py", workflow)
            self.assertIn("--github-annotation", workflow)
            self.assertNotIn("cat /tmp/cursor-stderr.log", workflow)

    def test_voc117_test_07_annotation_uses_stdout_without_raw_content(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch_path = Path(scratch)
            input_path = scratch_path / "response.json"
            output_path = scratch_path / "verdict.md"
            failure_record = scratch_path / "failure.json"
            provider_text = "Requested model is not available; secret-like tail"
            input_path.write_text(
                json.dumps(
                    {
                        "is_error": True,
                        "subtype": "error_during_execution",
                        "result": provider_text,
                    }
                ),
                encoding="utf-8",
            )
            completed = subprocess.run(
                [
                    sys.executable,
                    str(FIXTURE_CONFIG / "extract-cursor-result.py"),
                    str(input_path),
                    str(output_path),
                    "--allow-waiting",
                    "--github-annotation",
                    f"--failure-record={failure_record}",
                ],
                capture_output=True,
                text=True,
                check=False,
            )
            self.assertEqual(completed.returncode, 75)
            self.assertIn("reason=model_unavailable_or_invalid", completed.stdout)
            self.assertIn("Raw provider output is withheld.", completed.stdout)
            self.assertNotIn(provider_text, completed.stdout)
            self.assertEqual(completed.stderr, "")
            self.assertFalse(output_path.exists())
            self.assertEqual(
                json.loads(failure_record.read_text(encoding="utf-8")),
                {
                    "failure_reason": "model_unavailable_or_invalid",
                    "failure_subtype": "error_during_execution",
                    "schema_version": 1,
                },
            )

    def test_voc117_test_08_bounded_failure_publisher_is_exact_sha_isolated(self):
        builder = read_fixture("config/build-review-failure-comment.py")
        self.assertIn("SAFE_REASONS = {", builder)
        self.assertIn('"model_unavailable_or_invalid"', builder)
        self.assertIn('FAILURE_RECORD_KEYS = {"failure_reason", "failure_subtype", "schema_version"}', builder)
        self.assertIn("PR base/head pair changed before failure publication.", builder)
        for workflow in (self.review, self.plan_review):
            self.assertIn("--failure-record=/tmp/", workflow)
            self.assertIn("actions/upload-artifact@", workflow)
            self.assertIn("actions/download-artifact@", workflow)
            self.assertIn("retention-days: 1", workflow)
            self.assertNotIn("needs.review.outputs.failure_reason", workflow)
            self.assertNotIn("needs.plan-review.outputs.failure_reason", workflow)
            self.assertIn("build-review-failure-comment.py", workflow)
            self.assertIn("ref: ${{ job.workflow_sha }}", workflow)
            self.assertIn("permission-pull-requests: write", workflow)


if __name__ == "__main__":
    unittest.main()
