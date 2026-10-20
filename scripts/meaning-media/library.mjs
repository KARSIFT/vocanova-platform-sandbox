import { createHash, createHmac, randomUUID } from "node:crypto";
import { Buffer } from "node:buffer";
import { URL } from "node:url";
import { setTimeout } from "node:timers/promises";
import { mkdir, readFile, rename, stat, writeFile } from "node:fs/promises";
import { dirname, resolve, sep } from "node:path";

export const PUBLIC_BUCKET = "vocanova-meaning-images";
export const ARCHIVE_BUCKET = "vocanova-meaning-originals";
export const CDN = "https://images.vocanova.site";
export const MAX_ORIGINAL_BYTES = 32 * 1024 * 1024;
const uuid = /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/;
const digest = /^[0-9a-f]{64}$/;
export const sha256 = (bytes) =>
  createHash("sha256").update(bytes).digest("hex");

export function validateIdentity(row) {
  if (!uuid.test(row.meaningId) || !digest.test(row.meaningRevision)) {
    throw new Error("Invalid meaning identity or revision");
  }
}

export function validateAlt(alt) {
  if (
    typeof alt !== "string" ||
    !alt ||
    alt !== alt.trim() ||
    alt.length > 500 ||
    ["\x00", "\r", "\n"].some((character) => alt.includes(character))
  )
    throw new Error(
      "Approved alt text must be trimmed and contain no NUL or line breaks",
    );
}

export function canonicalPublicOrigin(value = CDN) {
  const base = new URL(value);
  if (
    base.origin !== CDN ||
    base.username ||
    base.password ||
    base.search ||
    base.hash ||
    base.pathname !== "/"
  )
    throw new Error(`Import-ready images must be verified at ${CDN}`);
  return base.origin;
}

export function validateImage(row) {
  validateIdentity(row);
  validateAlt(row.alt);
  if (
    row.approved !== true ||
    !digest.test(row.sha256) ||
    !Number.isInteger(row.width) ||
    row.width < 1 ||
    row.width > 960 ||
    !Number.isInteger(row.height) ||
    row.height < 1 ||
    row.height > 640
  )
    throw new Error(
      "Only approved, bounded images with alt text may be published",
    );
  return row;
}

export function publicKey(row) {
  validateImage(row);
  return `meanings/${row.meaningId}/${row.meaningRevision}/${row.sha256}.webp`;
}

export function archiveKey(row) {
  validateIdentity(row);
  if (
    !digest.test(row.originalSha256) ||
    !["png", "jpeg", "webp"].includes(row.originalFormat)
  ) {
    throw new Error("Invalid original image identity");
  }
  return `originals/${row.meaningId}/${row.meaningRevision}/${row.originalSha256}.${row.originalFormat}`;
}

export function publicRecord(row) {
  validateImage(row);
  const { meaningId, meaningRevision, sha256, width, height, alt } = row;
  return {
    meaningId,
    meaningRevision,
    sha256,
    width,
    height,
    alt,
    approved: true,
  };
}

export async function readBounded(path, max = MAX_ORIGINAL_BYTES) {
  if ((await stat(path)).size > max)
    throw new Error("Image exceeds byte limit");
  const data = await readFile(path);
  if (data.length > max) throw new Error("Image exceeds byte limit");
  return data;
}

export function within(directory, relative) {
  if (typeof relative !== "string" || !relative)
    throw new Error("Missing artifact path");
  const root = resolve(directory);
  const path = resolve(root, relative);
  if (!path.startsWith(root + sep))
    throw new Error("Artifact path escapes working directory");
  return path;
}

export async function atomicJSON(path, value) {
  await mkdir(dirname(path), { recursive: true, mode: 0o700 });
  const temporary = `${path}.${randomUUID()}.tmp`;
  await writeFile(temporary, JSON.stringify(value, null, 2) + "\n", {
    mode: 0o600,
    flag: "wx",
  });
  await rename(temporary, path);
}

// Path-style S3 works with R2 (region 'auto'), S3 and compatible object stores.
// No SDK, token-bearing URLs, bucket listing, deletion or policy writes.
export function signRequest({
  endpoint,
  region,
  accessKeyId,
  secretAccessKey,
  method,
  bucket,
  key,
  body = Buffer.alloc(0),
  headers = {},
  now = new Date(),
}) {
  const base = new URL(endpoint);
  if (
    base.protocol !== "https:" ||
    base.username ||
    base.password ||
    base.pathname !== "/" ||
    base.search ||
    base.hash ||
    !/^[a-z0-9-]+$/.test(region) ||
    !accessKeyId ||
    !secretAccessKey
  ) {
    throw new Error(
      "A plain HTTPS S3 endpoint, region and credentials are required",
    );
  }
  if (
    !/^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$/.test(bucket) ||
    !/^[a-z0-9/.-]+$/.test(key) ||
    key.includes("..")
  ) {
    throw new Error("Unsupported publication destination");
  }
  if (!["HEAD", "GET", "PUT"].includes(method))
    throw new Error("Unsupported operation");
  const host = base.host;
  const path = `/${bucket}/${key}`;
  const timestamp = now.toISOString().replace(/[:-]|\.\d{3}/g, "");
  const date = timestamp.slice(0, 8);
  const payloadHash = sha256(body);
  const signed = {
    ...headers,
    host,
    "x-amz-date": timestamp,
    "x-amz-content-sha256": payloadHash,
  };
  const names = Object.keys(signed).sort();
  const canonicalHeaders = names
    .map((name) => `${name}:${String(signed[name]).trim()}\n`)
    .join("");
  const signedHeaders = names.join(";");
  const canonical = [
    method,
    path,
    "",
    canonicalHeaders,
    signedHeaders,
    payloadHash,
  ].join("\n");
  const scope = `${date}/${region}/s3/aws4_request`;
  const hmac = (key, value) => createHmac("sha256", key).update(value).digest();
  const signingKey = hmac(
    hmac(hmac(hmac(`AWS4${secretAccessKey}`, date), region), "s3"),
    "aws4_request",
  );
  const signature = hmac(
    signingKey,
    `AWS4-HMAC-SHA256\n${timestamp}\n${scope}\n${sha256(canonical)}`,
  ).toString("hex");
  signed.authorization = `AWS4-HMAC-SHA256 Credential=${accessKeyId}/${scope}, SignedHeaders=${signedHeaders}, Signature=${signature}`;
  return { url: `https://${host}${path}`, headers: signed };
}

export async function responseBytes(response, max) {
  if (Number(response.headers.get("content-length")) > max) {
    await response.body?.cancel();
    throw new Error("Remote image exceeds byte limit");
  }
  const chunks = [];
  let length = 0;
  for await (const chunk of response.body ?? []) {
    length += chunk.length;
    if (length > max) {
      await response.body?.cancel().catch(() => {});
      throw new Error("Remote image exceeds byte limit");
    }
    chunks.push(chunk);
  }
  return Buffer.concat(chunks);
}

export class S3Publisher {
  constructor(credentials, fetcher = globalThis.fetch) {
    this.credentials = credentials;
    this.fetcher = fetcher;
  }
  async request(method, bucket, key, body, headers = {}) {
    for (let attempt = 0; attempt < 3; attempt++) {
      const signed = signRequest({
        ...this.credentials,
        method,
        bucket,
        key,
        body,
        headers,
      });
      let response;
      try {
        response = await this.fetcher(signed.url, {
          method,
          headers: signed.headers,
          ...(method === "PUT" ? { body } : {}),
          redirect: "error",
          signal: globalThis.AbortSignal.timeout(60_000),
        });
      } catch {
        if (attempt === 2)
          throw new Error(`Object storage ${method} network request failed`);
      }
      if (response && ![429, 500, 502, 503, 504].includes(response.status))
        return response;
      await response?.body?.cancel();
      if (attempt === 2)
        throw new Error(
          `Object storage ${method} unavailable after three attempts`,
        );
      await setTimeout(250 * 4 ** attempt);
    }
  }
  async verify(bucket, key, bytes, type, cacheControl) {
    const response = await this.request("GET", bucket, key);
    if (response.status !== 200) {
      await response.body?.cancel();
      throw new Error(`Object verification failed (HTTP ${response.status})`);
    }
    if (
      response.headers.get("content-type") !== type ||
      response.headers.get("cache-control") !== cacheControl
    ) {
      await response.body?.cancel();
      throw new Error("Existing object has unexpected delivery headers");
    }
    const received = await responseBytes(response, bytes.length);
    if (sha256(received) !== sha256(bytes))
      throw new Error(
        "Remote object checksum mismatch; immutable object was not overwritten",
      );
  }
  async ensure(bucket, key, bytes, type, cacheControl) {
    const checksum = sha256(bytes);
    const head = await this.request("HEAD", bucket, key);
    const status = head.status;
    await head.body?.cancel();
    if (status !== 200 && status !== 404)
      throw new Error(`Object storage HEAD failed (HTTP ${status})`);
    if (status === 404) {
      const put = await this.request("PUT", bucket, key, bytes, {
        "content-type": type,
        "cache-control": cacheControl,
        "content-md5": createHash("md5").update(bytes).digest("base64"),
        "if-none-match": "*",
        "x-amz-meta-sha256": checksum,
      });
      await put.body?.cancel();
      if (put.status !== 200 && put.status !== 412)
        throw new Error(`Object storage PUT failed (HTTP ${put.status})`);
    }
    // Also re-read resumed/concurrently created objects; a local journal is never proof.
    await this.verify(bucket, key, bytes, type, cacheControl);
    return status === 404 ? "uploaded-or-concurrent" : "verified-existing";
  }
  async verifyCDN(key, bytes, publicBaseURL = CDN) {
    const origin = canonicalPublicOrigin(publicBaseURL);
    let response;
    try {
      response = await this.fetcher(`${origin}/${key}`, {
        redirect: "error",
        signal: globalThis.AbortSignal.timeout(60_000),
      });
    } catch {
      throw new Error("Image CDN verification request failed");
    }
    if (
      response.status !== 200 ||
      response.headers.get("content-type") !== "image/webp" ||
      !response.headers.get("cache-control")?.includes("immutable")
    ) {
      await response.body?.cancel();
      throw new Error(
        `Image CDN delivery verification failed (HTTP ${response.status})`,
      );
    }
    if (sha256(await responseBytes(response, bytes.length)) !== sha256(bytes))
      throw new Error("Image CDN checksum mismatch");
  }
}
