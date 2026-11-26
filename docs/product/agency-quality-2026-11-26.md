# Product-wide learner flow review

The review started from release 0.4.40, commit `33b9bd35`. Claude Fable 5.2 planned the audit and fixes. Claude Opus 5.5 implemented the changes in isolated source snapshots. Codex agents inspected the routes, wrote regression tests, and reviewed the integrated changes. The audit uses existing Vocanova controls, tokens, and navigation. Its checks follow the Agency Agents practices of clear next actions, actionable failure states, exact handoffs, and keyboard and mobile verification. [Agency Agents](https://github.com/msitarzewski/agency-agents)

## Route coverage

The 44 page patterns include redirects, confirmation pages, and the gated component gallery. The audit followed the normal destinations and inspected the corresponding feature tests for conditional states.

| Area | Page patterns | Flow checked |
| --- | --- | --- |
| Public entry | `/`, `/help`, `/connection-error`, `/signin`, `/login`, `/signup`, `/magic-link` | Landing actions, help and recovery, available sign-in methods, and destination preservation |
| Authentication and setup | `/auth/magic`, `/auth/password/reset`, `/auth/password/verify`, `/auth/email-change`, `/onboarding`, `/account-deactivated` | Link outcomes, account confirmation, setup completion, and support after account closure |
| Account | `/settings`, `/settings/profile`, `/settings/account` | Preferences, profile drafts, email capability, export, and deletion confirmation |
| Daily learning | `/home`, `/discover`, `/knowledge-map`, `/discovery`, `/plan` | Recommended next action, Journey and its alias, range and word selection, and plan preferences |
| Situations and lessons | `/discover/[situation]`, `/discover/[situation]/[word]`, `/courses`, `/learn/[lessonKey]` | Exact meanings, guidebooks, course recommendations, teaching, questions, and lesson completion |
| Practice and stories | `/practice`, `/practice/session/[sessionId]`, `/stories`, `/stories/[storyKey]`, `/stories/session/[sessionId]` | Vocabulary source, listening and typed recall, sentence assembly, story continuation, and related listening |
| Review and Mistakes | `/reviews`, `/review`, `/review/session` | Due queues, assisted answers, feedback, Mistakes, daily target, and optional review |
| Vocabulary and lists | `/vocabulary`, `/vocabulary/[word]`, `/vocabulary/check`, `/words`, `/words/[userWordId]`, `/lists`, `/lists/[listId]` | Search and map handoff, saving, self-check, private notes, list membership, and linked practice |
| Writing and progress | `/writing`, `/progress`, `/progress/sentences` | Save to write, feedback and retries, recorded progress, and history pagination |
| Development preview | `/component-gallery` | Shared controls and the production gate |

## Findings and changes

| Area | Observed behavior | Correction |
| --- | --- | --- |
| Discovery | A filter with no matches left the learner to clear multiple controls | Show all restores results, keeps the view, and focuses search |
| Sign-in and setup | Google sign-in lost newer learning destinations. Missing CSRF cookies could strand valid sessions. Expired reset links left a permanent failure on Save | Preserve safe learning destinations, recover the browser cookie, and focus a fresh reset action |
| Accounts | Sign-in could not restore an account as described. A deployment with no mail sender offered confirmation links | Explain immediate closure and scheduled anonymization. Use the actual sender capability for link requests and Account settings |
| Practice | Partial sentences were rejected and could lock editing. A failed specific lesson could silently become the whole course | Require a complete sentence before checking. Keep the chosen lesson visible and let the learner refresh or choose a different source |
| Home and session start | An optional review could disappear after the final checked answer. A saved-word read could block Home. Confirmed sessions had no link when client navigation stalled | Resume the exact review, localize the saved-word error, and expose a document link to each confirmed session |
| Vocabulary | Filtered and paginated state was lost after viewing a word. Situation links could open a different meaning from the selected row | Preserve the originating search state and exact meaning anchors |
| Saving and writing | An uncertain save could be retried with a new request identity. A mismatched receipt could show Saved. The writing action lost keyboard focus while saving | Replay the same request, verify the meaning in the receipt, and keep the writing action focused through retry |
| Sentence history | A rejected page cursor sent Retry to the same invalid page | Offer a direct return to the newest sentences |

The browser fixtures use synthetic learners and simulate missing content, stale navigation, and interrupted requests. Existing live smoke workflows record deployment identity and service health separately. Manual screen-reader behavior and learner outcomes need their own evidence.

## Local verification

- The focused acceptance passed 93 cases each in Firefox and WebKit at 360px, 430px, and 1280px. The wider Chromium run passed 772 cases with 10 configured skips. An existing Discovery reload case timed out once in that parallel run and passed three isolated reruns.
- Formatting, lint, Playwright typechecking, 369 web unit tests, 64 API client tests, and the Go suite passed. The generated API contract matched the generator. The web production build passed on retry after the existing Next.js security reminder.
- The local workspace validation stopped at 11 foundation failures. Every failing case also failed in a clean checkout of the starting `main`, using the same pinned Node and separate installed workspace dependencies.
