# WordUp and Duolingo hands-on audit

The comparison covers the Mac web apps, the owner’s recordings, and
Vocanova’s current delivery branch. The live sample completed ten WordUp
Discovery decisions, twenty WordUp due reviews, and three Duolingo sessions
through their results. Account statistics, images, and exercise text stay
outside this record.

## WordUp

| Surface | Observed interaction | What Vocanova can learn |
| --- | --- | --- |
| Home and collections | Home puts Today’s Reviews beside the Knowledge Map and a separate next word to learn. Its totals link to Knew, To Learn and Started. Review shows the full queue and presents a milestone after the twentieth item. | Show today's attainable goal alongside the full backlog. Keep self-reported knowledge, intention to learn and recorded practice distinct. |
| Discovery map | Switched between Cards, Tags and List. Cards combine picture, pronunciation, translation, definition and decisions. Tags make words compact and selectable. List adds definitions and inline decisions. All three use the same ranked range. | Offer quick scanning and focused study within the same range. Keep rank, state and return position stable between views. |
| Discovery decisions | Used Should Learn and Already Knew across words about everyday food, observation, networking, electrical measurement, business and civic life. Either choice advances the map. Should Learn leaves a Learn Now action on the previous card. A certificate promotion interrupted the sequence. | Give the learner an explicit route to study the exact saved sense now or continue browsing. Keep commercial and milestone interruptions out of that choice. |
| Meaning detail | Opened Learn more and inspected alternate senses, pronunciation, example sentences, translated text, notes and usage guidance. The exercise’s target sense appears first. Other meanings are available without changing the completed answer. | Keep word detail rich and organized by sense. Offer pronunciation, contextual usage, comparison, personal tools and a return to the exact exercise or map card. |
| Review prompts | Used spelling options, definition-to-word and word-to-definition choices, contextual gaps, audio-to-definition and a sentence self-check. Pictures accompanied several prompts. A spelling choice could sometimes be solved without understanding the meaning. | Vary what the learner retrieves. Match each cue to its goal, pair images with accessible text when appropriate, and use distractors that actually test that skill. |
| Review feedback | A correct answer led to a detail sheet with context and the suggested next date. The date menu offered shorter and longer intervals and an option to stop reviewing. Confirmed the recommended date and returned to the queue. | Keep the result, useful explanation and next date together. Always show that the selected date is saved or pending. |
| Help during review | Not Sure? and Look Up opened a short teaching sequence. It progressed through the meaning, usage guidance and an example before offering Learn more or Next Word. I Remember followed the self-report path. | Give unsure learners a supported way to see the meaning and return. Keep deliberate skips, assisted study, self-report and objective grades distinct. |
| Settings and practice | Inspected accent, reminders, translation, review cards, challenge types, example preferences, image style and appearance. Lexi offers vocabulary, writing and speaking. Recordings cover roleplay, topic writing corrections, scenarios, idioms, phrasal verbs, lists and exam entry points. | Connect preferences to observable behavior. Offer bounded practice with clear goals, working alternatives when audio is unavailable, and readable writing feedback with a rewrite path. |

## Duolingo

| Completed session | Scenarios and controls observed | What Vocanova can learn |
| --- | --- | --- |
| Mistakes | Opened a bounded set of recent errors. Played through gardening, travel, food and a private conversation using passage comprehension and contextual gaps. Used the audio-unavailable action on a typed listening exercise, then continued. Results led to a streak screen. | Make missed items easy to find. Offer listening alternatives without turning a skip into recall credit. Keep the result useful after a short session. |
| The Perfect Job | Read a conversation about work, travel and housing. Completed a detail check, true-or-false inference, phrase completion, word-bank assembly, and a final motivation question. The story revealed lines progressively and gave immediate feedback. | Let a story test what a sentence means in the conversation. Use varied interactions and provide access to previous lines and audio. |
| Share opinions about flavors | Completed illustrated gaps, passage-based meanings, and choosing a natural reply. Repeated target phrases in different situations. Submitted one wrong choice to inspect the correction. The session continued, labeled the missed exercise for a final retry, and reached Lesson Complete after correction. | Keep the first result and the later correction visible as separate accomplishments. Reuse a target phrase across different tasks and scenarios. |

The [earlier recording review](reference-video-review.md) adds the path,
guidebooks, sound discrimination, same/different and audio matching,
listening-only practice, radio, habits, social features and settings.
The live completion sample covers three session types. Speaking assessment,
premium roleplay, notification delivery and sync were outside that sample.
[Duolingo’s Practice tab guide](https://blog.duolingo.com/guide-to-duolingo-practice-hub/)
documents additional modes and their intended use.

## Differences that matter for Vocanova

- Vocabulary exploration already includes a ranked map, exact resume, search,
  named lists, and sense detail. The next change can connect those pieces so a
  learner can scan for a word, compare meanings, study one, practice it, and
  return to the same place.
- Daily Review already uses picture options, meaning choices and reviewed
  gaps. The sentence gap bank has twelve reviewed targets. The first thirty
  ranked example sentences are teaching material. New graded questions need
  their own answer and distractor review.
- Practice already has listening, typed recall, sentence assembly, and
  mistakes. The main Daily Review route records objectively graded attempts
  using the self-check prompt type. The existing mistake query admits
  multiple-choice review rows only, so a daily miss can disappear from
  optional practice. Use server-confirmed attempt provenance to distinguish
  objective misses from self-ratings before expanding this link.
- The repository has 30 guided lessons, 17 situation guidebooks, and six
  interactive stories. Some questions already support typing and listening.
  The missing connection is practice across these surfaces, including later
  correction of a miss and transfer of a word to a different situation.

## Delivery choices

1. **Connect the learning loop.** Combine range scanning, deeper sense
   teaching, assisted due review, actionable mistakes and several complete
   scenario packs on the existing Home, Journey and Review destinations.
   Share question controls and completion patterns while each server flow
   keeps its own rules for progress and scheduling.
2. **Start with questions and practice.** Build a richer, reviewed question
   bank and close the Review-to-Mistakes gap first. This brings new exercises
   to existing learners sooner and leaves the Discovery layout for later.
3. **Start with Discovery and content.** Build broader authored examples,
   contrasts and scenarios with better navigation between meanings and
   personal tools. This gives later question formats a stronger content base.

The first direction covers the owner’s requested redesign most directly.
Scope the first integrated scenario packs around work, travel, restaurants,
shopping, and an everyday problem. Review each exercise for sense, clarity
and plausible alternatives. Test at 360px, 430px and desktop, with keyboard,
both themes, enlarged text, missing media, resume, failed saves, retries and
legacy sessions. Confirm scheduling and progress behavior on PostgreSQL.
