// Separate, truthful evidence path for an operator's completed raw MCP PUTs.
// No storage writes or model calls; only local reads/writes and canonical CDN GET.
import { createHash } from "node:crypto";
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
  responseBytes,
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
  if (!receipt || typeof receipt !== "object")
    throw new Error("Missing MCP object PUT receipt");
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
      "MCP PUT receipt does not match the local image key, size, source SHA-256, single-part ETag/MD5 or metadata",
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

async function canonicalDelivery(key, bytes, fetcher) {
  let response;
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      response = await fetcher(`${CDN}/${key}`, {
        method: "GET",
        redirect: "error",
        signal: globalThis.AbortSignal.timeout(60_000),
      });
      break;
    } catch {
      if (attempt === 2)
        throw new Error("Canonical CDN receipt verification request failed");
      await setTimeout(250 * 4 ** attempt);
    }
  }
  if (
    response.status !== 200 ||
    response.headers.get("content-type") !== "image/webp"
  ) {
    await response.body?.cancel();
    throw new Error(
      `Canonical CDN receipt verification failed (HTTP ${response.status})`,
    );
  }
  const received = await responseBytes(response, bytes.length);
  if (sha256(received) !== sha256(bytes))
    throw new Error("Canonical CDN receipt checksum mismatch");
  return {
    method: "fresh-canonical-cdn-get-sha256",
    url: `${CDN}/${key}`,
    byteLength: received.length,
    sha256: sha256(received),
    contentType: response.headers.get("content-type"),
    cacheControl: response.headers.get("cache-control"),
    etag: response.headers.get("etag"),
    immutableCacheVerified: false,
    verifiedAt: new Date().toISOString(),
  };
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
  fetcher = globalThis.fetch,
  sharp,
}) {
  if (!["cloud", "local"].includes(originalArchive))
    throw new Error("Original archive mode must be cloud or local");
  const local = originalArchive === "local";
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
    receipts.transport !==
      (local
        ? "cloudflare-mcp-public-only-local-archive-v1"
        : "cloudflare-mcp-raw-put-v1") ||
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
      "Invalid or wrong-account/bucket MCP receipt collection (1..500 objects required)",
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
          "Prepared meaning is stale; do not accept MCP receipts",
        );
      if (planned.has(row.meaningId))
        throw new Error("Duplicate prepared meaning");
      planned.set(row.meaningId, row);
    }
    const manifestName = local
      ? "published-mcp-local.json"
      : "published-mcp.json";
    const manifestPath = resolve(workdir, manifestName);
    const manifest = await optionalJSON(manifestPath, {
      schemaVersion: 1,
      images: [],
    });
    if (manifest.schemaVersion !== 1 || !Array.isArray(manifest.images))
      throw new Error("Invalid MCP publication manifest");
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
          "Existing MCP publication metadata conflicts with the approved current plan",
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
          "Missing, duplicate or stale MCP receipt meaning identity",
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
      const publicProof = validateMCPObjectReceipt(receipt.public, {
        key: publicKey(row),
        bytes: optimized,
        contentType: "image/webp",
      });
      checked.push({ row, optimized, originalProof, publicProof });
    }
    const evidencePath = resolve(
      workdir,
      local
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
      throw new Error("Conflicting MCP publication evidence");
    for (const { row, optimized, originalProof, publicProof } of checked) {
      const delivery = await canonicalDelivery(
        publicKey(row),
        optimized,
        fetcher,
      );
      evidence.images[row.meaningId] = {
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
        publicPut: {
          ...publicProof,
          method: "operator-mcp-put-receipt-size-etag-md5",
        },
        publicDelivery: delivery,
      };
      // Evidence precedes import-ready metadata; a failed GET never adds a record.
      await atomicJSON(evidencePath, evidence);
      published.set(row.meaningId, publicRecord(row));
      await atomicJSON(manifestPath, {
        schemaVersion: 1,
        images: [...published.values()],
      });
    }
    return {
      verified: checked.length,
      published: published.size,
      prepared: planned.size,
      remaining: planned.size - published.size,
      privateVerification: local
        ? "local archive full-read SHA-256+MD5; no private cloud request"
        : "operator PUT receipt size+ETag/MD5; no private GET SHA-256",
      manifest: manifestName,
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
      }),
    ),
  );
}
if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main().catch((error) => {
    logError(`meaning-media MCP receipts: ${error.message}`);
    process.exitCode = 1;
  });
}
