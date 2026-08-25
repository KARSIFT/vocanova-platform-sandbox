from __future__ import annotations

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
VALIDATOR = REPOSITORY_ROOT / "tooling/governance/validate_python_bytecode_hygiene.py"

sys.path.insert(0, str(REPOSITORY_ROOT / "tooling/governance"))
from validate_python_bytecode_hygiene import (  # noqa: E402
    find_tracked_bytecode_artifacts,
    missing_required_gitignore_patterns,
    parse_gitignore_patterns,
    validate_gitignore_file,
)


class PythonBytecodeHygieneTests(unittest.TestCase):
    def run_validator(self, repository_root: Path) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [sys.executable, str(VALIDATOR), "--repository-root", str(repository_root)],
            check=False,
            capture_output=True,
            text=True,
            timeout=20,
        )

    def init_repo(self, root: Path) -> None:
        subprocess.run(["git", "init"], cwd=root, check=True, capture_output=True)
        subprocess.run(
            ["git", "config", "user.email", "test@example.com"],
            cwd=root,
            check=True,
            capture_output=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test User"],
            cwd=root,
            check=True,
            capture_output=True,
        )

    def test_clean_synthetic_repository_passes(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / "repo"
            root.mkdir()
            self.init_repo(root)
            (root / ".gitignore").write_text(
                "__pycache__/\n*.py[cod]\n", encoding="utf-8"
            )
            (root / "README.md").write_text("ok\n", encoding="utf-8")
            subprocess.run(["git", "add", "."], cwd=root, check=True, capture_output=True)

            result = self.run_validator(root)
            self.assertEqual(0, result.returncode, result.stdout + result.stderr)
            self.assertIn("Python bytecode hygiene validation passed.", result.stdout)

    def test_caller_gitignore_satisfies_required_patterns(self) -> None:
        errors = validate_gitignore_file(REPOSITORY_ROOT / ".gitignore")
        self.assertEqual([], errors)

    def test_tracked_pyc_path_is_detected(self) -> None:
        violations = find_tracked_bytecode_artifacts(
            ["infra/scripts/__pycache__/example.cpython-312.pyc"]
        )
        self.assertEqual(
            ["infra/scripts/__pycache__/example.cpython-312.pyc"], violations
        )

    def test_tracked_bytecode_suffixes_are_detected(self) -> None:
        violations = find_tracked_bytecode_artifacts(
            ["build/module.pyc", "build/module.pyo", "build/module.pyd"]
        )
        self.assertEqual(
            sorted(["build/module.pyc", "build/module.pyo", "build/module.pyd"]),
            violations,
        )

    def test_tracked_bytecode_fails_closed(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / "repo"
            root.mkdir()
            self.init_repo(root)
            (root / ".gitignore").write_text(
                "__pycache__/\n*.py[cod]\n", encoding="utf-8"
            )
            artifact = root / "infra/scripts/__pycache__/tracked.cpython-312.pyc"
            artifact.parent.mkdir(parents=True)
            artifact.write_bytes(b"tracked-bytecode")
            subprocess.run(
                ["git", "add", "-f", str(artifact.relative_to(root))],
                cwd=root,
                check=True,
                capture_output=True,
            )

            result = self.run_validator(root)
            self.assertEqual(1, result.returncode, result.stdout + result.stderr)
            self.assertIn("tracked Python bytecode/cache artifacts remain", result.stderr)
            self.assertIn("__pycache__", result.stderr)

    def test_missing_gitignore_patterns_fail_closed(self) -> None:
        errors = missing_required_gitignore_patterns(parse_gitignore_patterns("node_modules/\n"))
        self.assertIn(
            "repository .gitignore is missing required pattern: __pycache__/",
            errors,
        )
        self.assertTrue(
            any("bytecode coverage" in error for error in errors),
            errors,
        )

    def test_incomplete_gitignore_fails_closed(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / "repo"
            root.mkdir()
            self.init_repo(root)
            (root / ".gitignore").write_text("node_modules/\n", encoding="utf-8")
            (root / "README.md").write_text("ok\n", encoding="utf-8")
            subprocess.run(["git", "add", "."], cwd=root, check=True, capture_output=True)

            result = self.run_validator(root)
            self.assertEqual(1, result.returncode, result.stdout + result.stderr)
            self.assertIn("missing required pattern: __pycache__/", result.stderr)

    def test_validate_gitignore_file_reports_missing_file(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            errors = validate_gitignore_file(root / ".gitignore")
            self.assertEqual([f"missing repository .gitignore: {root / '.gitignore'}"], errors)

    def test_working_tree_delete_does_not_hide_indexed_bytecode(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / "repo"
            root.mkdir()
            self.init_repo(root)
            (root / ".gitignore").write_text(
                "__pycache__/\n*.py[cod]\n", encoding="utf-8"
            )
            artifact = root / "infra/scripts/__pycache__/tracked.cpython-312.pyc"
            artifact.parent.mkdir(parents=True)
            artifact.write_bytes(b"tracked-bytecode")
            relative = str(artifact.relative_to(root))
            subprocess.run(["git", "add", "-f", relative], cwd=root, check=True, capture_output=True)
            artifact.unlink()

            result = self.run_validator(root)
            self.assertEqual(1, result.returncode, result.stdout + result.stderr)
            self.assertIn("tracked Python bytecode/cache artifacts remain", result.stderr)
            self.assertIn("__pycache__", result.stderr)


if __name__ == "__main__":
    unittest.main()
