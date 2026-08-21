from importlib.util import module_from_spec, spec_from_file_location
from pathlib import Path
import sys
import textwrap
import unittest


REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_INFRA_ROOT = REPOSITORY_ROOT / "tooling/governance/fixtures/karsift-ai-infra"
CONFIG = FIXTURE_INFRA_ROOT / "config"


def load_module(name: str, path: Path):
    spec = spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise AssertionError(f"cannot load {path}")
    module = module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


policy = load_module("ready_for_review_reuse", CONFIG / "ready_for_review_reuse.py")

HEAD = "a" * 40
BASE = "b" * 40
PACKAGE = "specs/changes/VOC-104-ready-for-review-reruns-unchanged-exact-sha-ci"


class FixtureReadyForReviewReuseTests(unittest.TestCase):
    def test_positive_reuse_fixture(self):
        prior = policy.PipelineRunSummary(
            run_id=100,
            event="pull_request",
            head_sha=HEAD,
            status="completed",
            conclusion="success",
            jobs=(
                {"name": "ci / ci", "conclusion": "success"},
                {"name": "review / publish-review", "conclusion": "success"},
            ),
        )
        comment = {
            "id": 1,
            "user": {"login": policy.TRUSTED_BOT_LOGIN, "type": "Bot"},
            "body": textwrap.dedent(
                f"""\
                **Independent verification - bound to commit `{HEAD}`**
                task_id: `VOC-104-T00`
                package_path: `{PACKAGE}`
                authority_issue: `875`
                base_sha: `{BASE}`

                VERDICT: PASS
                """
            ),
        }
        body = textwrap.dedent(
            f"""\
            Implements task `VOC-104-T00`
            Package path: `{PACKAGE}`
            Closes #875
            """
        )
        decision = policy.evaluate_reuse_eligibility(
            event_action="ready_for_review",
            expected_head_sha=HEAD,
            expected_base_sha=BASE,
            live_head_sha=HEAD,
            live_base_sha=BASE,
            is_draft=False,
            head_ref="agent/voc-104-t00",
            pr_body=body,
            comments=[comment],
            pipeline_runs=[prior],
            current_run_id=200,
            result_path_exists=False,
        )
        self.assertEqual(decision.outcome, "reuse-evidence")


if __name__ == "__main__":
    unittest.main()
