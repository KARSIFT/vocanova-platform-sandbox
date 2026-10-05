# Learner experience improvement

Integrated delivery, 5 October 2026. This follows the accepted picture phase and
the owner's request for a candid product critique and substantial improvements.
It is not a production release or a measured learner-effectiveness claim.

## Design judgment and learner problems

The existing product has useful lessons, saved vocabulary, spaced review,
typed/listening practice, stories, guides, private lists and sentence feedback.
The weakness is how those capabilities fit together. The current experience
feels like several tools, with repeated cards, competing actions and too much
navigation between learning and writing. Technical acceptance of the earlier
phase does not establish excellence, style or ease for learners.

The review identified these observable improvements:

- Home should make the next daily action clear and expose sentence production.
- Journey should present each situation once, with lessons and related practice.
- Sessions should contain the task, an exit and predictable answer controls.
- Selecting an answer should not submit it before the learner chooses Check.
- Writing should allow explicit saving without leaving the selected scenario.
- Corrections should show the original and suggestion together near rewriting.
- Progress should lead with vocabulary, lessons and writing before rewards.
- The public introduction should demonstrate original picture-led teaching and
  an explicitly authored feedback example, without fictitious personal state.

These are design judgments based on source and rendered screens, not a user
study or a proficiency score. The prior reference recordings remain important
role models; original competitor media is not used in the implementation.

## Coherent visual direction

Retain the existing palette roles: neutral canvas #f5f5f5, white surfaces,
reading text #171717, blue action #1d4ed8 and lavender context #ede9fe, mapped
through existing generated tokens and their dark equivalents. Keep the system
sans-serif family and sentence case. Use compact headings and comfortable
left-aligned reading widths. One strong daily-action panel leads Home; flatter
rows and quiet utility sections provide contrast. Images explain words, not
decorate every section. Preserve their uncropped 3:2 frames and varied media.

Home, Journey and Progress remain the primary destinations. A focused lesson,
practice or story session has one explicit exit and a stable action rail instead
of account controls and primary navigation. Content reserves space for the rail.
Motion is not needed to explain these improvements.

## Behavior boundaries

Recommendations, progress, rewards, saving, first-attempt results and feedback
remain server-confirmed. Saved, known, list membership and review stage remain
distinct. Practice may prefer a completed or sufficiently taught lesson; merely
opening a lesson is not evidence of studying it. Explicit vocabulary choices
remain available. No historical session snapshot is rewritten for presentation.

The public feedback example is authored demonstration content, not live AI.
Pictures remain outside graded questions. Original drafts and exact uncertain
mutation retries remain protected. No provider activation, production deployment,
social competition or proficiency certification is part of this change.

## Acceptance evidence

Fresh root verification on 5 October:

- Web/API production build, formatting, lint/Go vet and types pass. The unit
  suites pass: 258 foundation, 52 client and 146 web helper tests. Full local
  `pnpm validate` stops at three existing Docker-backed PostgreSQL tests because
  the local Docker engine is unavailable; a full validation pass is not claimed.
- The broad browser matrix recorded 888 passes, 37 existing scope skips and two
  stale Home-label assertions. Those assertions now check the actual textual
  due count rather than the removed generic label. After the final build, the
  affected matrix passes **149 cases with one existing scope skip**, across
  360px, 430px and 1280px, including both themes, keyboard checks, accessibility,
  exact retries, missing/mismatched save responses, draft isolation, recent
  sentence errors/long text and focused sessions. Required hosted accessibility
  runs the complete matrix on the final PR revision.
- Fresh viewport and full-page screenshots were inspected for Home, Journey,
  Progress and teaching at 360px/1280px in both themes. Browser cases also
  capture writing and 430px layouts. Independent cross-flow review found a
  broken writing fragment shortcut; it was reproduced, corrected and verified
  by clicking through to the visible writing heading. No remaining actionable
  source finding was reported by the independent final review.
- The actual automated review on [PR #1487](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1487#issuecomment-5998071251)
  prompted memoizing the bounded comparison during typing, restoring the
  disabled-button cursor and directing Journey's browsing shortcut to its actual
  catalogue. A fresh build and **84 browser cases** pass after those refinements,
  including keyboard recovery when lessons/recommendations are unavailable.
- The authorized local Jev browser adapter read the local public preview and
  activated its visible example-feedback control at 360px without horizontal
  overflow. Native screenshot capture timed out, so it is not claimed as native
  image evidence. Rendered screenshot evidence comes from the browser harness.

The design used the repository skills and the official
[Anthropic frontend-design skill](https://github.com/anthropics/skills/blob/main/skills/frontend-design/SKILL.md)
and [Vercel web-design-guidelines](https://github.com/vercel-labs/agent-skills/blob/main/skills/web-design-guidelines/SKILL.md).
The original WordUp/Duolingo review informs the connected loop, focused sessions
and visible corrections. This is one consolidated delivery: open draft with
hold, require actual automated review and all applicable hosted checks, then use
the normal merge queue. [PR #1487](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1487)
is the consolidated delivery; its check record establishes hosted acceptance;
production promotion remains separate.

Fresh live staging inspection observed a transient Practice-page internal error;
a subsequent authenticated read returned HTTP 200 and reload rendered normally.
No server diagnostic established its cause, and no speculative contract repair
is claimed. Synthetic browser checks do not establish live feedback quality,
physical audio, physical mobile keyboard behavior or learner effectiveness.

This improves the interface and its verified behavior. Real learner usability,
feedback quality and learning outcomes remain separate evidence to collect.
