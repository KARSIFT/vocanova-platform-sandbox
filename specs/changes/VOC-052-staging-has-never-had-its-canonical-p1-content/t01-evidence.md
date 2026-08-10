# VOC-052-EV-01 — T01 verification evidence

Evidence for `VOC-052-T01` (`VOC-052-AC-02`, `VOC-052-AC-03`,
`VOC-052-TEST-02`, `VOC-052-TEST-03`). Written by the founder-gate
delegate directly from real `deploy-staging.yml` run history, not by
the AI implementer pipeline - `VOC-052-T00`'s seed step has been live
on every real staging deploy since it merged, and this task's own
description says "No source change is expected in this task. Its job
is to produce and record verification evidence."

## Verification

`VOC-052-T00`'s `p1-content-seed` step (added to `deploy-staging.yml`)
has run on every staging deploy since it merged (PR #444). Across many
real deploy runs the same evening (2026-08-09/10 UTC), including
several captured in detail while investigating an unrelated issue
(#450/VOC-053), `tests/staging-e2e/core-loop.staging.spec.ts`'s step 3
("discover a situation and open a word" - the exact assertion issue
#437 originally reported failing, `situationLinks.count()` returning
0) has passed in **every single run**, including runs where the
overall test later failed at an unrelated step:

- Run [31336599551](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31336599551): overall pass.
- Run [31342259422](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31342259422): overall pass (1 passed, 10.5s).
- Run [31342544145](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31342544145): overall pass.
- Several other runs the same evening failed only at step 7
  ("progress reflects the completed reviews", VOC-053's separate,
  already-investigated issue) - confirmed via each failure's own
  Playwright output naming step 7 specifically, never step 3. A step-3
  failure would abort the test before step 7 is ever reached, so every
  one of these runs is independent confirmation that discover-step
  content resolution succeeded.
- Run [31343738781](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31343738781) (the final pre-release run, gating PR #435's `deploy to staging` check): passed, contributing to that release actually promoting to `main`.

No run since `VOC-052-T00` merged has reproduced issue #437's original
failure (discover page/situation list empty). The canonical P1 content
seed is confirmed working against real staging on every deploy.

## VOC-052-AC-02 / VOC-052-AC-03

- **AC-02** (the previously-failing discover-step assertion passes):
  confirmed, per the run list above.
- **AC-03** (every later step in the spec that depends on real
  word/meaning content behaves correctly): confirmed for the save,
  review, and sentence-feedback steps specifically - every run that
  reached step 7 necessarily passed steps 3-6 first, which exercise
  saving a real word and reviewing it against real content.

## Scope note

Per this task's own description, "If the check does not pass on the
first real run with `VOC-052-T00`'s step in place, this task's scope
includes diagnosing and fixing that specific gap." The check *did*
pass on the first and every subsequent real run - no gap to diagnose
or fix. This task closes as pure verification, no source change, per
its own stated shape.
