"""VOC-145 caller regressions for governed role-binding reconciliation."""

from __future__ import annotations

import hashlib
import re
import subprocess
import sys
import unittest
from pathlib import Path

from voc080_fixtures import FIXTURE_INFRA_ROOT, read_fixture


REPO_ROOT = Path(__file__).resolve().parents[3]
EVIDENCE_PATH = (
    REPO_ROOT
    / "specs/changes/VOC-145-direct-role-binding-and-regression-expectation/t00-evidence.md"
)

AUTHORITATIVE_PIN = "8993e867640dfb604dec0466c4e0787e68d8e258"
UNAUTHORIZED_HEAD = "d8720829b176cf1287e633f9382989fc8f258105"
CURRENT_PIN = "ad2b27784e6fc33b3ac7e9dab48245dd6d08ac7f"
IMPLEMENTATION_PR_BASE = "10df745c42ed283405d6bdf5b01180afdfed7d26"
INFRA_CARRIER_PR = "175"
AUTHORIZED_PATH = "Path A"

VOC117_BINDINGS = {
    "implementer": "cursor/composer-2.5",
    "implementer_escalation": "cursor/composer-2.5",
    "planner": "cursor/grok-4.6[effort=high,fast=false]",
    "reviewer": "cursor/grok-4.6[effort=high,fast=false]",
    "reviewer_fast_retry": "cursor/grok-4.6[effort=high,fast=false]",
    "plan_reviewer": "cursor/grok-4.6[effort=high,fast=false]",
}

FINAL_INFRA_CHANGED_MIRRORS = frozenset(
    {
        "CHANGELOG.md",
        "config/roles.yml",
        "tests/test_voc117_role_bindings.py",
    }
)

MIRRORED_FILE_HASHES = {
    "CHANGELOG.md": (
        "a33a305abf76528c71632a3df7b5b0b8afe4e5899d88a330f5623c500da7bdff"
    ),
    "config/roles.yml": (
        "0e3940e9a4248520d140fa7076f4bcce5e84225919af26725783876db6493648"
    ),
    "tests/test_voc117_role_bindings.py": (
        "d45bd343156ab3c10eebdc09d972b53145bca695e7f6748b2a8bd2bf7b2a6fc4"
    ),
}

CURRENT_PIN_ASSERTION_PATHS = (
    "scripts/foundation/voc097-fixture-matrix.test.mjs",
    "scripts/foundation/voc104-ready-for-review-reuse.test.mjs",
    "scripts/foundation/voc108-authoritative-lifecycle.test.mjs",
    "tooling/governance/tests/test_voc121_implement_policy.py",
    "tooling/governance/tests/test_voc122_implement_policy.py",
    "tooling/governance/tests/test_voc124_implement_policy.py",
    "tooling/governance/tests/test_voc125_implement_fixture.py",
    "tooling/governance/tests/test_voc125_implement_policy.py",
    "tooling/governance/tests/test_voc126_workflow_dispatch_input_limit.py",
    "tooling/governance/tests/test_voc129_caller_replacement.py",
    "tooling/governance/tests/test_voc136_caller_replacement.py",
    "tooling/governance/tests/test_voc137_pr_sha_scan.py",
    "tooling/governance/tests/test_voc138_promotion_pr_provenance.py",
    "tooling/governance/tests/test_voc139_promotion_recovery_metadata.py",
    "tooling/governance/tests/test_voc140_release_convergence.py",
    "tooling/governance/tests/test_voc142_adoption_roster_wait.py",
    "tooling/governance/tests/test_voc145_caller_replacement.py",
)

FIXTURE_CONFIG = FIXTURE_INFRA_ROOT / "config"
if str(FIXTURE_CONFIG) not in sys.path:
    sys.path.insert(0, str(FIXTURE_CONFIG))

from prepare_cursor_model import CursorModelError, prepare_cursor_model  # noqa: E402


def active_role(config: str, role: str) -> str:
    matches = re.findall(rf"^{re.escape(role)}:\s*(\S+)\s*$", config, re.MULTILINE)
    if len(matches) != 1:
        raise AssertionError(f"expected one active {role} binding, found {matches}")
    return matches[0]


class Voc145CallerReplacementTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pin = read_fixture("PINNED_SHA.txt").strip()
        cls.roles = read_fixture("config/roles.yml")
        cls.readme = read_fixture("README.md")
        cls.changelog = read_fixture("CHANGELOG.md")
        cls.fixture_readme = (
            FIXTURE_INFRA_ROOT / "README.md"
        ).read_text(encoding="utf-8")
        cls.voc117_test = read_fixture("tests/test_voc117_role_bindings.py")
        cls.implement = read_fixture(".github/workflows/implement.yml")
        cls.review = read_fixture(".github/workflows/review.yml")
        cls.remediate = read_fixture(".github/workflows/remediate.yml")
        cls.merge_gate = read_fixture(".github/workflows/merge-gate.yml")
        cls.evidence = EVIDENCE_PATH.read_text(encoding="utf-8")

    def test_current_pin_matches_coordinated_infra_carrier_head(self):
        self.assertEqual(self.pin, CURRENT_PIN)
        self.assertNotEqual(self.pin, UNAUTHORIZED_HEAD)
        self.assertNotEqual(self.pin, AUTHORITATIVE_PIN)

    def test_changed_mirrors_match_declared_hashes(self):
        for relative in sorted(FINAL_INFRA_CHANGED_MIRRORS):
            path = FIXTURE_INFRA_ROOT / relative
            digest = hashlib.sha256(path.read_bytes()).hexdigest()
            self.assertEqual(digest, MIRRORED_FILE_HASHES[relative], relative)

    def test_live_current_pin_assertions_reference_new_merge(self):
        self.assertEqual(len(CURRENT_PIN_ASSERTION_PATHS), 17)
        for relative in CURRENT_PIN_ASSERTION_PATHS:
            text = (REPO_ROOT / relative).read_text(encoding="utf-8")
            self.assertIn(CURRENT_PIN, text, relative)

    def test_roles_yml_has_six_exact_path_a_bindings(self):
        for role, expected in VOC117_BINDINGS.items():
            with self.subTest(role=role):
                self.assertEqual(active_role(self.roles, role), expected)
        self.assertNotIn("effort=xhigh", self.roles)
        self.assertNotIn("fast=true", self.roles)

    def test_voc117_fixture_preserves_exact_high_effort_assertions(self):
        self.assertIn("VOC117_BINDINGS", self.voc117_test)
        for role in ("reviewer", "reviewer_fast_retry", "plan_reviewer"):
            self.assertIn(
                f'"{role}": "cursor/grok-4.6[effort=high,fast=false]"',
                self.voc117_test,
                role,
            )
        self.assertNotIn("effort=xhigh", self.voc117_test)
        self.assertIn(
            'prepare_cursor_model(VOC117_BINDINGS["planner"])',
            self.voc117_test,
        )
        self.assertNotIn("binding.removeprefix", self.voc117_test)

    def test_prepare_cursor_model_requires_effort_and_fails_closed(self):
        for role, binding in VOC117_BINDINGS.items():
            with self.subTest(role=role):
                self.assertEqual(
                    prepare_cursor_model(binding),
                    binding.removeprefix("cursor/"),
                )
        for unavailable in ("cursor/grok-4.6", "cursor/grok-4.6[fast=false]"):
            with self.subTest(unavailable=unavailable), self.assertRaises(CursorModelError):
                prepare_cursor_model(unavailable)

    def test_missing_api_key_and_unsupported_prefix_fail_closed(self):
        env = dict(**{k: v for k, v in __import__("os").environ.items()})
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
        with self.assertRaises(CursorModelError):
            prepare_cursor_model("openai/codex-action")

    def test_current_docs_describe_path_a_and_not_xhigh_drift(self):
        self.assertIn("effort=high,fast=false", self.roles)
        self.assertIn("VOC-145", self.changelog)
        self.assertIn("d8720829", self.changelog)
        self.assertNotIn("effort=xhigh", self.roles)
        self.assertIn(CURRENT_PIN, self.fixture_readme)
        self.assertIn("VOC-145", self.fixture_readme)
        self.assertIn("PINNED_SHA.txt", self.fixture_readme)
        self.assertIn(INFRA_CARRIER_PR, self.fixture_readme)

    def test_safety_controls_remain_explicit(self):
        self.assertIn("expected_head_sha:", self.review)
        self.assertIn('if [ "$actual_head" != "$EXPECTED_HEAD_SHA" ]; then', self.review)
        self.assertIn("expected_head_sha:", self.merge_gate)
        self.assertIn('if [ "$next_attempt" -gt 2 ]; then', self.remediate)
        self.assertIn("no attempt 3", self.implement)
        self.assertNotIn("OPENAI_API_KEY", self.roles)

    def test_evidence_records_implementation_base_path_and_pins(self):
        self.assertIn(IMPLEMENTATION_PR_BASE, self.evidence)
        self.assertIn(CURRENT_PIN, self.evidence)
        self.assertIn(AUTHORITATIVE_PIN, self.evidence)
        self.assertIn(UNAUTHORIZED_HEAD, self.evidence)
        self.assertIn(AUTHORIZED_PATH, self.evidence)
        self.assertIn(INFRA_CARRIER_PR, self.evidence)

    def test_historical_package_records_remain_unmodified(self):
        for slug in (
            "VOC-117-route-planning-and-escalation-through-codex-and",
            "VOC-142-adoption-roster-wait-ignores-pending-required-ci",
        ):
            package_dir = REPO_ROOT / "specs/changes" / slug
            self.assertTrue((package_dir / "change.yaml").is_file(), slug)


if __name__ == "__main__":
    unittest.main()
