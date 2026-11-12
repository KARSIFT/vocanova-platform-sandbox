# Vocanova learner usability study

Status: **prepared, not conducted**. The owner will arrange 3–5 A2–B1 English
learners. No participant observations, completion rates or usability scores have
been collected. Automated accessibility checks do not answer these questions.

## Purpose and setup

Check whether learners understand the vocabulary explanations, find a different
meaning of a word, resume learning, recover unfinished changes and interpret
progress. Test the released interface rather than explaining the intended design.
Use four core tasks (another meaning, story resume, interrupted edit and progress)
plus two rotating optional tasks per participant; allow approximately 35–45
minutes including consent and discussion. Spread the three definition examples
and the list/email/empty-catalog tasks across participants rather than rushing
every task in one session. Prefer their usual phone; add
a desktop session if that is a device they normally use. Record viewport, theme,
interface language, browser and the exact web/API release identity.

Use test accounts with prepared learning history on staging. Confirm the release
under test before recruiting. Never ask for a participant's Google password,
OAuth code, personal email inbox or existing private learning account. Test email
recovery with an owner-controlled test inbox and enabled capabilities. If password
entry is disabled in that environment, mark that task **not tested**; do not enable
it or claim a result from a mock. A moderator can show a prepared recorded state
as a separate comprehension task, labelled as such.

The owner should recruit a mix of A2 and B1 learners, including people unfamiliar
with Vocanova. Ask about their usual English learning and device use. Self-reported
level is background information, not a proficiency assessment. Include assistive
technology users where possible; document missing coverage honestly. Use participant
aliases such as P01. Do not put participant names, recordings, account details or
raw notes in the public repository.

## Moderator preflight and reproducible state

Complete a private setup sheet before each session. A task is ready only after
the moderator verifies the specified state on the exact test release. The product
does not have a general staging toggle for session expiry or an empty catalog.
Never use developer credentials to create these states during a learner session.

| Setup item           | Record and verify privately before the session                                                                                                                                                                                                                                                                                                                                                                                       |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Meaning task         | Candidate: `light`, whose source catalog includes separate noun, verb and adjective uses. Open `/vocabulary/light` on the tested release and record the actual meaning IDs, parts of speech, selected source explanation and anchor routes. Use only meanings returned by that release; if they are absent, choose another verified word or mark not tested. Do not presume that two meanings always have different parts of speech. |
| Definition examples  | Record each selected word/meaning ID, exact learner explanation and expected sense in plain language. Include `abandon` and one verified technical example such as `protein`; rotate examples between participants.                                                                                                                                                                                                                  |
| Story                | In an owner-controlled test account, start the selected story normally, stop at an agreed line/question, and record its session ID, saved step, route and expected next task. Reopen the library and verify Continue returns there. No production account resets.                                                                                                                                                                    |
| List                 | Create a disposable list with named meanings using ordinary controls. Record its ID and initial name/revision. Verify it can be practised before assigning the task.                                                                                                                                                                                                                                                                 |
| Progress             | Record actual server-confirmed states and counts for the prepared test account. If all four word states cannot be prepared, test the available subset and record the gap. Never edit progress rows to manufacture a study result.                                                                                                                                                                                                    |
| Sign-in interruption | Use an explicitly isolated test environment with a prearranged expiring test session or a documented fixture that produces 401 only for the selected save. Verify the real recovery route and same-user restoration first. A local fixture result remains fixture evidence. If no suitable setup exists, mark the task not tested and study the ordinary save path instead.                                                          |
| Email recovery       | Check which capability is enabled, the test inbox, cooldown and error simulation in the isolated environment. Do not disable production providers or intercept a participant's account. Mark unavailable variants not tested.                                                                                                                                                                                                        |
| Empty topics         | Use the documented synthetic browser fixture `e2e_journey_situations=empty` only in the local test harness. It has no claimed staging/production behavior. If that harness cannot be used in a session, mark the task not tested; a recording is a separately labelled comprehension task.                                                                                                                                           |
| Repeat/reset         | Use a fresh disposable account/session/list where feasible. Record any owner-operated test reset procedure and confirm its target before the session. Never run catalog or learner-data deletion against production.                                                                                                                                                                                                                 |

## Consent and introduction

Read: “We are testing the website, not your English. There are no grades. You may
skip a task or stop whenever you want. Please say what you are trying to do and
what you expect to happen. I may stay quiet while you work. Your feedback will help
us improve the website.”

Ask permission before recording screen or voice and explain who will see it, where
it will be stored and when it will be deleted. Participation must remain possible
without recording. Agree a retention period with the owner rather than assuming
indefinite storage. Never record passwords, tokens, inbox contents or unrelated
browser tabs. Pause recording before sign-in.

## Moderator rules

- Give one task at a time and let the learner choose their path. Do not name the
  button or destination that solves the task.
- Ask “What are you looking for?” or “What did you expect?” rather than suggesting
  an action. Allow silence before prompting.
- If the participant is stuck for about two minutes or wants to stop, offer to move
  on. Record assistance precisely. Assisted completion is not independent success.
- Ask for a plain-language explanation before teaching or correcting. Do not infer
  understanding from a click, a smile or a correct multiple-choice guess.
- Keep connectivity failures separate from navigation/confusion. Keep first
  attempts and later retries separate. Avoid presenting alternative designs until
  the primary task is complete.

## Tasks and evidence

Prepare the exact test account state and selected word/senses in advance. Use the
same initial state for each participant and reset only the disposable test account.
Ask for their own sentence without requiring personal information.

| Task                        | Neutral participant prompt                                                                                                           | Prepared state and observable evidence                                                                                                                                                                                                                                                                    |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Understand an explanation   | “Find **abandon**. Tell me what it means here in your own words, then give a simple example.”                                        | Confirm the reviewed learner explanation and part of speech. Record paraphrase and example against the intended sense; do not mark it correct merely because the same wording is repeated. Add one long technical definition and one ordinary editorial definition, rotating order.                       |
| Find another meaning        | “This word can be used another way. Find that meaning and show me which kind of word it is.”                                         | Pick a word with at least two actual teaching senses/parts of speech; verify those exist. Start from search/map or a personal list. Record selected sense, wrong turns and whether the correct meaning receives focus.                                                                                    |
| Resume a story              | “You started this story earlier. Carry on from where you stopped.”                                                                   | Seed an in-progress story. Record first action, overview detours, correct resumed line/question and whether the learner recognizes the current task. Include one new/completed story as a follow-up rather than implying all states behave alike.                                                         |
| Organize and practise words | “Find your travel words and practise them. After that, give the list a name you would recognize.”                                    | Existing list with several meanings. Record whether learning is found before management, whether settings can be discovered, and whether confirmed rename is understood. Do not ask for deletion of personal data.                                                                                        |
| Recover interrupted editing | “Change your practice preference/name. Your sign-in has expired before it could save. Sign in again and finish what you were doing.” | Moderator uses a disposable short-lived session. Tell the learner a sign-in interruption will occur; do not fake a security alarm. Record lost edits, restored-draft understanding, pending versus confirmed state and duplicate attempts. Run Profile/preferences separately and label which was tested. |
| Correct an email address    | “You requested a sign-in or confirmation message, but the address is wrong. What would you do?”                                      | Use only the enabled recovery flow with a test inbox. Record ability to change address, understand resend cooldown and recover from a temporary resend error. Never interpret a generic accepted response as proof an email was sent.                                                                     |
| Interpret progress          | “Explain what these colors and numbers say about your words. Which words still need practice?”                                       | Seed new, already-known, learning and learned states. Record distinctions, misconceptions and next action; distinguish activity/rewards from proficiency.                                                                                                                                                 |
| Empty writing catalog       | “You want to practise writing, but no topics are available. Show what you would do next.”                                            | A documented staging fixture or separate controlled test environment with successful empty response. Label fixture-based findings. Record understanding and useful alternative; do not portray a server error as an empty catalog.                                                                        |

Counterbalance the starting destination for the meaning task and the order of the
three definition examples. Do not try every state with every person if fatigue
changes behavior; record exactly what was and was not observed. For keyboard users,
retain their usual input method and note focus traps, hidden controls and overlays.

## Observation and synthesis

Use [the blank session sheet](learner-usability-observation-template.md). Record
independent/assisted/failed/not-tested per task, first action, elapsed time if
practical, wrong turns, exact assistance, paraphrase, and a short anonymized quote
with consent. Time is descriptive; with 3–5 participants it is not a population
benchmark. Do not average different task variants or treat this small study as
statistically representative.

After each session, separate observed behavior from the moderator's interpretation.
For definition comprehension, compare the paraphrase/example with the reviewed
sense and ask an English teacher/editor to adjudicate uncertain cases. Tag whether
the definition, unfamiliar word, navigation or task wording caused difficulty.

Prioritize any loss of edits, unintended action, keyboard blocker or misleading
state immediately. Repeated confusion in two participants is a reason to investigate,
not proof of a universal problem. A severe issue observed once still matters. Keep
individual cases visible alongside counts and include missing coverage. Turn
confirmed issues into bounded changes, retest affected tasks with fresh learners
where possible, and retain the release/version attached to every observation.

## Completion record

- Owner confirmation: owner will arrange learners; study preparation authorized.
- Participants recruited: pending owner.
- Sessions observed: **0**.
- Results/scores: **none**.
- Next action: owner arranges participants and sessions, using these materials;
  anonymized observations can then be analyzed and linked to follow-up fixes.
