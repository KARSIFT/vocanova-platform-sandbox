# Vocanova: a clear first five minutes

Status: **approved — Warm guided coach**, 13 November 2026.
Prepared 13 November 2026 against main `9e1c242f` / release 0.4.32.
The owner authorized the design, comparison, study preparation and subsequent
delivery workflow. The owner selected “Approve Warm guided coach and the proposal (Recommended)”
after reviewing the clickable concepts. This is a product decision, not a learner
study finding.

## Intent

Help an A2–B1 learner understand what Vocanova offers, take a useful next action,
remember a practical meaning and try an original sentence. Preserve practical
vocabulary, gentle encouragement and the distinction between activity and mastery.
Five minutes is a bounded design/study lens, not a promised completion time.

The prior release repaired consistency and layout. This work establishes a visible
target and observable learner tasks before expanding that target throughout the
product. It builds on the existing blue identity, serif vocabulary headwords,
Home/Journey/Progress destinations and the owner's WordUp/Duolingo references.

## Approaches considered

| Approach                                                          | Value                                                             | Tradeoff                                                                    |
| ----------------------------------------------------------------- | ----------------------------------------------------------------- | --------------------------------------------------------------------------- |
| Continue individual page polish                                   | Small releases, low integration cost                              | A coherent learning journey remains hard to evaluate                        |
| Compare a complete short journey, then extend its chosen patterns | Makes hierarchy, interactions and transitions reviewable together | Requires a deliberate design choice and learner observations                |
| Redesign the whole application immediately                        | Broad visual change                                               | More assumptions, more regression surface and weak evidence for the changes |

**Recommendation:** the complete short journey. All three prototypes share content,
navigation, activities and fixture limitations; their presentation varies.

| Direction                                      | Intended character                                                  | What to evaluate                                                          |
| ---------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| Calm dictionary                                | Crisp editorial rules, prominent meaning, restrained surfaces       | Does it feel useful and familiar, and is practice discoverable?           |
| Warm guided coach — recommended starting point | One clear next step, reassuring feedback, quieter secondary choices | Does guidance make the next action clear while preserving learner agency? |
| Visual situations                              | Larger contextual imagery and a practical reason to use the word    | Does the situation help understanding without delaying the task?          |

The recommendation reflects the product mission. No participant has selected a
winner. “Coach” describes guidance and tone; it does not introduce AI chat, a
persona, general tutoring or pronunciation grading.

## Working prototype

Source: [first-five-minutes prototype](../../design/prototypes/first-five-minutes/README.md).
Three interactive directions use _invite_ (verb), the existing definition and
approved original illustration. The prototype has Home, Journey, word detail,
save confirmation, ready-word self-check, sentence draft, explicitly fixed example
feedback, completion and Progress views. New, returning, completed and unavailable
Home scenarios are available in the outside preview controls, with light/dark modes.

No account, API call, audio, analytics or persistent storage is used. Reload resets
the disposable fixture. All apparent saving and review outcomes are sample state.
The writing draft is **not graded**; the feedback view explicitly shows a separate
fixed example and its correction. This is a design comparison artifact, not a
replacement implementation or evidence of production learning.

## Proposed journey and source boundaries

1. **Home:** retain the backend-selected next action. A new learner with no ready
   activity can discover useful words. A returning learner resumes/reviews what
   the server actually provides; never force every person through onboarding.
2. **Discover:** situations and the existing atlas remain available. Browsing and
   word navigation do not record knowledge. Teach the exact meaning and show
   optional details without crowding the primary action.
3. **Save:** only a confirmed exact-meaning mutation can display Saved. Preserve
   pending request identity and uncertainty recovery. Already know stays a separate
   self-assessment; it is not a save, assessment or mastery claim.
4. **Remember:** offer **Review ready words**, using the actual server queue. The
   prototype has only invite in that queue. Production must not promise the next
   queue item is the word just saved. Preserve self-check and multiple-choice
   behavior, the four Again/Hard/Good/Easy ratings, first-answer evidence and the
   existing schedule. No repeat assessment of the same canonical word that day.
5. **Use:** expose the existing exact saved word's sentence destination. Preserve
   its `userWordId`, sense, draft, submission identity, revision, reporting,
   feedback states and safe retries. Keep original text and actionable feedback
   ahead of optional tools. AI checks one target meaning and can make mistakes.
6. **Return:** summarize only server-confirmed activities, then expose the actual
   next action. One save or review does not establish mastery, proficiency,
   mission completion, points or a streak. Show a returned review date when
   available; otherwise say “Come back to see which words are ready.”

These are connected existing activities, **not one new atomic or durable session**.
The current PracticeStartRequest accepts a server journey, lesson or list, not an
arbitrary meaning. This design does not add that contract or claim an exact-word
practice mode. A new save can be due immediately; the at-least-24-hour boundary
applies after assessment, not merely after saving.

## Reusable implementation target after selection

Use the existing token source, Surface, `vn-btn` roles, MeaningPicture, focused
shell, accessible Link and sentence-feedback components. Refine these reusable
parts where the selected screens demonstrate a need; avoid a second design system.

Likely bounded integration points, to inspect again when writing the implementation
plan:

- `apps/web/src/app/(app)/home/page.tsx`: entry hierarchy while retaining the
  server-selected action and actual mission/queue distinctions.
- `apps/web/src/app/(app)/discovery/`: confirmed save/next-step presentation, preserving
  cursor, revision, pending intent and retry behavior.
- `apps/web/src/app/(app)/_components/word-detail-content.tsx`: exact-meaning
  writing entry and optional tools.
- `apps/web/src/app/(app)/words/[userWordId]/`: existing saved-meaning destination
  and `#sentence-practice` deep-link behavior.
- `apps/web/src/app/(app)/reviews/_components/review-session.tsx`: selected visual
  treatment only; preserve queue, grading, scheduling and idempotency.
- Shared sentence-feedback and completion components: readable original,
  correction, explanation and server-confirmed next step.

An implementation plan follows the owner's choice of direction and approval of
this concrete design. One delivery branch and one consolidated PR remain the
release model. Prototype files are not loaded by the production Next.js app.

## Quality target

- One understandable next action, with useful alternatives reachable.
- Headword, meaning, context and action appear in a stable reading order.
- No misleading progress, assessment, feedback, reward or future schedule.
- No clipped essential text, page overflow or obscured keyboard focus at 360px,
  430px, desktop and 200% text. Keep both themes and reduced motion.
- Confirmed, pending, empty, loading, failed and interrupted states use the same
  visual language. A success message cannot substitute for server confirmation.
- No automatic audio or compulsory animation. Physical devices and assistive
  technology remain separate evidence until observed.

## Learner study and measurement

[The comparison protocol](../../product/first-five-minutes-study-2026-11-13.md)
and blank CSV are prepared. The owner arranges participants. **Sessions: zero.**
First-exposure observations provide the primary task evidence. Repeating invite
in another design confounds content and task learning; later comparisons are
preference evidence only. Small formative samples do not establish a statistical
winner or learning/retention gains.

Proposed product measures include starting the appropriate next activity,
confirmed save, first recorded review, sentence-feedback completion and successful
return to the next step. The protocol specifies a minimal proposed event schema.
No analytics is implemented by this prototype; raw sentences, personal details,
query strings, account identifiers and private recordings must not be added to it.
Research can begin without production telemetry. Any telemetry implementation
needs a bounded purpose, current consent/privacy review and explicit data design.

## Evidence informing the process

- [Figma, 4 November 2026: compare working code prototypes](https://www.figma.com/blog/workflow-lab-build-with-code-on-the-canvas-using-the-figma-agent-and-code-layers/).
- [Figma, 23 October 2026: give agents real design-system context](https://www.figma.com/blog/how-to-put-your-design-system-to-work-with-the-figma-agent/).
- [NN/G, 30 October 2026: retain discovery and usability evaluation in AI-led development](https://www.nngroup.com/articles/ux-agentic-software-factories/).
- Repository product bible, learning workspace guidance and owner-supplied
  [reference recording review](../../product/reference-video-review.md).

These sources inform the workflow. They do not prove this prototype is intuitive,
that the recommended direction is best or that the product's learning is effective.
