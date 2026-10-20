import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath, URL } from "node:url";
import { Buffer } from "node:buffer";
import { createHash } from "node:crypto";
import { prepare, publish } from "../meaning-media/manage.mjs";
import {
  acceptMCPReceipts,
  validateMCPObjectReceipt,
} from "../meaning-media/accept-mcp-receipts.mjs";
import {
  S3Publisher,
  PUBLIC_BUCKET,
  ARCHIVE_BUCKET,
  CDN,
  sha256,
  signRequest,
  publicKey,
  archiveKey,
  publicRecord,
  validateImage,
  within,
  readBounded,
  atomicJSON,
} from "../meaning-media/library.mjs";

const row = {
  meaningId: "7ee5fc1b-8108-5fb4-a254-f75af84695f3",
  meaningRevision: "a".repeat(64),
  sha256: "b".repeat(64),
  width: 960,
  height: 640,
  alt: "Hands put a letter into an envelope.",
  approved: true,
  originalSha256: "c".repeat(64),
  originalFormat: "png",
};
const { Response, structuredClone } = globalThis;
const credentials = {
  endpoint: "https://objects.example.test",
  region: "us-east-1",
  accessKeyId: "EXAMPLE_ACCESS_KEY",
  secretAccessKey: "synthetic-test-secret",
};
const key = publicKey(row);
const bytes = Buffer.from("synthetic-object-checksum-data");
const cache = "public, max-age=31536000, immutable";
const storedResponse = (body = bytes, overrides = {}) =>
  new Response(body, {
    headers: {
      "content-type": "image/webp",
      "cache-control": cache,
      ...overrides,
    },
  });

test("public identity is meaning/revision/content-specific and excludes archive metadata", () => {
  assert.equal(
    key,
    `meanings/${row.meaningId}/${row.meaningRevision}/${row.sha256}.webp`,
  );
  assert.equal(
    archiveKey(row),
    `originals/${row.meaningId}/${row.meaningRevision}/${row.originalSha256}.png`,
  );
  assert.deepEqual(Object.keys(publicRecord(row)), [
    "meaningId",
    "meaningRevision",
    "sha256",
    "width",
    "height",
    "alt",
    "approved",
  ]);
  assert.notEqual(key, publicKey({ ...row, meaningRevision: "d".repeat(64) }));
  assert.notEqual(key, publicKey({ ...row, sha256: "d".repeat(64) }));
});

test("pending, oversized, malformed and inaccessible pictures cannot be marked ready", () => {
  for (const mutation of [
    { approved: false },
    { alt: " " },
    { alt: "a".repeat(501) },
    { width: 961 },
    { height: 0 },
    { height: 640.5 },
    { sha256: "not-a-hash" },
    { meaningId: "../../something" },
    { meaningRevision: "old" },
  ])
    assert.throws(() => validateImage({ ...row, ...mutation }));
  assert.doesNotThrow(() => validateImage({ ...row, width: 640, height: 640 }));
});

test("local artifact paths stay inside the chosen storage directory", () => {
  assert.equal(
    within("/tmp/library", "public/abc.webp"),
    "/tmp/library/public/abc.webp",
  );
  assert.throws(() => within("/tmp/library", "../library-other/private"));
  assert.throws(() => within("/tmp/library", "/etc/passwd"));
});

test("bounded file reads and atomically saved state keep checksum evidence", async () => {
  const directory = await mkdtemp(join(tmpdir(), "meaning-media-"));
  try {
    const path = join(directory, "state.json");
    await atomicJSON(path, { verified: false });
    await atomicJSON(path, { verified: true });
    assert.deepEqual(JSON.parse(await readFile(path)), { verified: true });
    await writeFile(join(directory, "image"), bytes);
    await assert.rejects(
      readBounded(join(directory, "image"), 2),
      /byte limit/,
    );
    assert.equal(
      sha256(await readBounded(join(directory, "image"))),
      sha256(bytes),
    );
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("SigV4 binds method, destination, immutable headers, payload and region", () => {
  const options = {
    ...credentials,
    method: "PUT",
    bucket: PUBLIC_BUCKET,
    key,
    body: bytes,
    now: new Date("2026-10-20T08:00:00Z"),
    headers: { "if-none-match": "*", "content-type": "image/webp" },
  };
  const signed = signRequest(options);
  assert.equal(
    signed.url,
    `https://objects.example.test/${PUBLIC_BUCKET}/${key}`,
  );
  assert.equal(signed.headers["x-amz-content-sha256"], sha256(bytes));
  assert.match(
    signed.headers.authorization,
    /20261020\/us-east-1\/s3\/aws4_request/,
  );
  assert.match(
    signed.headers.authorization,
    /SignedHeaders=content-type;host;if-none-match;x-amz-content-sha256;x-amz-date/,
  );
  for (const mutation of [
    { body: Buffer.from("other") },
    { region: "auto" },
    { bucket: ARCHIVE_BUCKET },
    { method: "GET" },
  ]) {
    assert.notEqual(
      signed.headers.authorization,
      signRequest({ ...options, ...mutation }).headers.authorization,
    );
  }
});

test("S3 signer refuses credential-bearing endpoints, cleartext, URL paths and destructive methods", () => {
  for (const mutation of [
    { endpoint: "http://objects.example.test" },
    { endpoint: "https://key:secret@example.test" },
    { endpoint: "https://example.test/path" },
    { endpoint: "https://example.test?token=secret" },
    { method: "DELETE" },
    { key: "../other" },
    { region: "bad/region" },
  ])
    assert.throws(() =>
      signRequest({
        ...credentials,
        method: "GET",
        bucket: PUBLIC_BUCKET,
        key,
        ...mutation,
      }),
    );
});

test("missing object is uploaded conditionally and fully re-read before success", async () => {
  const calls = [];
  const storage = new S3Publisher(credentials, async (_url, options) => {
    calls.push(options);
    if (options.method === "HEAD") return new Response(null, { status: 404 });
    if (options.method === "PUT") return new Response(null, { status: 200 });
    return storedResponse();
  });
  await storage.ensure(PUBLIC_BUCKET, key, bytes, "image/webp", cache);
  assert.deepEqual(
    calls.map((call) => call.method),
    ["HEAD", "PUT", "GET"],
  );
  assert.equal(calls[1].headers["if-none-match"], "*");
  assert.equal(calls[1].headers["x-amz-meta-sha256"], sha256(bytes));
  assert.equal(calls[1].headers["content-md5"].length, 24);
  assert.equal(calls[1].redirect, "error");
});

test("resume re-reads an existing object and never replaces it", async () => {
  const calls = [];
  const storage = new S3Publisher(credentials, async (_url, options) => {
    calls.push(options.method);
    return options.method === "HEAD" ? new Response(null) : storedResponse();
  });
  assert.equal(
    await storage.ensure(PUBLIC_BUCKET, key, bytes, "image/webp", cache),
    "verified-existing",
  );
  assert.deepEqual(calls, ["HEAD", "GET"]);
});

test("concurrent conditional PUT is accepted only after verifying winner bytes", async () => {
  const storage = new S3Publisher(credentials, async (_url, options) => {
    if (options.method === "HEAD") return new Response(null, { status: 404 });
    if (options.method === "PUT") return new Response(null, { status: 412 });
    return storedResponse();
  });
  await storage.ensure(PUBLIC_BUCKET, key, bytes, "image/webp", cache);
});

test("a corrupt existing object cannot pass on matching size or metadata alone", async () => {
  const methods = [];
  const storage = new S3Publisher(credentials, async (_url, options) => {
    methods.push(options.method);
    return options.method === "HEAD"
      ? new Response(null)
      : storedResponse(Buffer.alloc(bytes.length));
  });
  await assert.rejects(
    storage.ensure(PUBLIC_BUCKET, key, bytes, "image/webp", cache),
    /checksum mismatch/,
  );
  assert.deepEqual(methods, ["HEAD", "GET"]);
});

test("unexpected public headers fail verification instead of silently rewriting objects", async () => {
  const storage = new S3Publisher(credentials, async (_url, options) =>
    options.method === "HEAD"
      ? new Response(null)
      : storedResponse(bytes, { "cache-control": "private" }),
  );
  await assert.rejects(
    storage.ensure(PUBLIC_BUCKET, key, bytes, "image/webp", cache),
    /delivery headers/,
  );
});

test("access-denied response does not become a missing object or disclose provider body", async () => {
  const methods = [];
  const storage = new S3Publisher(credentials, async (_url, options) => {
    methods.push(options.method);
    return new Response("credential-sensitive provider diagnostic", {
      status: 403,
    });
  });
  await assert.rejects(
    storage.ensure(PUBLIC_BUCKET, key, bytes, "image/webp", cache),
    (error) => error.message === "Object storage HEAD failed (HTTP 403)",
  );
  assert.deepEqual(methods, ["HEAD"]);
});

test("CDN must return the approved WebP bytes and immutable caching without redirects", async () => {
  let called;
  const storage = new S3Publisher(credentials, async (url, options) => {
    called = { url, options };
    return storedResponse();
  });
  await storage.verifyCDN(key, bytes);
  assert.equal(called.url, `${CDN}/${key}`);
  assert.equal(called.options.redirect, "error");
  assert.equal(called.options.headers, undefined);
  await assert.rejects(
    storage.verifyCDN(key, Buffer.alloc(bytes.length), CDN),
    /checksum mismatch/,
  );
});

test("import-ready CDN checks reject other origins before any HTTP request", async () => {
  let requests = 0;
  const storage = new S3Publisher(credentials, async () => {
    requests++;
    return storedResponse();
  });
  for (const origin of [
    "https://media.example.test",
    "https://images.vocanova.site.evil.test",
    "http://images.vocanova.site",
    `${CDN}/preview`,
    `${CDN}?token=example`,
    "https://user:password@images.vocanova.site",
  ]) {
    await assert.rejects(
      storage.verifyCDN(key, bytes, origin),
      /Import-ready images/,
    );
  }
  assert.equal(requests, 0);
});

test("a later invalid alt or wrong public origin cannot cause the first cloud write or ready manifest", async () => {
  const directory = await mkdtemp(join(tmpdir(), "meaning-media-preflight-"));
  try {
    const second = {
      ...row,
      meaningId: "00000000-0000-0000-0000-000000000002",
    };
    const catalogPath = join(directory, "catalog.json");
    await atomicJSON(catalogPath, {
      schemaVersion: 1,
      meanings: [row, second],
    });
    let uploads = 0;
    let deliveries = 0;
    const storage = {
      async ensure() {
        uploads++;
      },
      async verifyCDN() {
        deliveries++;
      },
    };
    for (const alt of [
      "A letter\x00in a bag.",
      "A letter\nin a bag.",
      "A letter\rin a bag.",
    ]) {
      await atomicJSON(join(directory, "prepared.json"), {
        schemaVersion: 1,
        images: [row, { ...second, alt }],
      });
      await assert.rejects(
        publish({ catalogPath, workdir: directory, storage, limit: 1 }),
        /no NUL or line breaks/,
      );
      await assert.rejects(readFile(join(directory, "published.json")), {
        code: "ENOENT",
      });
    }
    await atomicJSON(join(directory, "prepared.json"), {
      schemaVersion: 1,
      images: [row],
    });
    await assert.rejects(
      publish({
        catalogPath,
        workdir: directory,
        storage,
        publicBaseURL: "https://media.example.test",
      }),
      /Import-ready images/,
    );
    await assert.rejects(readFile(join(directory, "published.json")), {
      code: "ENOENT",
    });
    assert.equal(uploads, 0);
    assert.equal(deliveries, 0);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("preparation and publication resume safely across an actual image and interrupted CDN delivery", async () => {
  const directory = await mkdtemp(join(tmpdir(), "meaning-media-pipeline-"));
  try {
    const meaning = {
      meaningId: row.meaningId,
      meaningRevision: row.meaningRevision,
    };
    const catalogPath = join(directory, "catalog.json");
    const candidatesPath = join(directory, "candidates.json");
    const source = fileURLToPath(
      new URL(
        `../../apps/web/public/images/meanings/${row.meaningId}.webp`,
        import.meta.url,
      ),
    );
    await atomicJSON(catalogPath, { schemaVersion: 1, meanings: [meaning] });
    await atomicJSON(candidatesPath, {
      images: [
        {
          ...meaning,
          source,
          alt: row.alt,
          approved: true,
          approvedBy: "synthetic-test-review",
        },
      ],
    });
    assert.deepEqual(
      await prepare({ catalogPath, candidatesPath, workdir: directory }),
      { prepared: 1, total: 1, missing: 0 },
    );
    const plan = JSON.parse(await readFile(join(directory, "prepared.json")));
    assert.equal(plan.images[0].width, 960);
    assert.equal(plan.images[0].height, 640);
    const original = await readFile(
      within(directory, plan.images[0].originalFile),
    );
    assert.equal(sha256(original), sha256(await readFile(source)));
    const calls = [];
    const storage = {
      async ensure(bucket, objectKey, objectBytes) {
        calls.push([bucket, objectKey, sha256(objectBytes)]);
      },
      async verifyCDN() {
        throw new Error("Synthetic CDN interruption");
      },
    };
    await assert.rejects(
      publish({ catalogPath, workdir: directory, storage }),
      /Synthetic CDN interruption/,
    );
    await assert.rejects(readFile(join(directory, "published.json")), {
      code: "ENOENT",
    });
    const journal = JSON.parse(await readFile(join(directory, "journal.json")));
    const state = Object.values(journal.entries)[0];
    assert.equal(state.originalVerified, true);
    assert.equal(state.publicVerified, true);
    assert.equal(state.deliveryVerified, false);
    storage.verifyCDN = async () => {};
    assert.deepEqual(
      await publish({ catalogPath, workdir: directory, storage }),
      { verified: 1, published: 1, prepared: 1, remaining: 0 },
    );
    assert.deepEqual(
      calls.map(([bucket]) => bucket),
      [ARCHIVE_BUCKET, PUBLIC_BUCKET, ARCHIVE_BUCKET, PUBLIC_BUCKET],
    );
    const importer = JSON.parse(
      await readFile(join(directory, "published.json")),
    );
    assert.deepEqual(importer.images, [publicRecord(plan.images[0])]);
    assert.equal(JSON.stringify(importer).includes("original"), false);
    assert.equal(
      (await publish({ catalogPath, workdir: directory, storage })).verified,
      0,
    );
    assert.equal(
      (await publish({ catalogPath, workdir: directory, storage, offset: 0 }))
        .verified,
      1,
    );
    await atomicJSON(catalogPath, {
      schemaVersion: 1,
      meanings: [{ ...meaning, meaningRevision: "e".repeat(64) }],
    });
    await assert.rejects(
      publish({ catalogPath, workdir: directory, storage, offset: 0 }),
      /stale/,
    );
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("MCP single-part PUT receipts bind source SHA, actual size, ETag/MD5, object key and metadata", () => {
  const receipt = {
    key,
    size: bytes.length,
    etag: `"${createHash("md5").update(bytes).digest("hex")}"`,
    sourceSha256: sha256(bytes),
    contentType: "image/webp",
    cacheControl: null,
  };
  const expected = { key, bytes, contentType: "image/webp" };
  const proof = validateMCPObjectReceipt(receipt, expected);
  assert.equal(proof.sourceSha256, sha256(bytes));
  assert.equal(proof.byteLength, bytes.length);
  assert.equal(proof.reportedCacheControl, null);
  for (const mutation of [
    { key: "other/file.webp" },
    { size: bytes.length + 1 },
    { sourceSha256: "d".repeat(64) },
    { etag: "0".repeat(32) },
    { etag: "a".repeat(32) + "-2" },
    { contentType: "image/png" },
    { cacheControl: "private\r\nheader" },
  ]) {
    assert.throws(
      () => validateMCPObjectReceipt({ ...receipt, ...mutation }, expected),
      /does not match/,
    );
  }
  assert.throws(() => validateMCPObjectReceipt(undefined, expected), /Missing/);
});

test("legacy preparation preserves the reviewed WebP without claiming native originals unavailable", async () => {
  const directory = await mkdtemp(join(tmpdir(), "meaning-media-legacy-"));
  try {
    const catalogPath = join(directory, "catalog.json");
    await atomicJSON(catalogPath, {
      schemaVersion: 1,
      meanings: [
        {
          meaningId: row.meaningId,
          meaningRevision: row.meaningRevision,
          existing: {
            src: `/images/meanings/${row.meaningId}.webp`,
            alt: row.alt,
          },
        },
      ],
    });
    await prepare({ catalogPath, workdir: directory, existing: true });
    const prepared = JSON.parse(
      await readFile(join(directory, "prepared.json")),
    ).images[0];
    assert.equal(prepared.originalProvenance, "reviewed-legacy-webp-preserved");
    assert.equal(prepared.originalFormat, "webp");
    assert.equal(prepared.originalSha256, prepared.sha256);
    assert.deepEqual(
      await readFile(within(directory, prepared.originalFile)),
      await readFile(within(directory, prepared.publicFile)),
    );
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});

test("MCP receipt acceptance verifies only canonical public bytes and reports weaker private proof truthfully", async () => {
  const directory = await mkdtemp(join(tmpdir(), "meaning-media-mcp-"));
  try {
    const catalogPath = join(directory, "catalog.json");
    const candidatesPath = join(directory, "candidates.json");
    const receiptsPath = join(directory, "mcp-receipts.json");
    const meaning = {
      meaningId: row.meaningId,
      meaningRevision: row.meaningRevision,
    };
    const source = fileURLToPath(
      new URL(
        `../../apps/web/public/images/meanings/${row.meaningId}.webp`,
        import.meta.url,
      ),
    );
    await atomicJSON(catalogPath, { schemaVersion: 1, meanings: [meaning] });
    await atomicJSON(candidatesPath, {
      images: [
        {
          ...meaning,
          source,
          alt: row.alt,
          approved: true,
          approvedBy: "synthetic-test-review",
        },
      ],
    });
    await prepare({ catalogPath, candidatesPath, workdir: directory });
    const plan = JSON.parse(await readFile(join(directory, "prepared.json")));
    const prepared = plan.images[0];
    const original = await readFile(within(directory, prepared.originalFile));
    const optimized = await readFile(within(directory, prepared.publicFile));
    const makeReceipt = (key, body) => ({
      key,
      size: body.length,
      etag: createHash("md5").update(body).digest("hex"),
      sourceSha256: sha256(body),
      contentType: "image/webp",
      cacheControl: null,
    });
    const receipt = {
      schemaVersion: 1,
      transport: "cloudflare-mcp-raw-put-v1",
      accountId: "a".repeat(32),
      publicBucket: PUBLIC_BUCKET,
      archiveBucket: ARCHIVE_BUCKET,
      objects: [
        {
          ...meaning,
          sha256: prepared.sha256,
          original: makeReceipt(archiveKey(prepared), original),
          public: makeReceipt(publicKey(prepared), optimized),
        },
      ],
    };
    await atomicJSON(receiptsPath, receipt);
    const calls = [];
    let corrupt = true;
    const fetcher = async (url, options) => {
      calls.push({ url, options });
      return new Response(
        corrupt ? Buffer.alloc(optimized.length) : optimized,
        {
          headers: {
            "content-type": "image/webp",
            "cache-control": "max-age=14400",
            etag: receipt.objects[0].public.etag,
          },
        },
      );
    };
    const args = {
      catalogPath,
      receiptsPath,
      workdir: directory,
      expectedAccountId: receipt.accountId,
      fetcher,
    };
    await assert.rejects(
      acceptMCPReceipts({ ...args, expectedAccountId: "b".repeat(32) }),
      /wrong-account/,
    );
    assert.equal(calls.length, 0);
    const missingOriginal = structuredClone(receipt);
    delete missingOriginal.objects[0].original;
    await atomicJSON(receiptsPath, missingOriginal);
    await assert.rejects(acceptMCPReceipts(args), /Missing MCP object/);
    assert.equal(calls.length, 0);
    await atomicJSON(receiptsPath, receipt);
    await assert.rejects(acceptMCPReceipts(args), /checksum mismatch/);
    await assert.rejects(readFile(join(directory, "published-mcp.json")), {
      code: "ENOENT",
    });
    corrupt = false;
    const result = await acceptMCPReceipts(args);
    assert.equal(result.verified, 1);
    assert.equal(result.manifest, "published-mcp.json");
    assert.equal(calls.length, 2);
    for (const call of calls) {
      assert.equal(call.url, `${CDN}/${publicKey(prepared)}`);
      assert.equal(call.options.method, "GET");
      assert.equal(call.options.redirect, "error");
      assert.equal(call.options.headers, undefined);
    }
    const output = JSON.parse(
      await readFile(join(directory, "published-mcp.json")),
    );
    assert.deepEqual(output.images, [publicRecord(prepared)]);
    assert.equal(JSON.stringify(output).includes("original"), false);
    const evidence = JSON.parse(
      await readFile(join(directory, "mcp-publication-evidence.json")),
    ).images[row.meaningId];
    assert.equal(evidence.original.privateGetSha256Verified, false);
    assert.equal(evidence.original.privateHTTPMetadataVerified, false);
    assert.equal(evidence.original.bucketExposureVerifiedByThisTool, false);
    assert.equal(
      evidence.original.method,
      "operator-mcp-put-receipt-size-etag-md5",
    );
    assert.equal(evidence.original.sourceSha256, sha256(original));
    assert.equal(
      evidence.publicDelivery.method,
      "fresh-canonical-cdn-get-sha256",
    );
    assert.equal(evidence.publicDelivery.cacheControl, "max-age=14400");
    assert.equal(evidence.publicDelivery.immutableCacheVerified, false);
    // The stronger S3 path is deliberately not relaxed for this receipt mode.
    const s3 = new S3Publisher(
      credentials,
      async () =>
        new Response(optimized, {
          headers: {
            "content-type": "image/webp",
            "cache-control": "max-age=14400",
          },
        }),
    );
    await assert.rejects(
      s3.verifyCDN(publicKey(prepared), optimized),
      /delivery verification failed/,
    );
    const count = calls.length;
    await atomicJSON(catalogPath, {
      schemaVersion: 1,
      meanings: [{ ...meaning, meaningRevision: "e".repeat(64) }],
    });
    await assert.rejects(acceptMCPReceipts(args), /stale/);
    assert.equal(calls.length, count);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
