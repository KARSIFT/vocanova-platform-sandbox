# First five minutes: three directions for Vocanova

Status: **prepared, not conducted**. Recruitment is pending the owner; sessions
observed: **0**. There are no learner results, scores or winning direction.
This is a formative study plan for 3–5 A2–B1 learners, not a learning experiment.
Use it alongside the [existing usability protocol](learner-usability-study-2026-11-12.md)
and [blank observation CSV](first-five-minutes-observation-template.csv).

## Question and hypotheses

Which entry experience helps a learner understand Vocanova and begin its
discover → save → recall → write loop with the least help?

The three clickable directions have the same word, explanation, example, task,
navigation destinations and demo limitations. Their entry hierarchy differs:

| Direction  | Hypothesis to investigate, not a finding                       | Possible tradeoff                                                                       |
| ---------- | -------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| Dictionary | A prominent word and meaning make the starting point familiar. | Learners may see a lookup tool without noticing practice.                               |
| Coach      | A clear next step makes the learning sequence easier to start. | Learners may follow instructions without understanding their goal or available choices. |
| Situations | A practical dinner context makes the reason to learn clear.    | Extra context may delay reaching the useful word.                                       |

Observe both helpful behavior and contrary cases. A preferred visual direction
does not establish better learning, comprehension, retention or real AI quality.

## Fixed content and scope

Use **invite**, verb, in all three directions. The teaching meaning is “To ask
someone to come to an event or do something with you.” The shared context is
inviting friends for dinner. The teaching example is “I want to invite my friends
to dinner.” The fixed feedback demonstration starts with “I want invite my friends
to dinner.” and suggests “I want to invite my friends to dinner.” Do not simplify
one direction's explanation or change its example,
feedback, task wording, starting state or facilitator help independently.

These are isolated design previews. All counts and progress are fixtures; any
saved confirmation is a preview state. Saving models eligibility for a first
scheduled review: a newly saved word can be ready immediately. “Review ready
words” opens a simulated queue containing only _invite_. The learner reveals the
meaning with “Show meaning” and chooses Again / Hard / Good / Easy; every rating
ends that word's review for the preview day, with no repeat review in that state.
This models the existing review flow; it does not create a new practice contract.
There is no real SRS record, scheduling/network request, AI evaluation, persistent
learning record or audio. Feedback is an explicitly labelled scripted sample for
the fixed example; it does **not** judge the learner's typed sentence.
“Reviewed once” describes a demo interaction, not mastery. Five minutes is the
study observation window, not a promised completion time or a forced app timer.

The production product still requires server-confirmed, idempotent learning
mutations. This study supplies no evidence about those guarantees, authentication,
data durability or production readiness. It does not replace the broader study.

## Participants, order and time

The owner arranges 3–5 learners, aiming for a mix of self-reported A2/B1 and people
new to Vocanova. Record broad prior experience privately; self-reported level is
not a proficiency test. Prefer the learner's usual phone and input method. Note
missing keyboard, assistive technology, device and language coverage honestly.
Allow approximately 25–35 minutes including consent, one first-use attempt,
two later comparisons, a returning-state task and discussion; take breaks or
omit secondary states when needed.

Assign order before the session, using this rotation rather than asking learners
to choose the first design. Direction names are moderator labels; do not introduce
one as the preferred or recommended design.

| Planned slot      | First exposure | Second     | Third      |
| ----------------- | -------------- | ---------- | ---------- |
| P01               | Dictionary     | Coach      | Situations |
| P02               | Coach          | Situations | Dictionary |
| P03               | Situations     | Dictionary | Coach      |
| P04, if recruited | Dictionary     | Situations | Coach      |
| P05, if recruited | Coach          | Dictionary | Situations |

These are planned slots, not recruited people. Three participants give each
direction one first exposure; four or five leave unequal group sizes. Report
actual denominators. This small sample cannot rank directions statistically.

Only the **first exposure** contributes primary task-completion, first-click,
goal-understanding and time evidence. Reusing _invite_ teaches both the content
and task sequence. Later attempts are **preference/navigation comparisons only**;
never pool their recall, comprehension, speed or completion with first-use data.
Counterbalancing reduces order bias but does not remove this carryover. If the
owner later wants learning comparisons, prepare fresh, editor-reviewed matched
words and a separate counterbalanced protocol before collecting those data.

## Moderator preflight

Keep the completed setup and observation sheet **outside this public repository**.

1. Record preview URL/version or commit, exact direction order, device, viewport,
   theme, browser and input method. Verify Home, Journey and Progress in each
   direction; follow through save, recall, writing, sample feedback and return.
2. Confirm all three show the fixed _invite_ content, same sample feedback and
   visible fixture/scripted labels. Verify typed text is not represented as graded.
   Record mismatches as setup defects and resolve them before a comparison.
3. Verify how the preview's controls select and reset **new**, **returning**,
   **completed**, and **unavailable** states. Record the exact safe reset procedure;
   reset only isolated preview state. Do not infer persistence from a refresh.
4. New state: no saved word or prior activity. Returning state: the same prepared
   sample history and ready review across directions. Verify all four review
   ratings and that an assessed word cannot be reviewed again in that preview
   day. Completed state (labelled “After practice”): truthful demo completion with
   a route home. Unavailable state (labelled “Unavailable next step”): an intentional failed-load state,
   distinct from successful empty data. Never manufacture these in production.
5. Keep state selectors and direction switching under moderator control. Explain
   that the outside toolbar is preview equipment, not product navigation. Confirm
   links do not take the learner into personal accounts or unrelated live flows.
6. Prepare a private notes file and optional recording setup, with an agreed
   deletion date. Close unrelated tabs. No sign-in, real email or credentials
   are needed. If a required preview state fails preflight, mark it not tested.

## Consent and shared introduction

Read the same introduction before the first direction:

> We are testing three website designs, not your English. There are no grades.
> You may skip anything or stop. Please tell me what you are trying to do and
> what you expect. These are demos: numbers and saved states are examples, and
> feedback is a fixed sample, not an assessment of your writing. Nothing here
> records real learning progress. You do not need an account or personal details.

Ask consent to participate and, separately, optional screen/voice recording.
Explain who can see notes or recordings, their private storage and deletion date.
Allow participation without recording. Request permission before retaining short
anonymized literal quotes; do not include identifying details. Stop recording if
personal data appears. Do not put names, contact details, filled sheets or
recordings in Git. The owner manages recruitment privately.

## First exposure: identical five-minute task

Start on the new learner Home of the assigned first direction. Before giving
task hints, ask: **“What do you think you can do here? What would you try first?”**
Record the learner's own words and first product click. Do not teach the goal.

Give this neutral first task verbatim:

> Tomorrow you want to ask friends to come to dinner. You have five minutes to
> prepare with this website. Show me how you would start.

Let the learner choose a route. If they reach later steps naturally, observe
without interrupting. Use the following prompts **only when the next task has not
already happened**, one at a time, in this same order for every first exposure.
Log the timestamp and exact prompt used; prompted discovery is not spontaneous.

| Step              | Shared prompt                                                                                             | Observe, without supplying the answer                                                                                                                                                                                 |
| ----------------- | --------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Meaning           | “Find invite. Tell me what it means here in your own words.”                                              | Route, first click, sense, paraphrase and where the learner looks. A click or copied definition alone does not show understanding.                                                                                    |
| Save              | “Keep this word so you could practise it again.”                                                          | Whether save is found, fixture confirmation understood and next step recognized.                                                                                                                                      |
| Recall in review  | “Try remembering it using the practice on this site.”                                                     | Finding Review ready words, using Show meaning and choosing Again / Hard / Good / Easy. Check what the rating and finished review mean to the learner. Do not call this retained learning or a real scheduled result. |
| Own sentence      | “Write a sentence with invite about a situation you choose. It can be imaginary.”                         | Entry route, task understanding and whether the word's meaning is available. Avoid personal details; do not transcribe the full sentence into the CSV.                                                                |
| Scripted feedback | “Look at the sample feedback. What sentence is it about? What does it tell you?”                          | Whether the learner recognizes the fixed example and that their writing was not assessed. The moderator must not imply that arbitrary input receives real feedback.                                                   |
| Return            | “You are finished for now. Go back to where you would start next time. What would you expect to do then?” | Home route, next action, interpretation of fixture progress and any expectation of real saving or scheduling.                                                                                                         |

Start elapsed time when the first task is read; capture step times if practical.
At five minutes, record the reached step and elapsed time before offering an
optional continuation. Tag all later actions **after window**. Do not rush the
learner or count a moderator finish as their success. If stuck for about two
minutes or wanting to stop, offer to skip; log exact assistance. Keep task prompts
separate from navigation hints. Classify each step independent / assisted /
failed / not tested, with a reason. A task prompt can elicit an independent
action; a hint identifying the control or route makes the action assisted.

## Later comparisons and returning state

Reset the next direction to the identical new state and repeat the same invite
content and task wording. Label these exposures comparison-only in notes. Do not
score faster repeated performance as an improvement. After all three, ask which
starting point they would choose, why, what felt confusing and what they would
change. Retain contrary preferences and distinguish attractiveness, familiarity,
understanding and task support. Avoid “Which was easiest?” as the only question.

Then select one direction using a preassigned rotation, not the learner's
favorite, and load its returning fixture. Read: **“Imagine you used this site
yesterday. You have a few minutes now. Show what you would do next and tell me
why.”** Record whether they find the next task and interpret the sample saved,
reviewed and completed states. This is a **scenario comprehension task after prior
exposure**, not fresh returning-user behavior or evidence of real next-day recall.
Record the direction, fixture and prior exposures with the observation.

As an optional short follow-up, show completed and unavailable states and
ask **“What does this tell you? What would you do next?”** Do not mix these with
the timed first-use task. An unexpected crash, broken reset, missing content or
network failure is a **setup/data-load failure**, not learner failure. Pause the
timer, record lost time and mark affected steps not tested if a clean retry is
not possible. A deliberate unavailable scenario can yield recovery-usability
observations; label its scenario separately from a real incident. Its “Try again”
control returns to a prepared fixture, not proof that a real request recovered
or saved progress survived a failure.

## Observation, privacy and synthesis

Copy the blank CSV privately; one row represents one observed task or event.
Leave unknown values blank. Record an alias, direction, exposure number,
new/returning/completed/unavailable scenario, first-click target, wrong turns, exact
assistance, elapsed seconds, observed behavior, optional anonymized literal quote
and separate interpretation. Do not prefill expected outcomes as observations.
Keep consent, schedule, release/device setup and any recording link in a private
session cover sheet; do not put credentials or contact information in it.

For primary evidence, report individual first-exposure cases and counts with the
actual direction/state denominator. Separate assisted outcomes, omitted steps,
after-window continuation, setup failures and later preferences. Time describes
these sessions, not a population benchmark. Compare paraphrases with the intended
sense; refer uncertain language judgments to an editor/teacher. No comprehension
or learning rate can be inferred from dummy counts or scripted feedback. A severe
blocker seen once deserves action; repeated confusion is a reason to investigate,
not a universal claim. Choose bounded follow-ups and retest with fresh learners.

### Measurement plan, not implemented analytics

No automatic analytics are added by these materials. Manual observation is enough
for this study. If instrumentation is later authorized, use an explicit allowlist:

| Proposed field                  | Allowed values or rule                                                                                                                                                                    |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| study_version / preview_version | Fixed study version and build/commit identifier, no URLs                                                                                                                                  |
| event_name                      | view_home, open_word, save_demo_confirmed, start_review_demo, reveal_meaning, rate_review_demo, open_sentence, view_scripted_feedback, complete_demo, return_home, load_error, retry_load |
| direction / scenario            | dictionary, coach, situations / new, returning, completed, unavailable                                                                                                                    |
| exposure_index / step           | Integer 1–3; fixed task identifier                                                                                                                                                        |
| elapsed_seconds / outcome       | Relative duration; independent, assisted, failed, not_tested                                                                                                                              |
| review_rating                   | Fixed enum again, hard, good, easy; only for rate_review_demo                                                                                                                             |
| error_kind                      | Fixed enum such as controlled_load_error or unexpected_load_failure                                                                                                                       |

Use study-local aggregates or private manual aliases only. Do not send or store
sentences, typed text, search terms, emails, account/user IDs, tokens, full URLs,
query strings, IP/device fingerprints, absolute timestamps or recordings as
analytics payloads. No automatic text capture, session replay or keystroke logging.
Do not attach manual literal quotes to event payloads. Define consent, retention,
access and deletion before any later telemetry implementation; inspect actual
payloads rather than assuming the allowlist is enforced.

## Completion record

- Preparation authorized; owner arranges recruitment and sessions.
- Participants recruited: **pending owner**. Sessions observed: **0**.
- Findings, winner, learner scores and production evidence: **none**.
- Next action: owner verifies previews and fixtures, schedules 3–5 learners and
  stores completed observations privately; only then synthesize observed issues.
