"""VOC-139 caller regressions for promotion head-hash binding and metadata."""

from __future__ import annotations

import copy
import hashlib
import json
import os
import stat
import subprocess
import tempfile
import unittest
from pathlib import Path

import yaml

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
CALLER_WORKFLOWS = REPO_ROOT / ".github/workflows"
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-139-promotion-recovery-cannot-validate-an-accumulated/t00-evidence.md"
)

AUTHORITATIVE_PIN = "123735c80fec813a5b46a004f3e1122bd425cde2"
VOC139_INFRA_PIN = "599436835371f27fac52ec6b47a18b36257366ac"
CURRENT_PIN = "8993e867640dfb604dec0466c4e0787e68d8e258"
PROTECTED_COMPARISON_ANCHOR = "b9e74fc2db4691c48c637639b265d527de9f4505"
IMPLEMENTATION_PR_BASE = "ecb3d6d8e30628a9691928ea4594523f7193b961"
PROMOTION_BASE_SHA = "0d0b0cdf0692d0349f380e9cae3285b4c7916b05"
PROMOTION_HEAD_SHA = "4812fb91ab1b674f9a9ec03906f90c0edf50421d"

NO_CHANGE_PATHS = (
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
    "scripts/foundation/voc112-navigation-benchmark-run.mjs",
    "scripts/foundation/validate-workspace.mjs",
    ".agents/skills/vocanova-repo-navigator/SKILL.md",
    "package.json",
)

VALID_PAYLOAD = {
    "state": "open",
    "base": {
        "ref": "main",
        "sha": "0" * 40,
        "repo": {"full_name": "KARSIFT/vocanova-platform-sandbox"},
    },
    "head": {
        "ref": "develop",
        "sha": "1" * 40,
        "repo": {"full_name": "KARSIFT/vocanova-platform-sandbox"},
    },
}


def sha256_at_revision(revision: str, relative: str) -> str:
    content = subprocess.check_output(
        ["git", "show", f"{revision}:{relative}"], cwd=REPO_ROOT
    )
    return hashlib.sha256(content).hexdigest()


def metadata_script(pipeline_text: str) -> str:
    pipeline = yaml.safe_load(pipeline_text)
    steps = pipeline["jobs"]["promotion-pr-metadata"]["steps"]
    return next(
        step["run"]
        for step in steps
        if step.get("name") == "Resolve immutable promotion PR metadata"
    )


class Voc139PromotionRecoveryMetadataTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.pipeline_text = (CALLER_WORKFLOWS / "pipeline.yml").read_text(
            encoding="utf-8"
        )
        cls.fixture_pipeline = read_fixture(
            "templates/project-repo/.github/workflows/pipeline.yml"
        )
        cls.metadata_scripts = {
            "caller": metadata_script(cls.pipeline_text),
            "fixture": metadata_script(cls.fixture_pipeline),
        }
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")
        cls.readme = (FIXTURE_INFRA_ROOT / "README.md").read_text(encoding="utf-8")
        cls.ops_doc = (
            REPO_ROOT / "docs/operations/11-devops-and-ci-cd.md"
        ).read_text(encoding="utf-8")
        cls.skills_doc = (
            REPO_ROOT / "docs/development/agent-skills.md"
        ).read_text(encoding="utf-8")

    def _run_metadata(
        self,
        script: str,
        payload,
        *,
        repository="KARSIFT/vocanova-platform-sandbox",
        gh_exit=0,
    ):
        with tempfile.TemporaryDirectory() as directory:
            workspace = Path(directory)
            gh_path = workspace / "gh"
            gh_path.write_text(
                "#!/usr/bin/env bash\n"
                "set -euo pipefail\n"
                'expected="repos/${GITHUB_REPOSITORY}/pulls/${PROMOTION_PR_NUMBER}"\n'
                'if [ "$#" -ne 2 ] || [ "$1" != "api" ] || '
                '[ "$2" != "$expected" ]; then\n'
                "  exit 99\n"
                "fi\n"
                'if [ "$GH_FIXTURE_EXIT" -ne 0 ]; then\n'
                '  exit "$GH_FIXTURE_EXIT"\n'
                "fi\n"
                "printf '%s\\n' \"$GH_FIXTURE_PAYLOAD\"\n",
                encoding="utf-8",
            )
            gh_path.chmod(gh_path.stat().st_mode | stat.S_IEXEC)
            output_path = workspace / "github-output"
            environment = os.environ.copy()
            environment.update(
                {
                    "PATH": f"{workspace}:{environment['PATH']}",
                    "PROMOTION_PR_NUMBER": "1090",
                    "GH_TOKEN": "fixture-token",
                    "GH_FIXTURE_PAYLOAD": json.dumps(payload),
                    "GH_FIXTURE_EXIT": str(gh_exit),
                    "GITHUB_OUTPUT": str(output_path),
                }
            )
            if repository is None:
                environment.pop("GITHUB_REPOSITORY", None)
            else:
                environment["GITHUB_REPOSITORY"] = repository
            result = subprocess.run(
                ["bash", "-c", script],
                cwd=workspace,
                text=True,
                capture_output=True,
                env=environment,
            )
            result.github_output = (
                output_path.read_text(encoding="utf-8")
                if output_path.exists()
                else ""
            )
            return result

    def test_pin_advances_to_reviewed_voc139_infra_merge(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, VOC139_INFRA_PIN)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_seven_voc112_paths_remain_frozen_against_protected_anchor(self):
        for relative in NO_CHANGE_PATHS:
            working = REPO_ROOT / relative
            anchor = subprocess.check_output(
                ["git", "show", f"{PROTECTED_COMPARISON_ANCHOR}:{relative}"],
                cwd=REPO_ROOT,
                text=True,
            )
            self.assertEqual(working.read_text(encoding="utf-8"), anchor, relative)

    def test_exact_accumulated_promotion_and_ordinary_pr_pass_with_pinned_anchor(self):
        self.assertEqual(
            subprocess.run(
                [
                    "git",
                    "merge-base",
                    "--is-ancestor",
                    PROMOTION_BASE_SHA,
                    PROMOTION_HEAD_SHA,
                ],
                cwd=REPO_ROOT,
            ).returncode,
            0,
        )
        self.assertNotEqual(
            sha256_at_revision(PROMOTION_BASE_SHA, "AGENTS.md"),
            sha256_at_revision(PROMOTION_HEAD_SHA, "AGENTS.md"),
        )
        reviewed_head = subprocess.check_output(
            ["git", "rev-parse", "HEAD"], cwd=REPO_ROOT, text=True
        ).strip()
        environment = os.environ.copy()
        environment.update(
            {
                "PR_BASE_SHA": PROMOTION_BASE_SHA,
                "PR_HEAD_SHA": reviewed_head,
                "VOC112_CAPTURE_PROVENANCE_MODE": "pr-validation",
                "VOC112_PROMOTION_PR": "true",
            }
        )
        command = [
            "node",
            "--test",
            "scripts/foundation/voc112-navigation-benchmark.test.mjs",
        ]
        promotion = subprocess.run(
            command, cwd=REPO_ROOT, env=environment, text=True, capture_output=True
        )
        self.assertEqual(promotion.returncode, 0, promotion.stderr)
        environment.pop("VOC112_PROMOTION_PR")
        ordinary = subprocess.run(
            command, cwd=REPO_ROOT, env=environment, text=True, capture_output=True
        )
        self.assertEqual(ordinary.returncode, 0, ordinary.stderr)

    def test_real_metadata_bodies_succeed_without_git_repository(self):
        for name, script in self.metadata_scripts.items():
            with self.subTest(name=name):
                result = self._run_metadata(script, VALID_PAYLOAD)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn(f"pr_base_sha={'0' * 40}", result.github_output)
                self.assertIn(f"pr_head_sha={'1' * 40}", result.github_output)
                self.assertIn(
                    'gh api "repos/$GITHUB_REPOSITORY/pulls/$PROMOTION_PR_NUMBER"',
                    script,
                )
                self.assertNotIn("headRepository.nameWithOwner", script)

    def test_real_metadata_bodies_fail_closed_on_identity_negatives(self):
        invalid_cases = []
        for side, value in (("base", "production"), ("head", "feature")):
            payload = copy.deepcopy(VALID_PAYLOAD)
            payload[side]["ref"] = value
            invalid_cases.append((f"wrong-{side}-ref", payload))
        for side in ("base", "head"):
            payload = copy.deepcopy(VALID_PAYLOAD)
            payload[side]["repo"]["full_name"] = (
                "evil-fork/vocanova-platform-sandbox"
            )
            invalid_cases.append((f"wrong-{side}-repo", payload))
            payload = copy.deepcopy(VALID_PAYLOAD)
            del payload[side]["repo"]["full_name"]
            invalid_cases.append((f"missing-{side}-repo", payload))
            payload = copy.deepcopy(VALID_PAYLOAD)
            payload[side]["sha"] = "short"
            invalid_cases.append((f"malformed-{side}-sha", payload))
            payload = copy.deepcopy(VALID_PAYLOAD)
            del payload[side]["sha"]
            invalid_cases.append((f"missing-{side}-sha", payload))
        payload = copy.deepcopy(VALID_PAYLOAD)
        payload["state"] = "closed"
        invalid_cases.append(("closed", payload))

        for script_name, script in self.metadata_scripts.items():
            for case_name, payload in invalid_cases:
                with self.subTest(script=script_name, case=case_name):
                    result = self._run_metadata(script, payload)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertEqual(result.github_output, "")

    def test_real_metadata_bodies_require_context_and_existing_pr(self):
        for name, script in self.metadata_scripts.items():
            with self.subTest(script=name, case="missing-repository"):
                result = self._run_metadata(script, VALID_PAYLOAD, repository=None)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("GITHUB_REPOSITORY", result.stderr)
            with self.subTest(script=name, case="missing-pr"):
                result = self._run_metadata(script, {}, gh_exit=1)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(result.github_output, "")

    def test_metadata_jobs_have_no_checkout(self):
        for text in (self.pipeline_text, self.fixture_pipeline):
            pipeline = yaml.safe_load(text)
            steps = pipeline["jobs"]["promotion-pr-metadata"]["steps"]
            self.assertFalse(any("uses" in step for step in steps))

    def test_fixture_readme_records_voc139_pin_and_hash_contract(self):
        self.assertIn(CURRENT_PIN, self.readme)
        self.assertIn("587269f547c93a899ca7b5504825ab5304d7a266", self.ops_doc)
        self.assertIn("587269f547c93a899ca7b5504825ab5304d7a266", self.skills_doc)

    def test_evidence_records_implementation_pr_base_and_infra_merge(self):
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(VOC139_INFRA_PIN, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)


if __name__ == "__main__":
    unittest.main()
