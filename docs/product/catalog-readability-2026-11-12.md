# Catalog readability and meaning coverage — 12 November 2026

The source audit covers all **10,000 ranked entries** and all **10,051 canonical
meanings** represented by the checked-in seed sources. A targeted semantic pass
reviewed **320 distinct candidates** and supplies **257 original plain-English
explanations**, while retaining 63 already-authored explanations. The visible
explanation changes; the selected meaning, grading identity and approved media do
not change.

## Audit method and actual coverage

The deterministic audit reads the frozen ranked catalog, original curriculum,
reviewed supplemental uses, historical WordNet alias metadata and the new
explanation overlay. It makes no database or network calls and assigns no CEFR
rating. Run from the repository root:

```sh
go -C apps/api run ../../scripts/content/audit-readability.go -root ../.. > /tmp/vocanova-readability-audit.json
```

The JSON contains a row for every canonical meaning: identity/revision, rank,
source, definition word count, punctuation flags, repeated-definition count,
available teaching senses/parts of speech, historical WordNet sense membership,
and explanation coverage. Word count and punctuation are review aids, not
measures of learner understanding or proof that a definition is unsuitable.

| Source                        | Ranked entries | All canonical meanings |
| ----------------------------- | -------------: | ---------------------: |
| Vocanova original curriculum  |             60 |                     92 |
| Vocanova editorial core       |          3,288 |                  3,288 |
| WordNet reference definitions |          6,652 |                  6,652 |
| Vocanova supplemental uses    |              0 |                     19 |
| Total                         |         10,000 |                 10,051 |

The ranked inventory contains 148 definitions longer than 20 words and 172
short comma/semicolon candidates of at most five words. Their union, including
the reported “abandon — forsake, leave behind” example, is 320 entries. Each
candidate received an individual same-sense decision; punctuation by itself was
not treated as a defect. The four long original editorial explanations for
pinball, queer, electroshock and bondsman were deliberately retained, with
individual reasons in the [review manifest](catalog-readability-review-2026-11-12.json).

There are 257 new explanations for reference definitions. Examples include:

| Word    | Existing canonical definition                                                     | New display explanation                                                                                           |
| ------- | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| abandon | forsake, leave behind                                                             | To leave someone or something behind.                                                                             |
| virus   | A technical description beginning “(virology) ultramicroscopic infectious agent…” | A tiny germ that multiplies only inside living cells and may cause disease.                                       |
| protein | A 46-word chemical description                                                    | A substance found in living things and foods such as meat and beans that helps bodies grow and repair themselves. |

The independent pass checked all 257 proposed originals against the source sense
and part of speech and assessed all 63 retention decisions. It made six targeted
copy corrections before acceptance. The original side-by-side review file SHA-256
was `dfe05b0ad054002b2ceca2eb635119a38aa097fac99c87f104afa62c251eeb30`.
The committed manifest records the same final copy and explicit decisions. This
is model-assisted editorial review, not a human certification that every word in
the wider inventory is understandable to every A2–B1 learner.

## Meaning and media consistency

The original `shortDefinition`, word/meaning IDs, part of speech, ordering,
source-sense selection and semantic alias metadata remain unchanged. New text is
an optional `learnerDefinition` display projection looked up only when its exact
canonical revision matches. Existing authored explanations win when already
provided by the Word Detail service. Unknown or changed identities fall back to
the canonical definition. No migration, catalog reseed, media regeneration or
learner-state write is needed.

Search, the word map, Discovery, starting self-check, saved words, Home, lists,
situations and word details can display the optional plain explanation. Word
details keep the source under “Reference definition”. Review preserves its
canonical options/answers and shows an explanation after revealing the answer.
Existing authored learner-definition fields are still fetched by Word Detail;
other projections add only the exact reviewed reference overlay and do not add
new database joins or claim to expose every authored explanation everywhere.
Authored lesson/story/practice snapshots keep their existing reviewed teaching
text and immutable grading payloads.

## Polysemy findings

The audit distinguishes available teaching senses from historical dictionary
membership. The checked-in canonical source has 37 meanings on words with more
than one available teaching sense. Its supplemental catalog already adds 19
uses across 12 words, preserving separate meaning IDs, parts of speech and
examples. Word details and comparison links expose the actual available senses;
no additional meanings are invented from spelling, synonym lists or an image.

The historical WordNet sidecar reports multiple sense memberships for 7,745 of
the canonical meanings' word forms. These older reference senses include uses
that may be rare, inappropriate for this product or already represented by other
forms. This is an editorial inventory aid, not a requirement to import every
reference sense. A word with one current teaching sense is not claimed to have
only one meaning in English. Adding a genuine new use still needs an authored
meaning, distinct stable identity, examples, semantic review and media acceptance.

## Verification and limits

Focused tests verify exact-revision lookup, authored precedence, stale-source
fallback, unchanged canonical revisions in API projections, map display context,
source-object isolation and empty-list array preservation. UI helper checks cover
whitespace/case deduplication and canonical fallback. Browser acceptance adds
read-only lazy word details with an authored explanation absent from the map
card; root records the combined browser/build results with the delivery PR.

The deterministic full report's SHA-256 is
`0a4acc9ae34fdda0249aa7679a122d2023a306223fee6ac1e6f4def4ce98fb43`.
Its input checksums include the unchanged ranked catalog
`7d7b12577b7f00653f8be87d773a5a9c6e85c1639ff87b14db5c4a497e4e6c63`,
unchanged alias sidecar
`e70d8f957c0b2694db758e957e96964632ffd844fc0ad124075c4074aa1fd8b8`,
and reviewed explanation data
`35a3a6edacd7976a2f53144e98ff4e4d34439a2d6d6d892fa930cd815fb428a3`.
The script reproduces the row-level report; local private working artifacts are
not published as learner data.

All ranked entries have mechanical audit coverage. The 320 identified candidates
have targeted semantic decisions; the remaining 9,680 ranked entries have not
received a new full human readability or sense-completeness certification in
this delivery. Observed A2–B1 learner testing remains the evidence needed to
assess comprehension rather than infer it from a lookup or readability metric.
