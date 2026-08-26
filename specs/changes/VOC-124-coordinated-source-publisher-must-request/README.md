# VOC-124 — Coordinated source publisher must request workflow-write permission

| Field | Value |
|-------|-------|
| Package | `VOC-124` |
| Title | Coordinated source publisher must request workflow-write permission |
| Path | `specs/changes/VOC-124-coordinated-source-publisher-must-request` |
| Status | `draft` |
| Risk | `R4` (draft proposal; coordinated source-publisher token mint and `tooling/governance/` fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1013](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1013) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-121/VOC-123 made the governed implementer preserve authorized nested
`karsift-ai-infra` edits, bundle the exact committed head, and publish them
from a clean `publish-source` job. That publisher still cannot push an
authorized infrastructure commit that changes a GitHub Actions workflow.

GitHub refuses a GitHub App token that creates or updates workflow files
unless the minted token requested `workflows: write`. The live
`publish-source` mint asks for contents, issues, and pull-requests only.

### Live reproduction (2026-08-26)

| Item | Value |
|------|-------|
| Adopted task that hit the defect | [#1003](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1003) (`VOC-122-T00`) |
| Pipeline run | [`32958526215`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32958526215) |
| `publish-source` job | [`98147443377`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32958526215/job/98147443377) |
| Nested bundle head | `f90eb630743c8c523e2e6e8dff017acbb31a7f43` |
| Infrastructure base | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` |
| Bundle download / verify / token mint | succeeded |
| Failure | GitHub rejected the App-token branch push: `refusing to allow a GitHub App to create or update workflow ... without workflows permission` because the commit changes `.github/workflows/recover-actions-checks.yml` |
| Caller PR | [#1012](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1012) — intentionally draft and must remain unmerged; staged mirror/evidence still pins the prior infrastructure SHA |
| Defect locus | `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` `publish-source` mint: `permission-contents` / `permission-issues` / `permission-pull-requests` without `permission-workflows: write` |
| Installation | `karsift-ai-infra-bot` installation `148001476` already has repository `workflows: write`; no secret, account, or repository-setting change is required |

An ordinary VOC-122 retry cannot repair this. The caller executes
`implement.yml@main`, so the same broken token request would reject its own
workflow-file fix.

## Required outcome (summary)

Use one largest-safe coherent task to repair the infrastructure publisher
token mint, with the minimum required caller fixture/pin/evidence coordination:

1. Request `permission-workflows: write` only on the cross-repository
   `publish-source` App token that must publish authorized infrastructure
   workflow changes. Do not add it to the caller `publish` token.
2. Preserve the existing GitHub App boundary, exact bundle/head/base
   verification, isolated bare-repository publication, fail-closed credential
   behavior, branch lease, protected-branch controls, retry limits, and
   independent exact-revision review.
3. Add deterministic tests proving an authorized source bundle that changes
   `.github/workflows/**` receives that token permission, while missing App
   credentials, invalid bundles, stale bases/leases, and other publication
   failures still fail closed. Keep the caller publisher's workflow-file
   refusal and missing-workflows-permission defense.
4. Use one explicit, narrow, auditable bootstrap for this self-hosting repair
   because the live publisher cannot publish its own fix. Exhaust the
   exception after the exact reviewed repair merges. Do not use it to publish
   VOC-122's source bundle.
5. Correct current-state workflow/PR text in the same carrier path that
   inaccurately says human approval is still required under active A-004.
   Do not rewrite historical audit records.
6. Document exact rollback and live evidence, then retry the existing
   `VOC-122-T00` carrier rather than creating a replacement VOC-122 task or PR.
   Keep #1012 draft until that authoritative infrastructure merge exists, then
   update it to the exact infrastructure merge SHA with truthful evidence.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task. No OpenAI credential or execution
route is needed or authorized. Do not print credential values.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Request workflow-write on the infrastructure publisher token, with tests, bootstrap, docs, caller pin, and existing VOC-122 retry | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because the change mutates
coordinated second-repository publication of nested infrastructure workflow
files. The path-based classifier and independent verifier remain authoritative;
this draft proposal is not a determination.
