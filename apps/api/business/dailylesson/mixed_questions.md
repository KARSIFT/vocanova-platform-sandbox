# Reviewed daily-question sets

The allowlist in `mixed_questions.go` is intentionally narrow. It supplements
text questions with existing reviewed authored content; it is not a heuristic
that turns any example or available image into a question. Definitions and
headwords must match pinned meaning revisions, and examples must match the
exact active authored sentence. The semantic alias check applies to every pair.

The picture set is parcel, menu, key card and appetizer. The committed reviewed
pictures show a wrapped postal parcel, a restaurant menu, a card unlocking a
hotel door, and a small food dish. The options represent distinct physical
objects/scenes; none is another sense or synonym of the target. Every option
requires an approved `meaning_image_assets` publication for that exact revision.
All four equivalent definitions and immutable image hashes/URLs are frozen with
the opaque option IDs. Missing/stale/unapproved metadata falls back to text.
The existing reviewed manifest and WebP bytes are used by the PostgreSQL tests;
this verifies actual catalog eligibility without asserting live publication.

The gap sets are:

- “Our flight leaves from _____ twelve.”: gate, syllabus, luggage, appetizer.
  A numbered airport boarding location is the only suitable choice.
- “My _____ stopped working.”: key card, syllabus, luggage, appetizer.
  Only the electronic door key functions as a device that can stop working.

These sentences are copied exactly from the existing authored seed; no new
examples or provider generation are introduced. Each set freezes four word
options. Headword, slug and pronunciation identity are omitted before grading.
The original headword, slug, definition and meaning revision are returned with
confirmed feedback for post-answer teaching. A changed example, definition,
headword or missing candidate disqualifies the whole set. For a word eligible
for both picture and gap, its position in the frozen session alternates the
preferred kind. Existing snapshots and action receipts never reroll.

Content coverage is deliberately bounded: four eligible picture meanings and
two gap meanings, one shared. Other catalog meanings use four text definitions.
This is production catalog content coverage, not evidence of learner efficacy.
