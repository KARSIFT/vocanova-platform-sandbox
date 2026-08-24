"""VOC-117 role bindings and parameterized Cursor model routing regressions."""

from __future__ import annotations

import os
import re
import subprocess
import sys
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CONFIG = ROOT / "config"
if str(CONFIG) not in sys.path:
    sys.path.insert(0, str(CONFIG))

from prepare_cursor_model import (  # noqa: E402
    CursorModelError,
    prepare_cursor_model,
)

VOC117_BINDINGS = {
    "implementer": "cursor/composer-2.5",
    "implementer_escalation": "cursor/composer-2.5",
    "planner": "cursor/grok-4.6[fast=false]",
    "reviewer": "cursor/grok-4.6[fast=false]",
    "reviewer_fast_retry": "cursor/grok-4.6[fast=false]",
    "plan_reviewer": "cursor/grok-4.6[effort=high,fast=false]",
}


def active_role(config: str, role: str) -> str:
    matches = re.findall(rf"^{re.escape(role)}:\s*(\S+)\s*$", config, re.MULTILINE)
    if len(matches) != 1:
        raise AssertionError(f"expected one active {role} binding, found {matches}")
    return matches[0]


class Voc117RoleBindingsTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.roles = (ROOT / "config/roles.yml").read_text(encoding="utf-8")
        cls.plan = (ROOT / ".github/workflows/plan.yml").read_text(encoding="utf-8")
        cls.review = (ROOT / ".github/workflows/review.yml").read_text(encoding="utf-8")
        cls.plan_review = (
            ROOT / ".github/workflows/plan-review.yml"
        ).read_text(encoding="utf-8")
        cls.implement = (ROOT / ".github/workflows/implement.yml").read_text(
            encoding="utf-8"
        )

    def test_voc117_test_00_six_exact_role_bindings(self):
        for role, expected in VOC117_BINDINGS.items():
            with self.subTest(role=role):
                self.assertEqual(active_role(self.roles, role), expected)

    def test_voc117_test_00_no_active_openai_codex_binding(self):
        for line in self.roles.splitlines():
            stripped = line.strip()
            if stripped.startswith("#") or ":" not in stripped:
                continue
            key, value = stripped.split(":", 1)
            value = value.strip()
            if not value:
                continue
            self.assertFalse(
                value.startswith("openai/"),
                msg=f"active binding {key} must not use OpenAI/Codex: {value}",
            )

    def test_voc117_test_01_planner_parameterized_cursor_model(self):
        stored = VOC117_BINDINGS["planner"]
        self.assertEqual(prepare_cursor_model(stored), "grok-4.6[fast=false]")
        self.assertIn("prepare_cursor_model.py", self.plan)
        self.assertIn("--require-api-key", self.plan)

    def test_voc117_test_02_review_roles_use_grok_standard(self):
        for role in ("reviewer", "reviewer_fast_retry", "plan_reviewer"):
            stored = VOC117_BINDINGS[role]
            cli_model = prepare_cursor_model(stored)
            with self.subTest(role=role, cli_model=cli_model):
                self.assertTrue(cli_model.startswith("grok-4.6"))
                self.assertIn("fast=false", cli_model)
        plan_cli = prepare_cursor_model(VOC117_BINDINGS["plan_reviewer"])
        self.assertIn("effort=high", plan_cli)
        self.assertIn("prepare_cursor_model.py", self.review)
        self.assertIn("prepare_cursor_model.py", self.plan_review)

    def test_voc117_test_03_missing_api_key_fails_closed(self):
        env = os.environ.copy()
        env.pop("CURSOR_API_KEY", None)
        result = subprocess.run(
            [
                sys.executable,
                str(CONFIG / "prepare_cursor_model.py"),
                "--require-api-key",
                VOC117_BINDINGS["planner"],
            ],
            capture_output=True,
            text=True,
            env=env,
            check=False,
        )
        self.assertEqual(result.returncode, 1)
        self.assertIn("missing_cursor_api_key", result.stderr)
        self.assertNotIn("CURSOR_API_KEY", result.stdout + result.stderr)

    def test_voc117_test_03_unsupported_prefix_fails_closed(self):
        with self.assertRaises(CursorModelError) as ctx:
            prepare_cursor_model("openai/codex-action")
        self.assertEqual(str(ctx.exception), "unsupported_provider_prefix")

    def test_voc117_test_04_current_state_comments_not_dormant_routes(self):
        header = "\n".join(self.roles.splitlines()[:20])
        self.assertIn("VOC-117", header)
        self.assertIn("Grok 4.6 Standard", header)
        self.assertIn("historical only", header)
        self.assertNotRegex(
            header,
            r"(?i)openai/codex.*current|currently.*openai/codex",
        )

    def test_voc117_test_05_implementer_and_reviewer_remain_distinct(self):
        self.assertNotEqual(
            VOC117_BINDINGS["implementer"],
            VOC117_BINDINGS["reviewer"],
        )


if __name__ == "__main__":
    unittest.main()
