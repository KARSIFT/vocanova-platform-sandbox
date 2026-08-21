# VOC-104-T01 — Controlled draft-to-ready evidence

evidence_id: `VOC-104-EV-01`
source_transition_claimed: `true`

Package: `specs/changes/VOC-104-ready-for-review-reruns-unchanged-exact-sha-ci`
Change: `VOC-104`
Task: `VOC-104-T01`
Observed: `2026-08-21` (UTC)

The operator used the already reviewed T00 PR as the controlled source. Its
draft-to-ready transition kept the base and head unchanged, selected exact-SHA
reuse, preserved the required CI context, re-evaluated merge policy, and merged.
The separate deterministic evidence carrier was created by auto-advance; no
general implementer run was started for this operator-owned task.

## Allowlisted source metadata

| Field | Observed value |
| --- | --- |
| Source pull request | `878` — merged |
| Source base SHA | `aedb3d35a65919a1364be385ab33f0d6c9fbd20f` |
| Source head SHA | `df888cd51462fa8ea05c45002f4ae1f34c068678` |
| Prior successful pipeline run | `32527436805` |
| `ready_for_review` pipeline run | `32528668647` |
| Source merge commit | `6ef5929032fb58aaa558f3529b7e62d7fe62393a` |
| Shared reuse-policy SHA | `a592dd8fa8ea1718c0f2f632b648213b53a47e57` |
| Reuse decision | `true` |

## Source transition outcomes

| Required observation | Outcome |
| --- | --- |
| `ready-for-review-reuse / decide (ready_for_review)` | PASS — success |
| Required `ci / ci` context | PASS — success |
| `Record exact-SHA CI evidence reuse` | PASS — success |
| Project checkout | PASS — skipped |
| Shared-infrastructure checkout | PASS — skipped |
| Full application validation | PASS — skipped |
| Model review publisher | PASS — skipped |
| `merge-gate / report-status` | PASS — success |
| Unique App-authored pre-merge attestation | PASS — present before merge |
| `merge-gate / auto-merge` | PASS — success |

## Authoritative proof state

Proof state is intentionally not duplicated as a mutable markdown flag. Before
qualification, the adjacent
`.karsift/live-evidence/VOC-104-T01.result.json` is absent and the task remains
waiting. After the read-only `verify-ready-for-review-reuse` workflow succeeds
on the exact carrier head, the repository reconciler alone writes that result.
It contains only the contract-allowlisted workflow, run, job, SHA, conclusion,
and bounded timing metadata. The trusted reconcile comment binds the resulting
commit, and that post-reconcile head must receive a fresh independent review
before merge.

Any later evidence edit changes the carrier SHA and invalidates that binding.
The generated result must then be removed and requalified; prose never overrides
the machine record.

## TEST-09 evidence split

- The positive live optimized path is recorded here as `VOC-104-EV-01` and
  satisfies `VOC-104-TEST-08` plus the positive side of `VOC-104-TEST-09`.
- The unsafe-path/full-path matrix is recorded in `VOC-104-EV-00`
  (`t00-evidence.md`) through `VOC-104-TEST-03` to `VOC-104-TEST-06`; T01 does
  not manufacture a second live failure to duplicate those deterministic
  fail-closed cases.

No workflow logs, artifacts, secrets, tokens, sessions, OAuth data, cookies,
email addresses, or other user identifiers were read or recorded.
