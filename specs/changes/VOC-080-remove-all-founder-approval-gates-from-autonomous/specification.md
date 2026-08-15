# VOC-080 — Remove All Founder Approval Gates from Autonomous Engineering Workflows: Specification

## Objective and requirement source

Remove every founder-approval gate from VocaNova's autonomous engineering
workflows so agents and workflows progress when deterministic checks,
independent verification, scope, and other non-founder gates pass — under
any risk class — without waiting for a founder `approved` comment.

Requirement source:
[GitHub issue #627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627)
(effective request date 2026-08-15; founder `m-e-h-r-d-a-a-d`).

This request **supersedes** VOC-075 / issue #573 “approve only R4” for
engineering-workflow gates. It does **not** remove independent verification,
CI, risk classification, protected-path checks, rollback requirements,
secrets isolation, least-privilege credentials, builder/verifier separation,
or the obligation to ask when product requirements are genuinely ambiguous.

**Transition authority (issue #627):** this governance replacement is
evaluated under the authority effective before it. It cannot authorize its
own adoption. One final founder approval under current (A-003 / VOC-075)
policy is required for the exact transition revision. Once that revision is
effective, no later founder approval gate remains.

## Confirmed findings (from issue #627 and drafting-time repo read)

| Finding | Evidence / location |
|---------|---------------------|
| R4 never auto-merges; founder `approved` merges any risk | `karsift-ai-infra` `merge-gate.yml` report-status + `approve-and-merge`; calling `pipeline.yml` passes `founder_username` |
| Founder comment can proceed despite missing/non-PASS independent verification | `merge-gate.yml` approve-and-merge step comments (2026-08-14 override) — **forbidden** by #627 AC |
| Unparseable risk → founder approval | merge-gate `risk=unknown` branch |
| Residual `automatic_merge_allowed: false` → founder approval | merge-gate Path 3 exclusion + status text |
| Plan packages draft unauthorized; human adoption flip | planner prompts; `adopt.yml` verifies adopted fields then opens tasks |
| No `workflow_dispatch` on adopt | `adopt.yml` `on: workflow_call` only; VOC-040 / AGENTS.md recovery via `gh run rerun` |
| Bot merge can leave package merged as draft | #627 cites plan PR #625 / VOC-079; recovery PR #626 still under founder architecture |
| Release still carries founder_username / comment retry path | `pipeline.yml` release job; `release.yml` header (auto_release primary since 2026-08-08) |
| Canonical docs still say R4 requires founder approval | `AGENTS.md`, `CLAUDE.md`, `approval-matrix.md`, `change-risk-classification.md`, A-003 §R4, DOC-15 |

## Scope and non-goals

In scope:

1. **Authority text** — successor amendment (recommended A-004) and
   transition-state / activation record; supersede A-003 / VOC-075 only where
   they require founder approval for engineering-workflow gates.
2. **`karsift-ai-infra` runtime** — `merge-gate.yml`, `adopt.yml` (and/or
   plan-merge adoption path), `plan.yml` / `plan-review.yml` comments,
   `release.yml`, `remediate.yml` as needed, prompts/README that claim
   founder approval is required, role/config notes.
3. **Autonomous adoption** — independently reviewed plan package that passes
   governance + deterministic validation transitions to
   `status: adopted` / `implementation_authorized: true` (and matching
   nested fields) automatically; records exact revision, review evidence,
   resolved/deferred decisions, risk, and authority provenance; idempotent;
   recoverable via explicit dispatch without replaying an old GitHub event.
4. **Merge behavior** — R0–R4 eligible PRs merge after CI, governance,
   scope, and independent-verification gates pass; no founder comment to
   override failed/missing gates; unparseable/inconsistent risk fails for
   correction.
5. **`automatic_merge_allowed`** — retire or neutralize as a
   founder-attention mechanism (`VOC-080-DEP-02`).
6. **Release/deploy** — preserve auto develop→main promotion and
   push-to-main production deploy; remove residual founder-comment /
   environment-reviewer requirements on repository-controlled paths;
   failed deploys stay fail-closed until remediation checks pass.
7. **Caller-repo reconciliation** — AGENTS.md, CLAUDE.md, DOC-15/16,
   matrices, templates, PR template, `pipeline.yml` wiring/comments,
   repository-settings documentation, and any autonomy markers that must
   move in lockstep.
8. **Tests + rehearsal + activation** — harness coverage; sandbox/dry-run
   proof; one-final-founder-approval activation; post-activation path for
   VOC-079 / #624.

Non-goals / explicitly excluded:

- Weakening or removing independent verification, CI, risk classes,
  protected-path floors, rollback, secrets isolation, or least privilege.
- Replacing founder gates with another standing human approval role (EHR
  remains exceptional-only).
- Allowing implementer self-review of the same exact revision.
- Rewriting historical A-003 / VOC-075 / audit narratives as if prior rules
  never existed.
- Implementing VOC-079 nginx cutover content.
- Snapshot-then-recheck-drift promotion tasks (not applicable; this is new
  governance/workflow content).
- Adopting or authorizing this package from within the draft; claiming
  post-transition rules authorize this package's own land.

## Risk and protected areas

Builder assessment: **R4** (draft proposal for the reviewing human at
adoption — not a determination).

Path floors from `.github/approved-policy/protected-paths.yaml` that this
package is expected to touch include:

| Path | Floor |
|------|-------|
| `docs/governance/amendments/` | R4 |
| `docs/governance/a003-transition-state.yaml` | R4 |
| `docs/operations/15-…operating-model.md` | R4 |
| `scripts/governance/`, `tooling/governance/` | R4 |
| `AGENTS.md`, `CLAUDE.md`, `docs/governance/` (non-amendment) | R3 |
| `.github/workflows/` | R3 |
| `specs/templates/` | R3 |

Cross-repo `KARSIFT/karsift-ai-infra` workflow changes are outside this
repository's path classifier but are **in scope for the package's required
outcome** and carry equivalent governance/authority risk.

Protected effects: agent authority, CI/CD merge/release control, adoption
authority, production-release control, governance amendments. EHR is not
triggered by drafting. Application code, migrations, and product features
are out of default scope.

Under **active A-003** (until activation), R4 founder authority remains
required for this package's adoption and transition activation.

## Decisions, contradictions, security, and privacy

`VOC-080-D00` (recorded for traceability; formal acceptance at adoption):
No autonomous engineering workflow may wait on a founder `approved` comment
after the transition is effective. Deterministic checks + independent
verification + scope/governance gates are sufficient for progression at
every risk class, including R4. R4 remains a meaningful risk class with
stronger evidence, validation, verification, rollout, monitoring, and
rollback — not a founder-comment gate.

`VOC-080-D01`: An independently reviewed plan package that passes
governance and deterministic validation must become adopted /
implementation-authorized automatically (or via idempotent reconcile), and
must not remain silently merged as `draft`.

`VOC-080-D02`: No merge/release/deploy path may use founder approval to
override a failed or missing deterministic or independent-verification gate.
Unparseable or inconsistent risk fails closed for correction.

`VOC-080-D03`: This package is evaluated and activated under pre-transition
authority. Post-transition rules do not authorize VOC-080 itself.

Contradiction with VOC-075 / A-003 / AGENTS.md “R4 founder approval”:
explicit supersession for **engineering-workflow gates** after activation;
historical records preserved; cite issue #627.

Adoption decisions (resolved on PR #628; implementation details remain subject
to exact-revision independent verification):

1. **`VOC-080-DEP-00` — Amendment vehicle.** Use an **A-004** successor
   that supersedes A-003 only where founder approval is required for
   engineering-workflow gates; A-003 remains historical; activation flips
   `authority_model` / transition-state when evidence is complete.
2. **`VOC-080-DEP-01` — Product/legal R4 decisions.** All approval workflow
   gates removed entirely; founder still answers genuine product/legal/
   strategy ambiguities as **requirement clarification** before a package
   has stable AC — not as an `approved` comment on merge/adopt/release.
   “Initial public / major launch go/no-go” may require requirements to be
   clarified, but is not a founder approval gate.
3. **`VOC-080-DEP-02` — `automatic_merge_allowed`.** Choose **(a)**:
   neutralize — all new packages draft `true` including R4; merge-gate no
   longer treats `false` as founder-attention; optional historical backfill;
   keep the field for audit/compat unless adoption prefers full retire (b).
4. **`VOC-080-DEP-03` — Cross-repo sequencing.** Land infra behavior
   PRs in `KARSIFT/karsift-ai-infra` first (or lockstep), caller
   `pipeline.yml` + docs after inputs stabilize; this package is the
   authorizing change package for the required behavior in both places.
5. **`VOC-080-DEP-04` — Rehearsal.** Prove on
   `vocanova-platform-sandbox` and/or infra self-ci/smoke before activation
   closes; record run URLs for adopt, R4 auto-merge, reconcile-dispatch,
   release, and unparseable-risk fail-closed.
6. **`VOC-080-DEP-05` — Dependabot / non-agent PRs.** Recognized dependency
   bots follow the documented green-CI exception without founder approval;
   unrecognized bot identities fail closed.
7. **GitHub environment reviewers.** Confirm production (and any other)
   environment reviewer settings that still require founder click-approve
   are removed or documented as out-of-band ops with immediate
   doc follow-up per AGENTS.md settings rule.
8. **Risk.** R4 accepted.

Security / privacy: no new personal-data handling. Bot App credentials stay
secret-scoped. Audit evidence must not embed secret values. Separation of
builder and independent verifier remains mandatory.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None.
- **Analytics:** None.
- **Accessibility:** None. Governance / workflow / documentation package.
