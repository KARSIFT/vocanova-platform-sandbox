// Ensures squash-discarded VOC-112 capture commits exist as local Git objects in
// full checkouts before foundation tests run in default local provenance mode.

import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const fixturesDirectory = path.join(
  repositoryRoot,
  "scripts/foundation/fixtures",
);
const captureFixtures = [
  "voc112-navigation-benchmark-traces.json",
  "voc112-skill-discovery-evidence.json",
];

function readCaptureRevisions() {
  const revisions = new Set();
  for (const fixtureName of captureFixtures) {
    const fixture = JSON.parse(
      readFileSync(path.join(fixturesDirectory, fixtureName), "utf8"),
    );
    if (typeof fixture.subject_revision === "string") {
      revisions.add(fixture.subject_revision);
    }
    for (const row of fixture.discoveries ?? []) {
      if (typeof row.subject_revision === "string") {
        revisions.add(row.subject_revision);
      }
    }
  }
  return [...revisions];
}

export function hydrateVoc112GitObjects() {
  const provenanceMode = process.env.VOC112_CAPTURE_PROVENANCE_MODE ?? "local";
  if (provenanceMode !== "local") {
    return;
  }
  if (
    spawnSync("git", ["rev-parse", "--git-dir"], { cwd: repositoryRoot })
      .status !== 0
  ) {
    return;
  }
  const isShallow =
    execFileSync("git", ["rev-parse", "--is-shallow-repository"], {
      cwd: repositoryRoot,
      encoding: "utf8",
    }).trim() === "true";
  if (isShallow) {
    return;
  }
  for (const revision of readCaptureRevisions()) {
    if (!/^[a-f0-9]{40}$/.test(revision)) {
      continue;
    }
    const exists =
      spawnSync("git", ["cat-file", "-e", `${revision}^{commit}`], {
        cwd: repositoryRoot,
      }).status === 0;
    if (exists) {
      continue;
    }
    const fetch = spawnSync("git", ["fetch", "origin", revision], {
      cwd: repositoryRoot,
      encoding: "utf8",
    });
    if (fetch.status !== 0) {
      throw new Error(
        `VOC-112 foundation tests require capture commit object ${revision}`,
      );
    }
  }
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  hydrateVoc112GitObjects();
}
