# VOC-040 — Impact Analysis

## Security and privacy

None beyond what is already true today. This package documents an existing
recovery procedure (`gh run rerun --failed` against a specific already-completed
`adopt` run) that uses only permissions a repository maintainer who can already
re-run any failed Actions run and edit `change.yaml` on the target branch already
has. It does not request, grant, or document any new permission, credential, App
installation, or secret. It does not change `adopt.yml`'s own refusal behavior
(still refuses to open task issues for a still-draft `change.yaml` — see
`specification.md`'s "Non-goals").

## Data and migrations

None. No schema, table, or migration is touched.

## Analytics and accessibility

None. No analytics event is added; `AGENTS.md` is not a user-facing UI surface, so
no accessibility impact applies.

## Risks, dependencies, and evidence

- `VOC-040-R00`: The documented procedure could go stale if `adopt.yml`'s verify
  step is ever changed (in `karsift-ai-infra`, outside this repository's control)
  to read a frozen `change.yaml` snapshot instead of the target branch's current
  tip, or if GitHub changes `gh run rerun --failed`'s semantics. Mitigated only by
  `VOC-040-T00`'s documentation itself flagging the dependency explicitly (see
  `specification.md`'s open question 2) — this package has no mechanism to detect
  or alert on such a change, since it lives in a different repository.
- `VOC-040-R01`: Documenting a workaround could reduce the perceived urgency of
  the structural fix (a `workflow_dispatch` entry point, or a plan.yml guardrail)
  that would actually reduce how often this mistake happens, rather than just
  making it recoverable after the fact. Mitigated by `VOC-040-T00`'s documentation
  explicitly stating it is a workaround, not a structural fix, and by
  `specification.md`'s open question 1 flagging the follow-up for the reviewing
  human rather than treating this package as a complete resolution of issue #301's
  underlying gap.
- `VOC-040-DEP-00`: see `change.yaml`'s dependency entry — depends on properties
  of `karsift-ai-infra`'s `adopt.yml` and of GitHub Actions' run-retention window,
  neither controlled by this repository.
- `VOC-040-EV-00`: to be produced by `VOC-040-T00` (the merged `AGENTS.md` diff
  itself, satisfying `VOC-040-TEST-00`'s review procedure). Does not exist yet;
  this package is a draft proposal, not evidence of the task having been
  performed.
