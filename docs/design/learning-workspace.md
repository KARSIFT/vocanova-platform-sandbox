# VocaNova learning workspace

The visual direction is a calm, welcoming place to practise practical English.
The learner should recognise the same product from the public introduction through
sign-in, daily practice, and account settings. Real words and situations give the
interface its character; progress and next actions give it structure.

## Visual direction

Keep the established blue identity with a restrained palette: white surfaces
(`#ffffff`), a light neutral canvas (`#f5f5f5`), dark reading text (`#171717`),
blue actions (`#1d4ed8`), and soft lavender context (`#ede9fe`). These are roles
in the existing generated palette, not a second token source. Dark mode maps the
same roles to the existing dark surfaces and readable foregrounds.

Use a system sans-serif stack, a compact type scale, comfortable line height,
and left-aligned reading content. Headlines should introduce the page without
displacing its primary action. Large numbers belong to real progress, not
decorative marketing statistics. Borders group related information; shadows
and rounded corners should be restrained.

## Layout and interaction

- The public page shows vocabulary in a recognisable situation and explains the
  discover → remember → practise loop. Its learning action is easy to find.
- Authentication shares the brand and visual language of the learning workspace.
  Only configured sign-in methods appear. Visual changes do not activate password
  registration; its verification and recovery email requirements still apply.
- Home keeps the backend-selected next action prominent. Secondary vocabulary
  and practice information can use additional desktop space without competing
  with the daily mission.
- Home, Journey, and Progress remain the three primary destinations. Mobile uses
  thumb-reachable bottom navigation; desktop places navigation beside or above
  the workspace. There is one accessible primary navigation at each size.
- Journey presents situations as places to use language. Word content remains
  readable, and saving or review outcomes remain backend-authoritative.
- Settings groups learning preferences and makes profile, account security, and
  appearance easy to find. Light, Dark, and System remain device preferences.

The layout adapts to the task: a broader overview for Home and Journey, a narrower
reading width for word detail and focused practice. Do not stretch every form
merely because a larger screen is available.

## Learning flow refinements

Keep daily goals distinct from the review queue available now. The mission's
backend-selected action remains authoritative; presentation must not imply that
more words are due just because the learner has a larger daily target.

Review completion should lead naturally into optional sentence practice, with a
clear route back home. Keep the word and its meaning separate from the writing
instruction. Put the writing area and actionable feedback ahead of secondary
draft-storage details, while retaining the privacy guidance and draft controls.

Journey labels use readable categories and level ranges, such as A2–B1. Topic
icons should identify the situation rather than repeat a generic work icon.
Progress shows dated, chronological records returned by the API. Missing dates
are not evidence of rest days, and a protected streak day is not evidence of a
completed mission. Explain the points balance as rewards for learning activity,
not a language proficiency score. Unavailable settings should not appear as new
choices; previously stored values must survive unrelated edits.

## Working with design skills

Use [Anthropic's frontend-design](https://github.com/anthropics/skills/tree/main/skills/frontend-design)
to establish the visual plan and critique the rendered result. Use
[Vercel's web-design-guidelines](https://github.com/vercel-labs/agent-skills/tree/main/skills/web-design-guidelines)
to review the implementation's usability, semantics, and interaction states.
Repository requirements and this product direction take precedence when generic
skill preferences conflict, including sentence case and the existing palette.
These skills are development tools, not runtime dependencies.

For subsequent visual changes:

1. Inspect the relevant current screens in a browser.
2. State the intended hierarchy and reuse the shared presentation primitives.
3. Implement with real content, including empty, error, and long-content states.
4. Inspect desktop and 360px/430px mobile layouts in light and dark modes.
5. Verify keyboard focus, 44px controls, overflow, contrast, and core learning
   behaviour with the repository's browser and accessibility checks.

Public staging screens can be inspected directly. Local authenticated browser
fixtures are synthetic: they verify rendering and interaction, and do not prove
live identity-provider or database behaviour. Keep that distinction explicit in
review evidence.
