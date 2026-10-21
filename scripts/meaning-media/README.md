# Meaning image library tools

These tools prepare, archive and publish approved image files. They **do not call
an image model**, create placeholders, approve generated pictures or edit learner
data. Generate pictures with the supported image-generation workflow and inspect
their exact meaning, accessibility description and visual variety before approval.

The initial canonical export contained **10,051 meanings**: 10,000 map
meanings, 92 curriculum meanings (60 already in the map), and 19 supplemental
uses. It had 92 reviewed existing pictures, leaving 9,959 missing. These are
the initial source-tree counts verified on 20 October 2026; re-export after content changes.
An exported queue is not evidence of completed image coverage.

## Export a fresh queue

Run from the repository root with its declared Go/Node/pnpm versions installed:

```sh
(cd apps/api && go run ../../scripts/meaning-media/export-catalog.go \
  -root ../.. -out ../../.work/meaning-media/catalog.json)
```

The exporter imports the authoritative supplemental uses and shared Go revision
function. It reads the authored curriculum before the map, matching canonical
seeding precedence. Repeated identities must match exactly. Each revision binds
the meaning ID, canonical spelling, part of speech and short definition. Output
order is deterministic: reviewed existing pictures, map rank, remaining meaning
IDs. Source attribution and original catalogue files remain unchanged.

The working directory is deliberately ignored by Git. Do not commit generated
originals, working queues, upload journals or credentials into the public repo.

## Prepare existing pictures locally

```sh
node scripts/meaning-media/manage.mjs prepare-existing \
  --catalog .work/meaning-media/catalog.json \
  --workdir .work/meaning-media/library
```

All existing reviewed WebP files are completely decoded, bounded to 960×640 and
copied byte-for-byte. No further lossy compression is applied. These records use
`reviewed-legacy-webp-preserved`: the file in this publication plan is the reviewed
WebP, rather than the native generated original. Native originals may be attached
through a separate archive manifest after exact source-to-current-WebP matching
and separate upload verification; their discovery does not change the already
verified publication plan's original file hashes or receipts.

New generator output uses a local candidate manifest:

```json
{
  "images": [
    {
      "meaningId": "the-exact-uuid-from-the-export",
      "meaningRevision": "the-exact-64-character-revision-from-the-export",
      "source": "/absolute/path/to/generated-original.png",
      "alt": "A concise description of the approved scene.",
      "approved": true,
      "approvedBy": "the-actual-reviewer-or-review-process"
    }
  ]
}
```

```sh
node scripts/meaning-media/manage.mjs prepare \
  --catalog .work/meaning-media/catalog.json \
  --candidates .work/meaning-media/reviewed-candidates.json \
  --workdir .work/meaning-media/library
```

Only mark `approved: true` after visual inspection. A generated file's existence
does not establish meaning accuracy. Candidate source paths resolve from the
current working directory; absolute paths are recommended. Supported originals
are single-frame PNG, JPEG and WebP, at most 32 MiB and 32 million pixels. The
existing Next.js-pinned Sharp dependency validates full decoding, applies EXIF
orientation, fits inside 960×640 without cropping/enlarging, strips metadata and
encodes WebP at quality 82. Public files over 1 MiB are rejected for review.

Each completed image is saved before continuing, so interrupted batches resume.
Conflicting prepared artwork for a meaning requires a separate replacement batch;
the tool never silently replaces prior approval. Run one preparation/publication
process per working directory.

Local outputs are portable storage files:

- `public/meanings/{meaningId}/{meaningRevision}/{sha256}.webp`: approved image.
- `private/originals/{meaningId}/{meaningRevision}/{originalSha256}.{format}`:
  original bytes; never expose this directory publicly.
- `prepared.json`: local approval metadata and paths; not a ready database import.

The `public/` tree can be copied to a static file server or another object store
without changing content keys. Keep archive storage private and back it up
separately. The tools do not configure servers, DNS, bucket policies or backups.

## Current workflow: local originals and public optimized images

Keep originals on this computer in a durable private archive. Upload only the
optimized `public/meanings/...` WebP files to the public object store. Preparation
still keeps a working copy under `private/`; publication independently re-reads
the durable archive and does not require that working copy to remain present.

Before accepting public receipts, preserve each prepared original under:

```text
/private/meaning-originals/
  originals/{meaningId}/{meaningRevision}/{originalSha256}.{format}
```

The portion beginning `originals/` is exactly `archiveKey(preparedRow)`. Preserve
the original bytes unchanged, keep the archive root private, and back it up
separately. Use regular files or hard links; symlinks in the root, intermediate
directories or final file are rejected. The adapter does not copy, replace,
delete or upload originals. A prepared original must be a valid fully decodable
PNG/JPEG/WebP whose full SHA-256 matches the immutable meaning/revision/source key.
Reviewed legacy WebP originals retain their truthful legacy provenance; preserving
a separately recovered native PNG does not change an existing prepared contract.

Record normalized receipts only from actual completed **public** MCP PUTs:

```json
{
  "schemaVersion": 1,
  "transport": "cloudflare-mcp-public-only-local-archive-v1",
  "accountId": "the-actual-32-hex-account-id",
  "publicBucket": "vocanova-meaning-images",
  "objects": [
    {
      "meaningId": "the-exact-uuid-from-prepared-json",
      "meaningRevision": "the-exact-revision-from-prepared-json",
      "sha256": "the-optimized-image-sha256",
      "public": {
        "key": "meanings/meaning-id/revision/optimized-sha256.webp",
        "size": 12345,
        "etag": "the-actual-single-part-32-hex-put-etag",
        "sourceSha256": "the-optimized-sha256-checked-before-the-mcp-put",
        "contentType": "image/webp",
        "cacheControl": null
      }
    }
  ]
}
```

Do not include `archiveBucket` or `original` receipts in this mode. A collection
contains 1–500 exact current prepared meanings. Receipt values must come from
successful public uploads, never be fabricated from a plan. The storage upload
itself remains an operator step; this acceptance command makes no cloud writes.

```sh
node scripts/meaning-media/accept-mcp-receipts.mjs \
  --catalog .work/meaning-media/catalog.json \
  --receipts .work/meaning-media/completed-public-put-receipts.json \
  --workdir .work/meaning-media/library \
  --account-id EXPECTED_CLOUDFLARE_ACCOUNT_ID \
  --original-archive local \
  --local-archive-root /private/meaning-originals
```

The complete plan must match the fresh canonical catalogue. Before the first
HTTP request, every supplied receipt and original is checked against its exact
identity and source bytes; missing, corrupt or aliased originals fail closed.
Public PUT key, size, source SHA-256, content type and single-part ETag/MD5 must
match the prepared WebP. The only network requests are fresh GETs to
`https://images.vocanova.site`, without redirects, with bounded bytes and full
SHA-256 verification. No archive bucket or private cloud endpoint is contacted.
Network retries remain bounded to three fresh 60-second requests.

Only `published-mcp-local.json` is the import-ready manifest for this mode. It
has the same public image contract as other publication modes and excludes local
paths, original hashes and reviewer identity. Repeating a receipt collection
revalidates the local originals and public bytes without duplicating entries.

The separate private `mcp-local-publication-evidence.json` records the archive
root, each identity-derived relative original key, independently read byte length,
SHA-256 and MD5, receipt-file hash and fresh canonical public delivery proof.
Original evidence uses `local-archive-full-read-sha256-md5` and explicitly scopes
`cloudOriginalUploadedByThisCommand: false` and
`cloudOriginalPresenceChecked: false`. It does not assert that no historical
cloud copy exists. The actual public cache header is recorded;
`immutableCacheVerified: false` remains truthful for this raw MCP path.

```sh
(cd apps/api && go run ./cmd/meaning-media-import \
  --manifest ../../.work/meaning-media/library/published-mcp-local.json)
```

Import still requires the operator's `DATABASE_URL` and independently validates
live canonical revisions. This workflow does not deploy the application or
establish backups. Historical `published-mcp.json`,
`mcp-publication-evidence.json`, and the S3 mode's stronger evidence remain
separate and unchanged.

### Import after a release

Both deployment bundles build `cmd/meaning-media-import` from the release checkout
as a static Linux/amd64 operator binary. Staging installs it at
`/opt/vocanova/apps/api/bin/meaning-media-import`; the separate manual production
release installs it at `/opt/vocanova/production/apps/api/bin/meaning-media-import`.
The application runtime image remains unchanged. Deployment neither copies the
private selected manifest nor invokes this tool automatically.

After migrations and canonical seeding, transfer the reviewed import-ready
manifest separately, verify its full SHA-256 against the selected delivery
checkpoint, and run the installed tool with `--manifest /absolute/path/to/manifest.json`
in the target tier's operator database environment. Keep database credentials out
of command output. Check the tool's checked-record count against the manifest;
the inserted count depends on existing matching target records. An exact repeat
must report zero additions. Then verify current-revision metadata through the
authenticated API and inspect the affected learner screens. A successful code
deploy, upload or local rehearsal does not replace these target checks.

### Direct REST receipts with local originals

When an operator uses the Cloudflare REST API instead of MCP, select that source
explicitly. Do not relabel REST responses as MCP PUT receipts. The acceptance
adapter does not obtain credentials or make Cloudflare API calls, and REST mode
requires local originals; it cannot enable cloud-original uploads.

Normalize the actual API results into the same collection and public metadata
shape shown above, with these differences:

- Collection `transport` is `cloudflare-rest-public-only-local-archive-v1`.
- Every `objects[]` entry requires `operation: "uploaded"` for a completed PUT or
  `operation: "reconciled-existing"` for a matching existing object's metadata
  readback. The latter reports no new PUT. Preserve actual API result metadata,
  response digests and readback evidence privately; an upload plan alone is not
  a receipt.
- `public.key`, `size`, `etag`, `contentType` and `cacheControl` come from the actual
  object result/readback. `sourceSha256` binds the exact locally checked public
  bytes. Existing conflicts must be investigated rather than relabeled as success.

```sh
node scripts/meaning-media/accept-mcp-receipts.mjs \
  --catalog .work/meaning-media/catalog.json \
  --receipts .work/meaning-media/completed-rest-public-receipts.json \
  --workdir .work/meaning-media/library \
  --account-id EXPECTED_CLOUDFLARE_ACCOUNT_ID \
  --original-archive local \
  --local-archive-root /private/meaning-originals \
  --receipt-source cloudflare-rest
```

The API equivalent is `acceptMCPReceipts({ ...options, receiptSource:
"cloudflare-rest", originalArchive: "local" })`. The default remains
`cloudflare-mcp`; a receipt transport cannot switch modes implicitly. Missing or
unknown REST operations and mismatched transports fail before any HTTP request.
All canonical revision, source SHA-256, size, ETag/MD5, image decoding, local
archive, symlink, manifest-conflict and fresh CDN full-byte checks remain required.

REST mode writes only `published-rest-local.json` and the private
`rest-local-publication-evidence.json`. Public import records keep the same schema;
historical MCP manifests/evidence remain untouched. REST evidence stores
`publicObject`, not `publicPut`: `operation` and `newPutReported` distinguish
`operator-rest-put-receipt-size-etag-md5` from
`operator-rest-existing-metadata-readback-size-etag-md5`. These are operator
receipt claims checked against local bytes; the separate fresh CDN GET proves
current public delivery. The command makes no storage writes in either case.
It reports uploaded versus reconciled receipt counts without treating a metadata
readback as a new upload. The observed CDN cache header remains recorded without
asserting immutable caching. Use the REST manifest explicitly when importing or
composing the selected delivery manifest, and retain the separate receipt history.

## Optional legacy workflow: cloud originals through portable S3

This mode uploads originals to a private cloud bucket. It remains available for
separately authorized cloud-archive deployments; the current owner workflow above
keeps originals local and publishes only optimized public pictures.

Use operator-provided environment variables. The tool never reads `.env` files or
credential profiles and never prints credentials or raw provider diagnostics.

| Variable                             | Purpose                                                                          |
| ------------------------------------ | -------------------------------------------------------------------------------- |
| `MEANING_MEDIA_S3_ENDPOINT`          | Plain HTTPS origin, e.g. `https://ACCOUNT_ID.r2.cloudflarestorage.com`           |
| `MEANING_MEDIA_S3_REGION`            | Signing region, `auto` for R2                                                    |
| `MEANING_MEDIA_S3_ACCESS_KEY_ID`     | Scoped object read/write access key                                              |
| `MEANING_MEDIA_S3_SECRET_ACCESS_KEY` | Corresponding secret                                                             |
| `MEANING_MEDIA_PUBLIC_BUCKET`        | Optional; default `vocanova-meaning-images`                                      |
| `MEANING_MEDIA_ARCHIVE_BUCKET`       | Optional; default `vocanova-meaning-originals`                                   |
| `MEANING_MEDIA_PUBLIC_BASE_URL`      | Optional assertion of `https://images.vocanova.site`; other origins are rejected |

The archive and public buckets must be different. Preconfigure public image
delivery and private archive access before publication. Never put originals in
the public bucket. S3 operations use path-style SigV4 and only HEAD/GET/conditional
PUT. There are no list/delete/bucket-configuration operations or model calls.
Compatibility with a new S3 provider must be checked before a bulk run.
Storage endpoints remain portable, but an import-ready manifest always verifies
the exact canonical delivery origin `https://images.vocanova.site`, because the
backend constructs that URL. A preview/test host cannot produce a ready manifest.
Tests inject a synthetic fetch implementation while retaining the canonical URL.

```sh
node scripts/meaning-media/manage.mjs publish \
  --catalog .work/meaning-media/catalog.json \
  --workdir .work/meaning-media/library --limit 100
```

Default batches resume images absent from `published.json`; repeat until
`remaining` is zero. `--offset 0 --limit 100` explicitly re-verifies the first 100
prepared records. The maximum batch is 500. A lock prevents concurrent publication
to the same working directory; inspect the owning process before deleting a stale
lock after an abrupt termination.

For every selected image, publication validates the current exported revision,
local hashes and actual decodability; uploads/verifies the private original first;
then uploads/verifies the approved WebP; then verifies the public delivery URL.
Conditional `If-None-Match: *` prevents concurrent replacement. Uploads include
Content-MD5 and SHA-256 metadata. Existing, newly written and concurrently created
objects are fully fetched and SHA-256 checked; a journal or object header alone
never counts as verification. Network retries are bounded to three attempts, with
60-second request deadlines and strict response-byte bounds. Redirects are refused.
The complete plan's metadata is validated before the first cloud write, including
rejection of NUL, CR and LF in alt text, matching the backend import contract.

Public delivery uses `image/webp` and
`Cache-Control: public, max-age=31536000, immutable`. Originals use their actual
image type and `private, max-age=0, no-store`. A checksum/header mismatch fails
without overwriting the immutable object. A partial failure can leave safe orphan
objects, but cannot add an unverified image to the ready import manifest.

`journal.json` records upload/delivery verification. **Only `published.json`** is
the S3 path's public database import manifest; it excludes original paths and
reviewer identity. The transactional importer independently checks live canonical revisions:

```sh
(cd apps/api && go run ./cmd/meaning-media-import \
  --manifest ../../.work/meaning-media/library/published.json)
```

The importer requires an operator-supplied `DATABASE_URL`. Export freshness alone
cannot prove the live database has matching content. Publication does not import
automatically, deploy the application, mark missing pictures ready or replace the
existing fallback library.

## Optional legacy workflow: cloud-original MCP PUT receipts

This earlier mode requires receipts for both private cloud originals and public
optimized images. Its checks and historical evidence remain unchanged. Use the
explicit local-original mode above for the current owner storage policy.

The Cloudflare MCP raw object PUT can upload bytes without local S3 credentials,
but it does not set the S3 publisher's cache metadata. Its object GET is unsuitable
for independently verifying private binary SHA-256. Do not weaken the S3 checks
or claim that this path performed a private binary download.

An operator can instead record normalized receipts from completed MCP PUT results:

```json
{
  "schemaVersion": 1,
  "transport": "cloudflare-mcp-raw-put-v1",
  "accountId": "the-actual-32-hex-account-id",
  "publicBucket": "vocanova-meaning-images",
  "archiveBucket": "vocanova-meaning-originals",
  "objects": [
    {
      "meaningId": "the-exact-uuid-from-prepared-json",
      "meaningRevision": "the-exact-revision-from-prepared-json",
      "sha256": "the-optimized-image-sha256",
      "original": {
        "key": "originals/meaning-id/revision/original-sha256.png",
        "size": 123456,
        "etag": "the-actual-single-part-32-hex-put-etag",
        "sourceSha256": "the-source-sha256-checked-before-the-mcp-put",
        "contentType": "image/png",
        "cacheControl": null
      },
      "public": {
        "key": "meanings/meaning-id/revision/optimized-sha256.webp",
        "size": 12345,
        "etag": "the-actual-single-part-32-hex-put-etag",
        "sourceSha256": "the-optimized-sha256-checked-before-the-mcp-put",
        "contentType": "image/webp",
        "cacheControl": null
      }
    }
  ]
}
```

Use actual keys, byte counts, source hashes, content types and PUT result ETags;
the values above only illustrate the schema. `cacheControl` records returned
object metadata or `null` when absent, not a guessed CDN header. The collection
can name a subset of the plan, at most 500 images; every named image requires both
object receipts and its local original/optimized files. Receipts are operator
evidence, not cryptographic proof of an independent private read. Do not fabricate
them from local file metadata or an upload that has not actually succeeded.

```sh
node scripts/meaning-media/accept-mcp-receipts.mjs \
  --catalog .work/meaning-media/catalog.json \
  --receipts .work/meaning-media/completed-mcp-put-receipts.json \
  --workdir .work/meaning-media/library \
  --account-id EXPECTED_CLOUDFLARE_ACCOUNT_ID
```

This command makes **no cloud writes**. It validates the entire approved plan
against the fresh catalogue and every named local source before any HTTP request.
Receipt account, bucket, meaning/revision, object key, size, source SHA-256 and
single-part ETag/MD5 must match the actual prepared files. Missing, stale,
multipart or conflicting receipts fail. Public bytes are then fetched only from
`https://images.vocanova.site`, bounded and SHA-256 checked without redirects.
Network request failures get at most three attempts, each with a fresh 60-second
deadline and short backoff. HTTP rejections and invalid image bytes still fail
immediately; retries never substitute receipt metadata for actual public bytes.

The distinct `published-mcp.json` is the backend-ready manifest for this path.
The companion `mcp-publication-evidence.json` records source hashes and receipt
file hash, operator-reported PUT metadata, the actual public response cache
header (for example `max-age=14400`) and fresh public SHA-256 verification.
It explicitly records `privateGetSha256Verified: false`,
`privateHTTPMetadataVerified: false`, `bucketExposureVerifiedByThisTool: false`
and `immutableCacheVerified: false`. The operator must separately verify the
archive bucket's privacy and retain backups. This path does not establish the
S3 publisher's immutable/no-store headers or its independent private GET proof.
Existing `published.json` and its stronger S3 verification remain unchanged.

When recovered native originals exist, retain the separate
`native-originals-upload-plan.json` and portable `native-originals-manifest.json`
in private working storage. The native manifest links the public content hash to
the native original's immutable archive key, size, MD5, SHA-256 and exact matching
provenance. It is not a backend public-image import, and a prepared native archive
manifest does not establish cloud upload or independent private GET verification.
Retain its upload receipts and verification evidence separately. Preserve rejected
generation attempts outside the approved native upload plan.

Pass `published-mcp.json` to the same transactional backend importer only after
reviewing this evidence and the operator's private-bucket checks. A failed public
GET cannot add that picture to the ready manifest. Repeating the command
re-verifies all supplied receipts and safely resumes prior completed entries.

R2 conditional PUT/checksum support was checked against
[Cloudflare's S3 compatibility documentation](https://developers.cloudflare.com/r2/api/s3/api/).
Unit tests use synthetic object responses and one existing local image; they do
not establish live cloud credentials, CDN configuration or backup restoration.

## Verification

```sh
node --test scripts/foundation/meaning-media-publisher.test.mjs
pnpm exec prettier --check scripts/meaning-media/*.mjs \
  scripts/meaning-media/README.md scripts/foundation/meaning-media-publisher.test.mjs
```
