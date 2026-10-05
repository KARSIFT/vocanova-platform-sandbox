import { readFileSync, writeFileSync } from "node:fs";
import { createHash } from "node:crypto";
import { resolve } from "node:path";
import process from "node:process";
import { log } from "node:console";
import { fileURLToPath, URL } from "node:url";

const source = new URL("../src/lib/meaning-pictures.json", import.meta.url);
const destination = new URL(
  "../src/lib/meaning-pictures.runtime.json",
  import.meta.url,
);
/**
 * Keep editorial context out of client bundles; only reviewed media is visible.
 * @param {Record<string, {meaningId: string, src: string, status: string, alt: string}>} manifest
 * @param {(asset: URL) => Uint8Array} [readPicture]
 */
export function createRuntimeCatalogue(manifest, readPicture = readFileSync) {
  /** @type {Record<string, {alt: string, version: string}>} */
  const runtime = {};
  for (const [meaningId, picture] of Object.entries(manifest).sort(([a], [b]) =>
    a.localeCompare(b),
  )) {
    if (
      !/^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/.test(meaningId) ||
      picture.meaningId !== meaningId ||
      picture.src !== `/images/meanings/${meaningId}.webp`
    ) {
      throw new Error(`Invalid meaning picture identity: ${meaningId}`);
    }
    if (picture.status !== "ready" && picture.status !== "pending") {
      throw new Error(`Invalid picture status: ${meaningId}`);
    }
    if (picture.status === "ready") {
      if (typeof picture.alt !== "string" || !picture.alt.trim()) {
        throw new Error(`Missing ready picture alt text: ${meaningId}`);
      }
      // UUIDs identify meanings, while file bytes identify an artwork revision.
      // Revised artwork must bypass an older Next/image or browser cache entry.
      const version = createHash("sha256")
        .update(
          readPicture(new URL(`../public${picture.src}`, import.meta.url)),
        )
        .digest("hex")
        .slice(0, 16);
      runtime[meaningId] = { alt: picture.alt, version };
    }
  }
  return runtime;
}

if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  const runtime = createRuntimeCatalogue(
    JSON.parse(readFileSync(source, "utf8")),
  );
  const output = JSON.stringify(runtime, null, 2) + "\n";
  if (process.argv.includes("--check")) {
    if (readFileSync(destination, "utf8") !== output) {
      throw new Error(
        "Picture runtime catalogue is stale. Run pnpm --filter @vocanova/web generate:meaning-pictures.",
      );
    }
  } else {
    writeFileSync(destination, output);
  }
  log(
    `${Object.keys(runtime).length} ready meaning pictures: runtime catalogue ${process.argv.includes("--check") ? "verified" : "generated"}.`,
  );
}
