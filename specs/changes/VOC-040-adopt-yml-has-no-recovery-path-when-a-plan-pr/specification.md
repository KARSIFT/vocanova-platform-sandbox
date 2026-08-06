# VOC-040 — adopt.yml Has No Recovery Path When a Plan PR Merges Without Being Adopted First: Specification

## Objective and requirement source

Give this repository's operators a documented, supported way to recover when a
`plan/`-branch PR merges without first being adopted (its `change.yaml` edited to
`status: adopted` / `implementation.authorized: true`), instead of the recovery
depending on independently re-discovering an undocumented implementation detail of
`adopt.yml` under time pressure. Grounded in
[issue #301](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/301) in
full, including the live incident it describes (VOC-039's adoption: PR #299 merged
as a plain "merge" without the adoption edit; recovered via a follow-up PR #300 plus
manually re-running the original failed `adopt` job with `gh run rerun --failed`).
Not yet approved by a founder or technical steward — see `change.yaml`'s
`requirement_approval_status`.

## Scope and non-goals

This repository (`vocanova-platform-sandbox`) does not contain `adopt.yml` or
`plan.yml` — both are reusable workflows owned by the separate
`KARSIFT/karsift-ai-infra` repository and consumed here only via
`uses: KARSIFT/karsift-ai-infra/.github/workflows/adopt.yml@main` /
`.../plan.yml@main` in this repo's own `.github/workflows/pipeline.yml` (confirmed
by reading both files at drafting time: `adopt.yml` declares only
`on: workflow_call`, no `workflow_dispatch`, and `plan.yml`'s draft-PR-opening step
is likewise entirely infra-side). This package's own scope discipline (per its
governing planner prompt) forbids writing outside this new package's own directory
in *this* repository — it cannot edit `karsift-ai-infra`'s files even though the
untracked `karsift-ai-infra/` directory happens to be present in this working tree
as a local reference checkout (see this repo's `git status` at drafting time: `??
karsift-ai-infra/`, i.e. untracked, not part of this repository's own tracked tree).

Of the issue's three suggested improvements, this materially changes which are
achievable from inside this repository alone:

- **In scope for this package:** the first suggestion's documentation half —
  recording the `gh run rerun --failed`-on-the-original-run recovery procedure
  explicitly in a place this repository controls. The issue offers two candidate
  locations for this ("`adopt.yml`'s own header comment or in this repo's
  `AGENTS.md`"); only the second (`AGENTS.md`) is actually writable by this
  package, since `adopt.yml` itself lives in `karsift-ai-infra`. `VOC-040-T00`
  scopes this.
- **Out of scope, and flagged as an open question below, not silently dropped:**
  the issue's second suggestion (a `workflow_dispatch` entry point on `adopt.yml`
  itself, or a wrapper around it) and third suggestion (a checklist item or
  required-review note in `plan.yml`'s own draft PR body/template). Both require
  editing files that exist only in `KARSIFT/karsift-ai-infra`, a different GitHub
  repository this package has no access to and no authority over. Implementing
  either would require a separate change package (and, per that repository's own
  presumed governance — not itself examined here, since it is out of this
  repository's scope — a separate adoption decision) filed against
  `KARSIFT/karsift-ai-infra` directly, not against `vocanova-platform-sandbox`.

Non-goals: this package does not change `adopt.yml`'s or `plan.yml`'s actual
behavior (it cannot); it does not change the deliberate refusal behavior the issue
itself endorses as correct ("That refusal is the right behavior" — issue #301's own
words, echoing `adopt.yml`'s header comment); it does not retroactively re-litigate
VOC-039's specific recovery, which is already closed.

## Risk and protected areas

`AGENTS.md` is this repository's own top-level governance document, read into every
planner/implementer/reviewer prompt per this repository's own `.karsift/lessons.md`
convention for similar files. Adding a documentation section describing an existing,
already-proven-to-work recovery procedure (not introducing new behavior, new
authority, or a new automated action) is a low-risk, purely additive documentation
change. This package proposes `R1` (informational/documentation, no code or
production-behavior change) in `change.yaml`. This is a draft proposal for the
reviewing human at adoption time, not a determination — the reviewing human may
reasonably classify AGENTS.md edits higher given its role as a governance document
read by every future planning/implementation/review run; if so, that judgment
should be recorded at adoption time rather than assumed here.

## Decisions, contradictions, security, and privacy

No `VOC-040-D00`-style decisions are defined by this draft. The one substantive
choice this package makes — documenting via `AGENTS.md` rather than attempting the
other two suggested improvements — is not a product/founder judgment call; it is
forced by this repository's own file boundary (see "Scope and non-goals" above),
not a preference among equally-achievable options.

Security/privacy: none. This documents an existing, already-demonstrated-safe
procedure (`gh run rerun --failed` against a specific, already-completed workflow
run number) that requires the same `contents: write` / `issues: write` permissions
`adopt.yml` already runs with when GitHub Actions triggers it normally — no new
permission, credential, or secret is introduced, requested, or documented as
needed beyond what a repository maintainer re-running any existing failed Actions
run already has.

## Data, migrations, analytics, and accessibility

None. This package touches only `AGENTS.md`; no schema, migration, analytics
event, or accessibility surface is affected.

## Open questions

1. **The `workflow_dispatch` entry point and plan.yml checklist suggestions are
   not implementable from this repository.** As explained in "Scope and
   non-goals," both require a change to `KARSIFT/karsift-ai-infra`. This package
   flags, rather than silently drops, this gap: the reviewing human may want to
   separately file (or ask the karsift-ai-infra maintainer/founder to file) an
   equivalent request against that repository directly, since a documented
   workaround (this package's `VOC-040-T00`) closes issue #301's immediate ask but
   does not reduce how often the underlying merge-without-adopting mistake happens
   in the first place — only the second and third suggested improvements would do
   that. Whether that follow-up is worth doing is a founder/steward judgment this
   package cannot make.
2. **Whether the `gh run rerun --failed` procedure remains valid indefinitely.**
   It depends on two properties of `adopt.yml` and GitHub Actions that this
   package cannot guarantee going forward: that `adopt.yml`'s verify step re-reads
   `change.yaml` from the target branch's *current* tip (not a frozen snapshot
   from when the run was originally dispatched) and that the original run has not
   been garbage-collected by GitHub's own retention window (per issue #301's own
   framing: "it only works because the original run/job wasn't garbage-collected
   yet"). `VOC-040-T00`'s documentation must state this dependency explicitly
   rather than presenting the procedure as unconditionally reliable — see
   `acceptance-criteria.md`'s `VOC-040-AC-00`.
