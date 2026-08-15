# VOC-080 — Impact Analysis

## Security and privacy

- **Authority surface:** Removing founder-comment gates increases reliance on
  deterministic CI, independent verification, scope checks, and fail-closed
  risk parsing. Weakening any of those would be a Critical/High defect —
  this package must strengthen or preserve them, never trade them for speed.
- **Override removal:** Today's merge-gate Path 2 can merge despite
  missing/non-PASS independent verification when the founder comments
  `approved`. Removing that override is a security *improvement* relative
  to issue #627 AC; do not reintroduce a silent human bypass.
- **Bot credentials:** Autonomous adopt/merge/release continue to use the
  existing GitHub App installation token pattern. No new long-lived PATs.
  Evidence files must not contain secret values.
- **Self-approval:** Builder/verifier role separation remains mandatory.
  Activation must not allow the implementer-role occupant to verify its own
  exact revision.
- **Personal data:** None expected. No application data-plane change.

## Data and migrations

None. No database schema, seed, or data backfill. Package YAML field
transitions (draft→adopted) are repository metadata, not data migrations.

## Analytics and accessibility

- **Analytics:** None — evidence-backed non-applicability.
- **Accessibility:** None — no UI change.

## Risks, dependencies, and evidence

- `VOC-080-R00`: **Self-authorization / premature activation** — workflows
  begin skipping founder gates before the exact transition revision is
  approved under pre-transition authority. Mitigation: T00–T06 land
  mechanisms with transition markers still inactive where needed; T07 alone
  flips effective; package cannot authorize itself (`VOC-080-D03`).
- `VOC-080-R01`: **Silent merged-as-draft recurrence** (#625 class) if
  adopt-on-merge is event-fragile. Mitigation: atomic adopt path +
  idempotent reconcile dispatch (AC-04/AC-05); rehearsal in T06.
- `VOC-080-R02`: **Doc/behavior drift** — AGENTS.md or DOC-15 still claim
  founder approval after gates are removed (or the reverse). Mitigation:
  T04 same-PR reconciliation rule; AC-09; TEST-08.
- `VOC-080-R03`: **Unsafe auto-merge of broken R4** if independent
  verification or CI is skipped. Mitigation: AC-02/AC-07; merge-gate keeps
  checks_ok + PASS verdict hard requirements; tests for fail-closed.
- `VOC-080-R04`: **Cross-repo skew** — caller pins `@main` infra while
  docs/pipeline disagree with merged infra behavior. Mitigation:
  `VOC-080-DEP-03` sequencing; T04 after T01–T03; rehearsal on sandbox.
- `VOC-080-R05`: **Residual GitHub environment reviewers** still block
  deploy after workflows are clean. Mitigation: open question 7; T03/T04
  settings + doc lockstep.
- `VOC-080-R06`: **Rollback difficulty** if audit history is rewritten.
  Mitigation: preserve historical A-003/VOC-075; rollback restores prior
  workflow revisions + docs from known SHAs without rewriting audit.
- `VOC-080-DEP-00`–`DEP-04`: see `change.yaml` / `specification.md`.
- `VOC-080-EV-00`–`EV-07`: per-task evidence files in this package
  directory (`t00-evidence.md` … `t07-evidence.md`).
