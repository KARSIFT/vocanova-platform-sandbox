from pathlib import Path
import os
import subprocess
import sys
import unittest


REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_INFRA_ROOT = REPOSITORY_ROOT / "tooling/governance/fixtures/karsift-ai-infra"


class ReadyForReviewReuseFixtureSuiteTests(unittest.TestCase):
    def test_pinned_shared_infra_policy_suite(self):
        env = os.environ.copy()
        env["PYTHONPATH"] = str(FIXTURE_INFRA_ROOT / "config")
        result = subprocess.run(
            [
                sys.executable,
                "-m",
                "unittest",
                "discover",
                "-s",
                str(FIXTURE_INFRA_ROOT / "tests"),
                "-p",
                "test_*.py",
                "-v",
            ],
            cwd=FIXTURE_INFRA_ROOT,
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(
            result.returncode,
            0,
            result.stderr or result.stdout or "pinned reuse policy suite failed",
        )


if __name__ == "__main__":
    unittest.main()
