# Calm learning and broader word search — 7 October 2026

The owner requested a more stylish, less crowded and faster everyday experience,
including easy lookup beyond the lesson library. The delivery keeps Home,
Journey and Progress as the three primary destinations and retains the existing
blue identity, system typography and varied 3:2 teaching pictures.

## Observed problems and changes

Direct desktop and mobile inspection found repeated introductory copy, several
competing panels and an eagerly mounted writing editor on Home. The revised Home
puts one daily action first, a compact lesson recommendation second, and a
pictured saved word beside them on desktop. Writing opens deliberately from that
word. Home no longer requests a redundant current-user record or mounts sentence
feedback; its saved-word request fetches one spotlight word rather than three.
The mandatory middleware identity check remains in place.

Journey uses compact lesson recommendations and situation-specific icons.
Practice presents three clear mode choices, with selection help and completed
history expandable. Progress leads with vocabulary stages and dated activity;
lessons, sentence history and rewards expand on demand. Removing its duplicate
saved-word inventory removes one server request. Word pages show the meaning,
picture and example first; writing and personal tools are disclosures. Mounted
editors preserve drafts and pending retry identities when closed. Opening writing
records its fragment without adding browser history; session expiry retains the
fragment so the restored draft is visible after return.

Search remains in the normal app header. On the search page a single prominent
field replaces the header field; lesson filters and extra dictionary senses
expand on request. Secondary navigation follows the results, keeping the meaning
higher on small screens. A URL-derived form key fixes stale input values after
Clear filters and browser Back.

## Rendered layout comparison

The same isolated synthetic learner was captured at 360 pixels in light mode,
with fonts and images loaded, before and after the changes. Full-page heights
were Home 2,543 → 1,510 pixels, Journey 1,808 → 1,166, Practice 3,370 → 1,443,
and Progress 2,753 → 1,731. These describe the displayed default state, not a
speed benchmark; secondary information remains available through disclosures.
Direct review also covered 430 and 1,280 pixels and both themes. The dictionary
meaning and pronunciation fit above the mobile navigation in the inspected
360-pixel result. No screenshot review substitutes for interaction tests.

## Dictionary scope

Authenticated GET `/api/v1/dictionary?q=word` uses a bundled, immutable WordNet 3.0
excerpt: 82,710 headwords and 4,820 exception forms. It performs no external
lookup requests or learning mutations. Exact entries take precedence over source
exceptions and documented suffix rules; the resolved base word is displayed.
The query accepts a single English word of up to 48 characters with internal
apostrophes or hyphens. Lesson text search retains its broader 100-character
query and filters. Dictionary lookup is omitted while browsing, filtering lesson
words, or viewing the knowledge map. An exact lesson result takes precedence.

Dictionary results offer pronunciation and text reference, with the complete
original notice in an accessible source disclosure. They have no canonical
learning IDs, Save button, grading or invented teaching picture. General source
definitions are not graded A2–B1 teaching; coverage is broad, not exhaustive.
See [source, license and reproducible generation](../../apps/api/business/dictionary/SOURCE.md).
Missing entries, temporary failure, rate limit and expired authentication remain
distinct. The 20/minute requester limit is process-local, not a distributed quota.

## Reliability and acceptance

Server GET/HEAD requests now have an eight-second deadline covering response-body
consumption. Caller cancellation, cookies and 401 handling remain intact;
mutations are not timed out by this wrapper. This bounds a stalled read; it is
not evidence of faster normal network responses. Practice also exposes a working
Refresh action when lists are unavailable and no source selector is present.

Fresh local package and API checks pass: 260 foundation, 58 client and 186 web
helper tests, Go tests, formatting, lint, typechecks and production builds.
The focused browser matrix passed 117 cases at 360, 430 and 1280 pixels,
including both themes, keyboard controls, error recovery and exact writing retry.
The complete local sweep recorded 968 passes, 37 existing scope skips and 27
failures in old layout expectations or scan positioning. After repairing those
expectations without removing learning/state checks, all 93 affected learning
cases and 24 conversation cases passed against the same final production build.
The required hosted suite must verify the integrated test revision as a whole.

All 12 local Lighthouse audits passed: performance 100, accessibility 100 and
best practices 96 on Home, Journey, review and Progress at all three widths.
Separate cold-context probes used the local fixture API and three samples per
scenario. The Home probe configured with 4× CPU slowdown, 150 ms network latency
and 200 kB/s download recorded median first contentful paint 508 ms, load 2,061 ms
and 145 ms of long tasks. These are bounded lab observations, not production
latency or a measured improvement against a matched prior build. Staging will
exercise the real dictionary in its existing scoped synthetic-account journey.
Browser fixtures are synthetic; they do not establish live authentication,
provider feedback quality, real database durability or production latency.
