# Warm guided learning — 13 November 2026

The owner approved Warm guided coach and the [concrete proposal](../superpowers/specs/2026-11-13-first-five-minutes-design.md)
after comparing three clickable concepts. The [implementation plan](../superpowers/plans/2026-11-13-warm-guided-learning.md)
keeps the current learning model and applies the selected visual hierarchy to the
first learning journey.

## Delivered application behavior

- Home presents its existing server-selected activity on a calm reading surface
  with one filled blue action. Active review, mission targets, saved words and
  reviewed counts retain their existing sources and priority.
- Discovery adds brief guidance and an optional writing link after a confirmed
  Learn. It retains the original word and exact meaning while advancing normally.
  Lost receipts, replay, navigation and fresh-state reconciliation preserve the
  existing request identity and do not manufacture a saved record.
- Each saved meaning exposes Optional writing before Personal tools. Canonical
  and saved-word routes retain exact record IDs, independent drafts, keyboard
  disclosure behavior and existing fragments. Real feedback retains original and
  corrected text, explanation, revision, retry and report controls.
- Discovery defaults to quiet entry, including direct word links, while retaining
  an explicit sound-on session preference and manual Listen.
- Review uses shared reading surfaces and controls, starts automatic sound off,
  and describes the actual recorded result and returned schedule. Four ratings,
  first-answer rules, counts, locks and same-day scheduling remain unchanged.

Profile controls now wait for identity-scoped draft recovery before accepting
edits or Save. Slow JavaScript can no longer turn a pre-interactive save into a
native document navigation that drops the typed name. Typing during an actual
save still keeps a newer draft. The browser tab also uses the existing blue V
identity through a self-contained SVG icon.

There is no new API contract, database migration, authentication policy, combined
learning session, feedback provider, reward model or telemetry collector.

## Local acceptance

The final production build, lint, app/browser TypeScript, formatting and all
302 web helper/middleware tests passed. Changed application sources predate the final build; their hashes and exact
verification outcomes are retained with private evidence.

Final new-flow acceptance passed **447 cases** across Chrome, Firefox and WebKit
at 360px, 430px and desktop, including both themes and 200% text. Twelve intentional
skips run the 1024px Home breakpoint only in desktop projects rather than duplicate
it in phone projects. There were zero failures, flaky cases or retry attempts.
The final matrix includes Arial/Verdana entry stress after the first hosted run
exposed platform font differences; earlier runs remain historical evidence.

The surrounding Chrome regression suite passed **461 cases**, with one existing
phone-only desktop skip and no failures, flaky cases or retries. After the last
review-option spacing change, the affected review/hierarchy subset passed another
66 cases without skips or retries. After hosted findings, the original Discovery
entry regressions passed 36 cases across all nine engine/layout combinations;
Home image regressions passed 27 native cases plus three font-stressed cases.
Full Profile checks passed 63 cases with 18 existing representative-desktop
scope skips; the new blocked-JavaScript readiness case passed all nine engine/layout combinations. The exact manual-date continuation scenario also passed nine cases,
retaining the original request key/body and credited-review assertions. No accepted
run had failures, flaky cases or retry attempts.

Independent rendered review covered six widths (360, 430, 561, 768, 1024 and 1280)
and both themes: 12 complete state journeys, 220 viewport screenshots and 54 axe
scans, with no serious or critical findings. A final review-only pass after the
last spacing fix added 76 viewport captures and 18 scans across the same 12
contexts, all passing. A fresh follow-up after the hosted findings passed six
changed-screen contexts with 100 viewport captures and 20 axe scans, including
wider fonts, entry/completion controls, and Profile loading/ready/saving states.
All 100 images were inspected through 32 contact sheets plus selected original
PNGs. Actual images confirmed keyboard focus, intact enlarged words, full review
dates, answer readability and the correctness marker. Earlier
failed artifacts were retained rather than overwritten or counted as acceptance.

Independent source/screenshot review and reproducing regressions found and fixed:

- Unrequested Discovery audio with absent or unavailable preference storage.
- A confirmed final revisited Learn receipt erased by completion URL normalization.
- Clipped enlarged Discovery actions and a hidden review date.
- Review fieldset minimum sizing, squeezed correct-answer text, keyboard actions
  scrolling under the header, and a too-narrow Home main column at 1024px/200%.
- Discovery entry controls depending on narrow-font metrics, and Profile allowing
  native submission before its client recovery was ready.

Test-harness corrections preserved business assertions: save refresh setup,
settled fresh-map replay state, independent Personal tools expansion, explicit
sound opt-in, unambiguous banner/answer locators and WebKit's fractional-device
outline rasterization. Hosted findings also required scrolling a genuinely lazy
image into view after the Review/source checks, using a reverse keyboard
roundtrip at Firefox's last-control boundary, and checking one visible completion
message in the active main instead of globally matching an inactive streamed
copy. The duplicate-copy interpretation is guarded by an explicit assertion that
there is exactly one visible matching message. No scheduling logic was changed.
The existing hosted Profile overlap flake was not conclusively attributed; the
separate pre-interactive form bug was independently reproduced and fixed.

Manual synthetic browser operation exercised Home, confirmed Discovery saving,
canonical saved writing, submitted feedback, an incorrect real review response,
completion and return Home without playing audio. These fixtures establish
rendering and interaction, not production persistence, actual AI quality or
identity-provider behavior. The local map fixture and canonical catalog are
separate transports; the strict staging core-loop gate therefore additionally
follows a real confirmed Discovery Learn through its canonical meaning to the
exact saved writing record, using only the designated synthetic staging identity.
That gate must pass before production promotion.

## Research and release limits

The [prepared learner study](first-five-minutes-study-2026-11-13.md) has a blank
observation sheet and zero sessions. The owner arranges participants. Prototype
checks, source reviews and browser acceptance do not establish comprehension,
retention, a winning design or a measured product improvement.

The consolidated delivery PR records exact-head hosted CI and review acceptance.
The normal merge queue, strict staging and separate manual production workflow
remain release gates. Deployment identity and final hosted evidence are recorded
with that PR and the private operator handoff after those gates pass.
