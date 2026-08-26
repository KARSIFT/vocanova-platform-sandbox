# VOC-125 — Manual resume of an existing implementation PR must supply exact retry bindings

| Field | Value |
|-------|-------|
| Package | `VOC-125` |
| Title | Manual resume of an existing implementation PR must supply exact retry bindings |
| Path | `specs/changes/VOC-125-manual-resume-of-an-existing-implementation-pr` |
| Status | `draft` |
| Risk | `R4` (draft proposal; implement-recovery identity and `tooling/governance/` fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1020](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1020) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

A governed operator cannot safely resume an existing implementation carrier after
its bounded automatic remediation has stopped, even when the original task
issue and PR remain the only valid authority and carrier.

The reusable `implement.yml` already requires `expected_head_sha` and
`expected_base_sha` on `attempt != 1`. Automatic `remediate.yml` retries receive
those values from the pull-request review path. The documented caller
`action=implement` entrypoint exposes `attempt` but cannot supply a safe
existing-carrier recovery identity, and the caller `implement` job forwards
neither SHA.

Retrying as attempt 1 would violate the one-retry limit. Deleting the branch
or opening another PR would discard the existing carrier. That is not
acceptable.

### Live reproduction (2026-08-26)

| Item | Value |
|------|-------|
| Adopted task being resumed | [#1003](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1003) (`VOC-122-T00`) |
| Existing caller PR | [#1012](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1012) — draft; must remain the carrier |
| Pipeline run | [`32966618512`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32966618512) |
| Implement job | [`98170418081`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32966618512/job/98170418081) |
| Caller protected ref at dispatch | `develop@f0f7a54283c3527f40544fd236516dc3a5f4dc82` |
| Authoritative infrastructure ref | `KARSIFT/karsift-ai-infra@f406cc95a3f853e8aef5bf8bcf22d37a29d64547` |
| Existing caller branch/head | `agent/voc-122-voc-122-t00@0b7be8c531be8300d5a1d5534acc83bf4d6a1791` |
| Prior App-signed review base | `e910eb4a21d48bbb5b3e0c30b8ee647d64683dbe` |
| Dispatch | `action=implement`, `attempt=2`, no recovery identity |
| Failure | `Create implementation branch` finds the existing remote branch, then cannot establish the required exact-head remediation binding; stops before model resolution or Composer execution |
| Outcome | Run cancelled after the failure was identified; no implementation commit, source bundle, caller publication, or replacement PR |

Root cause: caller `.github/workflows/pipeline.yml` `workflow_dispatch` and its
`implement` job do not expose or forward a verified existing-carrier identity
into `implement.yml`.

## Required outcome (summary)

Use one largest-safe coherent task to provide one fail-closed governed recovery
route for an existing implementation PR:

1. Derive immutable recovery identity from an explicit existing PR number and
   its App-signed exact-revision review/metadata. Do not trust free-form SHA
   inputs on the caller dispatch contract. If SHA inputs remain on
   `implement.yml` for automatic remediation, bind them to the live PR before
   any model or mutation step.
2. Resume the same task issue, branch, and PR. Never create a replacement
   carrier merely because automatic retry is exhausted.
3. Preserve the existing attempt number and one-retry maximum. Never
   reclassify attempt 2 as attempt 1 or allow attempt 3.
4. Fail closed for missing/malformed/stale head or base, wrong
   PR/branch/repository/task/package/authority issue, absent-or-foreign
   App-signed review evidence when a review exists, changed remote head, and
   already-closed/completed tasks.
5. Keep configured Cursor Composer for implementation and Cursor Grok for
   exact-revision independent review. Preserve source/caller publication
   leases, workflow-file permission isolation, risk classification, protected
   checks, and no credential fallback.
6. Add deterministic source and caller tests for a valid existing-carrier
   resume and every mismatch class above.
7. Update current-state workflow comments/docs and the caller fixture/pin if
   authoritative shared infrastructure changes.
8. After the exact reviewed repair is live, resume existing VOC-122 issue
   #1003 / PR #1012 through this route rather than creating a new VOC-122
   task or PR.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task. No OpenAI credential or execution
route is needed or authorized. Do not print credential values.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Bind existing-carrier resume identity, with tests, docs, caller pin, and existing VOC-122 retry | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because the change mutates
exact-head remediation bindings and implementation-carrier identity. The
path-based classifier and independent verifier remain authoritative; this
draft proposal is not a determination.
