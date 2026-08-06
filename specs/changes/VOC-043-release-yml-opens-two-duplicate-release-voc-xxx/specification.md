# VOC-043 — release.yml Opens Two Duplicate "Release: VOC-XXX" Issues for the Same Package: Specification

## Objective and requirement source

Make `karsift-ai-infra/.github/workflows/release.yml`'s `check-completion` job
idempotent per package, so that two (or more) near-simultaneous
`issues: closed` events for tasks belonging to the same completed package
never result in more than one open "Release: `<change_id>`" issue.

Grounded in [issue #328](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/328)
in full, including its confirmed reproduction (VOC-039's issues #303-305 all
closing within seconds of each other, producing duplicate issues #310 and
#311 with identical bodies), its likely-cause analysis (the existing-issue
check and the issue-create call are not atomic across concurrent workflow
runs), and its own suggested fix (make issue-opening idempotent per package,
via either a re-check immediately before create or a concurrency lock keyed
on the package/change_id). Not yet approved by a founder or technical
steward — see `change.yaml`'s `requirement_approval_status`.

## Scope and non-goals

In scope:
- `karsift-ai-infra/.github/workflows/release.yml`'s `check-completion` job:
  restructure so that concurrent runs triggered by near-simultaneous
  `issues: closed` events for tasks in the same package cannot both pass the
  "no existing 'Release: `$change_id`' issue" check before either creates
  one. The concrete mechanism (see "Open questions" below) is left to the
  implementer, but the observable requirement is fixed: at most one open
  "Release: `<change_id>`" issue per package, ever, regardless of how many
  roster-closing events fire close together.
- Preserving every other existing behavior of `check-completion` exactly:
  which issues are recognized as tracked tasks (title-prefix match against
  `specs/changes/<change_id>-*`), the roster-completeness check itself
  (`.karsift/tasks.json`, all-closed logic), the "no `.karsift/tasks.json`"
  early-exit, the `karsift:release` label creation/description, and the
  created issue's title/body content (including the roster summary and the
  "Deploy is out of scope here" note). This package changes *whether a
  second issue can be created*, not *what the first one says or how it's
  found*.

Non-goals:
- Any change to the `promote` job (the second job in the same file, reacting
  to the founder's `approved` comment). Its own idempotency handling for a
  related-but-different race (two concurrently-*approved* releases; see its
  "No commits between" branch) is not implicated by issue #328's
  reproduction, which is scoped entirely to issue *creation*, not approval or
  promotion.
- Any change to `merge-gate.yml`'s auto-merge path, or to the fact that
  multiple tasks in a package's roster can legitimately merge within seconds
  of each other. That is intended, working auto-merge behavior — the defect
  is `release.yml`'s response to it.
- Any change to the human-facing approval semantics: replying `approved` (as
  the configured founder identity) on a "Release: `<change_id>`" issue must
  continue to work exactly as it does today, whether there is one such issue
  or (until this fix ships) two.
- Manual cleanup of the two already-open duplicate issues (#310, #311) from
  the VOC-039 incident this issue documents. That is an operational action on
  already-created GitHub issues, not a workflow-logic change — flagged as an
  open question for the reviewing human, not performed by this package.
- Any change to a different repository's copy of `release.yml`. This fix
  lands in `karsift-ai-infra`'s reusable workflow; whichever mechanism this
  repository uses to consume that reusable workflow (a pinned ref, a vendored
  copy, or similar) determines when the fix actually takes effect here, and
  is unknown at drafting time (see "Open questions").

## Risk and protected areas

`karsift-ai-infra/.github/workflows/release.yml` is shared CI/CD workflow
infrastructure: it gates the one human decision required to promote any
completed package from `develop` to `main`, for every package in the fleet
this reusable workflow serves, not only this repository. A defect here has a
blast radius beyond a single package's own files — consistent with
`docs/governance/change-risk-classification.md` naming CI/CD/release-gating
workflow changes as at least an R3-floor category. This package proposes
`R3` (see `change.yaml`). It does not touch a secret, a schema migration, or
production application code directly, so no higher class is proposed, but it
does touch the mechanism that gates production promotion — the reviewing
human's own judgment governs the final class, not this proposal.
`scripts/governance/classify-change-risk.sh` has not been run against a
real, task-scoped file list at drafting time, consistent with how prior
packages (VOC-039 through VOC-042) handled this field — that computation
belongs to the task's own implementation PR.

## Decisions, contradictions, security, and privacy

No `VOC-043-D00`-style founder/product decision is defined by this draft.
The requirement (issue creation must be idempotent per package) is fully
specified by the issue; the only genuinely open implementation-level choice
(which concrete idempotency mechanism to use) is recorded as an open
question below, not invented as a silent decision.

No contradiction is recorded — issue #328's account of the root cause is
internally consistent with the workflow file's own logic (confirmed by this
package's own drafting-time read of `release.yml`) and is not disputed by any
other source.

Security/privacy: this fix does not change who can approve a release, what a
release issue's body discloses, or any authorization boundary — the
`promote` job's founder-identity check (exact comment body `approved` from
the exact configured founder login) is untouched. The defect and its fix are
both scoped to a duplicate, non-authoritative artifact (a second issue with
identical, non-sensitive content: a roster summary already visible in each
task's own issue). No secret, credential, or personal-data field is
introduced, removed, or exposed by either the defect or the fix.

## Data, migrations, analytics, and accessibility

None. This package touches only a GitHub Actions workflow file's job
structure; no schema, migration, analytics event, or accessibility surface is
affected.

## Open questions

1. **Concrete idempotency mechanism.** Issue #328 suggests two options: (a)
   an existing-issue re-check performed atomically with creation, or (b) a
   concurrency lock keyed on the package/change_id. In GitHub Actions
   specifically, a job-level `concurrency:` group is evaluated before the
   job's steps run and cannot reference a step output computed inside that
   same job — only `needs.*.outputs` from an upstream job in the same
   workflow run, or values computable directly from the event context (e.g.
   `github.event.issue.title`), are available at that point. Since
   `change_id` today is only derived via a shell `grep -oP` step (GitHub
   Actions expressions have no built-in regex/split function), the most
   direct way to key a `concurrency:` group on `change_id` is to split
   `check-completion` into two jobs: a small `identify` job (computes and
   outputs `change_id`/`package_path`/`is_task`, no `concurrency:` needed
   since it only reads) and a `check-and-open` job that `needs: identify` and
   sets `concurrency: { group: "release-check-${{ needs.identify.outputs.change_id }}", cancel-in-progress: false }`.
   This serializes only the runs that share the same `change_id`, and — because
   GitHub Actions queues (rather than cancels, per `cancel-in-progress: false`)
   concurrent runs sharing a group — the second run in the queue re-runs its
   existing-issue check only after the first run's `gh issue create` has
   already completed, closing the race. This package proposes that approach
   as `VOC-043-T00`'s design, but flags it here as an implementation-level
   choice for the implementer/reviewer to confirm rather than a locked
   specification detail: an implementer with a simpler correct alternative
   (e.g. a single repository-wide `concurrency:` group covering all
   `check-completion` runs regardless of package, trading a small amount of
   unnecessary cross-package serialization for a smaller diff) should record
   that tradeoff explicitly in the implementation PR rather than silently
   picking one without noting the choice.
2. **How this repository actually consumes `karsift-ai-infra`'s
   `release.yml`.** This package's diff lands in
   `karsift-ai-infra/.github/workflows/release.yml` (present in this
   checkout as an untracked directory at drafting time — confirmed via `git
   status`). Whether this repository's own workflows reference that file via
   a version-pinned ref (in which case a bump is a separate, later step) or
   some other mechanism this drafting pass did not inspect is unknown and
   not asserted either way — flagged for the implementer to confirm before
   this fix is treated as taking effect for this repository's own future
   releases, not merely committed to the infra template.
3. **Cleanup of the two already-open duplicate issues (#310, #311).** Not
   performed by this package (see "Non-goals"); flagged for the reviewing
   human as an independent operational action.
