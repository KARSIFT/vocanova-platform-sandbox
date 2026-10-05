# Performance and recovery — 6 October 2026

The owner reported slower use after the learner-experience changes and requested
a sustained bug investigation. Fresh public release identities showed staging on
`3e54a92a62e03706b941ea24b1644b02fab1073f` and production on
`af3eacb44edacedd4ee0fec0fa9d1f3026e9f742`. The redesign and resume correction
had reached staging only. Production promotion remains a separate operation.

## Reproduced delays and changes

Home, Journey and Progress unnecessarily started independent API reads in separate
waves. Writing also deferred its optional topic lookup. They now start independent
reads together, preserving mandatory errors, optional service fallbacks and
canonical meaning validation. Progress consistently treats optional reads' 401s
as expired authentication. No learner state is cached or optimistically advanced.

Local measurements used a production Next build, loopback fixture API, three
completed HTML responses per route and a cookie that delays every GET by 250 ms.
The mandatory middleware identity check remains a separate first wave. The
following are medians, not production latency or database measurements:

| Route                            |   Before |  After |     Reduction |
| -------------------------------- | -------: | -----: | ------------: |
| Home                             |   780 ms | 522 ms |           33% |
| Journey                          |   778 ms | 518 ms |           33% |
| Progress                         | 1,037 ms | 520 ms |           50% |
| Writing without a selected topic |   518 ms | 518 ms | None expected |

Real-source promise-barrier tests establish concurrent request starts without
using fragile elapsed-time assertions. Topic writing still waits for membership
validation before looking up its selected canonical meaning. Practice and lesson
routes now expose loading feedback during delayed transitions; keyboard navigation,
new lessons and resumed lessons are covered at all three supported widths.

Recommendation lookup previously loaded canonical content once for every one of
the 30 lessons. One batched query preserves active situation links, exact meaning
identities, canonical completeness, catalog ordering and frozen historical resume.
The repository path now makes five SELECTs rather than 34. SQLMock proves the
query bound and unchanged results; it does not establish database execution time.

## Recovery defects

Sentence checking now restores a missing CSRF cookie through the existing identity
endpoint, as other learning actions already do. It captures the exact logical
request before recovery and locks synchronously, retaining the draft and request
key across delayed recovery, authentication return and lost responses. Protected
writes still require a valid double-submit token. Failed preparation preserves
previous feedback and crisis guidance.

Temporary session-storage failures previously became 401s. Validation now separates
missing or invalid sessions from unavailable storage. Protected endpoints return a
generic, uncached 503 without running the handler, clearing cookies or exposing
storage details. Public sign-in still works with an old cookie. A protected page
shows a recoverable connection screen for identity transport/service/body failures;
genuine 401s still go to sign-in. A full-navigation retry keeps a normalized local
destination. The identity request has a 10-second deadline; a successfully read
identity with a missing onboarding field retains the conservative onboarding gate.

Mission and sentence-progress transactions previously requested a second database
connection for snapshot/history/streak reads. Three single-connection regressions
reproduced waits until their context deadlines. Those reads now use the transaction
connection, see its uncommitted changes and retain existing advisory locking and
commit/rollback ownership. No production pool setting was changed, and these waits
were not established as the cause of current production slowness.

## Acceptance boundaries

Focused acceptance includes 26 real-source page concurrency/error cases, 27 delayed
navigation browser cases, 24 sentence recovery cases and 12 connection recovery
cases. The full browser suite and hosted checks are delivery gates. Browser tests
use synthetic users and responses; they do not prove production authentication,
provider quality or persistence.

Initial local cold browser probes found no long tasks on Home, Journey, Progress
and Practice at 360 and 1280 pixels. Twelve initial Lighthouse audits met existing
budgets on the requested pages. The largest live shared script downloaded fully
in this check (Windows compressed 0.44–0.45 seconds, WSL 0.34 seconds); the previously
observed partial-transfer problem was not reproduced. These bounded observations
do not establish long-term network health or low-end-device performance.

Lighthouse now rejects redirected or failed documents before accepting scores and
owns a temporary browser profile outside the checkout, including under WSL. Budgets
and existing monitoring remain unchanged.

Local PostgreSQL/Docker execution is unavailable. The added CI read-reliability
job applies committed migrations and canonical seed to disposable PostgreSQL 16,
requires explicit PASS for recommendation coverage/privacy and single-connection
mission/sentence recovery, and rejects skips. Its hosted result is required before
claiming the changed SQL works against the real database. The existing controlled
authentication workflow separately supplies its disposable database evidence.

Future investigation can measure real database query timings and low-end-device
execution before changing monitoring imports, and assess bounded waiting for other
learning requests while preserving exact pending intents. This delivery does not
claim every possible product defect is removed.
