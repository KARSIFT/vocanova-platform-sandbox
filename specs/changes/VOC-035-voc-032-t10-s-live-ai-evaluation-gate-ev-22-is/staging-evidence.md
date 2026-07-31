# VOC-035 — Staging Evidence

**Adopted and implemented (2026-07-31). `T00`–`T02` merged (PRs #239, #240,
#241). `T03`'s live evaluation has run — see below. Result: FAIL, recorded
honestly, not retried until a pass appeared.**

## Status (2026-08-01)

- `VOC-035-T00`, `VOC-035-T01`, `VOC-035-T02`: merged. `GeminiFeedbackProvider`,
  `GeminiModerationProvider`, the `buildAIProviders` `gemini` branch, and
  `cmd/eval-live`'s `--provider gemini` flag all exist and pass their unit
  tests against a fake transport.
- `VOC-035-T03` (the T10-equivalent live Gemini evaluation, `EV-22`'s
  Gemini-provider analogue): **run twice, both FAIL.** `VOC-035-DEP-00`
  (Gemini API key) was already resolved. See the `EV-22`-equivalent section
  below for the full record, including a real defect found in `T00`'s shipped
  code (a hardcoded default model that Google had already blocked for new API
  keys by the time this ran) and a second, more informative failure once a
  working model was substituted.

## `EV-22`-equivalent — Live Gemini AI evaluation pass

**Result: FAIL.** Run twice, 2026-08-01, once `VOC-035-DEP-00` was confirmed
resolved and `T00`–`T02` had merged.

### Run 1 — `gemini-2.5-flash` (T00's shipped default): hard failure, real defect found

`gemini.go`'s `defaultGeminiModel` constant (used whenever `--model`/
`AI_PROVIDER_MODEL` is unset) is `"gemini-2.5-flash"`. Every one of 56
provider calls against this model returned HTTP 404 from Google's real API:

```
{
  "error": {
    "code": 404,
    "message": "This model models/gemini-2.5-flash is no longer available to
      new users. Please update your code to use a newer model for the latest
      features and improvements.",
    "status": "NOT_FOUND"
  }
}
```

Confirmed via a direct, isolated API call outside the evaluation harness
(same URL/header shape `gemini.go` itself uses) — this is a real, external
fact (Google restricting a model to existing users only, discovered after
this package's own planning/implementation), not a parsing or auth bug in
`T00`'s code. `parseGeminiTextResponse` correctly mapped every 404 to a
non-nil error per its own fail-closed design (`VOC-035-D03`) — the harness
behaved correctly given what the real API returned. `ProviderCalled=0/56`;
`structured_output_valid_first_response`, `overall_status_accuracy`, and
`clearly_correct_accuracy` all `0.000`. `Provider`/`Model` report fields
rendered as `unknown`/`` — a separate, cosmetic pre-existing gap in
`live_eval.go`'s `providerName`/`providerModel` helpers (hardcoded type
switches with no `*GeminiFeedbackProvider` case), not itself evidence of
anything about Gemini's quality.

- Provider/model as configured: `gemini` / `gemini-2.5-flash` (unresolved default)
- Dataset/spec: `initial-dataset-v1` / `doc09-v1`
- Duration: 6.5s: Total=56, Validated=56, ProviderCalled=0, Matched=0
- Cost: `$0.00` (free tier; `CostCeilingUSD=0.25`, not exceeded)
- Result: **FAIL** — every threshold `0.000`

### Run 2 — `gemini-flash-latest` (working model, substituted for diagnosis only): still FAIL, but with real signal

A raw API probe (outside the harness) confirmed `gemini-flash-latest` and
`gemini-3.5-flash` both return HTTP 200 for this API key; `gemini-2.0-flash`
returned HTTP 429 (`generate_content_free_tier_requests` quota limit `0` for
that specific model on this key's free tier); `gemini-2.5-flash-lite`
returned the same 404-for-new-users error as the original default.

Re-running the full 56-case evaluation with `--model gemini-flash-latest`:

- Duration: 1m16s; ProviderCalled=11/56 (up from 0/56); Matched=10/56;
  correctness 5/28 correct, 1/28 needs-improvement
- Latency: min 165ms, mean 1.36s, p50 249ms, p95 6.85s, max 12.0s — the wide
  spread (sub-200ms alongside multi-second calls in the same run) is
  consistent with free-tier per-minute rate-limiting intermittently
  throttling a 56-request burst fired with no pacing between calls, not a
  content-quality problem — `VOC-035-D04` already flagged that the retry/
  timeout budget copied from OpenCode's paid-tier assumptions was unverified
  against Gemini's free-tier characteristics, and this run is exactly the
  finding that flag anticipated.
- Cost: `$0.00` (free tier; ceiling not exceeded)
- Result: **FAIL** — `structured_output_valid_first_response` 0.196 (11/56,
  need ≥0.990), `overall_status_accuracy` 0.179 (10/56, need ≥0.900),
  `clearly_correct_accuracy` 0.179 (5/28, need ≥0.950)

### Disposition

Per the founder's explicit decision (2026-08-01): **this FAIL is recorded as
the result, not retried further.** No rate-limit-pacing fix was pursued this
round. `VOC-035-AC-10`'s `Result` is `fail`. `VOC-032-T10` (issue #186)
**remains open** — Gemini did not resolve it. Two real, separately-actionable
follow-ups are on the table for whoever picks this up next, neither decided
here:

1. Fix `gemini.go`'s stale `defaultGeminiModel` default (a real, narrow code
   defect independent of the rate-limit question).
2. Add request pacing/backoff for Gemini's free tier if a future attempt at
   this provider is worth pursuing.

**This result was recorded honestly, whichever way it came out** — a failing
result is valid evidence, matching the discipline `VOC-032`/`VOC-034` already
established in this repository, not omitted or silently retried until a pass
appeared.

## Rollback triggers

Per this package's (currently unauthorized, draft) `implementation-plan.md`
§Deployment and rollback / `release-plan.md` §Rollback, initiate rollback on:

- A merged `T00`–`T02` PR found to introduce a defect after merge.
- A live evaluation result showing Gemini cannot meet DOC-09 §18's latency
  budget or DOC-09 §23's thresholds (a finding to record, not a reason to
  silently loosen the budget or the test — and not itself a reason for a code
  rollback, since Gemini is opt-in and the default provider stays OpenCode).
- Any credential or secret value found in a committed file.
- Any of DOC-09 §25's rollback conditions, if an operator has set
  `AI_PROVIDER=gemini` in a real deployment.

## Rollback procedure

For a code-level defect: `git revert` of the specific merge commit, which
removes the Gemini branch/flag entirely. For an operational issue in a
deployment that has opted into `AI_PROVIDER=gemini`: reset the environment
variable to `opencode` (or unset it) and restart — no code deploy required.
Independent of either: the existing `AI_FEATURES_ENABLED` kill switch
disables all AI generation immediately regardless of provider. Never
automate a database rollback — this package touches no schema or persisted
state shape, so none is ever needed for this package's own scope.
