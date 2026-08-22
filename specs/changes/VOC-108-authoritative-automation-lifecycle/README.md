# VOC-108 — Authoritative automation lifecycle state

| Field | Value |
| --- | --- |
| Package | `VOC-108` |
| Status | `adopted` |
| Risk | `R4` actual task floor (`R3` at plan adoption; pinned governance fixtures establish R4) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#903](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/903) |
| Target branch | `develop` |
| Approval | `autonomously-adopted-after-independent-verification` |
| Implementation authorized | `true` |
| `automatic_merge_allowed` | `true` |

## Problem

The VOC-107 delivery exposed a single lifecycle-state root: several automation
consumers treated historical or non-authoritative GitHub state as current truth.
An older failed check remained attached to a later successful exact SHA; a
cross-repository closing reference closed the caller task before its caller PR
merged; release eligibility treated issue state as completion evidence; and
overlapping release/check paths emitted contradictory decisions around the same
promotion PR. The implementation was eventually delivered, but the races wasted
validation/model time and temporarily advanced release state too early.

## Required outcome

1. Select the newest authoritative evidence for each required logical check on
   the exact repository, PR, base SHA, and head SHA; obsolete failures must not
   poison a later successful run and a later failure must not be hidden by an
   older success.
2. Use non-closing references in cross-repository PRs and reserve caller task
   closure for the caller repository's verified merge path.
3. Treat a roster task as complete only when its issue has immutable,
   App-authored evidence for the expected merged caller PR, package, task, and
   exact head—not merely because the issue is closed.
4. Give release promotion one idempotent final merge authority and suppress
   stale post-merge pending decisions.
5. Re-evaluate release eligibility when required external checks become terminal
   without rerunning expensive unchanged-SHA CI or independent review.
6. Add deterministic fixtures for duplicate runs, stale failure/later pass,
   later failure/older pass, cross-repository references, premature closure,
   release-evaluation races, and duplicate merge triggers.

## Task

`VOC-108-T00` implements the shared lifecycle helpers, workflows, documentation,
and deterministic fixtures in `KARSIFT/karsift-ai-infra`, with caller-side
contract/evidence updates only where required.

The adopted package does not broaden its own authority. It deliberately excludes the
scheduled-synthetics branch-selection footgun, stale operational-failure marker
lifecycle, Node runtime deprecations, historical issue cleanup, application
behavior, deployments, OAuth policy, secrets, databases, and monitoring
inventory.
