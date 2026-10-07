# Bundled English dictionary

This is a read-only excerpt of WordNet 3.0, not Vocanova lesson content. No
dictionary entry receives a canonical ID, saved state, review credit, or a
generated phonetic transcription. Coverage is broad but not comprehensive;
modern slang, some inflections, accented words and multiword phrases may be absent.
Definitions are source dictionary text, not graded A2–B1 explanations.

Source: the [official NLTK corpus distribution](https://github.com/nltk/nltk_data/blob/gh-pages/packages/corpora/wordnet.zip),
[version metadata](https://github.com/nltk/nltk_data/blob/gh-pages/packages/corpora/wordnet.xml),
and [WordNet license](https://wordnet.princeton.edu/license-and-commercial-use).
Retrieved 2026-10-07. This packaging does not imply endorsement by the source authors.

- Original ZIP: 10,775,600 bytes.
- Original SHA-256: `cbda5ea6eef7f36a97a43d4a75f85e07fccbb4f23657d27b4ccbc93e2646ab59`.
- Derivative gzip SHA-256: `814ced411a9fe144c21920ae4b39054ef8c2b047309b40288707eba7d877f091`.
- Derivative: 2,843,007 bytes compressed; 8,678,211 bytes JSON.
- 82,710 single-word headwords; 4,820 exception forms; 80,555 selected senses.

Regenerate with Python's standard library only:
`python3 generate.py /path/to/wordnet.zip`. The generator checks the pinned ZIP
checksum, follows source index sense order, keeps at most two senses per part of
speech, excludes entries outside the submitted-word grammar, and separates only
terminal quoted examples from definitions. Definitions and selected examples are
not rewritten. Gzip has no timestamp or filename, making output deterministic.
Exact entries take precedence; otherwise source exception forms and the
[documented WordNet single-word suffix rules](https://wordnet.princeton.edu/documentation/morphy7wn)
may resolve a known base form. The response shows that base form explicitly.
These rules are not a spelling checker and cannot establish that every supplied
inflection is conventional English.

The complete original notice is retained in `LICENSE.wordnet` and inside the
compressed data, and is returned with every dictionary response. The API loads the
immutable index once at startup; lookups perform no external requests or database
operations. No source provider code or runtime library was copied.

The earlier Free Dictionary API candidate was rejected after narrow local probes
timed out; it is not a runtime dependency or fallback.
