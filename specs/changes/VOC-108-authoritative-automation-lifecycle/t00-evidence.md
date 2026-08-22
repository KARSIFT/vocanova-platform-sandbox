# VOC-108-T00 — Evidence

Task: `VOC-108-T00` — Authoritative lifecycle evidence and idempotent advancement.

Evidence date: 2026-08-22

```yaml
gate_status: caller-merge-pending
shared_infra_merged: true
caller_pr_merged: false
live_lifecycle_claimed: false
shared_infra_pr: 102
shared_infra_exact_head_sha: 844f012ed65d161cca9a3dd4078867b8c00f2c3c
shared_infra_cleanup_pr: 103
shared_infra_cleanup_head_sha: 71a221c0a80fed7d8ee9f5a2eaf1f5cdcaee52df
shared_infra_cleanup_merge_sha: 0b57bb07f38eb66bf773b7208b258bcb3ffddd07
shared_infra_docs_pr: 104
shared_infra_docs_head_sha: 9dbc7195484eda8f09782171004a1ea071892871
shared_infra_review_remediation_pr: 105
shared_infra_review_remediation_head_sha: 3b70855d74aeb18e30ce757be7863c96a754e8a6
shared_infra_template_parity_pr: 106
shared_infra_template_parity_head_sha: f63d6b9360d47d78171e04fbb167a02be9531554
shared_infra_merge_sha: ee1b0a8ea8263a6671e753a6d3e80d15c855ddf4
shared_infra_self_ci_run: 32549356794
shared_infra_cleanup_ci_run: 32549968092
shared_infra_docs_ci_run: 32550666898
shared_infra_review_remediation_ci_run: 32551381890
shared_infra_template_parity_ci_run: 32552045142
```

## Implemented behavior

- Adoption, merge/reuse, and release select the newest authoritative attempt
  for each logical gate from complete paginated exact-SHA histories. Workflow
  run identity filters observer-generated check noise while retaining the
  pull-request, push, scheduled, and dispatched prerequisite workflows used by
  release.
- The merge gate publishes one strict App-authored marker only after the exact
  caller PR is observed merged. Auto-advance and release validate that same
  marker against live issue and PR state; issue closure alone is insufficient.
- Caller task PRs retain their local closing binding. Cross-repository text is
  generated and validated as a fully qualified non-closing reference; foreign
  qualified closing references are rejected by the merge gate.
- Authenticated pull-request API data must match the exact repository, PR,
  base, and head before check histories can be accepted. Task completion
  publishing is restricted to task branches, and a close event without the
  authoritative marker exits as a safe no-op.
- Automatic, reconcile, promotion-PR, third-party check, and external workflow
  wake-ups converge on one per-package serialized evaluator and one exact-head
  merge command. Terminal external workflow events wake evaluation without
  rerunning unchanged-SHA CI or the reviewer model.

## Shared-infrastructure evidence

| Evidence | Result |
| --- | --- |
| Shared PR | `KARSIFT/karsift-ai-infra#102` merged |
| Exact shared head | `844f012ed65d161cca9a3dd4078867b8c00f2c3c` |
| Whitespace follow-up | `KARSIFT/karsift-ai-infra#103` exact head `71a221c0a80fed7d8ee9f5a2eaf1f5cdcaee52df` merged |
| Authority-doc follow-up | `KARSIFT/karsift-ai-infra#104` exact head `9dbc7195484eda8f09782171004a1ea071892871` merged |
| Exact-SHA review remediation | `KARSIFT/karsift-ai-infra#105` exact head `3b70855d74aeb18e30ce757be7863c96a754e8a6` merged |
| Caller-template parity follow-up | `KARSIFT/karsift-ai-infra#106` exact head `f63d6b9360d47d78171e04fbb167a02be9531554` merged |
| Final consumed shared merge | `ee1b0a8ea8263a6671e753a6d3e80d15c855ddf4` |
| Hosted self-CI | run `32549356794`: actionlint, shellcheck, YAML parse, and 175 policy tests passed |
| Follow-up hosted self-CI | run `32549968092`: actionlint, shellcheck, YAML parse, and 175 policy tests passed |
| Authority-doc hosted self-CI | run `32550666898`: actionlint, shellcheck, YAML parse, and 176 policy tests passed |
| Review-remediation hosted self-CI | run `32551381890`: actionlint, shellcheck, YAML parse, and 179 policy tests passed |
| Template-parity hosted self-CI | run `32552045142`: actionlint, shellcheck, YAML parse, and 180 policy tests passed |
| Live selector replay | PR `#904` exact head selected the later successful attempt; obsolete failures did not poison the result |
| Cross-repository reference | PR `KARSIFT/karsift-ai-infra#102` passed closing-keyword-plus-target validation |

## Caller verification

The caller PR adds both event paths needed for release observation:

- `pull_request` promotion evaluation waits for the caller merge gate, so
  caller CI/review reaches terminal state before the release evaluator runs;
- `workflow_run: completed` observes repository-controlled external workflows,
  while `check_run: completed` covers non-Actions check providers.

`scripts/foundation/voc108-authoritative-lifecycle.test.mjs` runs the pinned
shared policy suite and validates this caller wiring. Governance validation,
risk classification (actual path floor `R4`), full workspace validation,
exact-SHA independent review, and caller merge remain to be recorded by the
governed PR checks.

## Live-proof boundary

This pre-merge file does not claim its own future merge or issue-close event.
After the exact caller PR merges, the App-authored task marker, task state,
release audit, promotion run, and idempotent duplicate/reconcile outcome are
recorded as sanitized GitHub issue comments. No logs, credentials, tokens,
sessions, OAuth material, secrets, or user identifiers are recorded here.
