# Meaning image library

Implementation checkpoint: 20 October 2026. This document describes the delivery
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
that hostname. Cloudflare R2 currently stores approved public WebPs and a separate
private archive. Provider-managed development URLs are disabled for both buckets;
the private archive has no public custom domain. Native generation files and
rejected attempts never enter the public image bucket.

The public image host has a dedicated Cloudflare cache rule for `/meanings/`
paths ending in `.webp`. Only successful 200/206 responses receive a one-year
edge TTL; 300–599 responses use no-store. Browser caching respects the origin
instead of inheriting that edge TTL. The expression includes GET, HEAD and PURGE
so single-file cache purges still match. This rule does not change application,
API or private archive caching. Verify successful CDN checksums and MISS-to-HIT
behavior after a single-image purge, plus repeated 404 cache bypasses; the rule's
configured TTL alone does not prove delivery behavior or whole-site performance.

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

The S3 publisher checks original/public bytes through authenticated GETs and
requires its explicit private no-store and public immutable cache headers. It
verifies canonical public delivery before emitting its import-ready manifest.

The distinct Cloudflare MCP receipt path validates the operator's successful raw
PUT receipts against actual local source SHA-256, size and single-part ETag/MD5.
It then fetches the public image from the canonical origin and verifies SHA-256.
Its separate manifest/evidence records the actual public response headers for
each verification; browser headers and edge cache policy are separate evidence.
It does **not** establish an independent private GET SHA-256 check or the
S3 publisher's immutable/no-store metadata. These stronger S3 gates are unchanged.

All 92 legacy native PNG originals were recovered through byte-identical
reproduction of their current WebPs, not receipt order or first approval. Their
177,279,634 bytes are locally SHA-verified, with exact receipt and transform
provenance. A separate portable native-original archive manifest links each
public content hash to the native archive key. All 92 native PNGs are uploaded. A fresh private API listing matches each
size and single-part MD5 ETag, with source SHA-256 checked before PUT. This is
not an independent private GET SHA-256 or restoration test. The portable index
is also archived privately under its content hash. The already verified WebP
publication receipts remain valid. Seven rejected native attempts are
preserved separately and excluded from the approved upload plan.

## Acceptance checkpoint

There are 113 uploaded public WebPs totaling 7,095,834 bytes, each freshly fetched
from the canonical CDN and SHA-256 verified. The private archive contains
92 preserved legacy WebPs, 113 native PNGs and the portable legacy-native index.
The 20-image generation pilot is visually reviewed, uploaded and CDN-verified;
remaining catalogue coverage is not finished.

See [release readiness](../product/release-readiness.md) for separate local test,
hosted review/CI, metadata import, full coverage and deployment evidence. A public
object's existence is not proof that the learner-facing release has been deployed.
