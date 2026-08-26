"""VOC-125 caller pipeline dispatch regressions for existing-carrier resume."""

from __future__ import annotations

import re
import unittest
from pathlib import Path

REPOSITORY_ROOT = Path(__file__).resolve().parents[3]
CALLER_PIPELINE = REPOSITORY_ROOT / ".github/workflows/pipeline.yml"


class Voc125PipelineDispatchTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pipeline = CALLER_PIPELINE.read_text(encoding="utf-8")

    def test_workflow_dispatch_exposes_existing_pr_number_only(self):
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split("jobs:", 1)[0]
        self.assertIn("existing_pr_number:", dispatch)
        self.assertIn("existing open implementation PR to resume at attempt 2", dispatch)
        self.assertNotIn("expected_head_sha:", dispatch)
        self.assertNotIn("expected_base_sha:", dispatch)

    def test_implement_job_forwards_existing_pr_number(self):
        implement = self.pipeline.split("  implement:", 1)[1].split("\n  plan:", 1)[0]
        self.assertIn("existing_pr_number: ${{ inputs.existing_pr_number }}", implement)
        self.assertIn("attempt: ${{ inputs.attempt }}", implement)

    def test_dispatch_input_count_within_bound(self):
        dispatch = self.pipeline.split("workflow_dispatch:", 1)[1].split("jobs:", 1)[0]
        inputs = re.findall(r"^      [a-z0-9_]+:$", dispatch, re.MULTILINE)
        self.assertLessEqual(len(inputs), 25)


if __name__ == "__main__":
    unittest.main()
