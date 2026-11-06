# Frontend control and scenario coverage — 6 November 2026

This is a source-to-test coverage ledger, not a declaration that every possible
input, browser or device has passed. The earlier full-panel report records actual
execution results. Entries below describe assertions found in test source; a spec
name alone does not establish a pass. Follow-up additions require fresh execution
before their gaps can be closed.

The inventory contains 43 public/application page routes (excluding the private
`reviews/_components/daily-lesson-page.tsx` component). All are mapped below.
Dynamic routes need the setup listed; merely visiting their parent does not test
the controls inside a session. Component controls shared across routes still need
at least one complete behavioral test plus representative placement checks.

## How to read the ledger

- **Existing** means concrete automated assertions exist, not exhaustive coverage.
- **Gap** means an identified missing dimension or missing action/state assertion.
- **Verify** means inspection has not established complete coverage; do not convert
  it to covered without matching action/outcome assertions and execution evidence.
- Browser widths, themes and engines multiply scenarios. Existing Chromium
  projects are 360×640, 430×720 and 1280×720; theme loops are local to each spec.
- Frontend API/media fixtures establish UI behavior. Provider delivery, genuine
  audio decoding/quality, production persistence and physical-device usability are
  separate acceptance dimensions.

All spec names in the table refer to `apps/web/tests/e2e/`; source paths refer to
`apps/web/src/app/`, including its `(app)` directory for authenticated routes.

## Route/control ledger — initial source audit

This table and the G01–G09 batch definitions preserve the initial gap inventory.
They are not the latest execution status: follow-up sections below supersede
labels such as “missing” and “execution pending” where new evidence is recorded.

| Routes                                   | Controls and state transitions to exercise                                                                                                           | Existing assertion sources                                                                                                                                         | Targeted dimensions (follow-up status below)                                                                                                               |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| `/`                                      | Brand; header sign-in; Start learning; footer action; authored preview disclosures; keyboard navigation                                              | `landing-accessibility`, `learner-workspace`, `theme-preference`                                                                                                   | G02: activate every public link and assert destination; all-engine visual/text enlargement                                            |
| `/signin`, `/login`                      | Email/password entry; Google; send link; resend/edit address; signup/reset links; disabled capability combinations; failure/429/success and returnTo | `password-auth`, `signin-accessibility`, `auth-feedback-capabilities`, `signin-oauth-capability`, `password-field-clearance`                                       | G01: complete password success lifecycle on phones; G09: genuine provider completion                                                  |
| `/signup`                                | Password validation; Google-only capability; submit/pending/check-email; links back to signin                                                        | `password-auth`, `auth-feedback-capabilities`                                                                                                                      | G01: desktop-only password scenarios on phones; light/dark full lifecycle                                                             |
| `/auth/password/verify`                  | Explicit Verify email; no automatic token consumption; used/invalid token; no sign-in after verification; token removed                              | `password-auth`                                                                                                                                                    | G01: phone successful verification/replay; keyboard/theme states                                                                      |
| `/auth/password/reset`                   | Request link; password validation; Save new password; success, invalid/replayed token; sign-in link                                                  | `password-auth`, `full-panel-audit`                                                                                                                                | G01: phone successful reset/replay; existing phone audit only validation and invalid token                                            |
| `/auth/magic`, `/magic-link` | Incomplete/invalid link feedback; return-to destination; sign-in recovery | `auth-magic-accessibility`, `auth-magic-return-to`, `documented-route-aliases`, `auth-feedback-capabilities` | Missing activation of Back to sign in (preserve returnTo) and Back to home from incomplete/rejected consume states; G09 real email |
| `/auth/email-change`                     | Confirmation action; invalid/missing token; signed-out continuation; authenticated success                                                           | `email-change-confirmation`                                                                                                                                        | G03 request-form branch coverage is separate; G09 actual delivery                                                                     |
| `/connection-error`                      | Retry link with sanitized original destination; no redirect loop or external redirect                                                                | `auth-check-recovery`, `middleware-runtime-regression`                                                                                                             | G02 keyboard activation; G04 true offline/reconnect separate from server 503                                                          |
| `/account-deactivated`                   | Confirmed message; purge date; Go to sign in; malformed/absent date                                                                                  | `account-deletion-confirmation`                                                                                                                                    | G03 direct malformed/absent date presentation and link behavior                                                                       |
| `/help`                                  | Eight disclosure open/close; six learning/account destinations; brand; Go to learning; signin; support mailto                                        | `full-panel-audit`                                                                                                                                                 | G02 brand/top/signin activation and mailto contract; do not send support email                                                        |
| `/onboarding` | Name, language search/select, Back/Next, completion; save pending/error/retry | `onboarding-accessibility`, `onboarding-journey`, `core-loop` | Both-theme language search/Back/persist exists; remaining keyboard-only Continue/Back/Finish completion and physical IME/software keyboard |
| `/home` | Primary review/discovery/start/resume; recommendation; optional writing; progress/help links; secondary read retries | `home-today`, `home-next-action`, `home-mobile-accessibility`, `overview-hierarchy`, `learner-workspace`, `route-loading`, `navigation-prefetch`, `auth-check-recovery`, `secondary-navigation-controls` | Stories/Lists/empty-word search activation added, execution pending; Practice/Situations/Plan/latest headword already activate; All your words and Write a sentence with it links remain |
| `/discover`, `/knowledge-map` | Band selection, phone chooser, desktop map, range-word links, known/saved indicators, legacy redirect/query handling | `journey-word-map`, `journey-resilience`, `discover-accessibility`, `learning-direction`, `daily-discovery` | Map keyboard selection, phone chooser and range viewport exist; malformed/duplicate/out-of-range direct Journey band inputs remain distinct from Discovery header normalization; G05 |
| `/discovery`                             | Already know/Learn/Skip; undo/retry; sound toggle; explicit replay; next range/end; exit                                                             | `discovery-advance`, `pronunciation`, `learning-direction`                                                                                                         | G08 long-session memory across full discovery; G09 genuine playback                                                                   |
| `/discover/[situation]` | Situation word links; guidebook disclosure; lesson/story/practice/writing links | `unit-guidebooks`, `core-loop`, `daily-conversation`, `journey-practice-coherence` | All 17 phrase/example disclosures already exercised; guide writing and guided-lesson links remain href-only; full long-content visual states separate |
| `/discover/[situation]/[word]`           | Previous/next word; save; pronunciation; optional examples; sentence entry; meaning picture retry; personal knowledge/notes                          | `word-detail-disclosures`, `word-knowledge`, `meaning-teaching`, `visual-learning`, `pronunciation`, `core-loop`                                                   | G07 all active/pending/error panels; G08 retain/release media and drafts on navigation                                                |
| `/vocabulary` | Query submit/clear; filter selection; pagination; dictionary result and base form; save; back/forward query state | `vocabulary-search`, `vocabulary-search-navigation`, `dictionary-search`, `pronunciation` | Search/clear/back state exists; pending older search followed by newer submission/history restoration lacks explicit late-response assertion; G06 IME; G08 large fixtures |
| `/vocabulary/[word]`                     | Dictionary/canonical meaning; save/remove; audio; back search; missing word                                                                          | `pronunciation`, `dictionary-search`                                                                                                                               | G07 dynamic page visual matrix; G09 provider audio                                                                                    |
| `/vocabulary/check`                      | Already know/Want to learn/Skip; pending disable; same-choice retry; completion plan link                                                            | `learning-plan`                                                                                                                                                    | G05 other engines; G07 empty/completion layout                                                                                        |
| `/words`                                 | Saved filters/search/clear; page cursors; word detail; practice entry; unavailable/empty distinction                                                 | `saved-vocabulary`, `saved-vocabulary-filters`, `vocabulary-search-navigation`                                                                                     | G08 large saved library, repeated filters and memory; G04 route error boundary                                                        |
| `/words/[userWordId]`                    | Remove; writing save/check/retry; draft recovery; knowledge/notes; pronunciation; not-found recovery                                                 | `saved-vocabulary`, `word-knowledge`, `sentence-feedback-resilience`, `word-detail-disclosures`, `pronunciation`                                                   | G07 create saved fixture before layout tests; not-found return link activation                                                        |
| `/courses` | Recommendation; all-lessons anchor; situation/lesson cards; Find a word/Saved words/Practice/Stories/Lists/Plan links; empty fallback | `guided-lessons`, `lesson-recommendation`, `journey-resilience`, `journey-practice-coherence`, `secondary-navigation-controls` | Header/activities/footer activation added, execution pending; empty-catalog Go to Reviews activation remains; G07 empty/unavailable visuals |
| `/learn/[lessonKey]`                     | Start/resume; teaching next/save; choice/typed/sentence answer; check/continue; skip audio fallback; completion and optional writing                 | `guided-lessons`, `mixed-guided-lessons`, `lesson-save-review`, `lesson-practice-selection`, `learner-answer-controls`                                             | G07 screenshots for each step type, pending/error/completion; G08 repeated sessions/media cleanup                                     |
| `/plan` | Recommended start, change preferences, self-check and practice destinations | `learning-plan`, `learning-direction`, `secondary-navigation-controls` | Preferences/calendar/self-check/map/practice/reviews activation added, execution pending; unavailable daily-progress/recommendation presentation separate |
| `/reviews`, `/review`, `/review/session` | Start/resume; answer select/type/check; continue; save/skip; retry stale/unknown request; sound toggle; completed summary/next                       | `daily-discovery`, `review-readiness`, `review-hydration`, `review-stale-card`, `review-completion-summary`, `learner-answer-controls`, `documented-route-aliases` | G07 real started session needed, each answer state; G04 route boundary; G08 long review session                                       |
| `/practice`                              | Mode/source/list choice; Start; saved-session resume/fresh; retry unavailable sessions/lists; course/story/review links                              | `practice-hub`, `practice-sessions`, `personal-lists`                                                                                                              | G07 selection combinations and empty/error visuals; G08 repeated start/exit                                                           |
| `/practice/session/[sessionId]`          | Typed recall; assistance/show-answer; listening; token sentence assembly/removal/reorder; submit/next/retry; completion/repeat                       | `practice-sessions`, `learner-answer-controls`                                                                                                                     | G07 every mode layout; G09 actual playback and mobile interaction                                                                     |
| `/stories`                               | Story cards; resume/start; empty/unavailable retry; back                                                                                             | `story-practice`                                                                                                                                                   | G07 catalog empty/error and resumed card visuals                                                                                      |
| `/stories/[storyKey]`                    | Transcript; vocabulary; optional audio; Start/Resume                                                                                                 | `story-practice`                                                                                                                                                   | G07 story detail in both themes; G09 actual audible transcript                                                                        |
| `/stories/session/[sessionId]`           | Next line; answer/check/check again; finish/repeat; stale recovery                                                                                   | `story-practice`                                                                                                                                                   | G07 line/question/wrong/completion visuals; G08 session navigation memory                                                             |
| `/writing`                               | Meaning/topic selection; draft; save meaning before feedback; submit/retry; revision/report; pending navigation protection                           | `topic-writing`, `writing-flow-polish`, `sentence-feedback-*`                                                                                                      | G06 IME; G08 long/rapid draft changes; G09 real AI quality and latency                                                                |
| `/lists`                                 | Create/search/rename/delete/cancel/confirm; unavailable retry                                                                                        | `personal-lists`                                                                                                                                                   | G07 modal/inline form pending/error visual states; G08 many-list scaling                                                              |
| `/lists/[listId]`                        | Add/remove member; practice selected list; stale reload; deletion elsewhere; lost-action retry                                                       | `personal-lists`                                                                                                                                                   | G07 created list fixture and membership states; G08 large membership                                                                  |
| `/progress` | Milestones/activity; sentence disclosure/history; practice and next-step links; unavailable vs empty | `progress-accessibility`, `progress-mission-status`, `achievements`, `learner-workspace`, `learning-flow-refinement`, `secondary-navigation-controls` | Six secondary destinations added, execution pending; due Start review, saved lesson return, empty-history mission and failed-writing history links remain conditional navigation cases; G08 extensive history |
| `/progress/sentences`                    | Older sentences pagination; retry; write-first link; long text wrap                                                                                  | `sentence-history`                                                                                                                                                 | G07 empty/error/paginated visuals; G04 boundary; G08 many pages                                                                       |
| `/settings` | Daily preset/custom targets; email switch; theme selection; save/revert/pending retry; calendar date/time/download/help; links | `settings-*`, `theme-preference`, `calendar-reminder`, `secondary-navigation-controls` | Back/Home, Account, Profile and Help activation added, execution pending; G06 date controls/zoom/high contrast; G09 real calendar import |
| `/settings/profile`                      | Display name; save/reload; concurrent edits; unverified email; session expiry                                                                        | `profile`, `full-panel-audit`                                                                                                                                      | G01 phone unverified email; G07 validation/error visual states                                                                        |
| `/settings/account`                      | Email-change request/edit address; add-password request; export/download/retry; deactivate confirmation/cancel/submit                                | `profile`, `settings-account-accessibility`, `full-panel-audit`, `account-deletion-confirmation`                                                                   | G03 email request branches, deactivation held-pending/401; G01 phone add-password; G09 real export/provider delivery                  |

## Concrete follow-up batches

### G01 — complete mobile authentication and profile coverage

`password-auth.spec.ts` has six desktop-only scenarios. `profile.spec.ts` has four
initial desktop-only scenarios. Existing phone audit cases cover only reset
validation/invalid-token and profile save/expiry, not the whole lifecycle. Run
successful signup verification, wrong/unavailable/rate-limited and successful
login, successful reset/replay, unverified-email status and add-password request
at both mobile widths; preserve test isolation of one-use tokens. Include keyboard
and both themes where absent. Review `settings-successive-save` desktop skips too.

### G02 — activate public and secondary navigation

Help's brand/top-learning/signin/mailto were inspected rather than operated in the
first audit. Landing checks largely assert href, reachability and geometry. Add
actual keyboard navigation and verify destination plus returnTo/auth behavior.
Check mailto href/accessibility only; launching a mail client or sending a message
is unnecessary. Keep route aliases, malformed query strings and external-returnTo
sanitization distinct from ordinary navigation. Inspect app secondary links whose
current tests only assert href before marking their click behavior covered.

### G03 — account request, confirmation and recovery controls

Inspection found no existing e2e references to `New sign-in email`, `Use a different
address`, or `Sending confirmation link`. The token-confirmation tests are a
different flow. Add request validation, held request disabling, success/change
address, service error/retry and expiry without false success. Deletion's existing
case already covers cancel/focus return, a 500 preserving the phrase and retry,
and server-confirmed deactivation. Add held-pending locking and 401 behavior using
synthetic accounts only. Validate absent/invalid purge dates on the result page.

### G04 — real error boundaries and connectivity transitions

Server failure fixtures often render inline recovery UI; they do not prove that
Next route `error.tsx` reset buttons work. Explicitly trigger a render/data-shape
failure in a controlled local fixture, assert the boundary's Try again/Reload
controls, then recover. Cover Home, discover, review/reviews, words, stories,
writing, progress/history and settings/account. Global-error is a separate root
500 presentation using NextError and has no custom reset control. Add actual
browser offline→online recovery where safe, distinguishing aborted navigation,
failed reads and ambiguous writes. Do not duplicate confirmed mutations on retry.

### G05 — additional browser engines

Run the existing behavior suite on Firefox and WebKit with pinned Playwright
binaries. Chromium with a mobile viewport is not Safari, and WebKit emulation is
not a physical iPhone. Resolve browser-specific failures based on product behavior,
not by weakening assertions. Audit project-name gates before changing project
names: some specs explicitly accept only the three existing names.

### G06 — accessibility and responsive dimensions

No pre-follow-up e2e references were found for reducedMotion, forcedColors,
setOffline, zoom, landscape or screenshot assertions. Add text enlargement,
reduced motion, high contrast where supported, landscape/short-height keyboard
reachability, focus visibility and no clipped controls. Axe critical/serious scans
do not establish all WCAG criteria, semantic reading order, spoken announcements
or human screen-reader usability. Physical software-keyboard/IME and actual
VoiceOver/TalkBack/NVDA checks remain separate evidence requirements.

### G07 — dynamic-state visual inspection

Capture and inspect both themes for each dynamic step/control state, not just each
route's first screen. Use `conversation-basics` for the existing guided-lesson
fixture and `a-quiet-lunch` for story detail; start sessions through their UI and
use returned IDs. Create a list/save a meaning before list/saved-detail visits.
Review loading/error/empty, long text, disabled/pending, wrong answer, completion,
expanded disclosures and retained drafts. A screenshot saved to disk is not a
visual pass until inspected. Baseline comparisons alone can preserve an existing
visual defect.

### G08 — frontend performance and memory beyond navigation

Prior native measurements covered 17 route families and Home/Journey/Progress
cycling; the four-route Lighthouse matrix is narrower. Extend to active lesson,
review, practice, story, search/list pagination, writing drafts and media playback.
Use repeated cycles, post-warm-up retained heap/DOM/listeners, and controlled
large fixtures; record aborted samples explicitly. Measure alternatives under the
same data/network/CPU conditions. Do not infer a frontend improvement from the
58× database service result. Inspect media/object URL/event-listener cleanup and
retained draft/session state independently.

### G09 — live integrations and devices

Real email delivery, Google completion, account export data, AI feedback quality,
actual audible media, CDN behavior, calendar import and physical mobile/browser
behavior need separate evidence. The prior deployed staging core-loop/maturity
passes establish those workflows, not all provider branches or devices. Use
approved synthetic staging identities and existing deployment test facilities;
never mutate the owner's learning or deactivate their account. Record exact
unavailable hardware/access as blocked evidence, not passing or omitted coverage.

## Closure rule

Each gap closes only when its control/action, precondition, expected state,
browser/layout/theme, fixture realism and fresh result are recorded. Keep failures,
skips, not-run and externally blocked items visible. No coverage percentage is
claimed because there is not yet a complete enumerated Cartesian set of every
control/input/state/device combination. Finite acceptance coverage should be
explicit and reviewable rather than described as every possible condition.

## Follow-up evidence — 6 November 2026

The ledger above preserves the initial gap inventory. The follow-up added
`auth-mobile-lifecycle`, `account-request-resilience`, `dynamic-controls-audit`,
`public-navigation`, `frontend-layout-matrix`, dynamic visual-state tests and
controlled route-boundary recovery tests. Source additions alone do not close a
gap; execution and manual-inspection status are recorded separately here.

### Added control assertions

| Gap | Concrete follow-up assertions | Remaining evidence |
| --- | --- | --- |
| G01 | Both-theme signup validation/show-hide/pending/error/retry; explicit verification consent and token cleanup/replay; successful shared-mock password login; reset success/replay; add-password email-only state | Included in the full Chromium run below; real credential/email durability remains G09 |
| G02 | Help keyboard destinations and support mailto contract without sending a message; list practice/find-words/member links | Included in full Chromium; subsequent calendar-anchor repair and residual controls now have passing evidence across the expanded and focused runs below |
| G03 | Held-pending email requests, failure/retry/address change/expiry; deactivation pending locking and expiry | Included in full Chromium; real email and identity-provider operations remain separate |
| G04 | Eleven controlled render-boundary routes in both themes: keyboard retry and fallback link must reach healthy headings; actual offline settings save preserves draft, reconnect confirms and reload retains it | Included in full Chromium; other offline writes/provider failures are not implied |
| G06 | 28 entry routes × two themes × three widths; reduced motion, 200% root-font reflow, short landscape, keyboard focus, forced-colors controls and synthetic composition events | Entry 168, accessibility-mode 18, calendar-anchor 18 and calendar-reflow six cases passed on the final bundle; forced-colors contrast and physical accessibility limits below |
| G07 | Sixteen dynamic states × both themes × three widths: all 96 viewport captures inspected, with selected fullpage companions; entry and rebuilt repair review below | Uncaptured state/content combinations and physical devices remain explicit |
| G08 | Eight same-context dynamic journeys at two widths; discovery workload discrimination and heap-retainer investigation | Results below; no long-session memory ceiling or alternate-implementation performance claim |

### Automated execution

The earlier integrated continuation run finished with **325 passed and 29 failed**. All
**168 entry-layout cases passed**: 28 routes × two themes × three widths (360,
430 and 1280 pixels), each with normal-text and 200% root-font screenshots,
reduced-motion emulation, an axe critical/serious scan at normal text size,
visible heading and document-overflow assertions. Root-font enlargement is not
browser zoom. These entry checks do not establish every dynamic state or every
WCAG criterion.

The failed integration cases prompted product corrections and fixture/readiness
repairs. The subsequent full Chromium run completed with **1711 passed, 41 skipped
and zero failures in 21.7 minutes**. It used the earlier production bundle and
preceded the latest Journey/Progress/reset corrections and navigation additions.
Skips are not counted as passes.

The rebuilt focused run then finished **212 passed, 1 failed** out of 213:
168 entry-layout cases, 30 secondary-navigation cases, nine Journey regression
cases (including three reset-copy cases), and six Progress regression cases.
The failure was the Plan calendar link reaching the URL without bringing its
streamed target into view. Separate repeated diagnostics reproduced the race;
a native-anchor alternative was worse and was rejected. The later target-mount
repair passed the final calendar/navigation checks recorded below.

The separate calendar text regression ran before its grid repair and finished
**2 failed, 4 passed**: both desktop themes reproduced clipped date/time at 200%
root font, while both mobile widths/themes passed keyboard and width assertions.
The production build containing the container-aware grid and target-mount
repairs then completed with exit 0. The final expanded Chromium run finished
**318 passed, 6 failed out of 324**. All six failures were the magic-link test's
unscoped text selector also matching Next's route announcer. Scoping the assertion
to the actual main alert corrected the selector without changing product behavior.
A focused run then finished **12 passed, zero failed**: those six magic cases plus
six newly added invalid-deactivation-date/sign-in recovery cases.

The calendar-fix bundle therefore has passing evidence across these runs for:

| Calendar-fix bundle coverage set | Passing cases |
| --- | ---: |
| Entry layout, exact expected heading, reflow and reduced motion | 168 |
| Secondary navigation | 30 |
| Residual navigation, including the six corrected magic selectors | 48 |
| Residual interactions: keyboard onboarding, malformed Journey bands, late search | 12 |
| Calendar fragment navigation | 18 |
| Calendar enlarged-date/time reflow | 6 |
| Existing calendar download/error behavior | 9 |
| Accessibility modes | 18 |
| Journey/reset-copy regression and Progress reflow regression | 9 + 6 |
| Added invalid-deactivation-date and sign-in recovery | 6 |

Lint, typechecks and calendar formatting checks also passed. These are separate
run records: **there was no single green 324-case run**. Do not sum historical
runs into a unique-test total or erase earlier reproduced failures.

The Lighthouse matrix completed **84/84 passed**, with zero document, heading
or score failures, on that calendar-fix build. It predates the later four loading
ARIA fixes and writing focus-scroll CSS change; final-build refresh remains
planned. Details and measurement limits are below.

### Firefox findings and corrected verification

The first full Firefox run completed **1,782 passed, 41 skipped, 64 failed** out
of 1,887 cases in **27.3 minutes**. This is a completed failing run, not pending
execution or a passing engine gate. Investigation separated product defects from
test precision and readiness problems:

- Four loading skeletons (Home, Progress, Settings and Account settings) named
  generic `div` elements with a prohibited `aria-label`. Settings also supplies
  Profile's loading boundary. New held-loading checks reproduced the actual axe
  violation on **all five tested routes** against the old bundle. The repair uses
  named sections with status text. Thirty checks cover five routes, two themes
  and three widths, requiring the skeleton to remain present throughout the axe
  scan. All **30 corrected held-loading checks passed** in **6.9 minutes**.
- At 430px, Firefox could leave about 6px of the focused writing meaning chooser
  underneath the sticky header. Both a CSS scroll-margin arm and a focus-handler
  arm corrected the reproduced geometry. The narrower CSS option was selected;
  it preserves native disclosure behavior without a new focus handler. This
  counterfactual does not establish a corrected integrated-suite pass.
- Other failures concerned test precision or readiness: exact selectors and
  assertions, refresh response-commit readiness, an image fixture nonce,
  keyboard activation of anchors, and off-focus axe scan positions. These
  corrections received the follow-up execution recorded below; classification
  alone did not waive a failure. Settled-page theme scans now require the
  expected healthy H1, while the separate held-loading checks retain transient
  accessibility coverage.

The corrected production build, lint and **252 helper checks** passed. The
**522-case Firefox rerun** across 28 specs, including 13 helper consumers, finished
**506 passed, 15 failed, 1 skipped**. All 15 failures came from an invalid custom
fixture marker: the stateless teaching fixture always reported the saved state
as false. Replacing that setup signal with completion of the exact relevant RSC
response preserved the actual behavior assertions. The exact 15-case recovery
run then passed **15/15 in 27.8 seconds**. Separately, the new held-loading group
passed **30/30** as recorded above.

All identified failures from the full Firefox run now have passing final
follow-ups. **There was no single green full Firefox run**; these are distinct
execution records, not a sum of unique passing scenarios. Real-provider and
physical-device evidence remain separate; no local browser fixture result closes
those dimensions.

### WebKit findings and corrected verification

The full WebKit baseline completed **1,717 passed, 159 failed, 41 skipped** out
of **1,917 cases in 35.8 minutes**. Failures span 32 specs; this is not a green
engine gate. The original failures and intentional skips remain part of the
record. A corrected **855-case run** is now executing across the union of failed
specs, all keyboard-helper consumers and added regression coverage. Its final
result is pending; intermediate passes do not close this engine gate.

Two additional product defects were reproduced independently:

- Journey's native range select caused document width **518px at a 360px
  viewport with 200% text**, despite its visible box fitting. Scoped paint
  containment removed the overflow; rejected alternatives blanked the selected
  value. A wrapping minimum flex basis also exposed the numeric range at enlarged
  text sizes. Source fixes and geometry/value regression checks are included in
  the rebuilt rerun; final compiled verification remains pending.
- Native calendar date/time controls could be keyboard-focused without matching
  WebKit's host `:focus-visible`, leaving the intended outer focus ring absent.
  The scoped fix uses a 3px focus outline and offset with the brighter primary
  color. Final computed-style, contrast and interaction verification are pending;
  the earlier counterfactual screenshots are not proof of the final compiled style.

A native-controls counterfactual separated WebKit keyboard policy from product
behavior. On this macOS host, ordinary Tab reached text/select controls while
Option-Tab reached buttons, links and checkboxes too. Reverse Option-Shift-Tab
traversal also succeeded. Process-only full-keyboard arguments still omitted
links and were rejected as an incomplete solution. No persistent preferences
were changed. The explicit shared `pressTab` helper is now applied: only macOS
WebKit uses its alternate native keys; other engines retain Tab/Shift-Tab.
It replaces 36 calls across 15 existing files, with **24 actual keyboard consumer
specs** identified for rerun. Counts, focus/order assertions and actual keyboard
activation remain intact; there is no global key monkeypatch or replacement of
traversal with programmatic focus.

Other corrections address evidence precision: core-loop onboarding now awaits
the application's successful navigation; focus-retention branches use genuine
keyboard activation rather than assuming Safari pointer clicks focus buttons;
Progress and word disclosures scope selectors to the live main instead of hidden
React streamed copies; Help's delayed loading fixture is being made deterministic;
writing interaction uses an actual click before measuring pointer-focus geometry.
Mobile WebKit viewport-prefetch checks use verified document scrolling because
that engine rejects mouse-wheel transport. These changes preserve behavior,
server-result and accessibility assertions and require their corrected runs.
A remaining native radio test assumption is separately diagnosed: WebKit does
not wrap ArrowDown from the final radio, while adjacent arrow navigation changes
focus and selection normally. The actual product uses native radio inputs.

WebKit visual review examined **534 original captures through contact sheets**
plus **12 native-size detail crops** across 1280/360/430, including 28 entry routes
in both themes at normal/200% text, available dynamic states, and accessibility
mode/profile captures. This found the range-select defect above. Saved-writing
captures in both themes/all widths and landscape profile captures were missing
because those cases did not reach capture; they are not counted as approved.
Long contact-sheet pages do not establish line-by-line readability, native picker
behavior, physical-device rendering or screen-reader output. Exact source image
manifests and the independent visual report remain in private audit artifacts.

The initial enlarged-text run reproduced a shared-header defect: at 1280 pixels
with 200% root font size, the desktop header expanded to 1532 pixels. The wrapping
header fix and subsequent page/card/auth reflow corrections are covered by the
168 passing rebuilt-bundle entry-layout cases. This is stronger evidence than the
initial live-CSS hypothesis checks on 13 desktop routes.

### Visual inspection

All **96 dynamic viewport captures** were opened and inspected: 16 states × two
themes × 360, 430 and 1280 pixels. They cover lesson teaching/question/wrong
answer, practice question/assisted/completion, review question/correct/completion,
story dialogue/question/wrong answer, list pending/confirmation, saved-word
detail and writing. Selected fullpage companions were also inspected, including
both saved-writing themes at all widths and the story question at 430 pixels.
Private manifests record exact files and hashes. No additional actionable defect
was identified in those inspected dynamic states.

The focused saved-writing editor, focus ring, counter and actions are unobscured
in the actual viewport captures. Fullpage images can stitch sticky headers at
the capture scroll position; that artifact does not establish live obstruction.
Story questions initially remain below accumulated dialogue; the 430px fullpage
companions establish that prompt and choices exist below it. At 360px, some
practice inputs and lesson/review choices require scrolling. These observations
do not assert automatic question scrolling or simultaneous visibility of all
content, and do not substitute for physical software-keyboard testing.

All **112 entry PNGs at each mobile width** were opened as normal/enlarged pairs
(28 routes × two themes × two font sizes at each of 360 and 430 pixels). Tall
sheets were reduced for overview, so this is not every-glyph proofreading. The
old captures exposed Progress internal overlap and literal escaped reset copy.
Their initial very tall enlarged Settings/Vocabulary rasters also repeated early
content, leaving lower portions unverified in that capture batch.

Rebuilt CSS-scale review now verifies Journey and Progress in **12 pairs across
all three widths and both themes** (24 source PNGs); the formerly colliding map
labels, stage rows and activity cards reflow coherently. Four desktop
Settings/Vocabulary pairs were reviewed. Twelve native CSS-pixel strips from
eight enlarged mobile Settings/Vocabulary source PNGs establish the lower page
content in both themes at 360/430, removing the earlier repeated-raster gap.
This follow-up did not separately re-review every upper mobile region or normal
mobile file. The three rebuilt reset-copy regression cases passed.

That review found a **real enlarged desktop calendar defect**: year and AM/PM
segments clipped in the narrow settings column despite no document overflow.
Both original images and native crops confirmed it. After repair, all **12 calendar
images** (viewport/card × both themes × three widths) were opened and inspected.
The actual viewports show the complete date, complete time including meridiem,
picker icons and unobscured keyboard-focus outlines. The desktop controls now
stack within their narrow container. Tall card images retain known stitched
fixed-header/skip-link artifacts; the actual viewport companions establish live
control visibility. Together with all six passing reflow cases, this closes that
bounded clipping regression for these Chromium layouts, not physical native
picker or calendar-app acceptance.

`frontend-accessibility-modes` additionally exercises 667×375 landscape,
keyboard focus, reduced motion, forced-colors selection/feedback and synthetic
composition events. In forced-colors emulation, axe's computed white-on-white
contrast disagreed with the visibly painted system palette. Only that contrast
rule is excluded in forced mode; the spec subsequently disables forced colors
and runs the complete axe scan in the requested theme. This is **not an automated
forced-colors contrast pass**, a browser-zoom test, an OS IME test or an assertion
about screen-reader speech. Those limits survive the automated pass.

### Repeated dynamic memory measurements

A production-build Chromium harness completed eight UI journeys at 360 and 1280
pixels: search, discovery, practice, story, lesson, list, writing and media. Each
journey used one context/page across two warm-up cycles and 15 measured cycles,
with exact readiness/action assertions and a document-identity guard. All 16
scenario/width runs passed, producing 272 post-GC checkpoints, including 240
measured after warm-up. No page errors occurred. Listener counts stayed constant,
active object URLs returned to zero, and DOM counts were stable or bounded by
visible state. Retained heap nevertheless rose; discovery had the largest
15-cycle change, approximately 0.98 MiB. These observations do not establish
leak-free operation or a long-session memory ceiling.

Three additional 60-cycle discovery workloads completed at both widths, each
with two warm-up cycles (372 checkpoints total):

| Workload | 360px retained heap, first → last measured | 1280px retained heap, first → last measured | Interpretation |
| --- | --- | --- | --- |
| Advance word and navigate away/back | 6.245 → 9.360 MiB | 6.208 → 9.356 MiB | Continued growth; no plateau within 60 cycles |
| Mount/navigate without advancing | 5.816 → 6.726 MiB | 5.821 → 6.725 MiB | Smaller growth with navigation alone |
| Advance in place | 3.923 → 4.330 MiB | 3.881 → 4.294 MiB | Smallest growth under this workload |

These compare workloads, not interchangeable product implementations. DOM and
listener counts did not explain the discovery growth. A further 360px run with
warm-up/end heap snapshots found 60 new discovery map objects retained through
resolved React Server Component promises by Next's client back/forward cache.
The exact retaining-entry properties and rootward module references match the
installed framework cache implementation. No application-owned media/listener
leak was identified by this investigation.

Installed Next cache expiry is lazy, and its back/forward entries deliberately
ignore stale time. Its nominal 50 MiB LRU uses a 100-byte estimate for each such
entry; this is not a guaranteed renderer-memory cap. A five-minute idle wait would
therefore not establish release. No speculative cache clearing or framework patch
was applied. A future smaller-payload implementation would require equal-workload
measurement plus verified back/forward restoration before selection. Private
heap artifacts contain synthetic state and are excluded from public commits.

Media measurements exercised actual decoding of a local MP3 fixture and object
URL cleanup. They do not establish pronunciation quality, real CDN delivery,
hardware audibility or every provider branch. These remain G09 evidence items.

### External evidence currently blocked

Real email delivery remains unproven: the inspected staging configuration reports
password/magic-link capability disabled, and connected inbox access failed at the
connector transport. Authorization to send test messages does not establish that
one was sent or delivered. Keep this evidence blocked until capability and inbox
verification are available; do not count fixture email UI passes as delivery.

## Secondary-navigation source audit

`secondary-navigation-controls.spec.ts` adds **25 distinct controls**, grouped
into five source-route cases in both themes and all three configured widths
(30 test cases). Every case activates the real link with Enter, checks its exact
URL and requires the expected healthy level-one heading. The Plan reminder link
also requires the Calendar reminder heading in the viewport. Synthetic UUID
sessions isolate the tests; navigation does not mutate learner data. The E2E
TypeScript check passed. The rebuilt 30-case run had 29 passes and one failure:
the Plan calendar target was outside the viewport. After the target-mount repair,
all 30 passed in the final expanded run; the 18 dedicated fragment cases also
passed with delayed loading, direct-entry and viewport-hit-testing assertions.

| Source | Newly activated controls |
| --- | --- |
| Home / Explore your English and empty latest word | Short stories; Personal lists; Search for a word |
| Courses | Find a word; Saved words; Choose a practice; Short stories; Personal lists; Your learning plan |
| Plan | Change daily preferences; Set a calendar reminder; View reviews; Find your starting words; Open your vocabulary map; Choose a practice |
| Progress | Saved words; Open your word map; zero-known filter under About these stages; Browse courses under Your learning path; Write your first sentence and View sentence history under Sentence practice |
| Settings | Back to Home; Account security; Profile; Help and support |

Home Practice and Situations already reach healthy destinations in
`navigation-prefetch`; Your plan is activated through outage/recovery in
`auth-check-recovery`. They are not duplicated. The Courses recommendation's
catalogue anchor already activates by keyboard in `journey-practice-coherence`.
The self-check completion-to-Plan link already activates in `learning-plan`.
These are retained evidence, not new test claims.

The earlier vague Verify entries resolved to the following concrete residuals
(before the source additions in the next section):

- Magic-link result: successful returnTo consumption and incomplete-link
  accessibility exist. The Back to sign in link's preserved returnTo and Back to
  home link still need activation in incomplete/rejected-consume states. Alias
  rendering is covered; real delivery remains separately blocked.
- Onboarding: both-theme native-language search/selection, Back restoration and
  persisted completion exist. Only the language choice uses keyboard activation
  within that journey; a wholly keyboard-operated Continue/Back/Finish sequence
  and actual mobile IME/software keyboard are not established.
- Journey map: keyboard arrows/Enter, phone chooser, legacy redirect, next-range
  and return-anchor viewport assertions exist. Direct Journey inputs such as
  duplicate, negative, nonnumeric and above-maximum `band` still need explicit
  normalization assertions; Discovery's header-query tests are a different path.
- All 17 situation guides already open phrase/example disclosures and verify
  canonical/writing hrefs. One guide activates its story and meaning anchor.
  Writing and guided-lesson link activation remain missing, along with visual
  inspection of the long expanded guides; the gap is not absence of all 17 tests.
- Vocabulary search already covers clear, back and query/result restoration.
  A delayed older navigation resolving after a newer search or history action
  still lacks a focused assertion. The 15-cycle memory soak is not that race test.
- Conditional secondary navigation remains distinct from the default states:
  Courses empty-catalog Go to Reviews; Progress due Start review, saved lesson
  return, empty activity Go to today's mission and failed-writing history link;
  Home All your words and Write a sentence with it links (the latest-word
  headword already activates in `learner-workspace`). These require the
  matching shared-mock state before activation.
- Other precise return-link gaps in the 43-route ledger include saved-word
  not-found recovery, sentence-history empty Go to Home, and Profile's Back to
  Settings / Account security / Learning preferences. Existing page rendering,
  href, persistence or recovery assertions do not alone close these activations.

This residual list is based on the inspected assertions, not a blanket claim
that every unlisted state is covered. Dynamic visuals, large datasets, engine
coverage and live integrations retain their separately recorded evidence limits.

## Residual-navigation implementation

`residual-navigation-controls.spec.ts` now adds the remaining feasible navigation
controls identified above: **16 cases across both themes per project, 48 across
three widths**. Every activation checks the intended URL and a healthy destination
heading. The Home direct writing anchor must also reveal its textbox and open the
existing disclosure without the test opening it manually. Guide writing must show
the actual Daily Conversation topic/prompt. E2E TypeScript passed. All 48 cases
have passing final-bundle evidence across the expanded run and the six-case
magic-selector rerun; these are not represented as one uninterrupted green run.

| Controls | Required synthetic setup |
| --- | --- |
| Magic incomplete/rejected Back to sign in / Back to home | Local UUID session; rejected consume intercepted as 401 because shared mock always succeeds; returnTo preserved; no real token/email delivery |
| Guide writing / Start guided lesson | Existing daily-conversation situation and canonical lesson; no fabricated DOM |
| Home All your words / Write a sentence with it | Save canonical invite through shared mock API in the test's UUID session |
| Saved-word not-found / empty-history returns | Random missing saved-word ID; fresh UUID session with no learner sentences |
| Profile Back to Settings / Account security / Learning preferences | Existing synthetic profile |
| Progress Start review / in-progress lesson | Save invite and start actual shared-mock conversation-basics lesson session |
| Progress empty activity / failed-writing recovery | New explicit empty completionHistory hook; existing error sentence fixture cleared before activating recovery |
| Courses Go to Reviews with no catalogue | Existing empty lessons plus new explicit empty journey-situations hook |

The two narrowly scoped mock hooks are `e2e_progress_history_fixture=empty` and
`e2e_journey_situations=empty`; only their matching read response changes. Mock
syntax validation passed. A fresh shared mock is required for these new fixtures.
The executed controls above supersede their earlier residual labels. The separate
12-case interaction audit also passed keyboard-only onboarding, malformed direct
Journey query inputs and a late older search. Real devices/providers remain
separate evidence dimensions.

## Lighthouse runner verification

Independent review found and corrected two evidence-runner defects: empty screen
selection could report zero-audit success, and Lighthouse closed its internally
created page before post-audit heading inspection. Full mode now rejects an empty
filter and supplies a caller-owned Puppeteer Page, checks the visible expected H1
in the exact audited document, and closes it afterward.

A fresh installed-Chrome smoke run completed **9/9 audits** (Home, Sign in and
Discovery × 360/430/1280), **zero failed categories**, exit 0 in **47.744 seconds**.
Original thresholds stayed performance ≥85, accessibility ≥95 and best practices
≥90. Page-evidence foundation checks passed 2/2. This initially established the diagnostic subset and runner lifecycle.

The subsequent calendar-fix production-build run completed **84/84 audits**:
all 28 configured screens × all three widths, with **zero failed categories**,
exit 0 in **436.561 seconds (7 minutes 16.6 seconds)**. Exact final URL, successful
document status and a visible screen-specific H1 were checked in each actual
audited page before reporting scores. All 84 report timestamps were verified
against this run's start; none were stale diagnostic files.

| Measure | Observed range / maximum |
| --- | --- |
| Performance score | 100–100 |
| Accessibility score | 100–100 |
| Best-practices score | 96–100 |
| Largest contentful paint | Maximum 563.8 ms, Journey at 430px |
| Total blocking time | Maximum 0 ms |
| Cumulative layout shift | Maximum 0 |

The original thresholds and Lighthouse simulated throttling/storage reset were
unchanged. These are local production-build entry measurements with synthetic
API data, not production network, provider, large-library or physical-device
benchmarks. Accessibility scores do not establish all WCAG or screen-reader
criteria. Private logs and the fresh-report summary preserve every screen/layout
result. This run predates the subsequent loading-ARIA and writing focus-scroll
repairs; a Lighthouse refresh on their final build remains planned.

## Current closure status and acceptance queue

- **G02 residual controls:** `residual-navigation-controls` contains 48 cases;
  `residual-interaction-controls` adds 12 covering keyboard-only onboarding in
  both themes, malformed/duplicate/out-of-range Journey bands, and a held older
  search resolving after the newer search. All 60 have passing evidence across
  the final expanded and focused runs above. The interaction tests do not
  establish physical IME or every search/history race.
- **Closed public recovery branch:** account-deactivated valid-date
  rendering is exercised by the deletion journey and no-date layout is captured.
  `public-navigation` now adds invalid `purgeAfter` suppression and actual Tab /
  Enter activation of Go to sign in, requiring the healthy login page and theme
  preservation (two themes × three widths). All six passed in the focused run;
  this is actual action/outcome evidence, not inference from the layout pass.
- **G06 calendar closed for tested Chromium layouts:** production build, six
  enlarged-text cases, nine existing calendar cases and 18 fragment cases passed.
  All 12 viewport/card images were visually reviewed. The reproduced desktop
  clipping and streamed-target arrival regressions are verified repaired within
  that scope; engine/device and actual calendar-import evidence remain separate.
- **G05 engines and G08 load performance:** full Firefox completed with
  1,782 passed / 41 skipped / 64 failed. Its identified failures have passing
  final follow-ups: the affected rerun was 506 passed / 15 failed / 1 skipped,
  the exact 15 corrected fixture cases then passed, and all 30 held-loading cases
  passed. This does not constitute one green full Firefox run. Full WebKit was
  1,717 passed / 159 failed / 41 skipped; its corrected 855-case run is executing,
  with no final passing claim. The earlier 84-entry Lighthouse matrix passed;
  refresh after the latest fixes is
  planned. Existing memory measurements compare workloads; they are not proof
  that a new implementation is faster or that retained memory has a long-session
  ceiling.
- **G09 actual providers:** staging run 42194703269 passed core-loop/maturity
  against real staging, including six pronunciation cases with browser decoding
  and second-request cache checks. First calls may already be cached; there was
  no human auditory assessment. Its sentence smoke accepted a verdict **or an
  error alert**, so it does not establish successful AI evaluation. It did not
  verify browser export or current CDN image body/decode.
- **G09 new acceptance, not yet executed:** `provider-data-acceptance.spec.ts`
  reuses the CI-minted reserved synthetic identity after server identity checks.
  It requires numeric GitHub run/attempt identifiers and includes their unique
  pair in the authored sentence so a later deployment does not reuse an earlier
  stored feedback result. It permits one provider submission, requires a completed
  correct verdict and correct target-word use, downloads and validates the actual
  learner export against public schema/that test record, then deletes the file.
  Its CDN case requires an approved canonical asset and explicitly skips if none
  is configured. The opt-in staging step still needs execution; a provider error
  must fail this acceptance rather than be counted as success. One labelled
  sentence would establish only bounded semantic smoke, not broad AI quality.
- **Externally unavailable:** real email remains blocked as described above;
  Google provider completion, calendar-app import, hardware audio, physical
  mobile keyboards and actual VoiceOver/TalkBack/NVDA speech are not established
  by mock journeys, screenshot review, synthetic composition or axe checks.

The route count remains **43**: exact `page.tsx` routes, including aliases and
dynamic patterns, with the private daily-lesson component excluded. The 28 entry
routes and 16 dynamic states are different coverage sets; neither count replaces
that route inventory or implies all state combinations are visually inspected.


## Final integration and isolated acceptance

The corrected WebKit follow-up finished **829 passed, 23 failed, 3 skipped**
out of 855 cases in **25 minutes**. The original full-run and this follow-up
remain separate records. Three calendar loading-fixture failures and three
native radio wraparound assumptions prompted narrower fixture/input corrections.
Other failures included pre-document browser transport stalls and late test
actions under severe host load; they require fresh execution, not relaxed
deadlines or a claim of application performance degradation.

Native WebKit does not wrap radio ArrowDown from last to first. The corrected
check verifies focus and selection at every adjacent ArrowUp step and then
ArrowDown, retaining theme, axe and geometry assertions. A bounded mock delay
now targets only GET Settings data so identity/layout reads can release the
loading shell. The actual loading, destination and calendar visibility checks
remain required.

Compiled visual review also caught the 360px enlarged calendar input partly
behind fixed navigation despite a passing center-point check. A controlled
comparison found six-rem scroll margin clears the entire input and its ring;
nearest scrolling did not, and a center-scrolling handler was unnecessary.
The final CSS adds no handler or listener. The regression now checks the whole
outline against fixed/sticky navigation and interior hit points. This final
refinement requires its own compiled verification.

A protected link's focus-prefetch timer was observed issuing redundant tagged
requests after actual navigation began. The final implementation cancels that
timer only for accepted navigation. A new regression records the actual
navigation, rejects a pre-existing prefetch, waits for the confirmed healthy
destination and requires no redundant tagged request or write. The final
request-count outcome must come from execution; source alone is not a speed
measurement.

The final local red attempt was interrupted under continuing host overload;
it is not an aggregate red/green result. Integrated source checks passed web
lint, web and E2E typechecks, formatting and **252 web helper tests**. The latest
production build also completed. The complete inventory is **1,926 cases in
101 files** across three widths; intended skips are never counted as passes.

Final broad acceptance runs on isolated CI workers: the original Chromium
check plus complete Firefox and WebKit jobs, and all 84 Lighthouse entry
audits with unchanged budgets and exact document/heading checks. Passing and
failing Lighthouse reports are retained for seven days. The consolidated PR
stays held until relevant checks and review are resolved. Actual PR and staging
execution records establish the final outcomes; this ledger does not imply
those jobs have passed before they execute. The reserved-account staging
provider/export/CDN gate and the external limitations above remain separate.
