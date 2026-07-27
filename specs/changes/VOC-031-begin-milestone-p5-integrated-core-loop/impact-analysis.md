# VOC-031 — Impact Analysis

## Security and privacy

`VOC-031-R00`: cross-learner exposure of onboarding/settings/account data.
`user_onboarding_profiles`, the newly-public `user_settings` fields, and
`users.display_name` are all requester-owned personal state. Mitigate with
the existing authenticated requester context, service-level query scoping
(every new endpoint is implicitly self-scoped, no ID parameter to
enumerate), and exact-SHA review.

`VOC-031-R01`: this package adds the **first public write surface** onto
`user_settings` (previously internal-only) and onto `users` (previously
read-only outside auth). A missing or weakened `RequireAuth`/`CSRFMiddleware`/
`Idempotency-Key` check on either new endpoint would be a new cross-user or
CSRF write vulnerability where none existed before. Mitigate with the exact
same pattern as every existing P1/P2 write (`SaveUserWord`, `SubmitReview`),
tests asserting 401/403/409 on missing auth/CSRF/idempotency, and exact-SHA
review specifically comparing the new endpoints' middleware chain against
the established pattern.

`VOC-031-R02`: retroactively degrading an already-shipped, already-accepted
flow via the cross-feature/reliability/UX-consistency work (`T04`, `T05`,
`T08`). Unlike P4 (VOC-030), which added new writes *inside* existing
transactions, this package touches shared frontend components and
navigation/retry logic used by *every* P1–P4 screen at once — a bug here has
a wider blast radius per mistake, even though no backend business-logic
transaction is modified. Mitigate with: no change to any P1–P4 backend
transaction boundary in `T04`/`T05`/`T08`; integration tests asserting each
existing screen's core data and behavior is byte-for-byte unchanged after
adopting the shared components; and exact-SHA review of `T04`/`T05`/`T08`
independently, each scoped to frontend/reliability wiring only.

`VOC-031-R03`: retry/recovery logic (`T05`) introduces a new way to get
idempotency wrong — if a retried request reuses a stale `Idempotency-Key`
after the original request actually succeeded server-side, the client could
either falsely report failure (already-succeeded write treated as failed) or
mask a genuine duplicate-write bug. Mitigate with `VOC-031-AC-06`'s explicit
requirement that retry logic never silently resubmits a stale key in a way
that could mask a duplicate, and tests simulating exactly this
network-partial-failure scenario (request succeeds server-side, response is
lost client-side) for each of P1/P2/P3/`T02`'s writes.

## Data and migrations

`VOC-031-R04`: migration integrity for `user_onboarding_profiles`. Mitigate
with reviewed versioned Atlas SQL, the DOC-05 §18 ordering (added after
`grace_day_ledger`), the documented check/unique constraints, disposable
PostgreSQL forward/recovery rehearsal, explicit migration execution outside
API startup, and compatibility review confirming no existing
A1/P1/P2/P3/P4 table, column, or constraint is altered.

`VOC-031-R05`: the `VOC-031-D07` onboarding→`user_settings` seeding could,
if implemented incorrectly, overwrite a learner's already-customized
`daily_review_target` (e.g. a learner who used the internal default-20 path
and then, hypothetically, had a value set some other way before onboarding
completes). Mitigate with the explicit "never overwrite a non-default
existing value" rule in `VOC-031-AC-01`, and a dedicated test for the
already-customized case.

`VOC-031-R06`: the new accessibility/performance CI tooling (`T06`/`T07`)
could be installed as a non-blocking or warn-only job, giving the
appearance of automated coverage without actually gating merges — this
would directly violate the founder decision that these are not a
manual-only/spot-check fallback. Mitigate with `T06`/`T07` explicitly
requiring the new CI jobs to fail the build on a threshold/violation
failure, and independent-verifier confirmation that the job is
blocking, not advisory.

## Analytics and accessibility

Analytics: onboarding answers and settings values are structured
low-cardinality fields (enums, booleans, a bounded integer), never joined to
individual-identifying exports by this package; no learner sentence,
feedback, or free-text content beyond `displayName` (rendered back to the
same learner only) is touched. Accessibility is the dedicated subject of
`T06` (`VOC-031-R07`: an inaccessible or color-only-broken control anywhere
in the application, including this package's own three new screens) —
mitigated by installed axe-core/Playwright automation covering every route
at all three supported layouts, per the founder decision that rules out a
manual-only pass. `VOC-031-R08`: automation cannot verify every WCAG 2.2 AA
success criterion (e.g. some cognitive-accessibility and content-quality
criteria resist automated detection); any such residual gap must be
recorded as an explicit, named limitation in `staging-evidence.md`, never
silently treated as covered by the automated pass.

## Risks, dependencies, and evidence

- `VOC-031-R09`: three open decisions (`D06`–`D08`). `D06` (settings/account
  write-field boundary) blocks `T02`; `D07` (onboarding-seeds-`user_settings`)
  blocks `T00`; `D08` (routing-drift "do not rename" default) scopes `T08`.
  Founder adoption must resolve them into a `D09`/`D10` composite record
  before the affected tasks' evidence can be treated as final; this draft
  does not guess them.
- `VOC-031-R10`: this is the broadest milestone by design (DOC-12 §5 P5
  explicitly combines everything already built) — ten tasks touching every
  prior milestone's frontend surface plus two genuinely new backend
  modules. A task-ordering mistake (e.g. `T04`'s shared components landing
  after `T03`'s new screens, forcing rework) risks cross-task churn.
  Mitigate with the fixed `T00 → T01 → ... → T09` order and each task's own
  independent Claude Code review.
- `VOC-031-R11`: the DOC-12 §5 P5 gate's own wording ("works coherently in
  staging across supported layouts") is itself staging-dependent in a way
  no prior milestone's gate wording was as centrally — P4's gate
  ("missions accurately reflect completed behavior...") was primarily an
  in-repository-verifiable property with staging as one evidence item among
  several; P5's gate is *about* staging behavior. `VOC-031-DEP-02` (F3
  missing) is therefore the single largest completion risk for this
  milestone specifically, not merely a carried-forward item.
- `VOC-031-DEP-01`..`DEP-05`: dependencies recorded in `change.yaml`.
- `VOC-031-EV-00`..`EV-48`: migration, persistence, API, frontend,
  reliability, accessibility, performance, consistency, contract,
  evaluation, mock-inventory, staging, rollback, and exact-SHA review
  evidence referenced by the acceptance criteria.
