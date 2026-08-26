"""Caller fixture regressions for remediation ownership (VOC-106 / VOC-108)."""

from __future__ import annotations

from importlib.util import module_from_spec, spec_from_file_location
from pathlib import Path
import re
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[3]
FIXTURE_INFRA_ROOT = ROOT / "tooling/governance/fixtures/karsift-ai-infra"
CONFIG = FIXTURE_INFRA_ROOT / "config"
sys.path.insert(0, str(CONFIG))


def load_module(name: str, path: Path):
    spec = spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise AssertionError(f"cannot load {path}")
    module = module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


ownership = load_module("remediation_ownership", CONFIG / "remediation-ownership.py")
decision = load_module("decide_remediation", CONFIG / "decide-remediation.py")
verifier_runner = load_module(
    "verify_remediate_operator_ownership_runner",
    CONFIG / "verify-remediate-operator-ownership-runner.py",
)
verifier = load_module(
    "verify_remediate_operator_ownership",
    CONFIG / "verify_remediate_operator_ownership.py",
)


def contract(task_id: str, owner: str = "operator") -> str:
    return f"""schema_version: 1
task_id: {task_id}
ownership: {owner}
workflow_file: pipeline.yml
job_names:
  - verify-remediation-ownership / verify
events:
  - workflow_dispatch
branch: agent/voc-106-voc-106-t01
sha_lineage:
  mode: exact_pr_head
conclusion: success
"""


class RemediateOwnershipTests(unittest.TestCase):
    def make_package(self, root: Path, tasks: str) -> Path:
        package = root / "specs/changes/VOC-106-example"
        package.mkdir(parents=True)
        (package / "tasks.md").write_text(tasks, encoding="utf-8")
        return package

    def test_caller_template_exposes_read_only_remediation_verifier(self):
        template = (
            FIXTURE_INFRA_ROOT
            / "templates/project-repo/.github/workflows/pipeline-verify.yml"
        ).read_text(encoding="utf-8")
        self.assertIn("verify-post-promotion-workflow]", template)
        self.assertIn("  verify-remediate-operator-ownership:", template)
        verifier_block = template.split("  verify-remediate-operator-ownership:", 1)[1].split(
            "\n  verify-promotion-check-recovery:", 1
        )[0]
        self.assertIn("actions: read", verifier_block)
        self.assertIn("contents: read", verifier_block)
        self.assertIn("issues: read", verifier_block)
        self.assertIn("pull-requests: read", verifier_block)

    def test_ordinary_task_keeps_bounded_retry(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            self.make_package(root, "## VOC-106-T00 — ordinary\n")
            result = ownership.classify(root, "specs/changes/VOC-106-example", "VOC-106-T00")
            self.assertEqual(result.decision, "implement")
            self.assertEqual(
                decision.decide(
                    expected_sha="a",
                    current_sha="a",
                    review_state="FAIL",
                    ci_failed=False,
                    review_job_failed=False,
                    ownership_state="ORDINARY",
                ),
                "RETRY",
            )

    def test_valid_operator_contract_suppresses_review_and_ci_retry(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            package = self.make_package(
                root,
                "## VOC-106-T01 — proof\n\n- Automation ownership: operator\n",
            )
            contracts = package / ".karsift/live-evidence"
            contracts.mkdir(parents=True)
            (contracts / "VOC-106-T01.yaml").write_text(contract("VOC-106-T01"), encoding="utf-8")
            result = ownership.classify(root, "specs/changes/VOC-106-example", "VOC-106-T01")
            self.assertEqual(result.decision, "prepare-live-evidence")
            for ci_failed, review_state in ((False, "FAIL"), (True, "PENDING")):
                self.assertEqual(
                    decision.decide(
                        expected_sha="a",
                        current_sha="a",
                        review_state=review_state,
                        ci_failed=ci_failed,
                        review_job_failed=False,
                        ownership_state="OPERATOR",
                    ),
                    "ESCALATE_OPERATOR",
                )

    def test_malformed_or_conflicting_metadata_fails_closed(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            package = self.make_package(
                root,
                "## VOC-106-T01 — proof\n\n- Automation ownership: operator\n",
            )
            contracts = package / ".karsift/live-evidence"
            contracts.mkdir(parents=True)
            (contracts / "VOC-106-T01.yaml").write_text(
                contract("VOC-106-T01", "live-actions"), encoding="utf-8"
            )
            result = ownership.classify(root, "specs/changes/VOC-106-example", "VOC-106-T01")
            self.assertEqual(result.decision, "fail-closed")
            self.assertEqual(result.reason, "marker_contract_conflict")
            self.assertEqual(
                decision.decide(
                    expected_sha="a",
                    current_sha="a",
                    review_state="FAIL",
                    ci_failed=False,
                    review_job_failed=False,
                    ownership_state="FAIL_CLOSED",
                ),
                "ESCALATE_OPERATOR",
            )

    def test_remediate_workflow_uses_remediation_ownership_classifier(self):
        workflow = (FIXTURE_INFRA_ROOT / ".github/workflows/remediate.yml").read_text()
        self.assertIn("remediation-ownership.py", workflow)
        self.assertIn("ownership_state=", workflow)
        self.assertIn('echo "operator_escalation=true"', workflow)
        self.assertIn("Publish sanitized operator-ownership escalation", workflow)

    def test_hosted_verifier_extracts_base_sha_without_crashing(self):
        base_sha = "b" * 40
        valid = {
            "pull_requests": [
                {"number": 7, "head": {"sha": "a" * 40}, "base": {"sha": base_sha}}
            ]
        }
        self.assertEqual(verifier_runner.associated_base_sha(valid), base_sha)
        for malformed in (
            {},
            {"pull_requests": []},
            {"pull_requests": [None]},
            {"pull_requests": [{"base": None}]},
        ):
            self.assertEqual(verifier_runner.associated_base_sha(malformed), "")

    def test_authority_issue_line_accepts_carrier_punctuation_without_ambiguity(self):
        pattern = r"^Closes #([1-9][0-9]*)\.?$"
        self.assertEqual(re.findall(pattern, "Closes #885.", re.MULTILINE), ["885"])
        self.assertEqual(re.findall(pattern, "Closes #885", re.MULTILINE), ["885"])
        self.assertEqual(re.findall(pattern, "Closes #0.", re.MULTILINE), [])
        self.assertEqual(re.findall(pattern, "Closes #885 extra", re.MULTILINE), [])


if __name__ == "__main__":
    unittest.main()
