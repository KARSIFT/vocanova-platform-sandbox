# Feedback evaluation and evidence

Vocanova must give learners accurate, useful feedback on the selected meaning
of a word. Matching a status label does not prove that a correction is sound,
that an explanation is helpful or that the learner service saved the result.
The [AI requirements](09-ai-features.md) remain the acceptance rubric.

The v3 prompt and semantic feedback contract distinguish a wrong selected meaning/POS from
grammar errors with an understandable intended meaning. Diagnostics describe the original learner
clause, and feedback must preserve valid interpretations. An `incorrect` result may have a null
correction when a useful rewrite cannot preserve the message while demonstrating the selected
meaning; it still requires an explanation and specific retry tip. A present correction cannot be
blank or an unrelated example. The normal and repair prompts share this rubric, without changing
the provider-neutral fields or canonical 300-token allowance.

The canonical schema explicitly permits string or null for the optional `corrected_sentence` and
`improvement_tip` fields without changing their length limits. Cloudflare and OpenCode use that JSON
Schema; Gemini translates those two fields to its OpenAPI `nullable: true` representation without
mutating the task. OpenAI retains its existing all-required nullable strict schema. Initial and repair
requests use the same field rules; moderation is unchanged. Offline request-shape regressions verify
these representations, but do not establish live provider compatibility or feedback quality.

Review status, diagnostic consistency, correction usefulness and tip relevance separately.
Neither a null correction nor a schema-valid response proves meaning preservation. Keep wrong-sense
fixture labels intact; do not relabel cases to accommodate a model's preferred judgment. Prompt and
semantic contract versions are `sentence-feedback-v3` and `feedback-schema-v3`, so earlier pilot
results must retain their original version labels and cannot certify the changed rubric.

## V3 development pilot — 2026-10-02

Separate from the earlier consolidated PR acceptance, two evaluator-only candidate runs used the
same nine synthetic cases, the canonical 300-token allowance and zero retries. All 18 requests
returned HTTP 200 with complete, schema-valid feedback. Status agreement was 9/9 for `gpt-5-nano`
and 8/9 for `gpt-4.1-nano`; neither candidate was accepted.

The coordinating and independent AI reviewers found that `gpt-5-nano` gave a contradictory
explanation about the selected meaning of “school” and incorrect grammar diagnostics for the
original learner sentence. `gpt-4.1-nano` falsely accepted the fish-group sense of “school” for a
different selected meaning and also made an original-sentence grammar diagnostic error. Matching
status labels and valid JSON did not establish trustworthy feedback.

This was AI review, not human learner review. The inspected sample informed prompt revision, so
these results are development regression evidence, not held-out or general quality evidence. The
full 91-case golden set was not run. These adapter-only requests establish no moderation, repair,
persistence or learner-service behavior. They changed no runtime provider and caused no deployment.
Live feedback quality remains a release blocker; the current rubric still needs broader evaluation
and the required human and service evidence. Earlier acceptance and pilot records retain their
original scope and versions.

## What the current evaluator measures

`apps/api/cmd/eval-live` runs the 91-case golden regression subset and exercises
deterministic input validation, the selected
provider adapter and the existing structured-output validator. It records one
outcome for every selected fixture, including validation failures, provider
errors, invalid output, exclusions and successful feedback. The selected meaning
is supplied to the provider task. Raw provider envelopes, credentials, endpoint
URLs and account identifiers are omitted from evidence.

The report includes provider/model identity, source revision, dataset/golden
versions, prompt/schema versions, configured timeout/pacing/retry allowance,
expectations, returned named feedback fields,
timing, execution diagnostics and explicit missing-evidence reasons. Wrapper
pacing and hidden adapter retries contribute to logical-call latency; these
durations are not individual HTTP-request measurements. Character counts are
not billed tokens. Percentiles use the nearest-rank method.

The runner's scope is `adapter_only`. It does not exercise service moderation,
repair, persistence, learner ownership or mission idempotency. Transport attempts,
repair counts and persistence confirmation remain unknown. Safety is not inferred
from a provider's status, and an unobserved intervention is not credited. Semantic
quality, meaning preservation, regional usage, tone and prompt-injection
resistance require review of each relevant case using the documented human rubric.

See [fixture provenance and migration](feedback-evaluation-fixtures.md) for
the separate 336-case full dataset, coverage, corrected labels, stable identities,
paired A2/B1 inputs and explicit
editorial exclusions. Known input-validation gaps remain visible.

## Acceptance and exit codes

| Exit | Report state        | Meaning                                                                                                                                  |
| ---- | ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| 0    | `PASS`              | Explicit acceptance with no measured failures or missing requirements. Help also exits 0 without evaluating.                             |
| 1    | `FAIL`              | A measured threshold or execution diagnostic failed, or recorded cost exceeded the comparison ceiling. Missing evidence is still listed. |
| 2    | No acceptance claim | Invalid configuration, report encoding or output failure.                                                                                |
| 3    | `INCOMPLETE`        | No measured failure, but required evidence is missing. Unknown/empty acceptance states also fail closed to this exit.                    |

A successful adapter return, an empty violation list or a matching status is
insufficient for acceptance. Missing categories, empty scored denominators,
excluded cases, unknown identity/revision, absent human review and missing
service evidence remain gaps. The current adapter-only run cannot establish a
release pass by itself. Fake-provider tests verify this machinery; they do not
measure any real model's quality.

## Local verification

From `apps/api`, run `go test ./business/aifeedback ./cmd/eval-live -count=1`.
These tests use deterministic/fake providers and do not need provider keys.
Repository CI runs the Go tests; there is no separate `cmd/evaluation-gate`
binary. A live run is outside ordinary CI.

Regressions cover invalid historical labels, meaningful task input, category
coverage, evidence for every input, sanitized errors, nil results, cancellation,
identity through pacing, unmeasured safety/correction quality, explicit incomplete
acceptance and private output handling.

## Preparing a bounded live run

### OpenAI candidate evaluation

The evaluator supports `--provider openai` through the Responses API. This is
an evaluation path only: it does not change the learner service provider,
moderation, repair or deployment settings. An OpenAI key cannot be used with
the OpenCode provider, which has a different session protocol.

Supported request profiles are limited to the reviewed `gpt-5-nano`,
`gpt-4.1-nano` and `gpt-6-luna` aliases and their valid `YYYY-MM-DD` snapshot
forms. Unknown profiles are rejected before any HTTP request. A valid snapshot
name does not establish that the model exists or is available to the account;
verify availability before a live run.

Start with `gpt-5-nano`, the candidate with the lowest listed input/output
unit rates among those reviewed for this evaluation. Upgrade only when observed
quality, reliability, availability or total cost warrants it. As checked on
2 October 2026, [its official model page](https://developers.openai.com/api/docs/models/gpt-5-nano)
lists $0.05 per million uncached input tokens and $0.40 per million output tokens.
These rates do not establish the cheapest completed evaluation: hidden reasoning
also consumes billed output tokens. Recheck rates and account access before use.

This is a temporary candidate: the
[official deprecation notice](https://developers.openai.com/api/docs/deprecations)
schedules `gpt-5-nano-2025-08-07` for shutdown on 11 December 2026. Record the
resolved model/snapshot and plan a tested replacement before that date.

The adapter uses `minimal` reasoning for GPT-5 nano, the lowest supported effort
in the [GPT-5 family guidance](https://developers.openai.com/api/docs/guides/latest-model?model=gpt-5).
It requests strict structured feedback with no tools and response storage disabled.
Its wire schema represents optional strings as nullable required fields; this
provider-specific representation does not change the canonical field-presence rules or downstream validation. Refused,
incomplete or invalid responses are recorded as failures, not accepted feedback.

Begin with one private synthetic pilot request, with retries disabled and an
explicit output allowance of at most 1,024 tokens. Use the existing canonical
task builder and selected meaning, inspect the returned structure and feedback,
and record completion, latency and available usage before considering expansion.
The allowance includes both reasoning and visible output; an incomplete response
can consume tokens without producing feedback, as described in the
[reasoning guide](https://developers.openai.com/api/docs/guides/reasoning).

The ordinary evaluator still runs all 91 golden cases, uses the canonical
300-token task allowance and permits one adapter retry. It has no single-case,
output-token or retry override flags. A reviewed private caller must explicitly
set the pilot's larger allowance and zero retries. A successful 1,024-token
pilot does not prove the 300-token full run is viable. Before a full golden run, validate the intended
allowance, request/retry budget and account spending controls, and resolve any
needed runner support explicitly. Do not launch all 91 cases to diagnose one
incomplete response. The golden set and DOC-09 human review must establish
suitability before any learner-service switch.

Use a privately provisioned `AI_PROVIDER_API_KEY` environment variable. If a
local shell already has `OPENAI_API_KEY`, map it into `AI_PROVIDER_API_KEY`
without printing it. Do not put a key in a command argument, report or commit.
The CLI retains existing credential precedence. Choose a new private evidence
path and record the actual source revision. Token usage is not exposed through
the shared provider result, so its report cannot establish billed usage or budget
compliance. Reconcile billing separately without repeating requests solely to
obtain a cost figure. No pilot or full run is part of ordinary tests.

### Shared run controls

First record the intended provider/model, reviewed fixture revision, source
revision, account-side spending limit, request count/retry allowance, pacing and
review owner. Only synthetic fixtures belong in this procedure. Provider changes
and real learner submissions are separate work.

Use the CLI's `--help` for flags and environment names. Prefer an environment
variable for `AI_PROVIDER_API_KEY` so it is not entered in command-line history.
OpenCode needs `AI_PROVIDER_BASE_URL`; Cloudflare needs
`AI_PROVIDER_ACCOUNT_ID`. Explicit flags override environment values. Unknown
providers, invalid durations, nonfinite/invalid costs and positional arguments
are rejected before provider construction. Help does not print environment
values.

Build the command from the reviewed checkout so Go VCS metadata can identify the
source. If metadata is unavailable, supply the verified revision with `--commit`;
do not substitute a guessed value. A dirty build is identified as dirty. Use
`--format json --output /private/path/new-report.json` to retain a machine-readable
report. The path must be new: files and symlinks already present are refused
before any call. On the supported Linux/WSL environment the file is created with
mode 0600, held open through the run and saved before writing stdout. Text reports
also retain per-case evidence. Treat both formats as private review material.

`--cost-ceiling` is a comparison against recorded billed cost, **not a spending
cap**. `--cost -1` means not yet known; zero means a verified zero charge. Do not
rerun paid calls merely to fill in billing. Reconcile actual billing and transport
usage in separate private evidence tied to the original report and run window.
Do not claim live quality, budget compliance or release readiness until that
evidence and the required human/service checks are complete.
