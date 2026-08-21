from importlib.util import module_from_spec, spec_from_file_location
from pathlib import Path
import sys
import tempfile
import textwrap
import unittest


REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_INFRA_ROOT = REPOSITORY_ROOT / "tooling" / "governance" / "fixtures" / "karsift-ai-infra"
CONFIG = FIXTURE_INFRA_ROOT / "config"


def load_module(name: str, path: Path):
    spec = spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise AssertionError(f"cannot load {path}")
    module = module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


policy = load_module("live_evidence_reconcile", CONFIG / "live_evidence_reconcile.py")
ownership = load_module("auto_advance_ownership", CONFIG / "auto_advance_ownership.py")
verifier = load_module(
    "verify_auto_advance_live_evidence",
    CONFIG / "verify_auto_advance_live_evidence.py",
)


PACKAGE = "specs/changes/VOC-102-auto-advance-dispatches-implementer-for-operator"


def contract_yaml(task_id: str, ownership_value: str = "operator") -> str:
    return textwrap.dedent(
        f"""\
        schema_version: 1
        task_id: {task_id}
        ownership: {ownership_value}
        workflow_file: pipeline.yml
        job_names:
          - verify-auto-advance-live-evidence / verify
        events:
          - workflow_dispatch
        branch: agent/voc-102-voc-102-t01
        sha_lineage:
          mode: exact_pr_head
        conclusion: success
        """
    )


class AutoAdvanceOwnershipTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.auto_advance = (
            FIXTURE_INFRA_ROOT / ".github/workflows/auto-advance.yml"
        ).read_text()
        cls.verify_workflow = (
            FIXTURE_INFRA_ROOT / ".github/workflows/verify-auto-advance-live-evidence.yml"
        ).read_text()

    def test_voc102_test_00_contract_path_is_authoritative(self):
        with tempfile.TemporaryDirectory() as scratch:
            package = Path(scratch) / "pkg"
            package.mkdir()
            (package / "tasks.md").write_text("# tasks\n", encoding="utf-8")
            contract_dir = package / ".karsift/live-evidence"
            contract_dir.mkdir(parents=True)
            (contract_dir / "VOC-102-T01.yaml").write_text(
                contract_yaml("VOC-102-T01"),
                encoding="utf-8",
            )
            result = ownership.classify_next_task(str(package), "VOC-102-T01", "")
            self.assertEqual(result.decision, "prepare-live-evidence")
            self.assertEqual(result.ownership, "operator")

    def test_voc102_test_01_operator_and_live_actions_recognized(self):
        for value in ("operator", "live-actions"):
            with self.subTest(value=value):
                with tempfile.TemporaryDirectory() as scratch:
                    package = Path(scratch) / "pkg"
                    package.mkdir()
                    contract_dir = package / ".karsift/live-evidence"
                    contract_dir.mkdir(parents=True)
                    (contract_dir / "VOC-099-T01.yaml").write_text(
                        contract_yaml("VOC-099-T01", value),
                        encoding="utf-8",
                    )
                    result = ownership.classify_next_task(str(package), "VOC-099-T01", "")
                    self.assertEqual(result.decision, "prepare-live-evidence")
                    self.assertEqual(result.ownership, value)

    def test_voc102_test_02_operator_next_task_does_not_dispatch(self):
        with tempfile.TemporaryDirectory() as scratch:
            package = Path(scratch) / "pkg"
            package.mkdir()
            contract_dir = package / ".karsift/live-evidence"
            contract_dir.mkdir(parents=True)
            (contract_dir / "VOC-102-T01.yaml").write_text(
                contract_yaml("VOC-102-T01"),
                encoding="utf-8",
            )
            result = ownership.classify_next_task(str(package), "VOC-102-T01", "")
            self.assertEqual(result.decision, "prepare-live-evidence")
            self.assertNotEqual(result.decision, "implement")

    def test_voc102_test_03_ordinary_next_task_dispatches(self):
        with tempfile.TemporaryDirectory() as scratch:
            package = Path(scratch) / "pkg"
            package.mkdir()
            (package / "tasks.md").write_text(
                textwrap.dedent(
                    """\
                    ## VOC-000-T01 — ordinary task

                    - Status: pending
                    """
                ),
                encoding="utf-8",
            )
            result = ownership.classify_next_task(str(package), "VOC-000-T01", "")
            self.assertEqual(result.decision, "implement")

    def test_voc102_test_04_malformed_contract_fail_closed(self):
        with tempfile.TemporaryDirectory() as scratch:
            package = Path(scratch) / "pkg"
            package.mkdir()
            contract_dir = package / ".karsift/live-evidence"
            contract_dir.mkdir(parents=True)
            (contract_dir / "VOC-102-T01.yaml").write_text(":\n  bad\n", encoding="utf-8")
            result = ownership.classify_next_task(str(package), "VOC-102-T01", "")
            self.assertEqual(result.decision, "fail-closed")

    def test_voc102_test_05_missing_or_unrecognized_ownership_fail_closed(self):
        cases = [
            ("schema_version: 1\ntask_id: VOC-102-T01\n", "missing ownership"),
            (
                contract_yaml("VOC-102-T01").replace("ownership: operator", "ownership: bot"),
                "unrecognized ownership",
            ),
        ]
        for body, label in cases:
            with self.subTest(label=label):
                with tempfile.TemporaryDirectory() as scratch:
                    package = Path(scratch) / "pkg"
                    package.mkdir()
                    contract_dir = package / ".karsift/live-evidence"
                    contract_dir.mkdir(parents=True)
                    (contract_dir / "VOC-102-T01.yaml").write_text(body, encoding="utf-8")
                    result = ownership.classify_next_task(str(package), "VOC-102-T01", "")
                    self.assertEqual(result.decision, "fail-closed")

    def test_voc102_test_06_contradictory_metadata_fail_closed(self):
        tasks_md = textwrap.dedent(
            """\
            ## VOC-102-T01 — operator task

            - Automation ownership: operator
            """
        )
        with tempfile.TemporaryDirectory() as scratch:
            package = Path(scratch) / "pkg"
            package.mkdir()
            (package / "tasks.md").write_text(tasks_md, encoding="utf-8")

            marker_only = ownership.classify_next_task(str(package), "VOC-102-T01", tasks_md)
            self.assertEqual(marker_only.decision, "fail-closed")
            self.assertEqual(marker_only.reason, "marker_without_contract")

            contract_dir = package / ".karsift/live-evidence"
            contract_dir.mkdir(parents=True)
            (contract_dir / "VOC-102-T01.yaml").write_text(
                contract_yaml("VOC-102-T01", "live-actions"),
                encoding="utf-8",
            )
            conflict = ownership.classify_next_task(str(package), "VOC-102-T01", tasks_md)
            self.assertEqual(conflict.decision, "fail-closed")
            self.assertEqual(conflict.reason, "marker_contract_conflict")

            duplicate_tasks = tasks_md + "\n- Automation ownership: live-actions\n"
            duplicate = ownership.classify_next_task(str(package), "VOC-102-T01", duplicate_tasks)
            self.assertEqual(duplicate.decision, "fail-closed")
            self.assertEqual(duplicate.reason, "duplicate_automation_marker")

            prose = tasks_md + "\nNarrative mentions operator ownership in prose only.\n"
            ordinary = ownership.classify_next_task(str(package), "VOC-000-T00", prose)
            self.assertEqual(ordinary.decision, "implement")

    def test_voc102_test_07_last_task_is_out_of_scope_for_classifier(self):
        # The advance job no-ops before classification when there is no next task.
        self.assertIn(
            "No next task after",
            self.auto_advance,
        )

    def test_voc102_test_11_evidence_path_is_strict_and_idempotent_helpers(self):
        self.assertEqual(
            ownership.derive_evidence_relative_path("VOC-102-T01"),
            "t01-evidence.md",
        )
        self.assertEqual(
            ownership.derive_evidence_relative_path("VOC-031-T07a"),
            "t07a-evidence.md",
        )
        with self.assertRaises(ValueError):
            ownership.derive_evidence_relative_path("VOC-102")
        body = ownership.carrier_pr_body(
            change_id="VOC-102",
            task_id="VOC-102-T01",
            package_path=PACKAGE,
            issue_number=866,
            evidence_relative_path="t01-evidence.md",
        )
        self.assertIn(f"Package path: `{PACKAGE}`", body)
        self.assertTrue(
            ownership.is_valid_carrier_pr(
                pr_title="VOC-102: VOC-102-T01",
                pr_body=body,
                change_id="VOC-102",
                task_id="VOC-102-T01",
                package_path=PACKAGE,
            )
        )

    def test_voc102_test_12_permission_boundary(self):
        advance_block = self.auto_advance.split("  advance:", 1)[1].split("  prepare-live-evidence:", 1)[0]
        publisher_block = self.auto_advance.split("  prepare-live-evidence:", 1)[1].split("  fail-closed:", 1)[0]
        implement_block = self.auto_advance.split("  implement:", 1)[1]
        self.assertIn("issues: read", advance_block)
        self.assertIn("pull-requests: read", advance_block)
        self.assertIn("contents: read", advance_block)
        self.assertNotIn("actions:", advance_block)
        self.assertIn("create-github-app-token", publisher_block)
        self.assertIn("auto-advance-carrier-publisher.py", publisher_block)
        self.assertNotIn("implement.yml", publisher_block)
        self.assertNotIn("secrets: inherit", publisher_block)
        self.assertIn("uses: KARSIFT/karsift-ai-infra/.github/workflows/implement.yml@main", implement_block)
        verify_permissions = self.verify_workflow.split("    permissions:", 1)[1].split("    steps:", 1)[0]
        self.assertIn("actions: read", verify_permissions)
        self.assertNotIn("actions: write", verify_permissions)
        self.assertNotIn("create-github-app-token", self.verify_workflow)

    def test_voc102_test_13_verifier_fail_closed_and_exact_head(self):
        source_ok = verifier.verify_source_run(
            run={
                "repository": {"full_name": "KARSIFT/example"},
                "event": "issues",
                "head_branch": "develop",
                "path": ".github/workflows/pipeline.yml",
                "conclusion": "success",
            },
            repository="KARSIFT/example",
            integration_branch="develop",
        )
        self.assertTrue(source_ok.ok)

        source_bad = verifier.verify_source_run(
            run={
                "repository": {"full_name": "KARSIFT/example"},
                "event": "workflow_dispatch",
                "head_branch": "develop",
                "path": ".github/workflows/pipeline.yml",
                "conclusion": "success",
            },
            repository="KARSIFT/example",
            integration_branch="develop",
        )
        self.assertEqual(source_bad.reason, "wrong_event")

        implement_bad = verifier.verify_no_implement_job(
            [{"name": "auto-advance / implement / implement", "conclusion": "success"}],
            "VOC-102-T01",
        )
        self.assertEqual(implement_bad.reason, "implement_job_executed")

        pr_body = ownership.carrier_pr_body(
            change_id="VOC-102",
            task_id="VOC-102-T01",
            package_path=PACKAGE,
            issue_number=866,
            evidence_relative_path="t01-evidence.md",
        )
        carrier_ok = verifier.verify_carrier_state(
            pr={
                "title": "VOC-102: VOC-102-T01",
                "body": pr_body,
                "headRefName": ownership.branch_name("VOC-102", "VOC-102-T01"),
                "headRefOid": "a" * 40,
            },
            prs_on_branch=[{"number": 1}],
            comments=[{"body": ownership.WAITING_MARKER_PREFIX}],
            change_id="VOC-102",
            task_id="VOC-102-T01",
            package_path=PACKAGE,
            evidence_exists=True,
            current_ref="a" * 40,
        )
        self.assertTrue(carrier_ok.ok)

        stale = verifier.verify_carrier_state(
            pr={
                "title": "VOC-102: VOC-102-T01",
                "body": pr_body,
                "headRefName": ownership.branch_name("VOC-102", "VOC-102-T01"),
                "headRefOid": "a" * 40,
            },
            prs_on_branch=[{"number": 1}],
            comments=[{"body": ownership.WAITING_MARKER_PREFIX}],
            change_id="VOC-102",
            task_id="VOC-102-T01",
            package_path=PACKAGE,
            evidence_exists=True,
            current_ref="b" * 40,
        )
        self.assertEqual(stale.reason, "stale_pr_head")

        self.assertIn('name: verify-auto-advance-live-evidence / verify', self.verify_workflow)


if __name__ == "__main__":
    unittest.main()
