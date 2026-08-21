# VOC-105 — Ready-for-review reruns unchanged exact-SHA CI and model review: Specification

## Objective and requirement source

Stop redundant full CI and independent model review on `ready_for_review` when
the PR base/head pair is unchanged and trusted exact-SHA evidence already
exists, as recorded in
[GitHub issue #872](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/872).

Today the caller subscribes to `ready_for_review` so draft-aware merge-gate can
re-evaluate an unchanged SHA after evidence completes. That same event also
starts `ci` and `review` / `plan-review` unconditionally, repeating minutes of
compute for code and verdicts that have not changed.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time grounding:

| Item | Current state |
| ---- | ------------- |
| Caller trigger | `.github/workflows/pipeline.yml` includes `ready_for_review` among PR types |
| CI / review entry | `ci`, `review`, and `plan-review` run when `action != closed` |
| Merge-gate | Already refuses draft merge; re-evaluates after review jobs finish |
| Incident | PR #868 runs 32492018782 → 32492586621; PR #869 runs 32493543324 → 32494037984 |
| VOC-102 boundary | Explicitly deferred "duplicate exact-SHA reviews" as a separate follow-up |

## Scope and non-goals

In scope:

1. Detect the safe unchanged-SHA ready-for-review case before starting full CI
   or model review.
2. Reuse prior evidence only when every fail-closed precondition in
   `VOC-105-D01` holds.
3. In that safe case, skip full CI and model review; still run deterministic
   merge-gate re-evaluation so a non-draft PR can merge.
4. Preserve draft-never-auto-merge.
5. Fail closed to the normal full CI + review path whenever reuse is unsafe.
6. Never accept human/implementer comments as reusable verification authority.
7. Preserve exact-SHA stale-run protections and independent-role separation
   (implementer cannot approve or merge its own work; only trusted App-signed
   PASS/PASS WITH NON-BLOCKING FINDINGS counts).
8. Deterministic shared-infra policy tests plus calling-repo fixture/foundation
   coverage.
9. Controlled live proof of a draft → ready transition with unchanged base/head
   (operator-owned live evidence; metadata-only).
10. Update infra/caller docs only where current text would otherwise claim
    ready_for_review always re-runs full CI and model review.

Non-goals / explicitly excluded:

- Deprecated action inputs, Node runtime warnings, dependency alerts, and
  deterministic remediation preflight (separate roots per issue #872).
- Weakening merge-gate draft blocking, App-signed verdict trust, or
  expected_base_sha / expected_head_sha stale-run guards.
- Application, migration, signup-policy, secrets, database, or
  `infra/monitoring/` inventory ID changes.
- Granting the implementer general Actions credentials.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4** (semantic CI/CD verification-gate timing).
- **Measured path floor at drafting:** **R3** for `.github/workflows/` and
  related governance automation. Semantic escalation to R4 is proposed because
  incorrect reuse could skip independent verification.
- Protected areas: CI entry conditions; independent review / plan-review entry;
  merge-gate draft and exact-SHA gates; trusted App publisher identity for
  verdicts; live-evidence attestation binding where required.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

The R4 value is a **draft proposal for the reviewing human at adoption time,
never a determination**. The path-based classifier and independent verifier
govern each task PR.

## Decisions

`VOC-105-D00`: On `pull_request` action `ready_for_review`, the governed
pipeline MUST decide whether prior exact-SHA CI and independent-review evidence
may be reused before starting full CI or model review. The decision MUST be
deterministic and fail closed. Ordinary `opened` / `synchronize` / `reopened`
paths continue to run full CI and review as today.

`VOC-105-D01`: Reuse is permitted only when ALL of the following hold for the
current PR:

1. The event is `ready_for_review` (not a code-changing synchronize).
2. The live PR base SHA and head SHA equal the base/head pair that produced the
   reusable prior evidence (exact 40-character match).
3. All required checks for that exact head are successful (or an explicitly
   allowlisted equivalent success state already accepted by merge-gate today,
   such as SKIPPED where merge-gate already treats SKIPPED as non-blocking).
4. A trusted App-authored independent verification comment exists for that exact
   head (and base where the existing publisher already binds base), with
   VERDICT PASS or PASS WITH NON-BLOCKING FINDINGS.
5. The verdict binds the same package path / task ID / authority issue (or plan
   package identity for `plan/` PRs) that the current PR still carries.
6. If the prior review lifecycle required live-evidence attestation, that
   attestation is present and still valid for the exact head.
7. Identity and trust checks match the existing merge-gate / publish-review
   rules (App installation identity, isolated publisher check success,
   single-line safe scalars). Human, implementer, and non-App comments NEVER
   qualify.

If any condition fails or cannot be proven, reuse is refused and the normal
full CI + review path runs.

`VOC-105-D02`: When reuse is permitted, the pipeline MUST NOT invoke full CI
and MUST NOT invoke the model review / plan-review path. It MUST still run the
deterministic merge-gate re-evaluation required to merge a now-ready PR.
Draft PRs remain non-mergeable.

`VOC-105-D03`: When reuse is refused, behavior MUST match today's normal path:
full CI, then the applicable independent review (`review` for `agent/` heads,
`plan-review` for `plan/` heads), then merge-gate, with existing stale-run and
role-separation protections intact.

`VOC-105-D04`: Exact-SHA stale-run protections remain authoritative. A newer
push, base update, or mismatched `expected_head_sha` / `expected_base_sha`
invalidates reuse and any in-flight skip decision. Cancellation of superseded
runs remains an optimization only; correctness rests on exact-SHA guards.

`VOC-105-D05`: Independent-role separation is unchanged. The implementer cannot
approve or merge its own work. Only the trusted App-signed publisher path can
create reusable PASS evidence. Remediation MUST NOT treat a successful reuse
skip as a FAIL or as missing review.

`VOC-105-D06`: Deterministic shared-infra policy tests MUST cover positive
reuse, negative refuse-and-rerun cases (changed SHA, missing/failed checks,
missing/waiting/failing/malformed/untrusted verdict, missing live-evidence
attestation when required, identity/scope mismatch, human-comment rejection),
draft-never-merge, and regression that synchronize still always re-runs CI and
review. Calling-repo foundation fixtures MUST assert caller wiring for the
optimized ready_for_review path.

`VOC-105-D07`: Controlled proof uses a real draft → ready transition with
unchanged base/head after T00 is live. Latency/compute evidence records only
allowlisted Actions metadata (run IDs, job names, conclusions, boolean
reuse/skip decision). Forbidden: logs, artifacts, secrets, tokens, OAuth/
session/cookie material, user identifiers.

`VOC-105-D08`: Out of scope for this package — deprecated action inputs, Node
runtime warnings, dependency alerts, and deterministic remediation preflight.

## Open questions

1. **Exact skip mechanism placement (implementation detail):** Whether the
   reuse decision lives primarily in calling-repo `pipeline.yml` `if:`
   conditions fed by a reusable preflight job, or inside reusable
   `ci.yml` / `review.yml` / `plan-review.yml` early-exit helpers, is left to
   the implementer so long as `VOC-105-D00`–`D05` hold and docs stay truthful.
2. **Check-name allowlist stability:** Required-check identity should reuse
   merge-gate's existing check filtering rather than inventing a second
   allowlist unless tests prove a gap. Confirm at implementation against
   current `merge-gate.yml` behavior.

## Security and privacy

- No new secrets.
- Reuse authority is App-signed exact-SHA evidence only.
- Evidence and proof records are metadata-only.
- Incorrect reuse is the primary residual risk; fail-closed tests and exact-SHA
  guards are the mitigation.

## Data, migrations, analytics, and accessibility

- Data / migrations: None — evidence-backed non-applicability.
- Analytics: None — evidence-backed non-applicability.
- Accessibility: None — evidence-backed non-applicability.
