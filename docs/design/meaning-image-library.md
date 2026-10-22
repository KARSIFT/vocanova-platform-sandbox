# Meaning image library

Implementation checkpoint: 23 October 2026. This document describes the delivery
branch and verified storage work; it does not establish full image coverage or a
staging/production deployment.

## Learner experience

Pictures support the exact meaning and grammatical use alongside its short
definition, rather than illustrating a word spelling alone. They appear in teaching/discovery/detail contexts, not in
graded questions where the scene could reveal the answer. Missing or failed
pictures leave the learning text usable. Existing reviewed local pictures remain
a revision-checked fallback.

Preserve the established landscape picture frame while varying photographs,
illustrations and simple diagrams. Vary people, objects, environments, color and
visible action. Every candidate requires visual review for the exact definition,
clarity at mobile size and useful alt text. A successful generator response is
not approval. Candidate prompts, scenes, styles and generation/rejection receipts
are retained in private working storage; rejected attempts are preserved.

Do not print redundant target-word labels or explanatory captions inside pictures.
Only deliberately reviewed short semantic cues needed to disambiguate the scene
are allowed, such as `60 s` for a minute or a name when the meaning requires one.
The previous cat-label artwork was withdrawn from final delivery after its
text-free replacement passed validation.

## Content identity and delivery

The source union contains 10,051 meanings: 10,000 primary map entries, 32 additional
curriculum meanings and 19 supplemental uses. The initial queue had 92 reviewed
pictures and 9,959 missing pictures. A deterministic exporter preserves those
identities and exact source definitions; it does not invent ready placeholders.

Each meaning revision is the shared SHA-256 identity of its ID, canonical spelling,
part of speech and short definition. Public asset paths are immutable:

`meanings/{meaningId}/{meaningRevision}/{contentSha256}.webp`

The stable public origin is `https://images.vocanova.site`. Storage remains
portable: the same keys and metadata work with another S3-compatible store behind
that hostname. Cloudflare R2 currently stores only optimized public WebPs.
Original native files, reviewed legacy sources, rejected attempts and historical
receipts are retained in a portable private archive on the owner's computer.
The former originals and walkthrough R2 buckets were removed after verified local
preservation. Earlier cloud-original checkpoints below are historical and do not
describe currently existing buckets. Native generation files and rejected attempts
never enter the public image bucket.

The public image host has a dedicated Cloudflare cache rule for `/meanings/`
paths ending in `.webp`. Only successful 200/206 responses receive a one-year
edge TTL; 300–599 responses use no-store. Browser caching respects the origin
instead of inheriting that edge TTL. The expression includes GET, HEAD and PURGE
so single-file cache purges still match. This rule does not change application
or API caching. Verify successful CDN checksums and MISS-to-HIT
behavior after a single-image purge, plus repeated 404 cache bypasses; the rule's
configured TTL alone does not prove delivery behavior or whole-site performance.
The later REST public objects carry the one-year
`public, max-age=31536000, immutable` browser header, verified on actual CDN
responses, including every image in the latest 869-image addition.
Historical MCP objects and the first five REST uploads retain
their original browser metadata; the new header is not a claim about every asset.

The API returns approved metadata only when the meaning and canonical word are
active and the current revision matches. Discovery/detail embed image metadata
in existing reads. Other teaching steps coalesce lookups into authenticated batches
of at most 100 meaning IDs. There is no HTTP image-write route and no learner
progress mutation. Optional metadata reads have a one-second database deadline,
including pool waits; shorter caller deadlines still win. Frozen teaching steps require their exact meaning revision;
old pictures cannot silently explain a changed use.

A pending picture keeps its frame and a short loading status. After eight seconds
pending while visible, an explicit Retry picture control becomes available.
Normal image URLs stay unchanged; manual retry adds a fresh nonsecret
`vocanova_retry` query on the same trusted origin and path. Stored canonical
metadata is unchanged and no automatic retry request is issued. The frame stays
stable. Its named focus destination persists through late success, manual retry,
external-to-local fallback and terminal unavailability. If all known valid sources
fail, a terminal placeholder offers another explicit retry with a fresh query.
Initially missing or revision-ineligible images still render no frame. Cached
loads and late/stale events cannot restore obsolete loading state.

## Preparation and metadata import

New originals are fully decoded, orientation-corrected and fitted inside 960×640
without cropping or enlargement, then encoded as bounded WebP. Existing reviewed
WebPs are preserved byte-for-byte. The legacy publication plan identifies these
as `reviewed-legacy-webp-preserved`; it does not claim native originals unavailable.

An operator CLI imports verified public metadata in one database transaction.
Statements use batches of 500, retain stable lock ordering and check active
canonical revisions. Exact retries are no-ops; conflicting publication metadata
fails and rolls back the entire manifest. One forward migration adds the media
asset table; previous migration history and learner data remain unchanged.

The publication scripts and schemas are documented in
[the tooling guide](../../scripts/meaning-media/README.md). Working queues and
receipts stay outside public commits. Backend import checks live content again;
an exported queue alone cannot establish database freshness.

Code deployment does not import the private selected manifest automatically. The
final staging release must explicitly import its verified metadata into the
migrated target database, then confirm authenticated API readback and learner
appearance. Local SQL rehearsal and public object uploads are separate evidence.

## Separate verification paths

The current local-original workflow uploads only optimized public images. Its
explicit local archive mode independently reads each canonical original file,
checks full SHA-256 and MD5, verifies decoding, and rejects symlinks, nonfiles or
paths inconsistent with the exact meaning/revision/source hash. Public object
receipts must match the prepared WebP; a fresh canonical CDN GET and full SHA-256
check precede each import-ready record. The separate local manifest and evidence
contain no private-cloud verification claim or private-cloud request. Historical
evidence remains unchanged. See the tooling guide for the explicit mode and
configurable archive root.

The explicit REST receipt mode keeps its manifest and evidence separate from
MCP. It distinguishes a successful new PUT from reconciliation of an existing
object with matching content and metadata. It does not claim an existing-object
readback was a new upload. Both modes retain the same canonical identity, decoded
source, local archive and fresh public-byte verification gates. Publication
uses frozen content-addressed keys and bytes; conflicting existing metadata stops
the operation. Application-level same-content replay does not establish atomic
conditional creation by the REST service.

The optional legacy S3 publisher checks original/public bytes through
authenticated GETs and
requires its explicit private no-store and public immutable cache headers. It
verifies canonical public delivery before emitting its import-ready manifest.

The legacy cloud-original Cloudflare MCP receipt path validates the operator's
successful raw PUT receipts against actual local source SHA-256, size and
single-part ETag/MD5.
It then fetches the public image from the canonical origin and verifies SHA-256.
Its separate manifest/evidence records the actual public response headers for
each verification; browser headers and edge cache policy are separate evidence.
It does **not** establish an independent private GET SHA-256 check or the
S3 publisher's immutable/no-store metadata. These stronger S3 gates are unchanged.

All 92 legacy native PNG originals were recovered through byte-identical
reproduction of their current WebPs, not receipt order or first approval. Their
177,279,634 bytes are locally SHA-verified, with exact receipt and transform
provenance. A separate portable native-original archive manifest links each
public content hash to the native archive key. Historically, all 92 native PNGs
were uploaded and a private API listing matched each size and single-part MD5
ETag, with source SHA-256 checked before PUT. That historical check was not an
independent private GET SHA-256 or restoration test. These cloud copies and their
portable index have since been preserved locally and retired from R2; this upload
checkpoint is stale as a description of current storage. The already verified WebP
publication receipts remain valid. Seven rejected native attempts are
preserved separately and excluded from the approved upload plan.

## Acceptance checkpoint

The 23 October local checkpoint contains **1,939 selected published meanings** and
**1,939 prepared meanings** out of 10,051 canonical meanings. All prepared IDs are
in the selected delivery manifest; **8,112 meanings still need prepared artwork**.
The latest 1,000-attempt cloud batch retains 997 saved originals and complete
accessible returned receipts locally, two refused outcomes and one uncertain
submission boundary. Producer and author-disjoint full-image reviews originally yielded 805
pass/pass candidates and 192 held images. All 805 pass/pass candidates are
published after local approval and fresh CDN full-byte checks. They add
75,454,550 public bytes; their 2,015,312,954 native bytes remain local. A later
feedback-informed full-native review of ten held images approved seven unchanged
originals. Their original flags remain recorded alongside the new adjudications;
185 held images remained unresolved at that checkpoint. Those seven add 797,074
public bytes, with 19,369,832 native bytes kept local. Actual upload/readback receipts, fresh CDN
full-byte checks and independent selected-metadata checks pass. All previous
1,875 record fields and the three corrected replacements are preserved.
A subsequent root and independent full-native review of ten further held images
approved eight unchanged originals to accompany their exact short definitions.
The source flags and differing image-only judgments remain recorded. Division
and normally stayed held; 177 held images remained unresolved at that checkpoint.
The eight add 947,138 public bytes with 22,516,317 native bytes kept local. Actual
uploads, fresh canonical CDN full-byte checks and independent selection checks
pass; all previous 1,882 fields and row order plus three replacements survive.

A further feedback-informed root and independent review individually inspected
twenty full-native originals. The independent findings preserve eighteen candidate
recommendations and two holds; root approved nineteen, retaining the **someplace**
backpack-marking disagreement and uncertainty. **Personality** stays held for
word-like background menu marks. The **term** image's intrinsic **diameter** label
is a necessary semantic cue under the existing exception, not a target-word caption.
The nineteen add 1,797,792 public bytes while 47,957,438 native bytes stay local.
Actual upload/readback receipts, fresh canonical CDN full-byte verification and
independent selected-metadata verification pass; all previous 1,890 fields and
order plus the three replacements remain unchanged. That checkpoint contributed
839 published images totaling 78,996,554 public bytes, with 2,105,156,541 native
bytes local across thirteen receipt collections; 158 saved images remained held.

A subsequent feedback-informed review individually inspected thirty full-native
originals. Independent findings retain twenty-six candidates and four holds;
root approved twenty-five after its own thirty full-native views. **Cast**,
**cure**, **sin**, **charm** and **net** remain held. The additional root **sin**
hold preserves the independent disagreement over an ambiguous offering-plate
scene. All original flags, prompt misses and differing judgments remain recorded.
The **percent** diagram correctly shows 22 of 110 cells (20%), despite missing
the requested 100-cell grid. The **profit** diagram correctly shows 10 minus 8
equals 2; accounting roles come from the accompanying short definition, with no
detailed accounting-period claim.
These are reviewed teaching contexts, not image-only graded questions.
The twenty-five add 2,107,932 public bytes while 60,407,933 native bytes stay local.
Actual PUT/fresh object GET receipts, separate canonical CDN full-byte checks,
independent selected-metadata verification and root activation/readback pass.
Every previous 1,909 field, row order and top-level metadata is unchanged,
including all three replacements. That checkpoint contributed 864 published
images totaling 81,104,486 public bytes, with 2,165,564,474 native bytes local
across fourteen receipt collections; 133 saved images remained held.

The next thirty-original review ultimately approved **curious**, **boom**,
**cab**, **magazine** and **newspaper**. Root's preliminary fourteen approvals
became six after the historical flags and independent findings were considered,
then five after actual app-size review. All earlier judgments remain retained.
**Literally** stays held despite its independent native-size recommendation:
the necessary intrinsic sentence is too small for comfortable reading in the
actual 276×184 CSS-pixel frame at 360px in both themes. Root's thirty original
views, same-original reinspection and four subsequent app views are distinct.
The failed mobile gate superseded the six-candidate approval before canonical
preparation or upload; no redraw, zoom or custom enlargement was used.
The five add 619,628 public bytes while 13,666,930 native bytes stay local.
An initial partial upload and explicit operator reconciliation are preserved;
final receipts distinguish three existing-object reconciliations from two new
uploads. Separate canonical CDN full-byte checks, independent selected-metadata
verification and root activation/readback pass. All previous 1,934 fields, row
order and top-level metadata plus the three corrected replacements are unchanged.
This batch now contributes 869 published images totaling 81,724,114 public bytes,
with 2,179,231,404 native bytes local across fifteen receipt collections;
128 saved images remain held.
Unresolved held, refused and uncertain outcomes add no approved coverage.

The next subscription-only cloud run reports 991 saved images, seven refusals
and two uncertain submissions, with 991 producer and 991 independent reviews
and 154 disagreements preserved. Its 58 native packs and two final metadata ZIPs
total 2,341,067,929 bytes. None matched the two local download folders in the latest
local download scan; transfer, local intake and approval remain pending. These cloud
results add no prepared or selected coverage. Another batch stopped at its disk
floor with the executor unavailable: 848 saved, five refused, three uncertain and
144 untouched. Its last verified independent-review count is 344; 796 originals
are exported and 52 saved originals plus new reviews are unexported. Its final
reply has complete descriptors for two metadata ZIPs and links to 48 native packs
without their individual sizes/hashes. Those links do not establish transfer.
A further batch stopped at its disk floor; another reservation remains idle.
Capacity guards are unchanged.

The ten-image cloud pilot is included after independent visual review, verified
local original archival and fresh CDN-byte checks. The first 100-image cloud
batch retains its original review results: 95 pass, three flag and two reject.
Forty-four approved first-attempt images are included after verified local native
archival and fresh CDN checks; 51 approved first-attempt candidates remain
unprepared. All five original held decisions remain unchanged. The incremental
receiver verifies already-prepared batch entries against their exact reviewed
native and deterministically reproduced WebP instead of accepting overlapping
identities alone.
The explicitly bound second attempt for **admit** is also included after root
and independent full-image review, verified local archival, preparation,
publication and separate canonical CDN-byte verification. Its rejected first
attempt remains preserved and was never public. Four of the originally held
meanings remain unresolved.

The second 100-meaning batch is fully received and reviewed locally. All 100
originals, totaling 249,984,310 bytes, are archived on the owner's computer,
including the six held attempts. All 15 supplemental transfer ZIPs are received;
the original export family separately retains one received ZIP out of 51.
Full-image review and the preserved cloud decisions leave 94 eligible images and
six held: **indeed**, **mention**, **apparently**, **simply**, **whom** and **consider**.
All 94 eligible images are prepared, published and separately CDN-byte verified.
Each reviewed WebP was reproduced byte-for-byte from its exact archived native;
the application WebP was then independently reproduced during publication checks.
These 94 images preserve 236,181,622 bytes of native originals locally and add
9,051,612 bytes of optimized public artwork. Supplemental transfers remain
separate from the original source-pack ZIP receipts.

The earlier 100-image cloud batch003 reported 88 pass and 12 flag outputs. That
report and the subsequent 300-image assignment add no intake-proven local
artwork coverage. Historical cloud completion reports do not establish current
workspace availability or successful local preservation.

The selected delivery checkpoint binds canonical CDN-byte verification and full
local-original SHA-256 checks. Its three reviewed replacements for **word**,
**win** (the race scene) and **fine** remain unchanged, excluding superseded
artwork while preserving history. Separate selected MCP
(516) and REST (1,423) records preserve actual transport provenance. The latest
read-only R2 inventory contains **1,942 public objects / 175,361,158 bytes**,
including three superseded objects. Both retired buckets remain absent. Existing
managed Wrangler authentication provided public-only uploads; the latest
authenticated two-page storage inventory used the Cloudflare MCP connector.
Object counts do not establish active meaning coverage, and these
local checkpoints do not establish a staging or production deployment.

The retained 1,875-image presentation checkpoint passes 18 local website cases for barn,
vast and cookie at 360/430/1280 widths in both themes, including keyboard access,
overflow, image geometry and exact CDN response SHA-256. Two additional mobile
viewport checks confirm the full images remain visible after scrolling. Five
representative screenshots were actually inspected. These checks use real Next
pages and public CDN images with synthetic API/auth fixtures; they do not prove
live authentication, durable learner state or every artwork's visual quality.

The retained 1,909-image snapshot has separate civil/species/term checks. Initial
runs completed seventeen distinct original assertion sequences, three targeted
instrumented loads and two scrolled mobile viewport checks; root actually
inspected five screenshots. Explicit later failures record five-second and
thirty-second image-complete timeouts; the earliest failure masked its cause with
a closed-target error. A bounded follow-up completed all twelve original mobile
sequences, giving all eighteen requested combinations a completed sequence across
runs. Twelve loading-only and three default-network samples also passed, with
27 matching full CDN body hashes. Earlier intermittent failures remain unexplained;
this is not a failure-free matrix, performance acceptance or a demonstrated fix.
No application or CDN change was made. Capture request/image state if loading
stalls during the next required product check. Synthetic API/auth fixtures do not
establish live authentication, database durability or staging acceptance.
A separate single run against the 1,934-image snapshot completed twelve original
mobile assertion sequences for percent, profit and nerve at 360/430 widths in
both themes and two scrolled viewport assertions. Root inspected five actual
screenshots; the 20% two-of-ten-row cue and 10 minus 8 equals 2 are legible.
The final geometry aggregate was not written, so no numeric viewport proof is
claimed. The first desktop case, percent at 1280/light, failed the original
five-second image deadline with its frame in the viewport and no response,
request-finished or request-failed event before timeout. The remaining five
desktop cases were not run; there was no repeat. Default Chromium networking
used neither routing interception nor protocol overrides, so the earlier
allowlist alone cannot explain the recurrence. The failing network stage and
cause remain unknown; no network fix or full presentation/performance pass is
established. Historical failures remain retained, and synthetic API/auth
fixtures do not prove live authentication, durability or deployment.

The initial loading/manual-retry UI's 54-case pass is historical. Independent
review subsequently reproduced keyboard-focus loss on local fallback and terminal
unavailability; those failures are retained. The corrected stable frame passes a
new production build and **66 fixture-backed visual-learning cases in 47.8 seconds**,
twenty-two per 360/430/1280 layout with both themes covered and no retries, skips
or flaky results. Complete raw logs and the JSON reporter are retained; lint,
web and E2E typechecking, formatting and diff checks pass. Root actually inspected
all six pending screenshots; their appearance is retained across the correction.
Final independent code/evidence review found no blocking issues, including the
fallback and terminal focus paths; it did not rerun tests or claim full acceptance.
A separate real CDN GET
with the retry query returned 200 and the exact 19,926-byte percent WebP with
matching full SHA-256. This establishes query compatibility, not browser
performance or a new complete eighteen-case real-CDN matrix pass. The original
network cause remains unresolved.

Before this five-image publication, four actual app-size checks used temporary
QA WebPs from the unchanged optimizer and mocked image responses on the frozen
recovery build: **literally** at 360px/light and dark, **curious** at 430px/dark
and **cab** at 1280px/light. Root inspected all four viewport screenshots.
Technical rendering assertions passed; the separate sentence-legibility gate
held **literally**. These checks use synthetic API/auth and do not establish
real CDN performance, all five pictures' full viewport matrix or staging acceptance.
The earlier PostgreSQL rehearsal imported the then-current 1,070-image manifest
and separately exercised a synthetic 10,051-row fixture. Full artwork coverage,
the final manifest import rehearsal, one consolidated PR, required hosted checks
and live staging acceptance remain pending. Production promotion is separate.

Historical 20 October checkpoint: 113 public WebPs totaling 7,095,834 bytes had
been uploaded and freshly CDN SHA-256 verified, including the reviewed 20-image
pilot. At that time a private cloud archive held 92 preserved legacy WebPs, 113
native PNGs and the portable legacy-native index. Those counts and the presence
of a private bucket are stale; the current storage policy keeps originals and
history locally and only optimized public pictures in R2.

See [release readiness](../product/release-readiness.md) for separate local test,
hosted review/CI, metadata import, full coverage and deployment evidence. A public
object's existence is not proof that the learner-facing release has been deployed.
