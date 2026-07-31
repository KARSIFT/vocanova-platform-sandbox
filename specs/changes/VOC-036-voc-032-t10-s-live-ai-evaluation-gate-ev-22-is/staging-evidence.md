# VOC-036 — Staging Evidence

**Adopted and implementation-authorized (2026-08-01). `T00`–`T02` are present
in this checkout and their package tests pass; `T03` remains blocked by
credential dependency `VOC-036-DEP-00` and has no live-provider run yet.**

## Status (2026-07-31)

- `VOC-036-T00`, `VOC-036-T01`, `VOC-036-T02`: implemented in this checkout
  (`cloudflare.go`, Cloudflare wiring in `production.go`, and Cloudflare/pacing
  support in `cmd/eval-live`) and validated locally via:
  `go test ./business/aifeedback/... ./app/api/... ./cmd/eval-live/...` from
  `apps/api` (all passed).
- `VOC-036-T03` (the T10-equivalent live Cloudflare evaluation, `EV-22`'s
  Cloudflare-provider analogue): **still blocked by `VOC-036-DEP-00`** in this
  environment. `AI_PROVIDER_API_KEY` and `AI_PROVIDER_ACCOUNT_ID` were both
  unset, and a direct command attempt
  (`go run ./cmd/eval-live --provider cloudflare` from `apps/api`) exited with
  usage error:
  `eval-live: --account-id (or AI_PROVIDER_ACCOUNT_ID) is required`.
  No live Cloudflare API call occurred.

## Note on this package's own requirement-source evidence (recorded honestly, per this drafting pass's own findings)

This package's requirement source cites the founder's own account that
`VOC-035`'s Gemini live evaluation also failed (a stale, new-key-blocked
default model plus an unpaced free-tier rate-limit burst on the 56-case golden
set). At this package's own `base_sha`
(`c69b270a164bf2cb386f1b7637a7c3ab96af5bd0`), `VOC-035`'s own
`staging-evidence.md` does not yet contain that recorded failure — it still
shows `VOC-035-T03` as "Not yet attempted" / "blocked by `VOC-035-DEP-00`".
This is recorded here, honestly, as a fact this drafting pass found and could
not independently resolve from the repository alone (see `specification.md`
"Objective and requirement source" for the full detail and the reasoning for
why this package proceeds on the founder's own account regardless). This is
not a claim about `VOC-036`'s own Cloudflare live evaluation — it is a
transparency note about this package's own requirement-source evidence chain,
recorded here rather than silently omitted.

## `EV-22`-equivalent — Live Cloudflare Workers AI evaluation pass

**Attempt 1 status: blocked before provider call.**

Per `VOC-036-AC-11`/`VOC-036-TEST-11`, this evidence row remains
`pending — blocked by VOC-036-DEP-00` until a Cloudflare API token and account
ID are provisioned. Attempt 1 (this run) verified the command path and blocking
condition only:

- Command: `go run ./cmd/eval-live --provider cloudflare` (from `apps/api`)
- Exit: usage error (`exit status 2`)
- Blocking message: `--account-id (or AI_PROVIDER_ACCOUNT_ID) is required`
- Environment check: both required Cloudflare variables were unset in this run
  (`AI_PROVIDER_API_KEY_SET=no`, `AI_PROVIDER_ACCOUNT_ID_SET=no`)
- Live API calls made: **0** (blocked at argument/credential validation)

Once `VOC-036-DEP-00` is resolved, the next attempt must set a non-zero
`--request-interval` and record the full rendered `LiveEvaluationReport`
verbatim, including thresholds, latencies, provider-call count, cost/ceiling
fields, timestamps, pacing interval used, and operator notes.

**This result will be recorded honestly whichever way it comes out.** A
failing or still-blocked result is valid evidence and will be recorded as such,
not omitted or silently retried until a pass appears — matching the founder's
own explicit instruction for this package and the discipline
`VOC-032`/`VOC-034`/`VOC-035` already established in this repository.

## Rollback triggers

Per this package's `implementation-plan.md` §Deployment and rollback /
`release-plan.md` §Rollback, initiate rollback on:

- A merged `T00`–`T02` PR found to introduce a defect after merge.
- A live evaluation result showing Cloudflare cannot meet DOC-09 §18's latency
  budget or DOC-09 §23's thresholds, even with request pacing applied (a
  finding to record, not a reason to silently loosen the budget or the test —
  and not itself a reason for a code rollback, since Cloudflare is opt-in and
  the default provider stays OpenCode).
- Any credential or secret value (API token or account ID) found in a
  committed file.
- Any of DOC-09 §25's rollback conditions, if an operator has set
  `AI_PROVIDER=cloudflare` in a real deployment.

## Rollback procedure

For a code-level defect: `git revert` of the specific merge commit, which
removes the Cloudflare branch/flag/pacing addition entirely. For an
operational issue in a deployment that has opted into `AI_PROVIDER=cloudflare`:
reset the environment variable to `opencode` (or unset it) and restart — no
code deploy required. Independent of either: the existing
`AI_FEATURES_ENABLED` kill switch disables all AI generation immediately
regardless of provider. Never automate a database rollback — this package
touches no schema or persisted state shape, so none is ever needed for this
package's own scope.
