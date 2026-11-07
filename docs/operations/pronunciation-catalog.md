# Durable word pronunciation coverage

`pronunciation-catalog` is an explicit operator tool, not an HTTP/admin endpoint.
It reads public English catalog content and shares the application's exact
meaning/context/provider cache identity. It never creates a learner or changes
learning state. The API image includes the binary; its ordinary entrypoint is
unchanged.

## Audit first

Inside the existing API container, with its existing environment:

```sh
pronunciation-catalog
pronunciation-catalog --validate-audio
```

This defaults to SELECT-only audit and requires `DATABASE_URL`, but no speech
credential. `ready` means a matching persistent cache row, not independently
verified audible pronunciation. The optional `--validate-audio` audit streams one MP3 at a time through the
existing decode/signal gate and separately reports valid/invalid current clips.
It does not repair rows or prove correct pronunciation. Changed spelling, POS, definition, reviewed
context, model, voice or instruction version produce a different identity.

On 7 November 2026 a read-only production audit found 10,051 active English
meanings, all with nonempty definitions and approved current-revision images.
Only 14 had matching current word clips; 10,037 did not. There were 16 ready audio
rows overall; the other two were not current word identities and could be
examples or older identities. No pending/failed audio rows were present. Dedicated
speech was enabled with provider `openai`. This is a point-in-time metadata audit,
not a fresh check of all image URLs or a listening assessment. A separate read-only
validation of all 16 cached clips passed the complete decode/signal gate: durations
1.008–3.552 seconds, peak 0.2580–0.4530 and RMS 0.02299–0.07064. No cache repair
was needed; this still does not establish correct words or stress.

The validated audit reuses its canonical identity snapshot instead of querying
and hashing all meanings twice. A 20-iteration warm local comparison over 10,051
meanings and no cached clips retained identical reports, reduced total queries
from four to three per audit, catalog rows from 20,102 to 10,051 and Go cumulative
allocated bytes from 90,586,661 to 45,513,037 per audit. Metadata-only auditing
retained two queries and effectively unchanged allocations. This excludes audio
decoding and establishes neither production latency nor peak memory.

## Explicit bounded generation

After reviewing the audit, costs and a small pilot, an operator can explicitly
start a job with a fresh UUID and immutable limits. For example, a **five-attempt
pilot**, not a command run automatically by deployment:

```sh
pronunciation-catalog --generate --job OPERATOR-GENERATED-UUID \
  --max-generations 5 --max-characters 250 --max-audio-seconds 60
```

`PRONUNCIATION_PROVIDER=openai` and the existing dedicated
`PRONUNCIATION_API_KEY` must be available inside the container. Do not copy or
print credentials locally. If an operator environment already supplies a speech
key under another name, map it to the dedicated variable in that environment;
do not create a new key merely for this tool. Disabled AI configuration fails
closed. The command accepts no arbitrary provider URL/model or text input.

Save JSON stdout privately. Inspect/resume using the **same UUID and identical
budgets**:

```sh
pronunciation-catalog --job SAME-UUID
pronunciation-catalog --generate --job SAME-UUID \
  --max-generations 5 --max-characters 250 --max-audio-seconds 60
```

The read-only job report includes attempted keys, exact meaning revisions,
statuses and observed decoded durations. The database ledger is authoritative;
a lost local JSON file cannot reset spending. A fresh UUID explicitly authorizes
a separate budget; deployment never creates jobs automatically.

- One operator process holds the database-wide operator lock; requests run
  sequentially with the existing 12-second provider timeout.
- The same per-key transaction lock and pending lease used by public playback
  prevent duplicate generation. Ready rows are never replaced. Pending/failed
  rows from the same UTC day are skipped, preserving the public retry policy.
- Each attempt is durably reserved before contacting the provider. Failures,
  cancellations and interrupted attempts consume the attempt/input budget. A job
  never automatically retries the same cache key. A fresh job does not bypass a
  same-day failed/pending cache entry.
- Job budgets cannot be changed on resume. Attempt and character limits are
  checked transactionally. Input characters count spoken text, as in the public
  service; bounded definition/context instructions also incur provider tokens.
- `max-audio-seconds` stops subsequent calls once **observed decoded duration**
  reaches the limit. One response can cross it, and unavailable/undecodable
  responses may have unknown billed duration. This is not a hard dollar limit.
- Existing complete MP3 decoding and silence checks run before persistence.
  Spelling, POS and definition are rechecked under canonical row locks before
  completion. Example context is re-read and included in the cache key; a concurrent
  example edit can leave an obsolete ready key, which current reads/audits do not
  accept. Re-audit after content changes. Failed work stops the
  current invocation; no automatic provider retry/refund occurs.
- Public generation's 20-call/1,000-character daily budget is untouched. Operator
  job receipts account for this separately authorized work.

## Explicit recovery of one rejected clip

Ordinary generation skips a failed cache entry for the rest of its UTC day. The
original job never retries a key it already attempted, including after midnight.
Without targeted recovery, a later UTC day and a new explicitly budgeted job are
required; ordinary jobs still scan catalog order rather than target one meaning.
Continue other unattempted words by resuming their original job and budgets.

After inspecting a specific completed operator failure with positive observed
decoded duration, an operator may explicitly authorize **one** targeted recovery
during that same UTC day:

```sh
pronunciation-catalog --generate --job NEW-OPERATOR-UUID \
  --recover-failed-attempt ORIGINAL-FAILED-ATTEMPT-UUID \
  --max-generations 1 --max-characters 250 --max-audio-seconds 60
```

Both IDs must be canonical, non-nil UUIDs. Use a fresh job UUID, distinct from the
original attempt and its job. This requires the deployed recovery migration and
operator binary; it is never run by deployment. The same dedicated speech key,
provider and AI-enabled guards apply. No credentials belong in arguments.

Recovery binds its immutable budget and original attempt durably. The original
must be a completed failed attempt from an ordinary job, with positive observed
decoded duration and matching provider/canonical identity. Its current cache must
still be failed, contain no audio, identify that exact generation, have no
active lease, and retain an attempted day matching both the original attempt and
the current UTC date. Prior-day failures use the ordinary new-job policy instead. A failed HTTP request or an unobserved/undecodable response is not
made eligible merely by requesting recovery. Positive decoded duration does not
prove a particular defect, correct speech or a successful clip; review the
sanitized evidence before authorizing another paid request.

The original receipt and consumed budget remain unchanged. A new reserved attempt
and the recovery linkage are persisted before the provider call. One original
attempt can have only one recovery, even across different new job UUIDs. Recovery
attempts cannot themselves be recovered. Repeating the same recovery invocation
is rejected; changing its target or budgets is also rejected. The
ordinary generation mode rejects recovery jobs. Job reports include the optional
`recoveryOfAttempt` identity so the relationship remains inspectable after a lost
terminal response. Inspect it without generation using
`pronunciation-catalog --job RECOVERY-JOB-UUID`.

The operator lock, shared cache-key lock, generation ownership, canonical checks,
12-second provider timeout and existing audio validation still apply. Failed or
interrupted recovery consumes its allowance; there is no automatic retry/refund.
Ready clips are not replaced. Public daily generation limits remain unchanged.
Never delete a failed row, edit its attempted day or reset a receipt to force a
retry. Inspect the job report and current audit after recovery; an exhausted or
still-failed recovery is an explicit remaining coverage gap.

## Costs, pilot and rollout

The official [model pricing](https://developers.openai.com/api/docs/models/gpt-4o-mini-tts)
lists $0.60 per million input tokens and $12 per million audio output tokens.
Use `0.60 × input_tokens / 1,000,000 + 12 × audio_tokens / 1,000,000` for estimates;
MP3 bytes, spoken character counts and seconds are not billing tokens. For scale,
10,037 clips averaging 300 input tokens would cost about $1.81 for **input only**;
output is additional. This is an illustrative assumption, not a quote or spending
guarantee. Provider-side project controls remain the independent spending guard.

First run five clips, inspect provider usage/cost and per-request latency, decode
and listen to the resulting audio, then choose the next explicit job allowance.
Suggested next batch is at most 100 attempts with a reviewed input/output envelope.
At the configured maximum 1,000 attempts per job, completing all missing clips
requires at least 11 separately budgeted jobs; failed/changed clips require later
explicit review. Do not increase worker concurrency until the pilot establishes
throughput, rate limits and error behavior. Stop on provider/quality failure.

The pinned model is `gpt-4o-mini-tts-2025-12-15`, voice `marin`, instruction identity
`sense-en-US-v2`. OpenAI's [deprecation notice](https://developers.openai.com/api/docs/deprecations)
announces retirement on 6 January 2027. Plan a separately reviewed provider
migration before then; this tool does not change models or regenerate already
validated clips. Persisted cache bytes remain independent of provider request
availability while their identity/configuration stays stable.

Re-audit after each batch and require zero missing/pending/failed current word
identities before claiming complete stored coverage. Validate sampled browser
playback without installed English voices, speed controls, denied autoplay and
navigation cancellation. Heteronyms and a stratified sample of common/rare words
need listening review; decoded signal alone cannot prove pronunciation accuracy.
Image readiness likewise requires separate sampled public-byte/hash checks.
