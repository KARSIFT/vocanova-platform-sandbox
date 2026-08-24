"""VOC-117 role bindings and parameterized Cursor routing (caller fixture regressions)."""

from __future__ import annotations

import os
import re
import subprocess
import sys
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
        self.assertIn("historical only", header)

    def test_voc117_test_05_fixture_pin_is_recorded(self):
        pin = (FIXTURE_INFRA_ROOT / "PINNED_SHA.txt").read_text(encoding="utf-8").strip()
        self.assertRegex(pin, r"^[0-9a-f]{40}$")


if __name__ == "__main__":
    unittest.main()
