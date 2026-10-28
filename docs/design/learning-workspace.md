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

Use a system sans-serif stack for the interface and a system serif only for
vocabulary headwords, a compact type scale, comfortable line height, and
left-aligned reading content. Headlines should introduce the page without
displacing its primary action. Large numbers belong to real progress, not
decorative marketing statistics. Borders group related information; shadows
and rounded corners should be restrained.

### Word atlas (late October 2026)

The current direction treats vocabulary like a well-made pocket dictionary that
also knows where the learner is. It replaces the earlier calm study desk while
keeping its palette roles, control sizes and layered control styles.

- **Words are the hero.** Vocabulary headwords (discovery card, word pages,
  Home's latest word, saved words, Journey's next word) use the system serif
  stack in `--vn-headword-font` with the part of speech in italics beside them,
  like a dictionary entry. Interface headings and reading text stay in the sans
  stack. No web font is downloaded.
- **One filled action per screen.** `.vn-btn-primary` marks the backend-selected
  next step (Home's review/discovery panel, Journey's Discover words, the
  discovery Learn button). Other actions are outline (`.vn-btn-secondary`), text
  links (`.vn-text-link`) or chips. Home's practice path keeps its own action as
  an outline button so review and practice never compete.
- **The 10,000-word map is a picture of real progress.** Journey shows the
  selected range of 100 words (a 100-mark strip, counts in text, the next word
  and one action) beside an atlas of all 100 ranges: rows of 1,000 words, each
  square filled from the bottom by already-known (blue) then learning (lavender)
  words. From 40rem every square is a link of at least 44px, with one Tab stop
  and arrow-key movement; on phones the atlas is a picture and a native range
  chooser plus Previous/Next 100 change ranges. Colour never carries meaning
  alone: counts are written next to every visual.
- **Disjoint counts only.** Map counts (already known, learning, new) come from
  the knowledge map and never overlap. Progress shows saved words and
  self-assessed known words as two separate records, never a sum, and review
  stages describe saved words only. Home's goal rows show the day's mission
  values as returned; they are targets, not a promise that more words are due.
- **Immediate, honest feedback.** Every in-app link marks its navigation at
  once with a thin indeterminate bar (`NavigationProgress`); routes keep their
  loading skeletons and polite announcements. Nothing shows invented progress.
  Use `@/ui/link`: it prefetches on intent. Give primary navigation and
  prominent links into routes with a `loading.tsx` `prefetch="auto"`, so their
  loading state appears at once; only add a loading state where it can be
  prefetched, because a late one makes React hold the content back.
- **Quiet chrome.** Phones get a single 56px header row (symbol, search, settings,
  icon-only Log out with its text as the accessible name) above the content and
  the three-item bottom navigation. Focused sessions pin their decision buttons
  to the bottom edge on phones and reserve scroll space so they never hide the
  focused control.
- **Roles, not new colours.** `globals.css` defines `--vn-canvas`,
  `--vn-shadow-card`, `--vn-known`, `--vn-learning` and `--vn-unseen` from the
  generated palette; dark mode maps the same roles. Lavender marks context and
  learning, never a primary action.
- **Controls inherit authored utilities.** The `font: inherit` reset lives in
  `@layer base` and reusable control classes such as `.auth-input` in
  `@layer components`, so utilities (for example the password field's padding
  for its Show/Hide toggle) apply as written. Filled actions do not animate
  colour between disabled and enabled states.

## Layout and interaction

- The public page shows vocabulary in a recognisable situation and explains the
  discover → remember → practise loop. Its learning action is easy to find.
- Authentication shares the brand and visual language of the learning workspace.
  Only configured sign-in methods appear. Visual changes do not activate password
  registration; its verification and recovery email requirements still apply.
- Home reads in one order on phones: the next step (the blue panel with the
  backend-selected action and "Reviewed today"), today's goal (the mission's own
  review, new-word and optional sentence values and status), keep going (the
  practice path and, when review is next, a Discover row), then the latest saved
  word with its picture and a direct link to write a sentence with that meaning.
  Wide screens put the goal and the latest word in a rail beside the next step.
  The streak appears as quiet text only when it is above zero.
- Home, Journey, and Progress remain the three primary destinations. Mobile uses
  thumb-reachable bottom navigation; desktop places navigation beside the search
  field. There is one accessible primary navigation at each size.
- Journey is the one home of the word map: `/discover?band=N#range-words` opens
  a range, plain Journey opens the first range with new words, and the earlier
  `/knowledge-map?band=N` address redirects there. Range links, retries and
  sign-in returns stay on Journey. "Places to use English" situation cards stream
  in their own Suspense boundary and, with the other ways to learn, render
  whether or not the map or situations reads succeed.
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

## Vocabulary artwork

Mix realistic photos, illustrations and simple visual diagrams, as the owner
requested for the current collection and future pictures. Preserve the 3:2 picture
shape and readable teaching layout. Vary people, ages, appearance, clothing,
objects, colors, architecture, indoor/outdoor settings and times of day. The interface palette does not require every illustrated
person to wear blue or lavender or every background to resemble the same room.
Avoid repeating a small cast or template composition throughout the vocabulary.

Show clear actions and interactions when they explain the meaning: reaching,
handing over, entering, leaving, responding, repairing or changing direction.
Pose, gesture and composition should make the action legible in a still picture.
Use quiet object-focused scenes where that best explains a noun; do not add busy
motion that obscures the lesson. Keep one recognizable meaning, readable at
mobile size, with an accurate alternative description and no answer leakage in
graded exercises. Review batches together for variety as well as each image's
meaning accuracy.

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
