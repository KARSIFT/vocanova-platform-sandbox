import {
  readFile,
  writeFile,
  mkdir,
  realpath,
  open,
  unlink,
} from "node:fs/promises";
import { createRequire } from "node:module";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import process from "node:process";
import { log, error as logError } from "node:console";
import {
  PUBLIC_BUCKET,
  ARCHIVE_BUCKET,
  CDN,
  S3Publisher,
  sha256,
  readBounded,
  publicKey,
  archiveKey,
  publicRecord,
  validateIdentity,
  validateImage,
  validateAlt,
  canonicalPublicOrigin,
  atomicJSON,
  within,
} from "./library.mjs";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const json = async (path) => JSON.parse(await readFile(path, "utf8"));
const optionalJSON = async (path, fallback) => {
  try {
    return await json(path);
  } catch (error) {
    if (error.code === "ENOENT") return fallback;
    throw error;
  }
};

export async function imageProcessor() {
  // Use Next's already pinned optional image dependency, not an ad-hoc install.
  const require = createRequire(
    await realpath(resolve(root, "apps/web/node_modules/next/package.json")),
  );
  return require("sharp");
}

export async function validateAndOptimize(
  bytes,
  { preserve = false, sharp } = {},
) {
  sharp ??= await imageProcessor();
  const source = sharp(bytes, {
    failOn: "warning",
    limitInputPixels: 32_000_000,
    animated: false,
  });
  const metadata = await source.metadata();
  if (
    !["png", "jpeg", "webp"].includes(metadata.format) ||
    !metadata.width ||
    !metadata.height ||
    (metadata.pages ?? 1) !== 1
  ) {
    throw new Error("Only single-frame PNG, JPEG or WebP images are supported");
  }
  // A header probe alone does not establish that the complete image can decode.
  await source.clone().raw().toBuffer();
  if (
    preserve &&
    (metadata.format !== "webp" ||
      metadata.width > 960 ||
      metadata.height > 640 ||
      metadata.orientation)
  ) {
    throw new Error(
      "Reviewed legacy image is not a bounded, correctly oriented WebP",
    );
  }
  const optimized = preserve
    ? bytes
    : await source
        .rotate()
        .resize({
          width: 960,
          height: 640,
          fit: "inside",
          withoutEnlargement: true,
        })
        .webp({ quality: 82, effort: 5 })
        .toBuffer();
  if (optimized.length > 1024 * 1024)
    throw new Error("Optimized image exceeds 1 MiB; review before publishing");
  const decoded = await sharp(optimized, {
    failOn: "warning",
    limitInputPixels: 960 * 640,
  }).metadata();
  return {
    optimized,
    originalFormat: metadata.format,
    width: decoded.width,
    height: decoded.height,
  };
}

export function catalogue(data) {
  if (data.schemaVersion !== 1 || !Array.isArray(data.meanings))
    throw new Error("Unsupported catalogue");
  const map = new Map();
  for (const meaning of data.meanings) {
    validateIdentity(meaning);
    if (map.has(meaning.meaningId))
      throw new Error("Duplicate catalogue meaning");
    map.set(meaning.meaningId, meaning);
  }
  return map;
}

async function immutableFile(path, bytes) {
  await mkdir(dirname(path), { recursive: true, mode: 0o700 });
  try {
    await writeFile(path, bytes, { flag: "wx", mode: 0o600 });
  } catch (error) {
    if (error.code !== "EEXIST") throw error;
    if (sha256(await readBounded(path)) !== sha256(bytes))
      throw new Error("Existing local immutable image checksum mismatch", {
        cause: error,
      });
  }
}

export async function artifactPath(workdir, actual, expected) {
  if (actual !== expected)
    throw new Error("Prepared artifact path does not match immutable identity");
  const path = within(workdir, actual);
  // Resolve symlinks as well as lexical '..' before reading an artifact.
  return within(await realpath(workdir), await realpath(path));
}

export async function prepare({
  catalogPath,
  candidatesPath,
  workdir,
  existing = false,
  sharp,
}) {
  workdir = resolve(workdir);
  const map = catalogue(await json(catalogPath));
  const candidates = existing
    ? [...map.values()]
        .filter((row) => row.existing)
        .map((row) => ({
          meaningId: row.meaningId,
          meaningRevision: row.meaningRevision,
          source: resolve(root, `apps/web/public${row.existing.src}`),
          alt: row.existing.alt,
          approved: true,
          approvedBy: "existing-reviewed-library",
        }))
    : (await json(candidatesPath)).images;
  if (!Array.isArray(candidates) || candidates.length > 10051)
    throw new Error("Invalid candidate collection");
  const planPath = resolve(workdir, "prepared.json");
  const plan = await optionalJSON(planPath, { schemaVersion: 1, images: [] });
  if (plan.schemaVersion !== 1 || !Array.isArray(plan.images))
    throw new Error("Invalid existing publication plan");
  const records = new Map(plan.images.map((row) => [row.meaningId, row]));
  const seen = new Set();
  let prepared = 0;
  for (const candidate of candidates) {
    validateIdentity(candidate);
    const meaning = map.get(candidate.meaningId);
    if (!meaning || meaning.meaningRevision !== candidate.meaningRevision)
      throw new Error(
        "Candidate meaning is absent or stale; re-export catalogue",
      );
    if (seen.has(candidate.meaningId))
      throw new Error("Duplicate candidate meaning");
    seen.add(candidate.meaningId);
    if (
      candidate.approved !== true ||
      typeof candidate.approvedBy !== "string" ||
      !candidate.approvedBy.trim()
    ) {
      throw new Error(
        "Candidate requires an explicit completed visual approval",
      );
    }
    validateAlt(candidate.alt);
    if (
      existing &&
      meaning.existing.src !== `/images/meanings/${meaning.meaningId}.webp`
    )
      throw new Error("Unexpected reviewed legacy image path");
    const bytes = await readBounded(resolve(candidate.source));
    const image = await validateAndOptimize(bytes, {
      preserve: existing,
      sharp,
    });
    const row = {
      meaningId: meaning.meaningId,
      meaningRevision: meaning.meaningRevision,
      sha256: sha256(image.optimized),
      width: image.width,
      height: image.height,
      alt: candidate.alt.trim(),
      approved: true,
      originalSha256: sha256(bytes),
      originalFormat: image.originalFormat,
      approvedBy: candidate.approvedBy.trim(),
      originalProvenance: existing
        ? "reviewed-legacy-webp-preserved"
        : "provided-generator-output",
    };
    row.publicFile = `public/${publicKey(row)}`;
    row.originalFile = `private/${archiveKey(row)}`;
    const previous = records.get(row.meaningId);
    if (previous && JSON.stringify(previous) !== JSON.stringify(row))
      throw new Error(
        "Meaning already has a different prepared image; use a separate replacement batch",
      );
    await immutableFile(within(workdir, row.originalFile), bytes);
    await immutableFile(within(workdir, row.publicFile), image.optimized);
    records.set(row.meaningId, row);
    // Persist each completed asset so a partial batch can resume after failure.
    await atomicJSON(planPath, {
      schemaVersion: 1,
      images: [...records.values()],
    });
    prepared++;
  }
  return { prepared, total: records.size, missing: map.size - records.size };
}

export async function publish({
  catalogPath,
  workdir,
  limit = 100,
  offset,
  storage,
  publicBucket = PUBLIC_BUCKET,
  archiveBucket = ARCHIVE_BUCKET,
  publicBaseURL = CDN,
  sharp,
}) {
  // Object storage is portable; the backend's import-ready URL contract is fixed.
  canonicalPublicOrigin(publicBaseURL);
  if (
    !Number.isInteger(limit) ||
    limit < 1 ||
    limit > 500 ||
    (offset !== undefined && (!Number.isInteger(offset) || offset < 0)) ||
    publicBucket === archiveBucket
  )
    throw new Error(
      "Use a bounded batch (1..500), nonnegative offset and distinct public/private buckets",
    );
  workdir = resolve(workdir);
  await mkdir(workdir, { recursive: true, mode: 0o700 });
  const lock = await open(
    resolve(workdir, "publication.lock"),
    "wx",
    0o600,
  ).catch(() => {
    throw new Error(
      "Publication directory is locked; inspect the owning process before removing a stale lock",
    );
  });
  try {
    const map = catalogue(await json(catalogPath));
    const plan = await json(resolve(workdir, "prepared.json"));
    if (plan.schemaVersion !== 1 || !Array.isArray(plan.images))
      throw new Error("Invalid publication plan");
    const journalPath = resolve(workdir, "journal.json");
    const journal = await optionalJSON(journalPath, {
      schemaVersion: 1,
      entries: {},
    });
    if (
      journal.schemaVersion !== 1 ||
      !journal.entries ||
      typeof journal.entries !== "object"
    )
      throw new Error("Invalid publication journal");
    const manifestPath = resolve(workdir, "published.json");
    const published = await optionalJSON(manifestPath, {
      schemaVersion: 1,
      images: [],
    });
    if (published.schemaVersion !== 1 || !Array.isArray(published.images))
      throw new Error("Invalid published manifest");
    const records = new Map(
      published.images.map((row) => [row.meaningId, validateImage(row)]),
    );
    if (records.size !== published.images.length)
      throw new Error("Duplicate published meaning");
    for (const row of records.values()) {
      if (map.get(row.meaningId)?.meaningRevision !== row.meaningRevision)
        throw new Error(
          "Published meaning is stale; re-export and review the catalogue",
        );
    }
    let verified = 0;
    const seen = new Set();
    for (const row of plan.images) {
      validateImage(row);
      if (map.get(row.meaningId)?.meaningRevision !== row.meaningRevision)
        throw new Error("Prepared meaning is stale; do not publish");
      if (seen.has(row.meaningId))
        throw new Error("Duplicate prepared meaning");
      seen.add(row.meaningId);
      const previous = records.get(row.meaningId);
      if (
        previous &&
        Object.entries(publicRecord(row)).some(
          ([key, value]) => previous[key] !== value,
        )
      )
        throw new Error(
          "Meaning has conflicting published metadata; use a separate replacement batch",
        );
    }
    // Default resumes remaining assets; an explicit offset re-verifies a prior batch.
    const selected =
      offset === undefined
        ? plan.images
            .filter((row) => !records.has(row.meaningId))
            .slice(0, limit)
        : plan.images.slice(offset, offset + limit);
    for (const row of selected) {
      validateImage(row);
      if (map.get(row.meaningId)?.meaningRevision !== row.meaningRevision)
        throw new Error("Prepared meaning is stale; do not publish");
      const original = await readBounded(
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
        throw new Error("Prepared image checksum mismatch");
      const originalCheck = await validateAndOptimize(original, { sharp });
      const checked = await validateAndOptimize(optimized, {
        preserve: true,
        sharp,
      });
      if (
        originalCheck.originalFormat !== row.originalFormat ||
        checked.width !== row.width ||
        checked.height !== row.height
      )
        throw new Error("Prepared image metadata mismatch");
      const key = publicKey(row);
      const state = {
        meaningId: row.meaningId,
        meaningRevision: row.meaningRevision,
        sha256: row.sha256,
        originalVerified: false,
        publicVerified: false,
        deliveryVerified: false,
      };
      journal.entries[key] = state;
      await storage.ensure(
        archiveBucket,
        archiveKey(row),
        original,
        `image/${row.originalFormat}`,
        "private, max-age=0, no-store",
      );
      state.originalVerified = true;
      await atomicJSON(journalPath, journal);
      await storage.ensure(
        publicBucket,
        key,
        optimized,
        "image/webp",
        "public, max-age=31536000, immutable",
      );
      state.publicVerified = true;
      await atomicJSON(journalPath, journal);
      await storage.verifyCDN(key, optimized, publicBaseURL);
      state.deliveryVerified = true;
      await atomicJSON(journalPath, journal);
      // Backend importer receives no local paths, originals or approval identity.
      records.set(row.meaningId, publicRecord(row));
      await atomicJSON(manifestPath, {
        schemaVersion: 1,
        images: [...records.values()],
      });
      verified++;
    }
    return {
      verified,
      published: records.size,
      prepared: plan.images.length,
      remaining: plan.images.length - records.size,
    };
  } finally {
    await lock.close();
    await unlink(resolve(workdir, "publication.lock"));
  }
}

async function main() {
  const [command, ...values] = process.argv.slice(2);
  const options = {};
  for (let i = 0; i < values.length; i += 2) {
    if (
      !values[i]?.startsWith("--") ||
      !values[i + 1] ||
      values[i + 1].startsWith("--")
    )
      throw new Error("Arguments are --name value pairs");
    options[values[i].slice(2)] = values[i + 1];
  }
  if (!options.catalog || !options.workdir)
    throw new Error(
      "Required: prepare-existing|prepare|publish --catalog path --workdir path [--candidates path] [--limit 100] [--offset 0]",
    );
  const args = {
    catalogPath: options.catalog,
    candidatesPath: options.candidates,
    workdir: options.workdir,
  };
  if (command === "prepare-existing" || command === "prepare") {
    if (command === "prepare" && !options.candidates)
      throw new Error("--candidates is required");
    log(
      JSON.stringify(
        await prepare({ ...args, existing: command === "prepare-existing" }),
      ),
    );
  } else if (command === "publish") {
    const credentials = {
      endpoint: process.env.MEANING_MEDIA_S3_ENDPOINT,
      region: process.env.MEANING_MEDIA_S3_REGION,
      accessKeyId: process.env.MEANING_MEDIA_S3_ACCESS_KEY_ID,
      secretAccessKey: process.env.MEANING_MEDIA_S3_SECRET_ACCESS_KEY,
    };
    if (Object.values(credentials).some((value) => !value))
      throw new Error(
        "Four MEANING_MEDIA_S3_* environment variables are required; no credential files are loaded",
      );
    const storage = new S3Publisher(credentials);
    log(
      JSON.stringify(
        await publish({
          ...args,
          limit: Number(options.limit ?? 100),
          ...(options.offset === undefined
            ? {}
            : { offset: Number(options.offset) }),
          storage,
          publicBucket:
            process.env.MEANING_MEDIA_PUBLIC_BUCKET ?? PUBLIC_BUCKET,
          archiveBucket:
            process.env.MEANING_MEDIA_ARCHIVE_BUCKET ?? ARCHIVE_BUCKET,
          publicBaseURL: process.env.MEANING_MEDIA_PUBLIC_BASE_URL ?? CDN,
        }),
      ),
    );
  } else throw new Error("Unsupported command");
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main().catch((error) => {
    logError(`meaning-media: ${error.message}`);
    process.exitCode = 1;
  });
}
