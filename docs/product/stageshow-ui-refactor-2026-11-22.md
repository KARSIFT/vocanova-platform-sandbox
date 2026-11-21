# Stageshow interface refactor — 22 November 2026

This refactor applies the hierarchy, grouping and responsive layout principles of
[Stageshow 1.1.1](https://github.com/benjitaylor/skills/tree/299140fafd08df55a9c234e610c3f64af3f45531/stageshow)
to the existing Warm guided coach direction. The learner’s next action gets the
strongest visual emphasis. Related links and settings share a surface and thin
dividers. Secondary information uses native disclosures where appropriate.

| Area | Implementation |
| --- | --- |
| Public entry | The landing page demonstrates teaching and authored feedback with an existing VocaNova word image. Help groups its existing answers into disclosures. |
| Authentication and setup | Sign-in, signup, confirmation and recovery use a common reading-width shell. Onboarding shows the real step and answer choices. |
| Home, Journey and Review | Home emphasizes the next available activity. Journey groups situation links. Discovery and Review separate the word or question from navigation and confirmed feedback. |
| Activities | Practice, stories, courses, lessons, topic writing and the learning plan use consistent entry headings, grouped choices, and focused session frames. |
| Library and account | Saved words, search, lists, Progress, sentence history and Settings use shared headings and grouped rows. Word detail keeps the meaning and optional sentence practice distinct. |

`apps/web/src/ui/surface.tsx` defines the shared page widths, headings and
sections. Navigation, status and form data remain part of their existing route
and component contracts. Auth, scheduling, review counts, progress, and writes
still use the existing server paths. The learning demonstration contains
authored copy. Motion responds to interaction and respects reduced motion.

## Verification

- The Stageshow layout audit captured 174 combinations across 29 entry routes,
  360px, 430px and 1280px viewports, and both themes. It found no horizontal
  document overflow or layout findings. Screenshots were inspected on the
  landing page, Home, Journey, Discovery, Settings, and active learning flows.
- Web lint, formatting, Playwright type checks and 352 web middleware tests
  passed. The local production build succeeded on retry after Next.js emitted a
  security upgrade reminder for the existing dependency version.
- Playwright uses the synthetic API fixture and covers enlarged text, reduced
  motion, keyboard access, and learning, empty, recovery and loading states.
  The PR records browser results and hosted checks. Production authentication,
  real AI feedback and database durability require live acceptance evidence.
