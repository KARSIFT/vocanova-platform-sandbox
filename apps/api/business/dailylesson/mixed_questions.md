# Daily question variety

New sessions draw a private random question kind once and persist the resulting
question, opaque answer ID, alternative order and teaching content in the session
snapshot. The weights are picture 60, sentence gap 30 and text definition 10,
renormalized over modes actually eligible for the target. With pictures but no
reviewed sentence, pictures receive 60/70 of draws. Reloads and action/start
receipts never draw again. Existing sessions keep their original questions.
No live provider, image generation or external lookup is involved.

## Picture eligibility

Pictures use active canonical words and meanings across the existing approved
`meaning_image_assets` catalog, including meanings outside the former four-word
allowlist. Each candidate must have a valid approved publication whose revision
matches its current ID, spelling, part of speech and definition. The target is
queried separately; a missing or stale target image cannot be substituted.

Each construction looks up the exact expected target revision using the unique
meaning/revision index, returning at most one publication. It returns at most 64
alternative publication rows ordered by indexed meaning identity, restricted to
the same part of speech. It never sorts the full catalog randomly or loads its
10,000 rows into a request. A mutually compatible subset of at most 16 candidates supplies three
privately sampled distractors. Four unique meaning IDs, headwords, definitions
and image hashes are required. Shared catalog source senses and all-sense
WordNet aliases/near hypernyms disqualify pairs. The existing semantic check is
conservative but is not a guarantee that every image or distractor is ideal.

The existing approved immutable image URLs and metadata are frozen with the
options. Missing, invalid, unapproved, stale, duplicated or insufficient image
content excludes picture mode; an eligible reviewed gap or text remains usable.
A bounded pool can conservatively miss usable alternatives elsewhere in the
catalog. This is preferable to an expensive unbounded transaction scan.

## Reviewed sentence gaps

A lexical replacement cannot establish that a sentence has only one sensible
answer. The ranked corpus intentionally omits unreviewed WordNet examples.
Consequently sentence mode uses twelve reviewed sets copied exactly from the
existing authored seed, rather than converting arbitrary catalog examples.
`mixed_questions_gaps.go` pins all candidate meaning revisions and requires the
exact authored example to remain active. Every option is a noun; pairs also
pass the semantic alias and distinct-content checks. A single complete,
case-insensitive headword occurrence is replaced. Partial words, possessives,
hyphenated fragments and repeated complete occurrences are rejected.

| Target | Authored sentence | Alternatives | Contextual distinction |
| --- | --- | --- | --- |
| gate | Our flight leaves from gate twelve. | syllabus, luggage, appetizer | A numbered airport departure location. |
| key card | My key card stopped working. | syllabus, luggage, appetizer | An electronic access device can stop functioning. |
| syllabus | The reading list is in the syllabus. | gate, refund, appetizer | The course document contains a reading list. |
| appetizer | We shared an appetizer before dinner. | gate, syllabus, key card | A small food course precedes the main meal. |
| platform | Our train leaves from platform three. | syllabus, luggage, appetizer | A numbered train departure location. |
| refund | The shop sent my refund to my bank account. | gate, syllabus, appetizer | Returned money can be transferred to an account. |
| cash | Can I pay in cash? | gate, syllabus, appetizer | A form of payment in the fixed expression “pay in”. |
| form | Please fill in this form with your name and address. | gate, luggage, appetizer | A document with fields to complete. |
| size | Do you have this shirt in a smaller size? | gate, syllabus, appetizer | A clothing measurement that can be smaller. |
| voicemail | Please leave a voicemail if I do not answer. | gate, syllabus, appetizer | A recorded message when a call is unanswered. |
| signature | We need your signature at the bottom of the page. | gate, luggage, appetizer | A written name placed on a document. |
| receipt | The shop emailed me a receipt after I paid. | gate, luggage, appetizer | A purchase record sent after payment. |

These are original authored sentences with reviewed alternatives, not generated
examples. Changing any pinned meaning or exact example disables that gap set.
Before grading, gaps expose neither the target headword/slug nor pronunciation
meaning identity or image metadata. Correct options are opaque IDs. Confirmed
feedback restores the frozen original meaning for teaching. Wrong answers still
count toward review effort without granting confidence points.

## Verification limits

PostgreSQL tests use the original four committed WebP files and manifest for
baseline publication eligibility. Broader `debris` and `wit` regressions use real
frozen catalog definitions with explicitly synthetic publication metadata;
they prove catalog-wide selection and grading, not live image delivery or visual
quality. Deterministic private entropy exercises all three kinds without flaky
probability assertions. Content failures, all twelve gap sets, frozen reloads,
start receipts and wrong-answer effort/credit are covered separately.
