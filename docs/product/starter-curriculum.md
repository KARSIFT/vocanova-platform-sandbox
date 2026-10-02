# Thirty-lesson starter curriculum

The starter course contains 30 guided lessons with three selected meanings each. It reuses 52 existing meanings and adds 38 original meanings, for 90 distinct meaning targets. It preserves the original seven lesson keys, definitions and version-1 contexts. The canonical seed contains 17 situations, 89 words/phrases, 92 meanings, 148 examples, 200 notes and 92 situation links.

This is a suggested sequence of short lessons, not a 30-day deadline or a claim of A2/B1 certification. Learners can choose a relevant situation and return to saved progress. Level labels are editorial guidance about the task and language complexity. Specialist workplace and study vocabulary is explained through concrete examples; no corpus frequency rank or measured learning benefit is claimed.

## Course order

Each lesson uses teaching, meaning recall and original contextual questions. Existing lessons are marked with an asterisk; their stable keys remain unchanged. New lessons have separate keys from their situation slugs, so one situation can support several lessons without replacing earlier progress.

| Order | Lesson                         | Situation              | Targets                                                     |
| ----- | ------------------------------ | ---------------------- | ----------------------------------------------------------- |
| 1     | Start a friendly conversation  | Daily Conversation     | greeting; small talk; casual                                |
| 2     | Find a time that works         | Daily Conversation     | weekend plans; available; sounds good                       |
| 3     | Suggest something to do        | Daily Conversation     | suggest; join; arrange                                      |
| 4     | Make a plan with a friend*     | Daily Conversation     | invite; confirm; reschedule                                 |
| 5     | Keep a plan on track           | Daily Conversation     | cancel; on time; meet up                                    |
| 6     | Stay connected                 | Daily Conversation     | catch up; keep in touch; farewell                           |
| 7     | Order and pay*                 | Restaurant             | menu; appetizer; bill                                       |
| 8     | Plan a meal out or at home     | Restaurant             | reservation (table); take-out; tip                          |
| 9     | Choose clothes that fit        | Shopping               | size; try on; fit                                           |
| 10    | Return a purchase              | Shopping               | receipt; refund; exchange                                   |
| 11    | Deal with a home problem       | Home and Renting       | rent; landlord; repair                                      |
| 12    | Find the right service         | Public Transport       | bus stop; timetable; platform                               |
| 13    | Buy a ticket and change trains | Public Transport       | single ticket; return ticket; change                        |
| 14    | Find your flight*              | Airport                | boarding pass; gate; luggage                                |
| 15    | Move through the airport       | Airport                | security check; customs; layover                            |
| 16    | Arrive at your hotel*          | Hotel Check-in         | reservation (room); front desk; key card                    |
| 17    | Use hotel services             | Hotel Check-in         | amenities; wake-up call; check-out                          |
| 18    | Receive a parcel               | Deliveries             | parcel; address; delivery                                   |
| 19    | Pay for everyday things        | Everyday Payments      | cash; bank card; payment                                    |
| 20    | Complete a simple form         | Everyday Services      | form; signature; queue                                      |
| 21    | Ask for health services        | Health Appointments    | appointment; symptom; pharmacy                              |
| 22    | Handle a phone call            | Phone Calls            | signal; voicemail; call back                                |
| 23    | Exchange an email              | Email and Online Tasks | attachment; link; reply                                     |
| 24    | Talk about your working day    | Everyday Work          | shift; break; colleague                                     |
| 25    | Prepare for an interview*      | Job Interview          | resume; references; salary expectations                     |
| 26    | Send a job application         | Job Interview          | cover letter; qualifications; follow-up (interview message) |
| 27    | Leave a meeting with a plan*   | Work Meeting           | agenda; action items; deadline                              |
| 28    | Share ideas and progress       | Work Meeting           | brainstorm; stakeholder; update                             |
| 29    | Get ready for your course*     | University Class       | lecture; assignment; syllabus                               |
| 30    | Work with other students       | University Class       | group project; office hours; presentation                   |

The two reservation targets use the existing restaurant and hotel meaning IDs. The general workplace follow-up meaning and the university-specific deadline meaning remain in discovery but are not additional starter targets: near-equivalent teaching is not counted twice merely to expand the course. Existing content is not deleted.

## Teaching and practice boundaries

The 38 new meanings each have two original examples and a usage note. Short definitions identify one sense; notes add a useful grammar, usage or regional distinction. Health, payment and renting lessons teach communication, not medical, financial or legal advice. Classroom contexts distinguish a shared assignment, an instructor's help times, and an audience talk. Workplace contexts distinguish topic lists, assigned tasks, deadlines, freely offered ideas, interested people and progress news.

Typed practice asks for the word or phrase from the lesson, not every possible synonym. Exact case/space-normalized answers and explicitly reviewed spelling variants are accepted. A different valid English word receives target-recall guidance rather than a claim that the English itself is wrong. New noun spellings include take-out/takeout, check-out/checkout, follow-up/followup and voicemail/voice mail. Regional equivalents such as parcel/package and queue/line are explained in notes; they are not silently added by fuzzy matching.

Listening uses the three distinct definitions from the selected lesson. The noun résumé has a pronunciation override; device speech requires an explicit learner action. Listening does not record or score the learner's voice. Completing guided or independent practice does not itself save a word, award SRS mastery, add reward points or count as a scheduled review.

Sentence target validation explicitly supports the new phrasal verbs try on and call back with their regular inflections. The approved clothing forms also allow it/them between try and on; call back allows me/you/him/her/us/them. These are fixed forms for exact target/type/POS tuples. They do not permit arbitrary gaps, unrelated words, or general separable-verb matching. The new object-pronoun forms cannot cross punctuation between their words. Target presence still does not establish correct grammar or intended meaning, and the existing three-word sentence minimum remains.

## Stable content and saved sessions

New IDs are UUIDv5 using the standard URL namespace and prefix `vocanova/starter-30-v1/`. Frozen semantic keys are `situation/{slug}`, `word/{slug}`, `meaning/{slug}/{sense}`, `example/{slug}/{sense}/{1|2}`, `note/{slug}/{sense}/usage`, and `journey/{situation-slug}/{word-slug}/{sense}`. The 38 sense keys are pinned in `cmd/seed/starter_curriculum_test.go`. IDs are stored explicitly in the seed; wording and display-order changes do not regenerate them.

The 400 pre-expansion rows retain their IDs and content. Tests retain the complete historical ID inventory and verify new identities, references, allowed metadata and the original seven serialized lesson definitions. Existing saved-learning references therefore continue to point at the same content. The seed continues using its existing transactional upsert by primary key; repeated seeding is checked against a disposable database, not inferred from JSON alone.

New independent practice snapshots use content version `starter-90-v2`. Existing `starter-21-v1` snapshots remain readable and actionable with their original prompts and answers. The grading contract remains `exact-recall-v1`: the expanded catalog supplies explicit accepted forms to each new snapshot. Unknown content or grading versions still fail closed. Resuming or replaying an old action returns its saved session's current state; it does not rebuild the session from the expanded catalog.

## Verification scope

Catalog tests cover all 30 lesson triplets against the real seed, distinct meaning IDs, authored context fields and exactly one keyed correct option per question. Practice tests cover all 30 typed/listening groups, literal accepted/rejected spellings, private answer projection and historical snapshot compatibility. Seed-derived feedback tests pass every shipped example through the actual target loader and deterministic sentence validator. These checks do not establish linguistic correctness or learner benefit on their own; independent editorial review remains necessary.

The opt-in PostgreSQL checks exercise repeated seeding, preserved saved references, all guided lessons, independent practice, and persisted version-1 read/list/replay/answer/continue behavior. They require a disposable migrated database. Browser checks verify representative learner flows, not every content judgment. Release acceptance and exact integration results belong in the current release-readiness record; this curriculum document makes no live deployment or pilot-outcome claim.
