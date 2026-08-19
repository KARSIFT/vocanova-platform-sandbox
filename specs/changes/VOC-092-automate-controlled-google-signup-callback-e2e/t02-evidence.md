---
evidence_id: VOC-092-EV-02
task_id: VOC-092-T02
acceptance_criteria:
  - VOC-092-AC-09
  - VOC-092-AC-10
tests:
  - VOC-092-TEST-13
  - VOC-092-TEST-14
  - VOC-092-TEST-16
date: 2026-08-19
related_change: VOC-092
accountable_owner: unassigned
gate_status: repository-complete-exact-sha-review-pending
live_staging_synthetic_claimed: false
---

# VOC-092-T02 — Harness boundary documentation and VOC-088 evidence remediation

## Scope and outcome

Operations documentation now states what the repository-managed controlled-signup
OAuth callback E2E harness proves, what still requires human Google UI login, and
the periodic real-provider audit cadence. VOC-088 `t03-evidence.md` no longer
contains the self-referential `reviewed_sha` placeholder, and the scrubbed
allowlisted sign-in row explicitly records reaching onboarding/home without HTTP
503.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Operator procedure (harness boundary + audit) | `docs/operations/staging-controlled-signup.md` |
| Operations index link | `docs/operations/README.md` |
| Deterministic doc/evidence tests | `scripts/foundation/voc092-operator-docs.test.mjs` |
| VOC-088 evidence remediation | `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/t03-evidence.md` |
| This evidence | `specs/changes/VOC-092-automate-controlled-google-signup-callback-e2e/t02-evidence.md` |

## Acceptance mapping

| Acceptance criterion | Result | Evidence section |
| --- | --- | --- |
| AC-09 harness boundary documented | pass | § Documentation updates |
| AC-10 VOC-088 t03 findings remediated | pass | § VOC-088 evidence remediation |
| AC-09 human audit cadence | pass | § Documentation updates |

## Documentation updates

`docs/operations/staging-controlled-signup.md` adds **Repository-managed OAuth
callback E2E harness (VOC-092)** with explicit **What the harness proves** and
**What the harness does not prove** tables, local/CI run commands, and retention
of the unchanged staging OAuth-start synthetic. **Human sign-in verification**
now references the harness boundary and records the audit cadence: after auth
callback or allowlist policy changes and at least quarterly.

## VOC-088 evidence remediation

In `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/t03-evidence.md`:

- Removed `reviewed_sha: bind-at-independent-review` until an independent
  verifier binds the exact reviewed revision.
- Expanded the allowlisted first-time Google user row to
  `Reached staging onboarding/home without HTTP 503`.

## Deterministic validation

```bash
node --test scripts/foundation/voc092-operator-docs.test.mjs
node --test scripts/foundation/voc088-operator-procedure.test.mjs
```

Results on the task working tree:

- three VOC-092-T02 operator/documentation tests passed;
- three VOC-088-T03 operator/evidence structure tests passed (regression).

## Privacy and redaction

No repository secret value, real email address, OAuth authorization code, OAuth
`state`, session cookie, or mint token appears in the updated operations
documentation, foundation tests, or this evidence file.
