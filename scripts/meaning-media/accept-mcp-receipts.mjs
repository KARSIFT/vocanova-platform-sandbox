// Separate, truthful evidence paths for operator-recorded MCP or REST results.
// No storage writes or model calls; only local reads/writes and canonical CDN GET.
import { createHash } from "node:crypto";
import { Buffer } from "node:buffer";
import { performance } from "node:perf_hooks";
import {
  readFile,
  mkdir,
  open,
  unlink,
  lstat,
  realpath,
} from "node:fs/promises";
import { resolve, isAbsolute } from "node:path";
import { fileURLToPath } from "node:url";
import { log, error as logError } from "node:console";
import process from "node:process";
import { setTimeout } from "node:timers/promises";
import {
  CDN,
  PUBLIC_BUCKET,
  ARCHIVE_BUCKET,
  sha256,
  publicKey,
  archiveKey,
  publicRecord,
  validateImage,
  readBounded,
  atomicJSON,
  within,
} from "./library.mjs";
import { catalogue, artifactPath, validateAndOptimize } from "./manage.mjs";

const json = async (path) => JSON.parse(await readFile(path, "utf8"));
async function optionalJSON(path, fallback) {
  try {
    return await json(path);
  } catch (error) {
    if (error.code === "ENOENT") return fallback;
    throw error;
  }
}

async function readLocalOriginal(root, row) {
  const key = archiveKey(row);
  const file = within(root, key);
  // Canonical identity-derived paths only. Reject symlinks in the root, any
  // intermediate directory, or the final file, including in-root aliases.
  if (
    (await realpath(root)) !== root ||
    (await realpath(file)) !== file ||
    !(await lstat(file)).isFile()
  )
    throw new Error(
      "Local original archive must contain regular canonical files without symlinks",
    );
  const bytes = await readBounded(file);
  if (sha256(bytes) !== row.originalSha256)
    throw new Error("Local original archive checksum mismatch");
  return {
    bytes,
    proof: {
      method: "local-archive-full-read-sha256-md5",
      key,
      byteLength: bytes.length,
      sourceSha256: sha256(bytes),
      md5: createHash("md5").update(bytes).digest("hex"),
      contentType: `image/${row.originalFormat}`,
      cloudOriginalUploadedByThisCommand: false,
      cloudOriginalPresenceChecked: false,
      verifiedAt: new Date().toISOString(),
    },
  };
}

export function validateMCPObjectReceipt(receipt, { key, bytes, contentType }) {
  return validateObjectReceipt(receipt, { key, bytes, contentType }, "MCP PUT");
}

function validateObjectReceipt(receipt, { key, bytes, contentType }, source) {
  if (!receipt || typeof receipt !== "object")
    throw new Error(
      `Missing ${source === "MCP PUT" ? "MCP object PUT" : source} receipt`,
    );
  const etag =
    typeof receipt.etag === "string" &&
    receipt.etag.startsWith('"') &&
    receipt.etag.endsWith('"')
      ? receipt.etag.slice(1, -1)
      : receipt.etag;
  const md5 = createHash("md5").update(bytes).digest("hex");
  if (
    receipt.key !== key ||
    receipt.size !== bytes.length ||
    receipt.sourceSha256 !== sha256(bytes) ||
    receipt.contentType !== contentType ||
    !/^[0-9a-f]{32}$/.test(etag) ||
    etag !== md5 ||
    !(
      receipt.cacheControl === null ||
      (typeof receipt.cacheControl === "string" &&
        receipt.cacheControl.length <= 1024 &&
        !receipt.cacheControl.includes("\r") &&
        !receipt.cacheControl.includes("\n"))
    )
  )
    throw new Error(
      `${source} receipt does not match the local image key, size, source SHA-256, single-part ETag/MD5 or metadata`,
    );
  return {
    key,
    byteLength: bytes.length,
    sourceSha256: sha256(bytes),
    etag,
    md5,
    contentType,
    reportedCacheControl: receipt.cacheControl,
  };
}

async function canonicalDelivery(key, bytes, fetcher, onAttempts) {
  const attempts = [];
  const expectedSHA256 = sha256(bytes);
  for (let attempt = 0; attempt < 3; attempt++) {
    const detail = {
      attempt: attempt + 1,
      startedAt: new Date().toISOString(),
      phase: "fetch",
      outcome: "pending",
      status: null,
      bodyBytesRead: 0,
      expectedBodyBytes: bytes.length,
      expectedBodySHA256: expectedSHA256,
      url: `${CDN}/${key}`,
    };
    attempts.push(detail);
    await onAttempts(attempts);
    const started = performance.now();
    let response, delivery, failure;
    try {
      response = await fetcher(`${CDN}/${key}`, {
        method: "GET",
        redirect: "error",
        signal: globalThis.AbortSignal.timeout(60_000),
      });
      detail.status = response.status;
      detail.contentType = response.headers.get("content-type");
      detail.phase = "headers";
      if (response.status !== 200 || detail.contentType !== "image/webp")
        throw new Error(
          `Canonical CDN receipt verification failed (HTTP ${response.status})`,
        );
      detail.phase = "body-size";
      if (Number(response.headers.get("content-length")) > bytes.length)
        throw new Error("Remote image exceeds byte limit");
      detail.phase = "body";
      const chunks = [];
      for await (const chunk of response.body ?? []) {
        detail.bodyBytesRead += chunk.length;
        if (detail.bodyBytesRead > bytes.length) {
          detail.phase = "body-size";
          throw new Error("Remote image exceeds byte limit");
        }
        chunks.push(chunk);
      }
      detail.phase = "checksum";
      const received = Buffer.concat(chunks);
      detail.bodySHA256 = sha256(received);
      if (
        received.length !== bytes.length ||
        detail.bodySHA256 !== expectedSHA256
      )
        throw new Error("Canonical CDN receipt checksum mismatch");
      detail.outcome = "verified";
      delivery = {
        method: "fresh-canonical-cdn-get-sha256",
        url: `${CDN}/${key}`,
        byteLength: received.length,
        sha256: detail.bodySHA256,
        contentType: detail.contentType,
        cacheControl: response.headers.get("cache-control"),
        etag: response.headers.get("etag"),
        immutableCacheVerified: false,
        verifiedAt: new Date().toISOString(),
      };
    } catch (error) {
      failure = error;
      detail.outcome = "failed";
      detail.errorName = error.name;
      detail.errorMessage = String(error.message).slice(0, 512);
      if (error.code ?? error.cause?.code)
        detail.errorCode = String(error.code ?? error.cause.code).slice(0, 128);
      detail.willRetry =
        attempt < 2 && ["fetch", "body"].includes(detail.phase);
      // Do not let cancellation hide the original status/checksum/body error.
      void response?.body?.cancel().catch(() => {});
    }
    detail.finishedAt = new Date().toISOString();
    detail.elapsedMs = Math.round(performance.now() - started);
    await onAttempts(attempts);
    if (delivery) return delivery;
    if (!detail.willRetry) {
      const message =
        detail.phase === "fetch"
          ? "Canonical CDN receipt verification request failed"
          : failure.message;
      throw new Error(
        `${message} [key=${key}; phase=${detail.phase}; attempt=${attempt + 1}/3; bodyBytes=${detail.bodyBytesRead}/${bytes.length}]`,
        { cause: failure },
      );
    }
    await setTimeout(250 * 4 ** attempt);
  }
}

export async function acceptMCPReceipts({
  catalogPath,
  receiptsPath,
  workdir,
  expectedAccountId,
  publicBucket = PUBLIC_BUCKET,
  archiveBucket = ARCHIVE_BUCKET,
  originalArchive = "cloud",
  localArchiveRoot,
  receiptSource = "cloudflare-mcp",
  fetcher = globalThis.fetch,
  sharp,
  concurrency = 4,
}) {
  if (!Number.isSafeInteger(concurrency) || concurrency < 1 || concurrency > 4)
    throw new Error("CDN receipt concurrency must be an integer from 1 to 4");
  if (!["cloudflare-mcp", "cloudflare-rest"].includes(receiptSource))
    throw new Error("Receipt source must be cloudflare-mcp or cloudflare-rest");
  const rest = receiptSource === "cloudflare-rest";
  if (!["cloud", "local"].includes(originalArchive))
    throw new Error("Original archive mode must be cloud or local");
  const local = originalArchive === "local";
  if (rest && !local)
    throw new Error("REST receipt source requires local original archive mode");
  const sourceLabel = rest ? "REST" : "MCP";
  const expectedTransport = rest
    ? "cloudflare-rest-public-only-local-archive-v1"
    : local
      ? "cloudflare-mcp-public-only-local-archive-v1"
      : "cloudflare-mcp-raw-put-v1";
  if (local) {
    if (typeof localArchiveRoot !== "string" || !isAbsolute(localArchiveRoot))
      throw new Error(
        "Local original archive requires an absolute archive root",
      );
    localArchiveRoot = resolve(localArchiveRoot);
    if (
      !(await lstat(localArchiveRoot)).isDirectory() ||
      (await realpath(localArchiveRoot)) !== localArchiveRoot
    )
      throw new Error(
        "Local original archive root must be a real directory without symlinks",
      );
  } else if (localArchiveRoot !== undefined) {
    throw new Error(
      "Local archive root requires explicit local original archive mode",
    );
  }
  if (
    !/^[0-9a-f]{32}$/.test(expectedAccountId) ||
    (!local && publicBucket === archiveBucket)
  )
    throw new Error(
      "Expected Cloudflare account and distinct buckets are required",
    );
  const receiptBytes = await readBounded(receiptsPath, 16 * 1024 * 1024);
  const receipts = JSON.parse(receiptBytes.toString("utf8"));
  if (
    receipts.schemaVersion !== 1 ||
    receipts.transport !== expectedTransport ||
    receipts.accountId !== expectedAccountId ||
    receipts.publicBucket !== publicBucket ||
    (local
      ? Object.hasOwn(receipts, "archiveBucket")
      : receipts.archiveBucket !== archiveBucket) ||
    !Array.isArray(receipts.objects) ||
    receipts.objects.length < 1 ||
    receipts.objects.length > 500
  )
    throw new Error(
      `Invalid or wrong-account/bucket ${sourceLabel} receipt collection (1..500 objects required)`,
    );
  workdir = resolve(workdir);
  await mkdir(workdir, { recursive: true, mode: 0o700 });
  const lock = await open(
    resolve(workdir, "publication.lock"),
    "wx",
    0o600,
  ).catch(() => {
    throw new Error(
      "Publication directory is locked; inspect its owning process",
    );
  });
  try {
    const map = catalogue(await json(catalogPath));
    const plan = await json(resolve(workdir, "prepared.json"));
    if (plan.schemaVersion !== 1 || !Array.isArray(plan.images))
      throw new Error("Invalid publication plan");
    const planned = new Map();
    for (const row of plan.images) {
      validateImage(row);
      if (map.get(row.meaningId)?.meaningRevision !== row.meaningRevision)
        throw new Error(
          `Prepared meaning is stale; do not accept ${sourceLabel} receipts`,
        );
      if (planned.has(row.meaningId))
        throw new Error("Duplicate prepared meaning");
      planned.set(row.meaningId, row);
    }
    const manifestName = rest
      ? "published-rest-local.json"
      : local
        ? "published-mcp-local.json"
        : "published-mcp.json";
    const manifestPath = resolve(workdir, manifestName);
    const manifest = await optionalJSON(manifestPath, {
      schemaVersion: 1,
      images: [],
    });
    if (manifest.schemaVersion !== 1 || !Array.isArray(manifest.images))
      throw new Error(`Invalid ${sourceLabel} publication manifest`);
    const published = new Map();
    for (const row of manifest.images) {
      validateImage(row);
      if (
        published.has(row.meaningId) ||
        !planned.has(row.meaningId) ||
        Object.entries(publicRecord(planned.get(row.meaningId))).some(
          ([key, value]) => row[key] !== value,
        )
      )
        throw new Error(
          `Existing ${sourceLabel} publication metadata conflicts with the approved current plan`,
        );
      published.set(row.meaningId, row);
    }
    // Validate every receipt and local source before the first public HTTP read.
    const seen = new Set();
    const checked = [];
    for (const receipt of receipts.objects) {
      const row = planned.get(receipt.meaningId);
      if (
        !row ||
        (local && Object.hasOwn(receipt, "original")) ||
        seen.has(receipt.meaningId) ||
        row.meaningRevision !== receipt.meaningRevision ||
        row.sha256 !== receipt.sha256
      )
        throw new Error(
          `Missing, duplicate or stale ${sourceLabel} receipt meaning identity`,
        );
      if (
        rest &&
        !["uploaded", "reconciled-existing"].includes(receipt.operation)
      )
        throw new Error(
          "REST object operation must explicitly be uploaded or reconciled-existing",
        );
      seen.add(row.meaningId);
      const localOriginal = local
        ? await readLocalOriginal(localArchiveRoot, row)
        : null;
      const original = local
        ? localOriginal.bytes
        : await readBounded(
            await artifactPath(
              workdir,
              row.originalFile,
              `private/${archiveKey(row)}`,
            ),
          );
      const optimized = await readBounded(
        await artifactPath(workdir, row.publicFile, `public/${publicKey(row)}`),
        1024 * 1024,
      );
      if (
        sha256(original) !== row.originalSha256 ||
        sha256(optimized) !== row.sha256
      )
        throw new Error("Local prepared source checksum mismatch");
      const originalCheck = await validateAndOptimize(original, { sharp });
      const publicCheck = await validateAndOptimize(optimized, {
        preserve: true,
        sharp,
      });
      if (
        originalCheck.originalFormat !== row.originalFormat ||
        publicCheck.width !== row.width ||
        publicCheck.height !== row.height
      )
        throw new Error("Local prepared source metadata mismatch");
      const originalProof = local
        ? localOriginal.proof
        : validateMCPObjectReceipt(receipt.original, {
            key: archiveKey(row),
            bytes: original,
            contentType: `image/${row.originalFormat}`,
          });
      const publicProof = validateObjectReceipt(
        receipt.public,
        {
          key: publicKey(row),
          bytes: optimized,
          contentType: "image/webp",
        },
        rest ? "REST object metadata" : "MCP PUT",
      );
      checked.push({
        row,
        optimized,
        originalProof,
        publicProof,
        operation: receipt.operation,
      });
    }
    const evidencePath = resolve(
      workdir,
      rest
        ? "rest-local-publication-evidence.json"
        : local
          ? "mcp-local-publication-evidence.json"
          : "mcp-publication-evidence.json",
    );
    const evidence = await optionalJSON(evidencePath, {
      schemaVersion: 1,
      transport: receipts.transport,
      accountId: expectedAccountId,
      publicBucket,
      ...(local ? { localArchiveRoot } : { archiveBucket }),
      images: {},
    });
    if (
      evidence.schemaVersion !== 1 ||
      evidence.transport !== receipts.transport ||
      evidence.accountId !== expectedAccountId ||
      evidence.publicBucket !== publicBucket ||
      (local
        ? evidence.localArchiveRoot !== localArchiveRoot ||
          Object.hasOwn(evidence, "archiveBucket")
        : evidence.archiveBucket !== archiveBucket) ||
      !evidence.images ||
      typeof evidence.images !== "object"
    )
      throw new Error(`Conflicting ${sourceLabel} publication evidence`);
    const entryFor = (
      { row, originalProof, publicProof, operation },
      delivery,
    ) => ({
      meaningRevision: row.meaningRevision,
      sha256: row.sha256,
      receiptFileSha256: sha256(receiptBytes),
      original: local
        ? originalProof
        : {
            ...originalProof,
            method: "operator-mcp-put-receipt-size-etag-md5",
            privateGetSha256Verified: false,
            privateHTTPMetadataVerified: false,
            bucketExposureVerifiedByThisTool: false,
          },
      ...(rest
        ? {
            publicObject: {
              ...publicProof,
              operation,
              newPutReported: operation === "uploaded",
              method:
                operation === "uploaded"
                  ? "operator-rest-put-receipt-size-etag-md5"
                  : "operator-rest-existing-metadata-readback-size-etag-md5",
            },
          }
        : {
            publicPut: {
              ...publicProof,
              method: "operator-mcp-put-receipt-size-etag-md5",
            },
          }),
      publicDelivery: delivery,
    });
    const reusable = (item) => {
      const prior = evidence.images[item.row.meaningId];
      const delivery = prior?.publicDelivery;
      const latestAttempt = evidence.cdnAttempts?.[item.row.meaningId]?.at(-1);
      if (
        !published.has(item.row.meaningId) ||
        !prior ||
        !/^[a-f0-9]{64}$/.test(prior.receiptFileSha256 ?? "") ||
        delivery?.method !== "fresh-canonical-cdn-get-sha256" ||
        delivery.url !== `${CDN}/${publicKey(item.row)}` ||
        delivery.sha256 !== item.row.sha256 ||
        delivery.byteLength !== item.optimized.length ||
        delivery.contentType !== "image/webp" ||
        delivery.immutableCacheVerified !== false ||
        !Number.isFinite(Date.parse(delivery.verifiedAt))
      )
        return false;
      // Never hide a later failed/interrupted check behind an older successful
      // image entry. Legacy accepted entries have no per-attempt history.
      if (
        latestAttempt &&
        (latestAttempt.outcome !== "verified" ||
          latestAttempt.status !== 200 ||
          latestAttempt.bodySHA256 !== item.row.sha256 ||
          latestAttempt.bodyBytesRead !== item.optimized.length)
      )
        return false;
      const expected = entryFor(item, delivery);
      if (
        prior.meaningRevision !== expected.meaningRevision ||
        prior.sha256 !== expected.sha256
      )
        return false;
      for (const key of ["original", rest ? "publicObject" : "publicPut"])
        for (const [field, value] of Object.entries(expected[key]))
          if (field !== "verifiedAt" && prior[key]?.[field] !== value)
            return false;
      // Collection SHA may differ for an exact subset. The original attribution
      // stays untouched; reuse requires the same individually validated receipt,
      // current local original/public bytes and prior successful full CDN proof.
      return true;
    };
    const pending = checked.filter((item) => !(rest && reusable(item)));
    const reusedAccepted = checked.length - pending.length;
    let cursor = 0,
      freshCDNVerified = 0,
      stopped = false,
      failure;
    let writes = Promise.resolve();
    const persist = (operation) => {
      writes = writes.then(operation);
      return writes;
    };
    const worker = async () => {
      while (!stopped && cursor < pending.length) {
        const item = pending[cursor++];
        const { row, optimized } = item;
        const history = evidence.cdnAttempts?.[row.meaningId] ?? [];
        try {
          const delivery = await canonicalDelivery(
            publicKey(row),
            optimized,
            fetcher,
            (attempts) => {
              const last = attempts.at(-1);
              if (last.outcome === "failed" && !last.willRetry) stopped = true;
              return persist(async () => {
                evidence.cdnAttempts ??= {};
                evidence.cdnAttempts[row.meaningId] = [
                  ...history,
                  ...attempts.map((attempt) => ({
                    ...attempt,
                    receiptFileSha256: sha256(receiptBytes),
                  })),
                ];
                await atomicJSON(evidencePath, evidence);
              });
            },
          );
          await persist(async () => {
            evidence.images[row.meaningId] = entryFor(item, delivery);
            // Serialized evidence precedes import-ready metadata. Other workers
            // cannot race a stale manifest write over a successful acceptance.
            await atomicJSON(evidencePath, evidence);
            published.set(row.meaningId, publicRecord(row));
            await atomicJSON(manifestPath, {
              schemaVersion: 1,
              images: [...published.values()],
            });
            freshCDNVerified++;
          });
        } catch (error) {
          stopped = true;
          failure ??= error;
        }
      }
    };
    // Stop assigning work after a terminal failure, but drain bounded in-flight
    // GETs and durable writes before releasing the exclusive publication lock.
    await Promise.all(
      Array.from({ length: Math.min(concurrency, pending.length) }, worker),
    );
    await writes;
    if (failure) throw failure;
    if (freshCDNVerified + reusedAccepted !== checked.length) {
      throw new Error("Incomplete CDN receipt acceptance");
    }
    return {
      // verified counts validated receipt/local-byte bindings, including reuse;
      // freshCDNVerified is the number of new successful network body checks.
      verified: checked.length,
      freshCDNVerified,
      reusedAccepted,
      published: published.size,
      prepared: planned.size,
      remaining: planned.size - published.size,
      privateVerification: local
        ? "local archive full-read SHA-256+MD5; no private cloud request"
        : "operator PUT receipt size+ETag/MD5; no private GET SHA-256",
      manifest: manifestName,
      ...(rest
        ? {
            receiptSource,
            uploadedReceipts: checked.filter(
              ({ operation }) => operation === "uploaded",
            ).length,
            reconciledExistingReceipts: checked.filter(
              ({ operation }) => operation === "reconciled-existing",
            ).length,
            storageWritesByThisCommand: false,
          }
        : {}),
    };
  } finally {
    await lock.close();
    await unlink(resolve(workdir, "publication.lock"));
  }
}

async function main() {
  const options = {};
  const values = process.argv.slice(2);
  for (let i = 0; i < values.length; i += 2) {
    if (
      !values[i]?.startsWith("--") ||
      !values[i + 1] ||
      values[i + 1].startsWith("--")
    )
      throw new Error("Arguments are --name value pairs");
    options[values[i].slice(2)] = values[i + 1];
  }
  if (
    !options.catalog ||
    !options.receipts ||
    !options.workdir ||
    !options["account-id"]
  )
    throw new Error(
      "Required: --catalog FILE --receipts FILE --workdir DIR --account-id EXPECTED_ACCOUNT_ID",
    );
  log(
    JSON.stringify(
      await acceptMCPReceipts({
        catalogPath: options.catalog,
        receiptsPath: options.receipts,
        workdir: options.workdir,
        expectedAccountId: options["account-id"],
        originalArchive: options["original-archive"] ?? "cloud",
        localArchiveRoot: options["local-archive-root"],
        receiptSource: options["receipt-source"] ?? "cloudflare-mcp",
        concurrency:
          options.concurrency === undefined ? 4 : Number(options.concurrency),
      }),
    ),
  );
}
if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main().catch((error) => {
    logError(`meaning-media receipts: ${error.message}`);
    process.exitCode = 1;
  });
}
