from __future__ import annotations

import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
WORKFLOW = (ROOT / ".github/workflows/implement.yml").read_text(encoding="utf-8")
sys.path.insert(0, str(ROOT / "config"))

from implementer_source_carrier import (  # noqa: E402
    CarrierError,
    build_source_pr_body,
    nested_worktree_has_changes,
    validate_no_gitlink_paths,
)
from prepare_cursor_model import CursorModelError, prepare_cursor_model  # noqa: E402


class Voc121ImplementPolicyTests(unittest.TestCase):
    def test_workflow_preserves_helpers_before_nested_removal(self):
        self.assertIn("HELPER_DIR=/tmp/karsift-implement-helpers", WORKFLOW)
        self.assertIn(
            'cp karsift-ai-infra/config/prepare_cursor_model.py "$HELPER_DIR/prepare_cursor_model.py"',
            WORKFLOW,
        )
        self.assertIn("/tmp/karsift-implement-helpers/prepare_cursor_model.py", WORKFLOW)

    def test_workflow_bundles_nested_edits_before_removal(self):
        self.assertIn("git -C karsift-ai-infra bundle create /tmp/implementer-source.bundle", WORKFLOW)
        self.assertIn("has_source_changes=true", WORKFLOW)
        self.assertIn("publish-source:", WORKFLOW)

    def test_nested_gitlink_paths_are_rejected(self):
        with self.assertRaises(CarrierError):
            validate_no_gitlink_paths(["karsift-ai-infra"])

    def test_source_pr_body_is_non_closing_cross_repo_reference(self):
        body = build_source_pr_body(
            authority_repository="KARSIFT/vocanova-platform-sandbox",
            issue_number=996,
            change_id="VOC-121",
            task_id="VOC-121-T00",
            attempt=1,
        )
        self.assertIn("Relates to KARSIFT/vocanova-platform-sandbox#996.", body)
        self.assertNotRegex(body, r"(?i)\bcloses\b")

    def test_prepare_cursor_model_runs_from_preserved_copy_after_deletion(self):
        scratch = tempfile.TemporaryDirectory()
        helper_dir = Path(scratch.name) / "helpers"
        helper_dir.mkdir()
        helper_script = helper_dir / "prepare_cursor_model.py"
        helper_script.write_text(
            (ROOT / "config/prepare_cursor_model.py").read_text(encoding="utf-8"),
            encoding="utf-8",
        )
        env = os.environ.copy()
        env.pop("CURSOR_API_KEY", None)
        completed = subprocess.run(
            [
                sys.executable,
                str(helper_script),
                "--require-api-key",
                "cursor/composer-2.5",
            ],
            capture_output=True,
            text=True,
            check=False,
            env=env,
        )
        self.assertNotEqual(completed.returncode, 0)
        self.assertIn("missing_cursor_api_key", completed.stderr)
        self.assertNotIn("CURSOR_API_KEY", completed.stderr)

    def test_preserved_helper_still_validates_cursor_models(self):
        self.assertEqual(
            prepare_cursor_model("cursor/composer-2.5"),
            "composer-2.5",
        )
        with self.assertRaises(CursorModelError):
            prepare_cursor_model("opencode-go/foo")

    def test_nested_change_detection(self):
        self.assertTrue(nested_worktree_has_changes(" M config/foo.py\n"))
        self.assertFalse(nested_worktree_has_changes(""))

    def tearDown(self):
        if hasattr(self, "_scratch"):
            self._scratch.cleanup()


if __name__ == "__main__":
    unittest.main()
