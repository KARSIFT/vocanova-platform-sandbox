# Synthetic restore SQL contract

These files operate only in a disposable database named `vocanova_rehearsal`.
The runner owns isolation and distinguishes source from restored target; a database
name is an additional guard, not proof that a database is disposable.

Apply all committed forward migrations, load the canonical seed with explicit
column mappings, then execute `rehearsal-fixture.sql` once. It rejects a source
with existing users and intentionally does not upsert partial learner histories.
Use `psql -X -q -t -A --set ON_ERROR_STOP=1` for all three files.

## Nonempty reference state

The fixture inserts 15 records into 12 real application tables. Its only learner
is `restore-rehearsal@vocanova.invalid`, marked synthetic, with fixed ID
`d0c0a001-0000-4000-8000-000000000001`. No credentials, sessions, external identities,
tokens or provider requests are created. A2 onboarding and UTC settings specify a
five-review target and disabled notifications.

The saved word references the existing **catch up** meaning
`3d64c3c9-ede0-5ffd-b1ef-278f6b70e486`. One correct/good multiple-choice review takes
step 0 to step 1 at `2026-10-02 10:05:00Z`, due one hour later; all review counters
are one except consecutive incorrect, which is zero. These are historical fixture
timestamps, not a claim that the word remains future-due at execution time.

One original sentence has prewritten structured feedback. Provider/model values
explicitly say synthetic/fixed fixture; a succeeded stored outcome does not imply
that any provider ran or that feedback quality was measured. A daily mission is
open at 1/5 reviews with its one-word and one-sentence goals met. Activity agrees
with these source records. Four append-only reward events match current constants:
2 for saving, 5 for a good review, 3 for a sentence, 2 for feedback; running balances
are 2, 7, 10 and 12. No completed mission, streak reward or grace day is invented.
A real save idempotency claim retains the meaning/source fingerprint and owner.
Its historical timestamp does not prove present-day HTTP replay-window behavior.

## Verification and negative control

`rehearsal-check.sql` runs on the seeded source and again on the restored target.
It asserts explicit nonzero canonical counts and exact expected record values,
including identities, relationships, schedule, statuses, JSON, counters and
settings. It independently reconciles review, sentence, mission/activity and point
records. Consequently, two equally empty or equally malformed databases cannot
pass just because their digests agree. Expectations are static; the verifier never
imports its oracle from the dump or a source snapshot.

Selected schema protections must exist. Both ledger triggers must be enabled,
unconditional, row-level `BEFORE UPDATE OR DELETE` triggers for all columns, using
the expected function with no arguments. The exact PostgreSQL trigger event/type
bits are checked for both ledgers, including the empty grace ledger; an
UPDATE-only trigger cannot pass. Functional probes update one real settings row,
require the points-ledger trigger to reject both UPDATE and DELETE of a nonempty
record, and require the named result/rating check and composite review-ownership
FK to reject invalid writes. The probe transaction ends in `ROLLBACK`; failed probes
also abort it. A following read-only transaction verifies that settings, points,
review state and review meaning retain their original values. The runner must wait for successful psql exit before
accepting the final one-row JSON object or taking a post-check snapshot. Normal
output contains fourteen named true booleans and no data rows; errors use fixed
codes and must not be published with unsanitized PostgreSQL detail.

`rehearsal-missing-record.sql` is for the restored target only. It deletes exactly
one nonempty `daily_activity_summaries` record,
`d0c0a001-0000-4000-8000-000000000008`, after checking the synthetic identity. No
append-only protection is disabled. The subsequent verifier must fail with
`rehearsal_count_daily_activity_summaries`, and the runner's table comparison must
also detect the difference. A repeated fault application fails rather than
silently succeeding on an already empty table.

This is storage round-trip and selected-constraint evidence. It does not prove
production backups, RPO/RTO, off-host retention, live identity/provider behavior,
full API replay semantics or every application invariant. Update the explicit
canonical inventory and fixture contract deliberately when curriculum/schema or
reward policy changes. Root verification owns actual database execution; static
inspection alone is not a successful restore.
