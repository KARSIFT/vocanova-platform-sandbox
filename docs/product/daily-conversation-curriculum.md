# Daily Conversation curriculum

This bounded expansion adds language for making and changing plans to the canonical [seed](../../apps/api/cmd/seed/voc026-p1.json). It retains the original six meanings and gives each of the eighteen meanings two complementary examples and a fuller explanation that adds information to its concise definition. All examples are synthetic teaching content.

## Teaching levels and pathway

The situation uses the schema-supported `a2_b1` band. Word, meaning and example `difficulty_level` values use `a2` or `b1`; these are editorial teaching judgments, not CEFR certification or measurements of learner ability. A2 entries remain at the start and throughout the pathway. Longer definitions and examples provide support for B1 material. Reschedule and farewell are supported stretch/recognition items; farewell is not presented as the usual casual goodbye. Other situations keep their existing metadata.

The sequence follows a conversation from opening through suggesting, agreeing, adjusting and maintaining contact. It is not a strict difficulty ranking. The content API sorts core items before display order. The first seventeen entries are core and farewell is non-core, so the served order matches this pathway. Greeting and casual become core in place; their IDs are unchanged. Here core controls pathway placement, not proficiency or certification.

| Order | Target        | Editorial level | Intended meaning and teaching rationale                                                                                                                                            | Complementary patterns                                                  |
| ----- | ------------- | --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| 1     | greeting      | A2              | A2 entry and recognition: connects a noun to usable opening language; a spoken welcome and plural exchange complement one another.                                                 | friendly greeting in a spoken welcome / exchange greetings              |
| 2     | small talk    | B1              | Supported B1 social skill: complements make small talk with the uncountable noun phrase as subject; keep the explanation concrete.                                                 | make small talk / small talk as the subject                             |
| 3     | casual        | A2              | Supported A2 descriptive entry: casual chat complements the predicate was casual; stay within conversation rather than clothing or employment.                                     | predicate adjective / casual chat                                       |
| 4     | weekend plans | A2              | A2 entry chunk: a familiar opening question plus a changed-plan statement. Plural plans is usual for a general question, not mandatory in every context.                           | opening question / plans changed                                        |
| 5     | available     | A2              | Turns general weekend plans into a question about a specific time. The meaning is personal availability, not whether an item can be bought.                                        | available for lunch on Saturday / available after six with a time limit |
| 6     | invite        | A2              | A direct invitation gives learners a useful way to start a social plan. The two examples teach invite someone to an event and invite someone to do something.                      | invite someone to an event / invite someone to do something             |
| 7     | join          | A2              | Links invitations to taking part with friends. This sense is participating with people, rather than connecting objects or becoming a formal member.                                | join us for lunch / joined friends for a walk                           |
| 8     | suggest       | B1              | Adds a way to offer an idea without deciding for everyone. Complementary examples introduce suggest plus -ing and suggest plus an activity noun.                                   | suggest + -ing / suggest an activity noun                               |
| 9     | sounds good   | A2              | A short conversational response lets learners react naturally to a suggestion. The second example shows acceptance with a practical condition.                                     | That sounds good / acceptance with a weather condition                  |
| 10    | arrange       | B1              | Helps learners turn a suggestion into a practical plan with a time or place. Excludes the separate sense of placing objects in order.                                              | arrange a time to meet / arranged a picnic                              |
| 11    | meet up       | A2              | Adds the action of getting together, while the existing catch up focuses on exchanging news after time apart. A past-tense example is needed for natural follow-up conversation.   | meet up at a place / met up with someone                                |
| 12    | confirm       | B1              | Distinguishes an agreed plan from a tentative suggestion. Examples cover checking a meeting time and saying definitely that one will attend.                                       | ask to confirm a time / confirmed attendance                            |
| 13    | on time       | A2              | Connects agreement to arriving at the expected time. A phrase taught as a unit avoids requiring abstract grammar vocabulary.                                                       | arrive on time / reached the cinema on time despite a delay             |
| 14    | reschedule    | B1              | Gives a precise, useful word for changing a plan without abandoning it. The longer word needs a plain definition and a contrast with cancel, not an asserted certified CEFR level. | reschedule for a new day / rescheduled because of weather               |
| 15    | cancel        | A2              | Teaches a common plan change and contrasts it with choosing a new date. Keeps British spelling visible instead of teaching only the variant the matcher currently accepts.         | cancel an event / cancelled a booking                                   |
| 16    | catch up      | B1              | Supported B1 idiomatic social sense: invitation plus past narration and with someone. Differentiate exchanging news from merely meeting.                                           | invitation to catch up / caught up with a friend and heard news         |
| 17    | keep in touch | B1              | Provides a friendly way to end a visit while maintaining contact. The second example describes successful continued contact, rather than repeating a request.                      | request after a course / kept in touch through regular messages         |
| 18    | farewell      | B1              | Supported B1 register recognition, not the default casual goodbye: contrast say farewell with give someone a farewell, preserving everyday alternatives.                           | say farewell / give someone a warm farewell                             |

## Editorial decisions

- Join teaches the usual invitation pattern without claiming that join with someone is always wrong.
- Confirm includes verifying a planned detail as well as making an arrangement definite. Cancel includes arrangements such as bookings as well as events; reschedule supplies a different time.
- Suggest has a register note contrasting the conversational How about ...? pattern. On time explains expected arrival/start times; its usage note carries the separate in time distinction once.
- Sounds good! remains a natural two-word response in the teaching note. Sentence practice still requires at least three words, so the first practice example is That sounds good. The second uses natural let’s. Sound good and sounded good can preserve this evaluative sense; agreement still depends on context.
- Greeting now starts in a spoken social context. Farewell retains a single note about formal/important goodbyes and a separate say farewell to someone pattern. Weekend plans presents the usual plural without prohibiting a valid singular plan.
- Catch up means exchanging news after time apart; meet up means getting together and need not involve sharing news. Keep in touch concerns continuing contact, not arranging a particular meeting.
- Valid British cancelled and natural met up, kept in touch and caught up examples are retained. Content is not rewritten into unnatural English to fit a lexical matcher.

## Optional context practice

Daily Conversation includes an authored “Choose the word for the situation”
activity with three independently reviewed message/dialogue examples:
`invite` versus `join`, `sounds good` versus `keep in touch`, and `reschedule`
versus `cancel`. Each example has two choices; explanations appear after a
choice. Learners can try again, move to the next example, exit to the word list,
or restart after completion. Answer state is local and resets on a fresh page
load.

Completion links to the existing word pages. Saving remains explicit; an
already-saved word supports sentence practice, while an unsaved word must first
be saved. Context answers do not change points, missions, progress, saved words
or scheduled reviews. The activity introduces no API, schema or provider calls
and makes no claim about learning effectiveness.

The [content helper](<../../apps/web/src/app/(app)/discover/[situation]/_components/conversation-context-content.ts>)
requires the actual Daily Conversation situation and all six canonical meanings
to match their expected IDs, text and route slugs exactly, with no duplicate ID.
If any reference is missing or drifts, the optional activity is omitted and the
ordinary situation word list remains available. The
[practice component](<../../apps/web/src/app/(app)/discover/[situation]/_components/conversation-context-practice.tsx>)
uses those guarded references for its word-page links.

Full workspace validation passed, including four new real-seed helper checks.
All 21 focused browser checks passed across 360px, 430px and desktop, covering
both themes, keyboard navigation, explanations, retry/restart/exit, reload,
missing-content fallback and explicit saving before sentence practice. The full
browser matrix passed 314 tests with 37 existing skips. The activity shipped in
[PR #1476](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1476), and
[staging release checks](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36968236895)
passed on merged revision `e4cc4236`. The deployed synthetic journey covers the
general learning loop; it does not exercise these three examples. Synthetic checks do not
establish screen-reader speech or learning effectiveness.

## Stable identities and repeat seeding

All existing row IDs and foreign-key relationships are preserved, including the six original examples, notes and journey memberships. Greeting example wording, farewell definition and selected notes change in place. Display order changes do not change identities.

New UUIDv5 IDs use namespace `d9f0dfc7-bc05-5595-bbf2-035c4a61ca9c` and name prefix `daily-conversation-plans-v1/`. This is a documented convention for this slice, not an assertion about historical ID generation. Keys are frozen semantic identifiers; editorial wording and display order are excluded.

| Target slug   | Frozen sense key        |
| ------------- | ----------------------- |
| invite        | social-invitation       |
| join          | participate-with-others |
| available     | free-at-a-time          |
| suggest       | propose-an-activity     |
| arrange       | plan-a-social-event     |
| confirm       | make-a-plan-definite    |
| reschedule    | change-event-time       |
| cancel        | stop-a-planned-event    |
| on-time       | punctual                |
| sounds-good   | accept-a-suggestion     |
| meet-up       | meet-socially           |
| keep-in-touch | continue-communicating  |

For each new target, keys are `word/{slug}`, `meaning/{slug}/{sense}`, `journey/{slug}/{sense}`, `example/{slug}/{sense}/1`, `/2`, and `note/{slug}/{sense}/collocation`, `/register`, `/common-mistake`. These yield 96 rows. Each retained meaning’s second example uses `example/retained/{existing-meaning-uuid}/2`, adding six rows. Store explicit IDs in the seed; never generate replacement IDs on reruns.

The existing seed uses `ON CONFLICT (id) DO UPDATE`. Identity/reference checks and the frozen addition-ID regression catch accidental duplicate keys and replacement identities. The opt-in `TestCanonicalSeedPostgreSQLRerunPreservesLearningReference` uses `VOCANOVA_TEST_POSTGRES_DSN`, one pinned connection and temporary copies of the six content tables with their original check, unique and foreign-key constraints. It applies the full seed twice, checks counts after each pass, verifies a synthetic saved-learning reference and review state in a surrogate temporary table remain unchanged, and checks the effective PostgreSQL curriculum order. It skips when no disposable test DSN is configured. This is bounded integration evidence, not a production migration or a full learner journey; source checks or mocked SQL alone do not prove database idempotency.

## Inventory and validation

| Entity              | Before | After |
| ------------------- | -----: | ----: |
| Situations          |      7 |     7 |
| Canonical words     |     39 |    51 |
| Meanings            |     42 |    54 |
| Examples            |     42 |    72 |
| Usage notes         |    126 |   162 |
| Journey memberships |     42 |    54 |

Daily Conversation has 18 meanings, 36 examples and 54 notes. The other six situations retain six meanings and six examples each. Difficulty assignments cover these eighteen meanings; unchanged metadata elsewhere must not be represented as assessed content.

`cmd/seed` regressions cover inventory, the ordered eighteen-item path, two distinct valid-length examples per meaning, nonduplicative explanations, notes, referential/unique-key integrity and frozen addition IDs. The seed-derived feedback regression loads real canonical targets through the repository and checks every example and approved natural form. Lexical target presence does not establish grammar, intended sense, safety or learning success.

The three-word sentence minimum is unchanged. Phrase variants remain whole-token sequences: arbitrary intervening words and other senses are not inferred by the matcher. Feedback retains responsibility for judging the selected sense.

Before release, verify repeated seeding and existing learning references against a disposable database, then Journey, saving, recall and sentence practice at 360px, 430px and desktop with keyboard and light/dark themes. These checks do not establish real-provider feedback quality, learner outcomes or certification. Record actual release evidence separately in the product readiness notes.
