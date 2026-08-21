from importlib.util import module_from_spec, spec_from_file_location
from pathlib import Path
import sys
import tempfile
import textwrap
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


decide = load_module("decide_remediation", CONFIG / "decide-remediation.py")
ownership = load_module("remediate_ownership", CONFIG / "remediate_ownership.py")
verifier = load_module(
    "verify_remediate_operator_ownership",
    CONFIG / "verify_remediate_operator_ownership.py",
)


def contract_yaml(task_id: str, ownership_value: str = "operator") -> str:
    return textwrap.dedent(
        f"""\
        schema_version: 1
        task_id: {task_id}
        ownership: {ownership_value}
        workflow_file: pipeline.yml
        job_names:
          - verify-remediate-operator-ownership / verify
        events:
          - workflow_dispatch
        branch: agent/voc-106-voc-106-t01
        sha_lineage:
          mode: exact_pr_head
        conclusion: success
        """
    )


class RemediateOwnershipTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.workflow = (FIXTURE_INFRA_ROOT / ".github/workflows/remediate.yml").read_text()
        cls.verify_workflow = (
            FIXTURE_INFRA_ROOT / ".github/workflows/verify-remediate-operator-ownership.yml"
        ).read_text()
        cls.template_pipeline = (
            FIXTURE_INFRA_ROOT / "templates/project-repo/.github/workflows/pipeline.yml"
        ).read_text()

    def test_voc106_test_00_contract_path_is_authoritative(self):
        with tempfile.TemporaryDirectory() as scratch:
            package = Path(scratch) / "pkg"
            package.mkdir()
            (package / "tasks.md").write_text("# tasks\n", encoding="utf-8")
            contract_dir = package / ".karsift/live-evidence"
            contract_dir.mkdir(parents=True)
            (contract_dir / "VOC-106-T01.yaml").write_text(
                contract_yaml("VOC-106-T01"),
                encoding="utf-8",
            )
            result, _ = ownership.classify_task_for_remediation(
                str(package), "VOC-106-T01", ""
            )
            self.assertEqual(result, "operator")

    def test_voc106_test_01_operator_and_live_actions_recognized(self):
        for value in ("operator", "live-actions"):
            with self.subTest(value=value):
                with tempfile.TemporaryDirectory() as scratch:
                    package = Path(scratch) / "pkg"
                    package.mkdir()
                    (package / "tasks.md").write_text("# tasks\n", encoding="utf-8")
                    contract_dir = package / ".karsift/live-evidence"
                    contract_dir.mkdir(parents=True)
                    (contract_dir / "VOC-106-T01.yaml").write_text(
                        contract_yaml("VOC-106-T01", value),
                        encoding="utf-8",
                    )
                    result, _ = ownership.classify_task_for_remediation(
                        str(package), "VOC-106-T01", ""
                    )
                    self.assertEqual(result, value)
                    self.assertEqual(
                        decide.decide(
                            expected_sha="a" * 40,
                            current_sha="a" * 40,
                            review_state="FAIL",
                            ci_failed=False,
                            review_job_failed=False,
                            ownership=result,
                        ),
                        "ESCALATE_OPERATOR",
                    )

    def test_voc106_test_02_operator_waiting_remains_suppressed(self):
        self.assertEqual(
            decide.decide(
                expected_sha="a" * 40,
                current_sha="a" * 40,
                review_state="WAITING",
                ci_failed=False,
                review_job_failed=False,
                ownership="operator",
            ),
            "WAITING",
        )
        waiting_guard = self.workflow.index('decision" = "WAITING"')
        ownership_gate = self.workflow.index("remediate-ownership-classifier.py")
        retry_output = self.workflow.index('echo "should_retry=true"')
        self.assertLess(waiting_guard, ownership_gate)
        self.assertLess(ownership_gate, retry_output)

    def test_voc106_test_03_operator_fail_and_ci_escalate_without_retry(self):
        for ci_failed, review_state, reason in (
            (False, "FAIL", "review_fail"),
            (True, "PENDING", "ci_failure"),
        ):
            with self.subTest(ci_failed=ci_failed, review_state=review_state):
                self.assertEqual(
                    decide.decide(
                        expected_sha="a" * 40,
                        current_sha="a" * 40,
                        review_state=review_state,
                        ci_failed=ci_failed,
                        review_job_failed=False,
                        ownership="operator",
                    ),
                    "ESCALATE_OPERATOR",
                )
        self.assertIn("operator_escalation=true", self.workflow)
        self.assertIn("remediate-escalate-operator.py", self.workflow)
        retry = self.workflow.split("  retry:", 1)[1]
        self.assertIn("needs.decide.outputs.should_retry == 'true'", retry)

    def test_voc106_test_04_ordinary_fail_and_ci_still_retry(self):
        for ci_failed, review_state in ((False, "FAIL"), (True, "PENDING")):
            with self.subTest(ci_failed=ci_failed, review_state=review_state):
                self.assertEqual(
                    decide.decide(
                        expected_sha="a" * 40,
                        current_sha="a" * 40,
                        review_state=review_state,
                        ci_failed=ci_failed,
                        review_job_failed=False,
                        ownership="ordinary",
                    ),
                    "RETRY",
                )

    def test_voc106_test_05_stale_does_not_retry_or_consume_attempt(self):
        self.assertEqual(
            decide.decide(
                expected_sha="a" * 40,
                current_sha="b" * 40,
                review_state="FAIL",
                ci_failed=True,
                review_job_failed=False,
                ownership="operator",
            ),
            "STALE",
        )
        self.assertIn('initial_decision" = "STALE"', self.workflow)
        self.assertIn('echo "stale_run=true"', self.workflow)

    def test_voc106_test_06_malformed_contract_fail_closed(self):
        with tempfile.TemporaryDirectory() as scratch:
            package = Path(scratch) / "pkg"
            package.mkdir()
            (package / "tasks.md").write_text("# tasks\n", encoding="utf-8")
            contract_dir = package / ".karsift/live-evidence"
            contract_dir.mkdir(parents=True)
            (contract_dir / "VOC-106-T01.yaml").write_text(":\n  bad\n", encoding="utf-8")
            result, reason = ownership.classify_task_for_remediation(
                str(package), "VOC-106-T01", ""
            )
            self.assertEqual(result, "fail-closed")
            self.assertTrue(reason)
            self.assertEqual(
                decide.decide(
                    expected_sha="a" * 40,
                    current_sha="a" * 40,
                    review_state="FAIL",
                    ci_failed=False,
                    review_job_failed=False,
                    ownership=result,
                ),
                "FAIL_CLOSED",
            )
        self.assertIn("ownership_fail_closed=true", self.workflow)
        self.assertIn("remediate-fail-closed.py", self.workflow)

    def test_voc106_test_07_operator_ci_escalation_is_metadata_only(self):
        escalation_step = self.workflow.split(
            "- name: Record sanitized operator remediation escalation", 1
        )[1].split("- name: Record sanitized remediation ownership fail-closed", 1)[0]
        self.assertNotIn("/actions/jobs/", escalation_step)
        self.assertNotIn("/logs", escalation_step)
        self.assertNotIn("/artifacts", escalation_step)
        self.assertIn("remediate-escalate-operator.py", escalation_step)
        self.assertIn("should_retry: `false`", Path(CONFIG / "remediate-escalate-operator.py").read_text())
        self.assertNotIn("secrets: inherit", escalation_step)

    def test_voc106_test_11_verifier_is_read_only_and_fail_closed(self):
        verify_permissions = self.verify_workflow.split("    permissions:", 1)[1].split(
            "    steps:", 1
        )[0]
        self.assertIn("actions: read", verify_permissions)
        self.assertNotIn("actions: write", verify_permissions)
        self.assertNotIn("create-github-app-token", self.verify_workflow)
        self.assertIn("GITHUB_TOKEN: ${{ github.token }}", self.verify_workflow)
        self.assertIn("    name: verify", self.verify_workflow)
        runner_source = (
            CONFIG / "verify-remediate-operator-ownership-runner.py"
        ).read_text()
        self.assertNotRegex(runner_source, r'["\']logs["\']')
        self.assertNotRegex(runner_source, r'["\']artifacts["\']')

        jobs = [
            {"name": "remediate / decide", "conclusion": "success"},
            {"name": "remediate / retry", "conclusion": "skipped"},
        ]
        self.assertTrue(verifier.verify_source_jobs(jobs).ok)
        jobs[-1]["conclusion"] = "success"
        self.assertEqual(
            verifier.verify_source_jobs(jobs).reason,
            "implement_job_executed",
        )

        comments = [
            {
                "body": (
                    f"{ownership.OPERATOR_ESCALATION_MARKER_PREFIX} `VOC-106-T01`\n\n"
                    "should_retry: `false`\n"
                    "task_id: `VOC-106-T01`\n"
                    "package_path: `specs/changes/VOC-106-example`\n"
                    "pr_number: `1`\n"
                    "run_id: `123`\n"
                ),
                "author": {"login": "github-actions[bot]"},
            }
        ]
        self.assertTrue(
            verifier.verify_escalation_marker(
                comments,
                task_id="VOC-106-T01",
                package_path="specs/changes/VOC-106-example",
                pr_number=1,
                source_run_id=123,
            ).ok
        )


if __name__ == "__main__":
    unittest.main()
