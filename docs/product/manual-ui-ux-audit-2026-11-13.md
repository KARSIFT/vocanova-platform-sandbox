# Manual visual and functional audit — 13 November 2026

The owner requested a manual page-by-page review after noticing alignment and
styling inconsistencies. This pass starts from release 0.4.31 (`cf349f0f`).
Screenshots were visually inspected, rather than treating successful automated
checks as evidence of a coherent interface. Exact final CI and rollout evidence
belongs to the consolidated delivery PR.

## Observed corrections

| Observed problem                                                                 | Result                                                                                                                                                                                       |
| -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Landing columns overlapped and benefit labels clipped at 1024px with 200% text   | Content now wraps according to its reading width; normal desktop remains side-by-side and enlarged text stays readable without horizontal scroll.                                            |
| Sentence History's inline return link touched its page label                     | Return navigation and introduction occupy separate rows, including enlarged text.                                                                                                            |
| Plan controls touched on desktop and gained accidental indentation when wrapping | A single wrapping action group supplies consistent horizontal and vertical gaps.                                                                                                             |
| At 200% text, Practice's path action exceeded its indented row on narrow screens | The active action now has a full row inside the card; both Start and Continue stay bounded and keyboard accessible.                                                                          |
| Home's empty latest-word message used oversized dictionary typography            | Empty copy uses the normal interface font at the smaller label scale; real words keep their serif styling.                                                                                   |
| Account export differed from adjacent cards in padding, radius and heading size  | The export uses the same Surface and heading scale as peer account tools.                                                                                                                    |
| Every story card and several optional Plan actions looked equally primary        | Story browsing uses outline actions; one available continuation can be primary. Plan promotes a ready review or the recommendation, with preference controls secondary.                      |
| Practice, Lists, Plan and account submits used mismatched action shapes          | These actions use the existing shared button roles, preserving focus and disabled states.                                                                                                    |
| Typed recall put the instruction and complete definition into one large heading  | A short focused task heading precedes the intact server prompt in readable body text. No answer, picture or hidden teaching identity is exposed.                                             |
| Teaching and self-check words used interface typography                          | Actual teaching words use the existing headword font; question and completion headings remain interface text.                                                                                |
| Magic-link and email-change outcomes used a different brand shell                | They and their Suspense fallbacks share AuthShell and peer action styling. Token, CSRF, return destinations, retry and focus behavior are preserved.                                         |
| A failed synthetic export displayed only a generated HTTP status                 | The screen's recovery guidance replaces a bare generated status. Authored API guidance and 401 reauthentication are retained.                                                                |
| Shared action adoption initially hid selected audio/topic colors                 | Independent review caught the delivery-branch CSS regression. Component-layer defaults now allow authored selected/pressed utilities to apply; computed-color regressions cover both themes. |

No learning algorithm, grading identity, API DTO, provider flag, database migration
or dependency changed. The selected-state issue was caught before release and is
not presented as a previously observed production defect.

## Manual page and flow coverage

All 43 page route patterns were visited. They are not 43 independent rendered
screens: `/knowledge-map` redirects to Journey; `/login` and `/magic-link` are
aliases; `/review` and `/review/session` share the review flow. Source and final
URLs were recorded to avoid counting a sign-in redirect as a protected page.

| Page patterns                                                         | Manually operated or inspected                                                                                                                                                                                                                   |
| --------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/`, `/help`                                                          | Landing actions/learning preview; eight Help disclosures and learning destinations. No support message sent.                                                                                                                                     |
| `/signin`, `/login`, `/signup`, `/magic-link`, `/auth/magic`          | Signed-out entry, enabled methods, magic-link request/check-email/change-address, aliases and missing-link recovery. Signup form inspected.                                                                                                      |
| `/auth/password/reset`, `/auth/password/verify`, `/auth/email-change` | Missing-token states, reset request/check-email/change-address and account confirmation recovery. Real credentials were not changed.                                                                                                             |
| `/connection-error`, `/account-deactivated`                           | Recovery/confirmation layouts and destinations. The owner's account was not deactivated.                                                                                                                                                         |
| `/onboarding`                                                         | All five steps, language search, Back/Continue/target choices and synthetic completion request. The fixture's forced setup cookie was then removed; its override is not a production navigation failure.                                         |
| `/home`, `/discover`, `/knowledge-map`, `/discovery`                  | Overview, atlas/ranges, compatibility redirect, searchable word/range dialogs, direct navigation, a knowledge choice and subsequent word, optional-details failure/retry affordance. Automatic audio was switched off during the Discovery pass. |
| `/discover/[situation]`, `/discover/[situation]/[word]`               | Situation progress, word destinations, exact saved meaning and personal tools.                                                                                                                                                                   |
| `/vocabulary`, `/vocabulary/[word]`, `/vocabulary/check`              | Header search and results, exact word save, private note/save, self-check choices and ten-word completion.                                                                                                                                       |
| `/words`, `/words/[userWordId]`, `/writing`, `/progress/sentences`    | Empty/populated saved states, saved detail, optional sentence entry, feedback and saved history; topic chooser and initial writing states.                                                                                                       |
| `/courses`, `/learn/[lessonKey]`, `/plan`                             | Recommendations/direction controls, teaching, meaning and contextual questions, wrong-answer correction, full guided lesson completion and next actions.                                                                                         |
| `/reviews`, `/review`, `/review/session`                              | Empty/ready entry, actual started session, wrong-answer feedback, continuation and completion. Synthetic progress only.                                                                                                                          |
| `/practice`, `/practice/session/[sessionId]`                          | Mode/source controls, typed wrong/correct/assisted flow and completion; listening screen and audio-speed selected state. Additional journey/sentence-assembly states have fixture browser coverage.                                              |
| `/stories`, `/stories/[storyKey]`, `/stories/session/[sessionId]`     | Library/detail, actual started session, progressive dialogue, wrong-answer correction, final question and full completion.                                                                                                                       |
| `/lists`, `/lists/[listId]`                                           | Create, rename, membership addition, empty/populated details and optional management. No owner list changed.                                                                                                                                     |
| `/progress`, `/settings`, `/settings/profile`, `/settings/account`    | Progress/empty history, profile and preference saves, appearance, calendar-file preparation, email-change request and export failure recovery.                                                                                                   |

The screenshots cover narrow mobile, larger mobile and desktop views in light and
dark themes. Focused controls and scrolled action groups have viewport companions.
A fixed header or navigation bar stitched into a full-page screenshot is a capture
artifact, not evidence of a live overlap. Early transient reflow/redirect captures
were qualified and replaced where used for final visual conclusions.

## Verification

- New layout, learning hierarchy and selected-control regressions: **144 passed**,
  across three engines, three widths and both themes; zero skips/failures/flakes.
- Full hosted regression exposed the narrow Practice action overflow, which the
  local system font had kept inside the viewport despite exceeding its row.
  A stricter row-containment test failed before repair, then **18 passed** across
  all nine engine/layout projects and both themes for Start and Continue. Existing
  navigation tests now operate the retained empty-state action and still exercise
  the populated header. Settings retry tests require the authored recovery message
  instead of the replaced raw HTTP status; mutation assertions remain intact.
- Boundary audit: **168 passed** at 561/768/1024px in both themes, including
  enlarged text. The new Landing regression checks normal desktop columns,
  page/title/benefit reflow at 100% and 200%, and keyboard access. Six additional
  Landing checks passed at the standard widths. These final runs had zero
  failures, skips, flakes or retries.
- Existing entry-layout and dynamic visual-state suites: **198 passed** on Chromium
  at 360/430/1280px, both themes, with enlarged text, keyboard and axe assertions.
- Broader affected feature regressions: 321 cases passed initially; 12 failures
  were obsolete literal styling expectations. The corrected hierarchy suite passed
  all 42 cases, including the 30 overlapping cases; no behavioral assertion was
  removed. Those runs cover 333 unique case/project combinations.
- The audio-speed transition regression passed 20 consecutive larger-mobile cases
  after waiting for actual CSS transitions before taking its baseline. No arbitrary
  delay or retry was added. Full final CI remains independently required.
- Web helpers/middleware: **302 passed**, including generated-status fallback,
  retained authored API guidance and 401 destination behavior. Build, app/browser
  TypeScript, lint and formatting are recorded in the delivery evidence.
- Independent source review identified the selected-state issue and reviewed its
  correction. Independent screenshot review verified the visible spacing,
  typography, card and auth-shell repairs.

## Evidence boundaries

The comprehensive interaction pass uses isolated synthetic local learners. The
live production visit was read-only. The local export endpoint intentionally has
no implementation: the observed 404 establishes error presentation, not a broken
production export. Successful export behavior is covered by the existing explicit
API fixture and release acceptance, not that manual click.

Password/Google/email provider completion, actual email arrival, pronunciation
quality, physical mobile keyboards, screen readers and long-term database durability
are not established by this pass. Existing browser tests exercise synthetic auth,
error, concurrency and grading branches separately. No owner progress, settings,
credentials or private data were changed. Screenshots of synthetic accounts and
private working evidence stay out of the public repository.

The prepared learner study remains separate: the owner is arranging participants
and zero real learner sessions have been observed. This pass does not claim every
possible input/state/device combination or that the whole product is complete.
