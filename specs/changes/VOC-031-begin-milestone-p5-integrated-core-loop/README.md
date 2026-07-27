# VOC-031 — Begin Milestone P5: Integrated Core Loop

**Draft package — not adopted, not approved, and not implementation authority.**
Human adoption, resolution of the stated open decisions, and separate
implementation authorization are required before work begins. No authorization,
approval, activation, deployment, or closure field is set by this draft.

## Identity and lifecycle

- Package ID: `VOC-031`; canonical path:
  `specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/`.
- Lifecycle: `draft`; every authorization field in `change.yaml` remains at its
  unadopted default (`approval_status: not-approved`,
  `implementation_authorized: false`, `automatic_merge_allowed: false`,
  `production_impact: unknown`, `repository_adoption_status: not-adopted`).
- Proposed risk: **R3** (proposal only — not a determination). This milestone
  adds three new owned tables (`user_onboarding_profiles`, `email_change_links`,
  `account_deletion_requests`) and their migrations (`/apps/api/migrations`,
  `/apps/api/ent/schema` — R3 path floor), two new business modules
  (`apps/api/business/users`, `apps/api/business/accounts`), a real,
  irreversible account-deletion procedure against per-user data, an
  email-change flow that mirrors and extends the magic-link auth pattern, and
  an additive change to the already-shipped `GET /api/v1/me` response. The
  path-based classifier (`scripts/governance/classify-change-risk.sh`) floors
  these paths at R3; the implementation-time classifier, builder, verifier,
  and applicable human authority govern the actual class. One consequence
  inside this otherwise-routine-R3 package — production enablement of account
  deletion specifically — is flagged for explicit founder go/no-go at the
  future production-activation decision, distinct from this draft's own
  merge/staging scope (see specification.md's "Risk and protected areas").
  `VOC-031-D02`/`D03` are genuine open decisions (one this draft's own
  minimal-scope default, one a new product ambiguity this draft surfaces but
  does not resolve); this draft does not decide them.
- Decision owner: founder; target branch: `develop`; request source: free text
  (the DOC-12 §5 P5 paragraph plus the supplied request, which records five
  founder decisions already made — see specification.md `VOC-031-D01` —
  grounding DOC-03, DOC-05 §§6,13,16, DOC-06 §§6,7,9,14,15, DOC-07, DOC-08,
  and DOC-10 §7 in full).
- A-003 is active: routine R3 requires strengthened controls and exact-SHA
  independent verification but not standing steward/founder approval solely
  because it is R3. An R4 product/scope or material user-trust decision remains
  founder-controlled. Every P5 PR requires Claude Code review; authorization,
  migration, irreversible-action, and cross-screen-consistency findings block
  release. EHR is not presumed.

## Objective and requirement source

Begin DOC-12 §5 P5: combine everything built by A1/P1/P2/P3/P4 into one
coherent, reliable, mobile-first journey — cross-feature integration,
reliability/recovery, accessibility/performance, and final UX consistency.
Ground the design in DOC-03 (onboarding flow, UX principles, accessibility),
DOC-05 §§6,13,16 (`user_settings`, `user_onboarding_profiles`, idempotency,
account deletion), DOC-06 §§6,7,9,14,15 (auth conventions, authorization,
idempotency, account deletion, no-queue constraint), DOC-07 (API/DTO rules,
the already-named `POST /api/v1/account-deletion-requests`), DOC-08 (routing
table, Settings/account scope, Lighthouse thresholds), DOC-10 §7 (documented
end-to-end testing strategy), and DOC-11 §3 (kill switches). The A1
auth/session/requester context (VOC-025), the P1 content/learning foundation
(VOC-026), the P2 `reviews` module (VOC-027), the P3 `aifeedback` module
(VOC-028), and the P4 `missions`/`gamification` modules plus the
already-created but minimally-wired `user_settings` table (VOC-030) carry
forward and are the direct dependency this package activates and completes.

## Scope, non-goals, risk, and protected areas

Scope is a fixed ordered twelve-task sequence (`T00`–`T11`, detailed in
`tasks.md`): onboarding backend and frontend (`T00`–`T01`); Settings backend
(`T02`); email-change-verification backend (`T03`); account-deletion backend
(`T04`); Settings/account frontend (`T05`); a cross-feature reliability and
recovery pass over the full loop (`T06`); installing automated accessibility
testing — axe-core + Playwright, both net-new to this repository — across the
three supported layouts (360px, 430px, ≥1024px desktop) (`T07`); the first
full core-loop end-to-end Playwright suite (`T08`); installing automated
Lighthouse CI against the DOC-08 thresholds (`T09`); a final UX-consistency
design audit (`T10`); and evidence/mock-inventory/staging-evidence/gate
readiness (`T11`).

Excluded: renaming any already-shipped route to match DOC-08's documented
table, or building the documented-but-never-shipped `/words` routes (informational
carry-forward only, per the supplied request — see `VOC-031-D00`/`D08`); a
general Settings surface beyond the founder-directed field list;
`appLanguage` actually changing rendered UI language (no i18n infrastructure
exists — `VOC-031-D06`); onboarding resumability/partial-save (`VOC-031-D02`);
production deployment; real secrets; R1/R2/L1 work itself (this package only
produces P5's own gate evidence).

Protected: database migrations, Ent schemas, the new `users`/`accounts`
business modules, the already-shipped `GET /api/v1/me` response this package
additively extends, magic-link/session conventions this package reuses (but
does not modify) for the new email-change flow, and the real,
irreversible account-deletion procedure. Rollback must preserve learner-owned
settings/onboarding/session state and must never leave a partially-applied
account-deletion request in an inconsistent state (deactivated but never
purged, or purged without having been deactivated first).

## Verification, approvals, release, and closure

Every P5 PR requires Claude Code review bound to the exact final SHA;
authorization, migration, irreversible-action, cross-screen-consistency, and
accessibility/performance-automation findings block release. Run installed
commands (`pnpm validate`, `pnpm test`, `pnpm build`, the
`scripts/governance/*` checks as applicable, plus the Go format/vet/test/build
and web lint/typecheck/build suites discovered at the adopted base) and the
deterministic domain/migration/transaction/contract/consistency/accessibility/
performance tests this package adds. Staging validation and rollback
rehearsal are required before the DOC-12 P5 gate can be evaluated; live
staging evidence is blocked until the F3 staging environment exists
(`VOC-031-DEP-03`, following the same pattern VOC-025's `staging-evidence.md`
established for `VOC-025-DEP-01`). This draft grants no approval, merge,
activation, credentials, deployment, or closure authority, and the package is
not adopted.
