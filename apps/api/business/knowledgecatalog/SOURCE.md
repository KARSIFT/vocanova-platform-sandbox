# Ranked English knowledge catalog

The `english-10000-v1` catalog contains 10,000 distinct single-word English
forms, ordered into 100 bands of 100. Entries are real canonical learning
meanings, not placeholders or dictionary-only search results. Existing authored
curriculum content keeps all 89 word IDs, 92 meaning IDs and teaching text.
Authored phrases outside the ranked single-word selection remain available
through lessons and search.

## Sources and attribution

- Frequency counts: Hermit Dave's [FrequencyWords](https://github.com/hermitdave/FrequencyWords),
  English OpenSubtitles 2018 list. Content is CC BY-SA 4.0; the upstream MIT
  license applies to code. No upstream generator code is copied here.
- [Pinned frequency file](https://raw.githubusercontent.com/hermitdave/FrequencyWords/072bbed282316a23651aa7068c7173aa7898cf80/content/2018/en/en_50k.txt):
  50,000 rows, 622,749 bytes, SHA-256
  `5351ff405b1126ef555791dd4d9798a48e3e9a501a9fc481a9da957752cfb458`.
- Reference senses: Princeton WordNet 3.0, from the official
  [NLTK corpus distribution](https://github.com/nltk/nltk_data/blob/gh-pages/packages/corpora/wordnet.zip).
  Original ZIP: 10,775,600 bytes, SHA-256
  `cbda5ea6eef7f36a97a43d4a75f85e07fccbb4f23657d27b4ccbc93e2646ab59`.
- `editorial-core.json` contains original Vocanova definitions, including grammar
  missing from WordNet and reviewed corrections of unsuitable primary senses.
  They are labeled editorial content, never presented as verbatim WordNet text.

Frequency data and the derivative ranked selection/editorial additions use
[CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/), with full credit,
modification notice and license links in `LICENSE.frequency`. WordNet retains its
own complete `LICENSE.wordnet` notice, also embedded in the gzip and returned by
`SourceAttribution()`. Source definitions are distributed with this attribution
in the knowledge-map response and learning views. No source endorsement is implied.

## What the ordering means

Ranks follow descending frequency of the exact selected word forms in the
pinned subtitle list. Filtering creates gaps in source ranks; map ordinals are
contiguous from 1 to 10,000. The order is not a personalized relevance score,
contemporary population-wide frequency measurement, CEFR curriculum, or ranking
of individual senses. Some inflected forms are separate entries. Ambiguous
word forms can combine name and ordinary-word occurrences in source counts.
The inventory is broad, but it is not all English vocabulary.

The generator reads original case-preserved WordNet lemmas, filters unsuitable
name/abbreviation and low-support inflection collisions, and selects the
highest historically tagged eligible sense across parts of speech. Unreviewed
WordNet candidates require a positive tagged count. Princeton cautions that
[sense ordering is based on old concordances and is not an accurate indicator
of current usage frequency](https://wordnet.princeton.edu/documentation/cntlist5wn).
This is a selection heuristic, not evidence that a definition is pedagogically
appropriate or factually current.

The explicit exclusions and source-text gates remove default profanity/slurs,
marked offensive/archaic senses and suspect numeric factual glosses. These are
conservative inventory filters, not an accuracy guarantee; spelled-out obsolete
claims can survive mechanical checks. Independent semantic review and original
editorial corrections are therefore required alongside structural validation.
The broader reference tail is not claimed to have graded A2–B1 teaching text.

All unreviewed WordNet examples and quoted example tails are removed. Source
examples can describe another synonym, dated factual claims or unsafe actions.
Examples are optional; original authored curriculum examples are preserved.

## Version 1 content review

All 10,000 selected entries were individually inspected in independent model
review batches, including the 14 replacement entries introduced by exclusions.
The frozen selection has 60 existing authored entries, 3,288 original editorial
entries and 6,652 retained WordNet reference definitions. All 89 authored words,
92 meanings and their examples remain in the canonical seed, including material
outside this selection. This review is not a human certification or a promise
that every reference definition is suitable for a beginner.

The frozen catalog gzip is 999,343 bytes, SHA-256
`7d7b12577b7f00653f8be87d773a5a9c6e85c1639ff87b14db5c4a497e4e6c63`.
The independent all-sense alias sidecar is 677,585 bytes, SHA-256
`e70d8f957c0b2694db758e957e96964632ffd844fc0ad124075c4074aa1fd8b8`.
Changing version 1 membership or order requires an explicit migration; do not
replace this frozen inventory with a differently ranked file during deployment.

## Reproduction and identity

Run with Python's standard library and the two local source downloads above:

```sh
python3 apps/api/business/knowledgecatalog/generate.py /path/to/wordnet.zip /path/to/en_50k.txt
```

Both input hashes are enforced before parsing. The generator makes no network
requests, uses explicit deterministic ordering, and writes timestamp-free gzip.
Its output reports the derivative SHA-256, size and last source rank.

Existing authored word/meaning IDs win. New word IDs are UUIDv5 under the fixed
`https://vocanova.site/catalog/wordnet-3.0` namespace with `word:en:<word>` names;
new primary meaning IDs use `primary-meaning:en:<word>`. Rank and source-selection
changes do not change these IDs. Original WordNet sense keys remain separate
provenance metadata. After release, a genuine change of meaning requires an
explicit new meaning/version and data migration; routine wording corrections
must not silently retarget a learner's meaning.

The normal canonical seed command installs this public catalog in the same
transaction after the authored seed. Catalog-owned wording can be corrected on
rerun; authored rows and learner state are not overwritten. No runtime external
lookup, model download, image generation or corpus decompression is needed for
map reads. The gzip is a server seed asset, not a browser bundle dependency.

## Learner-state boundary

Map reads use a repeatable-read snapshot for band counts and the selected 100
items. A missing/incomplete seed is an error, never a fabricated complete map.

Discovery decisions lock the active learner before user-word rows, check an
opaque state revision and preserve private notes. Learn clears the self-assessment
and saves/resumes the same meaning. Already known records a self-assessment and
pauses an existing review row with `ignored` status, preserving review history
and counters. Neither choice creates review attempts, graded mastery, mission
credit or confidence points. Idempotency receipts store only namespaced request
fingerprints in the existing privacy-managed table; replay returns current state
and cannot restore an obsolete assessment.

## Multiple-choice alias guard

`semantic-aliases.json.gz` is generated from all original WordNet senses of the
current catalog and authored curriculum words, plus their lexical aliases.
It is independent of the selected teaching sense and editorial provenance.
`SemanticAliases` conservatively excludes candidate distractors sharing any
synset or a direct hypernym ancestor within two steps in either direction,
with case, phrase separator and source inflection normalization. Regular noun
plurals outside the inventory are resolved against known source noun senses.
Shared generic ancestors are not compared. This
prevents pairs such as baggage/luggage, bags/luggage and answer/reply from becoming a correct
answer and a supposedly wrong option merely because their definitions were
edited independently. Metadata is loaded once; no model or network call is used.

This deliberately over-excludes some polysemous words. It cannot prove that all
remaining options are semantically different; authored option review remains
necessary. Unknown spellings are only compared by normalization. The original
WordNet notice is retained inside this sidecar too.
