# VOC-045 — Impact Analysis

## Security and privacy

No new secret, credential, or attacker-controlled input is introduced. Both
affected call sites already operate on the authenticated caller's own
`user_id`; this fix changes only what value two existing `NOT NULL`
audit-timestamp columns receive on a genuine first INSERT. No personal-data
field is added, removed, or newly exposed. No authorization boundary changes.

## Data and migrations

No migration is required for this package's primary fix (see
`specification.md`'s open question 2 for the explicitly-flagged alternative
of also adding a schema-level `DEFAULT now()`, left to the reviewing human).
Existing `user_settings` rows are unaffected — they already have valid
`created_at`/`updated_at` values, since any row that exists today necessarily
went through a code path where the constraint was satisfied (e.g. gamification's
lazy-creation path when the onboarding path had not yet created a conflicting
row, or manual data). This fix only changes behavior for the previously-failing
INSERT path, which never persisted a row before (the transaction rolled back
on the constraint violation), so there is no pre-existing bad data to reconcile
or backfill. Rollback is a plain code revert with no data cleanup required.

## Analytics and accessibility

Not applicable. This fix touches only backend SQL upsert statements; no
analytics event, UI surface, or accessibility-relevant markup is affected.

## Risks, dependencies, and evidence

- `VOC-045-R00`: The confirmed defect is a full outage of new-user onboarding
  in production — every real, genuinely new user is blocked today. This is
  the primary risk this package addresses, not a risk this package
  introduces.
- `VOC-045-R01`: `specification.md`'s open question 1 (how
  `gamification/repository.go`'s `UpsertUserSettings` should source its
  timestamp, given it has no `time.Time` parameter today) could, if resolved
  by threading a new parameter through the function and its caller, touch
  `apps/api/business/gamification/service.go` as well — a slightly larger
  diff than the SQL-only `NOW()` alternative. Flagged so the reviewing human
  can weigh diff size against testability before `VOC-045-T01` is
  implemented.
- `VOC-045-R02`: If any other, currently-unidentified `user_settings` INSERT
  call site exists beyond the two named in the issue, it would share this
  same defect and remain broken until `VOC-045-T02`'s inventory grep finds
  and this package's fix (or a follow-up) covers it. Mitigated by making the
  repository-wide grep an explicit, evidenced task (`VOC-045-AC-03`) rather
  than an assumption.
- `VOC-045-DEP-00`, `VOC-045-DEP-01`, `VOC-045-DEP-02`: see `change.yaml`'s
  dependency entries — this package's implementation depends on the issue's
  reported reproduction and root cause remaining accurate, on the reviewing
  human resolving open question 1's implementation approach for
  `gamification/repository.go`, and on the reviewing human explicitly
  accepting or rejecting the schema-level `DEFAULT now()` alternative
  (open question 2).
- `VOC-045-EV-00` through `VOC-045-EV-03`: to be produced by `VOC-045-T00`
  through `T02` respectively (the fixed `CompleteOnboarding` behavior's
  passing verification, the fixed `UpsertUserSettings` behavior's passing
  verification, the new regression tests' passing run, and the call-site
  inventory). None exist yet.
