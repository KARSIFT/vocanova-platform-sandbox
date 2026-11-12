# UI/UX audit — 12 November 2026

## Direction

Keep Vocanova's calm learning workspace, practical English, existing tokens and
Home / Journey / Progress navigation. Improve continuity, clear next actions,
meaning-specific navigation and honest recovery before adding visual decoration.
Automated inspection supports these decisions; it does not establish how readily
A2–B1 learners understand the product.

## Recent primary sources

The requested research window is **12 October–12 November 2026**, inclusive.
Publication dates were checked on the articles; GitHub release timestamps were
also checked through its API. Release dates describe those updates, not the age
of their projects. No external skill or executable was installed.

| Published  | Source                                                                                                                                                            | Application to Vocanova                                                                                                                                                                 |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 30 October | [NN/G: The Role of UX in Agentic Software Factories](https://www.nngroup.com/articles/ux-agentic-software-factories/)                                             | Give coding agents product constraints and evaluate actual tasks before and after release. A generated interface is insufficient acceptance evidence.                                   |
| 27 October | [Automattic Design: Working towards higher-quality agentic experiences](https://automattic.design/2026/10/27/working-towards-higher-quality-agentic-experiences/) | Make intent, state and recovery legible; translate useful principles into reusable components and concrete project guidance. Its small pilot is practice evidence, not universal proof. |
| 26 October | [Paul Bakaus: Impeccable Skill 4.8.0](https://github.com/pbakaus/impeccable/releases/tag/skill-v4.8.0)                                                            | Scope findings to actual rendered surfaces and distinguish actionable failures from aesthetic advisories. Avoid treating automated style rules as product requirements.                 |
| 23 October | [NN/G: AI in UX survey](https://www.nngroup.com/articles/ai-in-ux-survey-26/)                                                                                     | Reported speed gains do not establish UX quality. Retain independent review and behavioral checks.                                                                                      |
| 16 October | [NN/G: Synthetic Users vs. Human Respondents](https://www.nngroup.com/articles/synthetic-users-vs-human-respondents/)                                             | Synthetic responses can underrepresent variation. Browser fixtures and AI reviews generate evidence about implementation, not representative learner research.                          |
| 7 November | [Leonxlnx: Taste Skill 2.5.1](https://github.com/Leonxlnx/taste-skill/releases/tag/v2.5.1)                                                                        | Check mobile collisions, accessible labels, focus and touch targets. The author's claims are self-reported; mandatory animation and photography do not fit every learning screen.       |

## Inspection scope

| Area                          | What was reviewed                                                                                                                                   |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Public introduction and entry | Value proposition, sign-in capabilities, continuation, onboarding, password and email confirmation, connection/deactivation states                  |
| Daily learning                | Home hierarchy and goals; Journey map; Discovery selection, details and decisions; saved words, search, word details, self-check and personal lists |
| Practice                      | Review entry/session/completion, guided lessons, practice sessions, story library/player and writing feedback                                       |
| Progress                      | Knowledge versus mastery, review stages, history, goals and rewards                                                                                 |
| Account and support           | Learning preferences, appearance, profile, reminders, export, email change, deactivation and Help                                                   |
| Shared experience             | Three destinations, active state, keyboard/focus, loading/recovery, responsive layout, light/dark colors and enlarged text                          |

Source inspection covered these families. Read-only live mobile observation
covered the landing, Home, Help, Word search, Journey, Progress and story library.
The live account's progress and preferences were not changed. Mutations and
failure paths use isolated local fixtures with browser audio muted.

## Implemented improvements

| Priority | Observed problem                                                                                        | Result                                                                                                                                                                                                       |
| -------- | ------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| P2       | Story Continue focused the title above the accumulating dialogue.                                       | Focus follows the new line or question.                                                                                                                                                                      |
| P2       | Lesson/story errors and unchanged saved-progress recovery left focus away from the task.                | Recovery alerts receive focus; successful recovery returns it to the current task.                                                                                                                           |
| P2       | Search, knowledge-map tiles, lists and lesson summaries identify a meaning but linked only to the word. | Links target the selected meaning's existing anchor, including later senses.                                                                                                                                 |
| P2       | Expired sign-in discarded unfinished setup and Settings changes.                                        | Validated, same-user drafts survive full navigation in the same tab for up to two hours. Confirmed saves, logout and deactivation clear resolved drafts. Pending edits and account boundaries are protected. |
| P2       | Connection errors and outages were described as invalid confirmation links.                             | Recoverable failures retain proof and offer a focused retry. Definitive invalid proof keeps the existing replacement/sign-in paths.                                                                          |
| P2       | Reporting feedback could never recover a missing CSRF cookie.                                           | Reports use the existing session-verified token recovery and retain stale-result guards.                                                                                                                     |
| P2       | AI limitations appeared after first submission or in Help.                                              | A short disclosure appears before submission. Completed feedback gives revision the primary visual action.                                                                                                   |
| P2       | Help's map link opened Word search and its Home guidance was stale.                                     | Help opens Journey's word map and describes review/discovery/practice.                                                                                                                                       |
| P3       | Long onboarding answers and pending email addresses could overflow.                                     | Summaries and addresses wrap; labels use simpler learner language.                                                                                                                                           |
| P3       | The deactivation action claimed immediate permanence despite a recovery period.                         | The button describes the immediate deactivation action; existing removal/recovery explanation remains.                                                                                                       |

## Remaining findings and product decisions

- **P2, content quality:** live search includes dictionary-style definitions such
  as “forsake, leave behind” and long technical descriptions. Audit definition
  readability and polysemy against the intended A2–B1 learner, with editorial
  review before catalog changes. A successful vocabulary lookup does not prove
  that the explanation is understandable.
- **P2, email recovery:** password signup/reset success states lack the resend
  and change-address controls already offered by magic-link entry. Reuse that
  established pattern while retaining non-enumerating responses and cooldowns.
- **P2, draft coverage:** standalone Profile and email-change forms still require
  separate decisions about preserving unsaved edits across reauthentication.
  This delivery preserves onboarding and learning preferences only.
- **P3, name consistency:** Profile/signup native UTF-16 `maxLength` differs from
  Settings/API code-point validation. Use one rule and test supplementary Unicode.
- **P3, hierarchy:** list management precedes learning content; an in-progress
  story presents Continue alongside Read and practise. Test whether learners
  prefer a compact management disclosure and a clearer single continuation.
- **P3, empty catalog:** Topic writing needs a dedicated empty state if the API
  supports a successful zero-topic response. Confirm that contract first.

Run a small observed A2–B1 usability study for finding a word's second meaning,
resuming a story, recovering from an interrupted save and explaining progress.
Measure completion, wrong turns, recovery and comprehension. Do not substitute
synthetic users for these observations or invent a usability score.

## Verification

Local acceptance passed **312 browser cases with six layout-specific skips** and
**294 unit tests**. A separate entry-page sweep passed **168 combinations**:
28 surfaces × three widths × two themes, with the expected heading, no horizontal
overflow, and no critical/serious WCAG-tagged axe findings. This sweep covers entry
states; the focused regressions cover the changed interactive paths. App and
browser TypeScript, web ESLint, token checks and the production build passed.

Delivery evidence is recorded with the consolidated PR. Targeted regressions
first reproduced focus, navigation, recovery, draft loss and long-address
overflow against earlier bundles. Checks include 360px, 430px and desktop,
keyboard focus, both themes, enlarged text, transient errors and confirmed saves.

Local API fixtures do not prove live email delivery, OAuth behavior, AI feedback
quality or database durability. Lighthouse laboratory results do not establish
production field performance. Rollout status must be stated separately from
local implementation acceptance.
