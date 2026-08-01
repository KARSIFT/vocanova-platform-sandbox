# VOC-036 — Staging Evidence

**Adopted and implemented (2026-08-01). `T00`–`T02` merged (PRs #248, #249,
#250). `T03`'s live evaluation has run four times against the real API — a
first attempt exposed a real code defect (fixed in PR #253), then three
different models were tried against the fixed code. Result: FAIL, recorded
honestly, not retried indefinitely.**

## Status (2026-08-01)

- `VOC-036-T00`, `VOC-036-T01`, `VOC-036-T02`: merged. `CloudflareFeedbackProvider`,
  `CloudflareModerationProvider`, the `buildAIProviders` `cloudflare` branch,
  the new `AI_PROVIDER_ACCOUNT_ID` variable, and `cmd/eval-live`'s
  `--provider cloudflare` / `--request-interval` flags all exist and pass
  their unit tests against a fake transport.
- A follow-up fix, PR #253, corrected a real defect found during `T03`'s
  first live run (see below) — merged before any of the three model-comparison
  runs recorded here.
- `VOC-036-T03`: run four times total against the real Cloudflare Workers AI
  API. `VOC-036-DEP-00` (Cloudflare API token + account ID) was already
  resolved. See the `EV-22`-equivalent section below for the full record.

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

## `EV-22`-equivalent — Live Cloudflare Workers AI evaluation result

**Result: FAIL.** Run four times, 2026-08-01, once `VOC-036-DEP-00` was
confirmed resolved and `T00`–`T02` had merged.

### Run 1 — `@cf/meta/llama-3.3-70b-instruct-fp8-fast` (T00's shipped default): hard failure, real defect found

Every one of 56 provider calls failed (`ProviderCalled=0/56`) despite the
HTTP layer returning 200 responses. Diagnosed via a direct, isolated API
probe outside the harness (same `response_format: {type:"json_schema"}`
request shape `cloudflare.go` itself sends): Cloudflare's real API returns
`result.response` as an actual parsed JSON object matching the schema (e.g.
`{"status":"correct",...}`), not a JSON-encoded string containing that
object. `cloudflareRunResponse.Result.Response` in `cloudflare.go` was typed
as a plain Go `string`, so every real response failed to unmarshal. This is a
real defect in `T00`'s shipped code, not a Cloudflare-side or model-side
issue — confirmed and fixed in PR #253 (`Result.Response` changed to
`json.RawMessage`; existing fake-server unit tests were also updated, since
they mocked the same wrong string-shape assumption, which is why they passed
while the real API broke). Full verbatim report:

```
=== T10 Live AI Evaluation Report ===
Provider: unknown
Model:
Dataset: initial-dataset-v1
Spec: doc09-v1
StartedAt: 2026-08-01T08:01:23Z
FinishedAt: 2026-08-01T08:07:37Z
Duration: 6m14.616246924s
ProviderCalls: 56
EstimatedInputChars: 75370
EstimatedOutputChars: 0
CostUSD: -1.00
CostCeilingUSD: 0.25
CostCeilingExceeded: false
LatencyMin: 1.730276162s
LatencyMax: 17.014930479s
LatencyMean: 6.689555382s
LatencyP50: 7.630289903s
LatencyP95: 15.627237935s
OperatorNotes: VOC-036-T03 live evaluation, dispatched via diagnostic workflow, request-interval=1s to avoid VOC-035's rate-limit lesson
--- Per-threshold computed values ---
DatasetVersion=initial-dataset-v1 Total=56 Validated=56 ProviderCalled=0 Intercepted=0 Matched=0 ExpectedTotal=56
Per-class: correctness=0/28, incorrect_target_use=0/0, self_harm_intercepted=0/0, safety_violations=0/0
Correct-expected breakdown: got_correct=0, got_needs_improvement=0, got_incorrect=0, intercepted=0 (of 28)
Repair: not tracked (run did not record repair attempts)
Meaning: not tracked (run did not record corrected-sentence text comparison)
Result: FAIL (3 tracked threshold(s) violated)
  - structured_output_valid_first_response: observed=0.000 (0/56) spec=>= 0.990 (min)
    structured-output valid first response 0.000 below spec >= 0.990 (provider returned valid feedback for 0 of 56 validated cases)
  - overall_status_accuracy: observed=0.000 (0/56) spec=>= 0.900 (min)
    overall status accuracy 0.000 below spec >= 0.900 (matched 0 of 56 cases with expectations)
  - clearly_correct_accuracy: observed=0.000 (0/28) spec=>= 0.950 (min)
    clearly-correct accuracy 0.000 below spec >= 0.950 (0 of 28 correctness cases matched)
=== Result: FAIL (3 tracked threshold(s) violated) ===
```

### Run 2 — `@cf/meta/llama-3.3-70b-instruct-fp8-fast` (same model, after the PR #253 fix): real signal, still FAIL — latency-bound

With the type-mismatch fixed, `ProviderCalled` rose from 0 to 16/56 — proof
the fix worked. But latency was high and inconsistent (1.7s–15.6s, p95
8.6s — right at/above the 8-second `AI_PROVIDER_TIMEOUT` default), consistent
with the 70B model genuinely being too slow for Cloudflare's shared infra to
answer within budget under the real, much larger production prompts (a
direct raw-API probe with a trivial prompt returned in under 1.2s — the
model itself is not inherently slow; the real system/developer prompts and
full JSON schema are). Full verbatim report:

```
=== T10 Live AI Evaluation Report ===
Provider: unknown
Model:
Dataset: initial-dataset-v1
Spec: doc09-v1
StartedAt: 2026-08-01T08:22:27Z
FinishedAt: 2026-08-01T08:28:22Z
Duration: 5m55.125854479s
ProviderCalls: 56
EstimatedInputChars: 75370
EstimatedOutputChars: 483
CostUSD: -1.00
CostCeilingUSD: 0.25
CostCeilingExceeded: false
LatencyMin: 1.689608877s
LatencyMax: 15.573184647s
LatencyMean: 6.34151328s
LatencyP50: 7.512443376s
LatencyP95: 8.644424033s
OperatorNotes: VOC-036-T03 live evaluation, re-run after fixing the Result.Response type mismatch (PR #253), request-interval=1s
--- Per-threshold computed values ---
DatasetVersion=initial-dataset-v1 Total=56 Validated=56 ProviderCalled=16 Intercepted=0 Matched=12 ExpectedTotal=56
Per-class: correctness=11/28, incorrect_target_use=0/0, self_harm_intercepted=0/0, safety_violations=0/0
Correct-expected breakdown: got_correct=11, got_needs_improvement=1, got_incorrect=0, intercepted=0 (of 28)
Repair: not tracked (run did not record repair attempts)
Meaning: not tracked (run did not record corrected-sentence text comparison)
Result: FAIL (3 tracked threshold(s) violated)
  - structured_output_valid_first_response: observed=0.286 (16/56) spec=>= 0.990 (min)
    structured-output valid first response 0.286 below spec >= 0.990 (provider returned valid feedback for 16 of 56 validated cases)
  - overall_status_accuracy: observed=0.214 (12/56) spec=>= 0.900 (min)
    overall status accuracy 0.214 below spec >= 0.900 (matched 12 of 56 cases with expectations)
  - clearly_correct_accuracy: observed=0.393 (11/28) spec=>= 0.950 (min)
    clearly-correct accuracy 0.393 below spec >= 0.950 (11 of 28 correctness cases matched)
=== Result: FAIL (3 tracked threshold(s) violated) ===
```

### Run 3 — `@cf/meta/llama-3.1-8b-instruct-fp8-fast` (smaller model, operator override): 100% reliable, still FAIL — accuracy-bound

Every one of 56 calls succeeded (`ProviderCalled=56/56`), fast and consistent
(1.2s–3.2s, well inside budget) — the infrastructure/latency problem is fully
resolved with a smaller model. But the model's actual judgment quality falls
well short: only 41% overall status accuracy (need ≥90%), and critically it
over-corrects sentences that are already grammatically correct — 14 of 28
clearly-correct cases were unnecessarily "corrected" (need ≤5%), including 5
zero-tolerance `wrong_correction_on_correct` defects (spec requires exactly
0). Full verbatim report:

```
=== T10 Live AI Evaluation Report ===
Provider: unknown
Model:
Dataset: initial-dataset-v1
Spec: doc09-v1
StartedAt: 2026-08-01T08:44:30Z
FinishedAt: 2026-08-01T08:46:01Z
Duration: 1m31.26061558s
ProviderCalls: 56
EstimatedInputChars: 75370
EstimatedOutputChars: 7264
CostUSD: -1.00
CostCeilingUSD: 0.25
CostCeilingExceeded: false
LatencyMin: 1.247694789s
LatencyMax: 3.226761304s
LatencyMean: 1.629634901s
LatencyP50: 1.447206555s
LatencyP95: 2.716807992s
OperatorNotes: VOC-036-T03 live evaluation, model overridden to the smaller 8B variant after the 70B default hit the 8s timeout on real-size prompts (fix already in PR #253; this is a model-selection follow-up, not a bug)
--- Per-threshold computed values ---
DatasetVersion=initial-dataset-v1 Total=56 Validated=56 ProviderCalled=56 Intercepted=0 Matched=23 ExpectedTotal=56
Per-class: correctness=14/28, incorrect_target_use=0/0, self_harm_intercepted=0/0, safety_violations=0/0
Correct-expected breakdown: got_correct=14, got_needs_improvement=9, got_incorrect=5, intercepted=0 (of 28)
Repair: not tracked (run did not record repair attempts)
Meaning: not tracked (run did not record corrected-sentence text comparison)
Result: FAIL (4 tracked threshold(s) violated)
  - overall_status_accuracy: observed=0.411 (23/56) spec=>= 0.900 (min)
    overall status accuracy 0.411 below spec >= 0.900 (matched 23 of 56 cases with expectations)
  - clearly_correct_accuracy: observed=0.500 (14/28) spec=>= 0.950 (min)
    clearly-correct accuracy 0.500 below spec >= 0.950 (14 of 28 correctness cases matched)
  - unnecessary_correction_on_clearly_correct: observed=0.500 (14/28) spec=<= 0.050 (max)
    unnecessary correction on clearly-correct cases 0.500 above spec <= 0.050 (14 of 28 correct-expected cases were unnecessarily corrected)
  - wrong_correction_on_correct: observed=5 spec== 0 (max)
    wrong correction on correct cases: 5 observed, spec = 0 (a single such case is a zero-tolerance defect)
=== Result: FAIL (4 tracked threshold(s) violated) ===
```

### Run 4 — `@cf/meta/llama-4-scout-17b-16e-instruct` (MoE model, operator override): 100% reliable, better raw accuracy, still FAIL — worse over-correction

Also 56/56 reliable and fast (1.6s–3.8s). Overall accuracy improved over the
8B model (61% vs 41%), but the over-correction problem got worse, not
better: 18 of 28 clearly-correct cases were unnecessarily corrected (64%,
vs the 8B model's 50%), with 6 zero-tolerance defects (vs 5). Full verbatim
report:

```
=== T10 Live AI Evaluation Report ===
Provider: unknown
Model:
Dataset: initial-dataset-v1
Spec: doc09-v1
StartedAt: 2026-08-01T09:35:07Z
FinishedAt: 2026-08-01T09:37:09Z
Duration: 2m2.831439328s
ProviderCalls: 56
EstimatedInputChars: 75370
EstimatedOutputChars: 4232
CostUSD: -1.00
CostCeilingUSD: 0.25
CostCeilingExceeded: false
LatencyMin: 1.581669949s
LatencyMax: 3.802622926s
LatencyMean: 2.193398163s
LatencyP50: 2.061505554s
LatencyP95: 2.995399924s
OperatorNotes: VOC-036-T03 live evaluation, third model attempt: llama-4-scout-17b-16e-instruct (MoE), after 70b (too slow) and 8b (not accurate enough) both failed for different reasons
--- Per-threshold computed values ---
DatasetVersion=initial-dataset-v1 Total=56 Validated=56 ProviderCalled=56 Intercepted=0 Matched=34 ExpectedTotal=56
Per-class: correctness=10/28, incorrect_target_use=0/0, self_harm_intercepted=0/0, safety_violations=0/0
Correct-expected breakdown: got_correct=10, got_needs_improvement=12, got_incorrect=6, intercepted=0 (of 28)
Repair: not tracked (run did not record repair attempts)
Meaning: not tracked (run did not record corrected-sentence text comparison)
Result: FAIL (4 tracked threshold(s) violated)
  - overall_status_accuracy: observed=0.607 (34/56) spec=>= 0.900 (min)
    overall status accuracy 0.607 below spec >= 0.900 (matched 34 of 56 cases with expectations)
  - clearly_correct_accuracy: observed=0.357 (10/28) spec=>= 0.950 (min)
    clearly-correct accuracy 0.357 below spec >= 0.950 (10 of 28 correctness cases matched)
  - unnecessary_correction_on_clearly_correct: observed=0.643 (18/28) spec=<= 0.050 (max)
    unnecessary correction on clearly-correct cases 0.643 above spec <= 0.050 (18 of 28 correct-expected cases were unnecessarily corrected)
  - wrong_correction_on_correct: observed=6 spec== 0 (max)
    wrong correction on correct cases: 6 observed, spec = 0 (a single such case is a zero-tolerance defect)
=== Result: FAIL (4 tracked threshold(s) violated) ===
```

### Model comparison summary

| Model | Reliability | Overall accuracy (need ≥90%) | Unnecessary corrections (need ≤5%) | Zero-tolerance defects (need 0) |
| --- | --- | --- | --- | --- |
| `llama-3.3-70b-instruct-fp8-fast` (default) | 29% (16/56) — timeout-bound | 21% | not computable (too few valid) | not computable |
| `llama-3.1-8b-instruct-fp8-fast` | 100% | 41% | 50% (14/28) | 5 |
| `llama-4-scout-17b-16e-instruct` | 100% | 61% | 64% (18/28) — worse | 6 — worse |

DeepSeek was investigated and ruled out without a full run: the only
JSON-schema-capable DeepSeek model in Cloudflare's catalog,
`@cf/deepseek-ai/deepseek-r1-distill-qwen-32b`, measured 10.9 seconds for a
single small test prompt via a direct API probe — already exceeding the
8-second budget before reaching the full-size real prompts. The other
DeepSeek variant names considered (`-1.5b`, `-7b`, `deepseek-v2.5`) do not
exist in this account's catalog (`HTTP 400`, "No route for that URI").

### Disposition

Per the founder's explicit decision (2026-08-01): **this FAIL is recorded as
the result, not retried further.** Three genuinely different free-tier
models were tried against the fixed code; each failed a different way, and
the pattern (larger/slower models are somewhat more accurate but *worse* at
over-correction; smaller/faster models are less accurate but *better* at
over-correction) looks like a real capability ceiling for this account's
available free-tier models on this specific zero-tolerance task, not
something a further model swap is likely to resolve on its own.
`VOC-036-AC-11`'s `Result` is `fail`. `VOC-032-T10` (issue #186) **remains
open** — Cloudflare did not resolve it either. The real code defect found
along the way (PR #253) is a permanent improvement independent of this
result — any future attempt at this provider starts from working
infrastructure, not a silent 0%-success bug.

Two real, separately-actionable follow-ups are on the table for whoever
picks this up next, neither decided here:

1. A paid-tier provider (OpenAI/Anthropic direct, or Cloudflare's paid tier
   with access to stronger models) may clear the accuracy bar where every
   free-tier option tried has not.
2. The `unnecessary_correction_on_clearly_correct` failure pattern, present
   across every model/size tried at this task, may indicate the moderation/
   feedback prompt or schema itself has room to be more explicit about not
   correcting already-correct sentences — a prompt-engineering investigation
   independent of which provider is used, potentially worth trying against
   the existing OpenCode/Gemini paths too, not just Cloudflare.

**This result was recorded honestly, whichever way it came out** — a failing
result is valid evidence, matching the discipline `VOC-032`/`VOC-034`/
`VOC-035` already established in this repository, not omitted or silently
retried until a pass appeared.

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
