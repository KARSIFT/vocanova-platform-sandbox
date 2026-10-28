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

## Prefetch, deferred monitoring and feedback — 28 October 2026

The owner again reported a slow-feeling app. Public staging and production
documents and API health answered in about 150ms on a reused connection from the
owner's region, with no multi-second stall, so the investigation measured the
authenticated flows locally: a production build, the mock API, synthetic
learners, installed Chrome, medians of 3–7 samples. These are lab results, not
field data, database timings or INP.

**Causes found.** Each protected page view started 24–26 router prefetches, and
every prefetch passed the identity check: one Home view made 32 API reads (27
identity checks), Journey 27; blocking prefetch cut Next server CPU per Journey
view from about 100ms to 30ms. Every route also shipped the error-monitoring
SDK in its critical JavaScript (about 52KB compressed of 211–237KB). Routes
without a loading state showed nothing for 355–530ms after a tap at 150ms API
latency. A situation word page read the word only after validating the
situation (three API waves), and each discovery choice waited for its POST and
then a full re-read of the 100-word range.

**Changes and results** (before → after, same build settings and fixtures):

| Measure                                        |                  Before |                   After |
| ---------------------------------------------- | ----------------------: | ----------------------: |
| Idle RSC prefetches, Home / Journey (1280px)   |                 26 / 24 |                 12 / 10 |
| API reads per view, Home / Journey             |                 32 / 27 |                 18 / 13 |
| Critical JS, Home / Journey (compressed)       |         236.5 / 211.5KB |         182.8 / 158.7KB |
| Mobile cold `load`, Home / Journey / search¹   | 2,040 / 1,674 / 1,633ms | 1,719 / 1,374 / 1,373ms |
| Situation word page at 150ms API latency       |         474ms (3 waves) |         319ms (2 waves) |
| Discovery choice to next word (desktop, 150ms) |                  374ms² |                   193ms |
| First acknowledgement of a tap (150ms API)     |                 0–530ms |                  4–94ms |
| Journey page height at 360px                   |                 4,927px |                 1,845px |

¹ 4× CPU slowdown, 1.6 Mbps/150ms emulated network, 50ms per API response; the
monitored build (well-formed fake DSN) measured 1,720 / 1,380 / 1,380ms.
² One baseline sample.

- Links prefetch on intent (a resting pointer or keyboard focus, not touch
  scrolling). Primary navigation and prominent links into routes with a loading
  state still prefetch on view: without that, a tap showed the loading state
  only after the identity check and React held content back, measured 524ms
  versus 377ms for Home → Practice. Middleware authentication is unchanged.
- The monitoring SDK loads from a three-export module after `load` and idle
  time, so its 51.8KB compressed chunk is off the critical path (a whole-package
  dynamic import would have been 162KB). Early errors are buffered (at most 20)
  and drained once; a failed download or initialisation is retried at most
  three times, at least 10 seconds apart, on the next error, report or `online`
  event, without unhandled rejections. Errors before the client bundle runs are
  missed, as they were with the earlier eager initialisation.
- A shared indeterminate bar acknowledges every in-app navigation within a frame
  and clears on arrival, cancellation, Back/Forward or a 15-second fallback.
- Discovery advances within a range from the server's reported state of the
  decided meaning; replays that report a different current state still re-read.

**Limits.** No real PostgreSQL or production authenticated timing was measured
(Docker was unavailable). Production's web server reaches the API through the
public edge, unlike staging's internal address; each identity check therefore
costs more there, which this change reduces but does not measure. Mobile FCP
under 4× CPU emulation was 20–40ms later after the change in three samples per
route, while `load` improved 15–18%; the cause was not isolated. Journey's
compressed document grew from 21.7 to 25.6KB because the map is rendered both
as links and as a phone picture with a range chooser.
