# Meaning image library

Implementation checkpoint: 22 October 2026. This document describes the delivery
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
The later 549 REST public objects carry the one-year
`public, max-age=31536000, immutable` browser header, verified on actual CDN
responses. Historical MCP objects and the first five REST uploads retain
their original browser metadata; the new header is not a claim about every asset.

The API returns approved metadata only when the meaning and canonical word are
active and the current revision matches. Discovery/detail embed image metadata
in existing reads. Other teaching steps coalesce lookups into authenticated batches
of at most 100 meaning IDs. There is no HTTP image-write route and no learner
progress mutation. Optional metadata reads have a one-second database deadline,
including pool waits; shorter caller deadlines still win. Frozen teaching steps require their exact meaning revision;
old pictures cannot silently explain a changed use.

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

The 22 October local checkpoint contains **1,070 selected published meanings** and
**1,070 prepared meanings** out of 10,051 canonical meanings. All prepared IDs are
in the selected delivery manifest; **8,981 meanings still need prepared artwork**.
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

Cloud batch003 reports 100 completed outputs with cloud review results of 88 pass
and 12 flag. Its artwork has not yet been downloaded locally. The next 300-image
batch is active in the existing durable subscription-only cloud task; generated,
reviewed and locally integrated totals for that assignment remain pending.

The selected delivery checkpoint binds canonical CDN-byte verification and full
local-original SHA-256 checks. Its three reviewed replacements remain unchanged,
excluding superseded artwork while preserving history. Separate selected MCP
(516) and REST (554) records preserve actual transport provenance. A fresh
read-only R2 inventory contains **1,073 public objects / 93,637,044 bytes**,
including three superseded objects. Both retired buckets remain absent. Existing
Wrangler authentication provided the readback while the MCP connector was
unavailable. Object counts do not establish active meaning coverage, and these
local checkpoints do not establish a staging or production deployment.

Historical 20 October checkpoint: 113 public WebPs totaling 7,095,834 bytes had
been uploaded and freshly CDN SHA-256 verified, including the reviewed 20-image
pilot. At that time a private cloud archive held 92 preserved legacy WebPs, 113
native PNGs and the portable legacy-native index. Those counts and the presence
of a private bucket are stale; the current storage policy keeps originals and
history locally and only optimized public pictures in R2.

See [release readiness](../product/release-readiness.md) for separate local test,
hosted review/CI, metadata import, full coverage and deployment evidence. A public
object's existence is not proof that the learner-facing release has been deployed.
