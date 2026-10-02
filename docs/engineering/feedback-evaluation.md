# Feedback evaluation and evidence

Vocanova must give learners accurate, useful feedback on the selected meaning
of a word. Matching a status label does not prove that a correction is sound,
that an explanation is helpful or that the learner service saved the result.
The [AI requirements](09-ai-features.md) remain the acceptance rubric.

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
