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

There is no new API contract, database migration, authentication policy, combined
learning session, feedback provider, reward model or telemetry collector.

## Local acceptance

The final production build, lint, app/browser TypeScript, formatting and all
302 web helper/middleware tests passed. All 11 manually changed application
sources predate the final build; their hashes are retained with private evidence.

Final new-flow acceptance passed **411 cases** across Chrome, Firefox and WebKit
at 360px, 430px and desktop, including both themes and 200% text. Twelve intentional
skips run the 1024px Home breakpoint only in desktop projects rather than duplicate
it in phone projects. There were zero failures, flaky cases or retry attempts.
An earlier 252-case cross-engine run also passed without skips or retries.

The surrounding Chrome regression suite passed **461 cases**, with one existing
phone-only desktop skip and no failures, flaky cases or retries. After the last
review-option spacing change, the affected review/hierarchy subset passed another
66 cases without skips or retries. A focused cross-engine review/Home run passed
42 cases with the same 12 intentional breakpoint skips before the final matrix.

Independent rendered review covered six widths (360, 430, 561, 768, 1024 and 1280)
and both themes: 12 complete state journeys, 220 viewport screenshots and 54 axe
scans, with no serious or critical findings. A final review-only pass after the
last spacing fix added 76 viewport captures and 18 scans across the same 12
contexts, all passing. Actual images were inspected, including keyboard focus,
full review dates, 200% answer readability and the correctness marker. Earlier
failed artifacts were retained rather than overwritten or counted as acceptance.

Independent source/screenshot review and reproducing regressions found and fixed:

- Unrequested Discovery audio with absent or unavailable preference storage.
- A confirmed final revisited Learn receipt erased by completion URL normalization.
- Clipped enlarged Discovery actions and a hidden review date.
- Review fieldset minimum sizing, squeezed correct-answer text, keyboard actions
  scrolling under the header, and a too-narrow Home main column at 1024px/200%.

Test-harness corrections preserved business assertions: save refresh setup,
settled fresh-map replay state, independent Personal tools expansion, explicit
sound opt-in, unambiguous banner/answer locators and WebKit's fractional-device
outline rasterization. No production logic was changed to accommodate those
harness errors.

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
