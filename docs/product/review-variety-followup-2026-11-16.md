# Review variety and daily-goal follow-up

The owner's three production annotations request a schedule menu that matches its
trigger width, one checked interval instead of a duplicated Automatic row, varied
review questions with pictures frequently used, and one daily-goal denominator.

## Design and implementation plan

1. Anchor the schedule popup to its own trigger. Show each interval once, check
   the saved automatic date's matching interval, and retain the actual automatic
   mutation until the learner explicitly chooses another date. Preserve unusual
   automatic dates, exact accessible dates, keyboard movement and touch targets.
2. Use server-confirmed daily effort/target for the visible review counter and
   progress bar, including older ten-word sessions and later remaining-word sets.
   Preserve frozen question identities and receipts. Where fewer words are due,
   explain availability rather than fabricate questions or completion.
3. Replace the four-picture allowlist with bounded selection from active meanings
   whose approved image revision matches current content. Prefer pictures, also
   randomly select sentence gaps and text questions, then freeze the result.
   Expand carefully reviewed authored gap contexts; do not generate unchecked
   examples during review. Four distinct nonalias choices, private entropy and
   concealed gap answers remain grading requirements.
4. Verify the annotated regressions, mobile/desktop, light/dark, keyboard and
   enlarged text. Run PostgreSQL grading/replay/effort tests and integrated checks;
   seek independent source review, then one consolidated follow-up PR through
   normal CI/merge queue, staging and the authorized production workflow.

## Competitor evidence

[WordUp's current official site](https://www.wordupapp.co/) describes vocabulary
through visuals, sound and real usage. [Duolingo's exercise-creation account](https://blog.duolingo.com/how-duolingo-experts-work-with-ai/)
describes turning authored language content into exercises, including targeted
sentence gaps. Its [flashcards article](https://blog.duolingo.com/duolingo-flashcards/)
adds active recall to recognition tasks. These support varied retrieval practice;
the picture preference here is the owner's product choice, not a claimed measured
WordUp question-frequency statistic.

## Acceptance evidence

Implementation and independent source review are complete. Picture selection now
uses current approved catalog publications; twelve authored gap contexts are
pinned and checked. Eligible question kinds use private weighted draws (picture
60, gap 30, definition 10), renormalized to available modes, and remain frozen on
reload. Existing saved questions retain their content.

Local API verification passed 2,782 test/subtest events across 33 test-bearing
packages with no failures or skipped tests (Go 1.26.10 / PostgreSQL 17). Build, types,
lint, formatting and vet passed. Chromium's affected-flow selection passed 287
cases with 4 existing scope skips. All 36 focused menu/goal cases passed separately on Firefox and WebKit; Firefox
required the existing 0.001px geometry tolerance for its 43.999992px serialization
of CSS 44px targets. Regressions also check live 200% text changes, vertical popup
clearance, complete keyboard access to the last option, automatic-date no-op
selection, unusual dates, legacy zones and 15-word goal continuity across saved
10/7-word sets.

Local foundation verification passed 333/337 with 4 macOS-specific shell harness
failures (BSD sed semantics, old system bash and Linux installer assumptions), whose
sources are unchanged by this delivery. Hosted Linux results, exact browser
counts, review outcomes and deployment identity are recorded in the consolidated
PR rather than inferred from these local results. Staging acceptance adds mandatory
live publication/CDN-byte/decoding checks for debris and wit outside the original
four-picture bank; the original four fingerprint checks and two gap checks remain.

Local synthetic browser tests do not prove provider quality or production
persistence. Extremely short viewports and arbitrary pinch-pan are not established
by the tested width/text matrix.
