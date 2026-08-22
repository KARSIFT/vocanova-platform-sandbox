# VOC-108 — Test Plan

## VOC-108-TEST-00 — Later pass supersedes historical failure

Given two correctly bound attempts for one logical check and exact SHA, select
the later successful terminal attempt and ignore the obsolete failure.

## VOC-108-TEST-01 — Later failure/pending supersedes historical pass

Given an older success and newer failed, cancelled, timed-out, or pending attempt,
fail closed and never reuse the older pass.

## VOC-108-TEST-02 — Identity, ambiguity, and pagination

Reject wrong repository, workflow, PR, base, head, missing identity, duplicate
ambiguous newest records, and truncated input; select correctly with more than
100 check runs/statuses and deterministic ID tie-breaking.

## VOC-108-TEST-03 — Cross-repository closing policy

Generate/update same-repository and cross-repository PR bodies. The local caller
task binding remains valid; cross-repository output has only a non-closing fully
qualified reference and policy rejects all GitHub closing-keyword variants.

## VOC-108-TEST-04 — Premature closure cannot advance

Close a roster issue without a valid caller-merge marker, including a simulated
foreign PR close. Auto-advance dispatches nothing and release creates no audit.

## VOC-108-TEST-05 — Valid caller merge advances exactly once

Provide one App-authored marker matching caller repo, issue, package, task, merged
PR, reviewed head, and live merge state. Auto-advance/release accepts it; repeated
events/reconcile remain idempotent.

## VOC-108-TEST-06 — Forged or stale markers fail closed

Reject human/free-form authors, duplicates, mismatched fields, stale head, wrong
PR/repository, closed-unmerged PR, and marker created before exact merge proof.

## VOC-108-TEST-07 — Terminal external check wakes cheap evaluation

Simulate staging/external check pending then successful. The terminal event wakes
release evaluation, reuses authoritative exact-SHA CI/review, and does not run the
full CI or reviewer-model jobs again. A terminal failure remains blocked.

## VOC-108-TEST-08 — Concurrent promotion triggers

Race automatic, reconcile, and duplicate terminal-check triggers. Serialization
and live state rechecks yield at most one effective exact-head merge; late runs
exit success/no-op and emit no post-merge pending comment.

## VOC-108-TEST-09 — Regression, policy, and privacy suite

Run shared self-CI and relevant caller foundation/governance checks. Assert App
auth, exact SHA/base, SHA-bound merge, workflow exclusions, fail-closed risk,
task ordering, two-attempt cap, and evidence privacy remain intact.
