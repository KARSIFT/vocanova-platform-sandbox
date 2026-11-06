# Full panel audit and performance experiments — 6 November 2026

Status: implemented and independently reviewed locally; hosted acceptance and
production release remain separate.

The owner requested a product-wide functional audit and measured comparisons of
alternative implementations for responsiveness and memory. The audit covered all
route families rather than selecting only a reported slow screen. Independent
agents inventoried coverage, inspected frontend/backend bottlenecks, prepared
experiments, implemented the selected change, and reviewed correctness.

## Functional coverage

The existing 85 browser spec files cover entry/authentication recovery, onboarding,
Home/navigation, Journey/discovery, vocabulary/search/saved words, personal lists,
reviews, guided lessons, independent practice, stories, writing/feedback, progress,
and account/settings. The configured layouts are 360px, 430px and 1280px Chromium.
Tests include keyboard, theme, empty/error, pending/retry and expired-session cases;
not every individual scenario is duplicated in every theme or layout.

`apps/web/tests/e2e/full-panel-audit.spec.ts` adds Help disclosure/link keyboard
coverage, personal-data export download/pending/error/expiry checks, and mobile
password recovery/profile persistence. Its export responses are intercepted per
test, so these cases establish frontend behavior, not live export completeness.

Local evidence:

- Fresh production web build: passed using Node 24.18.0 and pnpm 11.14.0.
- Web helper/middleware tests: 251 passed; API-client tests: 61 passed.
- Broad browser run: 1,349 passed, 41 intentional skips, two newly added test
  locator failures. The corrected audit file then passed 26 cases with four
  intentional desktop skips, including the separately exercised Help cases.
- Browser-test TypeScript check passed. The local runner used installed Chrome
  because the pinned Playwright browser download returned HTTP 403.
- All 12 Lighthouse audits passed existing performance/accessibility/best-practice
  budgets on Home, Journey, Reviews and Progress at the three configured widths.
  The audits used the fresh local production build and installed Chrome.
- Final optimized full Go suite against fresh migrated/seeded disposable
  PostgreSQL: 2,583 test/subtest passes, zero test failures or skips; 30 packages
  passed and ten had no test files. Installed Go 1.27.3 was used because pinned
  Go 1.26.5 could not be downloaded. This is diagnostic evidence, not the pinned
  CI acceptance result. `go vet ./...` and `go build ./...` also passed.
- The initial macOS foundation run had 304 passes and eleven failures. Seven
  failures passed with canonical temporary paths; the remaining four passed in
  a disposable Linux container with Node 24.18.0, Bash 5.2 and GNU sed. A full
  Linux attempt had 300 passes and fifteen missing-tool/dependency failures.
  There is no claim of a single integrated green local foundation run.

The seeded saved-vocabulary integration test incorrectly assumed the entire
catalog fit in two pages. It now traverses all cursors for a bounded fixture of
more than 100 distinct seeded words, checks page limits/uniqueness/completeness,
and distinguishes due-review word counts from ordinary saved-meaning counts.
Restricting its shared-database fixture to seeded catalog/authored words prevents
it from holding references to another package's temporary meanings.

## Knowledge-map summary experiment

The original summary performed a lateral latest-saved-row lookup for every one
of the 10,000 catalog entries. Historical removed rows are needed for an item's
revision and restore target, but contribute nothing to summary learning counts.

Four query shapes were compared on the same disposable PostgreSQL 16 database:
the original lateral lookup, `DISTINCT ON`, `ROW_NUMBER`, and a direct active-row
join. Fixtures included 0/100/1,000/5,000 meanings of history, mixed active,
removed, ignored, archived and known states, and tied historical timestamps with
different IDs. Comparisons alternated variant order after warm-up. The original
and selected candidate each have 20 measured samples per case; exploratory
sorting variants have five. Each statement was bounded to five seconds.

| History meanings | Original median | Active-row median | Original p95 | Active-row p95 |
| ---------------: | --------------: | ----------------: | -----------: | -------------: |
|                0 |       22.715 ms |         18.025 ms |    22.923 ms |      18.393 ms |
|              100 |      124.457 ms |         18.850 ms |   130.997 ms |      19.472 ms |
|            1,000 |      948.293 ms |         20.170 ms |   966.409 ms |      20.806 ms |
|            5,000 |    3,848.154 ms |         22.316 ms | 4,697.625 ms |      22.808 ms |

Exact summary equality for the selected active-row candidate passed for every case. At 5,000, the comparison materialized
each query once because evaluating the slow original twice in a symmetric
comparison exceeded the statement limit. `DISTINCT ON` showed planner-sensitive
regressions (218.672 ms at 100 histories); `ROW_NUMBER` reached 4,302 ms at 1,000
and timed out at 5,000. Neither was selected.

The selected change affects only the aggregate summary. The existing partial
unique index on active `(user_id, meaning_id)` rows guarantees at most one joined
row. Known-state precedence and status filters remain unchanged. Band items,
individual items, decisions, revision tokens, historical restore ordering and
transaction boundaries retain their original queries and behavior.

These are warm-cache SQL execution times, not production, HTTP or browser timings.
Shared-buffer hits in the final 5,000-case plan fell from 4,131,394 to 75,674;
buffer hits are access counts, not retained-memory measurements. There were no
temporary spills. Go allocation and browser-memory comparisons are separate.

## Full-service latency and allocations

The committed `BenchmarkKnowledgeMapReadMatureHistory` measures the complete
service read, including the unchanged 100-item query, revision construction,
decoding and local database transport. Both original and optimized sources
passed the new semantic regression. The original source was supplied through a
Go overlay; both versions used the same current benchmark and fresh isolated
schemas. Runs were serial, with one untimed warm-up and 20 timed reads per case.
No browser/build/test workload ran concurrently.

| History meanings |  Before mean | After mean | Speedup | Before p95 | After p95 | Before / after Go B/op |
| ---------------: | -----------: | ---------: | ------: | ---------: | --------: | ---------------------: |
|              100 |    95.920 ms |  26.603 ms |   3.61× |   97.18 ms |  27.05 ms |      208,262 / 207,300 |
|            1,000 |   736.599 ms |  30.439 ms |  24.20× |   947.4 ms |  31.86 ms |      207,067 / 207,423 |
|            5,000 | 3,823.971 ms |  65.784 ms |  58.13× |   4,019 ms |  66.47 ms |      206,241 / 207,038 |

Go allocation counts also remained approximately unchanged (2,958–2,965 before,
2,960–2,962 after). This establishes a large latency improvement without a
meaningful allocation increase; it does not establish lower retained process or
database memory. These are service/database measurements, excluding HTTP,
authentication and browser overhead. Hardware was Apple M5 Max, darwin/arm64,
with installed Go 1.27.3 and disposable PostgreSQL 16 in Docker.

Reproduce against a disposable migrated database using the keyword DSN described
in `docs/development.md`, from `apps/api`:

```sh
go test ./business/knowledgecatalog -run '^TestPostgreSQLMapSummaryPreservesHistoryAndUserBoundaries$' -count=1
go test ./business/knowledgecatalog -run '^$' -bench '^BenchmarkKnowledgeMapReadMatureHistory$' -benchtime=20x -benchmem -count=1
```

The before comparison substitutes `postgres.go` from `e4ffa8ce` using Go's
`-overlay`, keeping the new tests/benchmark in place. Neither the benchmark nor
regression should be treated as passing when the database fixture skips.

## Browser and memory evidence

Repeated production-build measurements used installed Chrome 156. The private harness
separates new-browser-context loads, warm reloads, interactions, and post-GC
navigation-cycle heap/DOM/listener counts. It checks route-specific content and
retains invalid samples, HTTP failures and cancellations rather than reporting
loading/error screens as successful results. External imagery is excluded from
these controlled comparisons, so they do not establish CDN/image memory behavior.

- Native baseline: 17 route families × three widths × three repetitions × cold
  and warm loads = **306 valid load measurements**. There were 35 valid and ten
  excluded interaction samples, plus 27 valid post-GC memory checkpoints.
- Eight Home/Journey/Progress navigation cycles per width kept DOM node and
  listener counts constant after warm-up. Final post-GC heap was 5.23–5.30 MiB;
  initial heap was 4.03–4.11 MiB. The last four checkpoints had stable DOM/listener
  counts. This bounded run does not prove the absence of longer-lived leaks or
  characterize image/GPU memory.
- A runtime counterfactual blocked only explicitly marked Next prefetch requests.
  It is not equivalent to a production source rewrite. The wide native baseline
  and narrower blocked route sweep were not ranked as a matched comparison.
- A separate matched comparison used three routes at 360px, CPU 4× slowdown,
  150 ms latency and 1.6 Mbps bandwidth, three repetitions per cache state.
  Blocking prefetch changed cold heading-ready medians by about −2 to +2 ms and
  saved roughly 3 KB per cold page. It did not establish a useful first-content
  improvement. Navigation samples with unresolved canceled RSC streams did not
  qualify for ranking. Existing selective prefetch remains unchanged.

A read-only production container snapshot showed approximately 85 MiB for web
and 35 MiB for API, each with a 512 MiB limit. This does not establish a leak or
long-term memory trend. Three public probes each of the landing page, sign-in and
API health returned HTTP 200. No owner learning state or account settings changed.

## Remaining acceptance boundaries

Hosted checks on the pinned versions remain required before delivery. The local
browser suite is synthetic and uses installed Chrome. Live OAuth/provider behavior, actual AI/audio
quality, physical devices and non-Chromium browsers are not established by the
synthetic suite. Production deployment remains a separate manual operation.
