// Ensures revision-bound VOC-112 evidence commits are present before local-mode
// provenance tests run under full-history checkouts (for example implement.yml).

import { execFileSync, spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

function collectSubjectRevisions() {
  const revisions = new Set();
  const fixturePaths = [
    "scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json",
    "scripts/foundation/fixtures/voc112-skill-discovery-evidence.json",
  ];
  for (const relativePath of fixturePaths) {
    const data = JSON.parse(
      readFileSync(path.join(repositoryRoot, relativePath), "utf8"),
    );
    if (typeof data.subject_revision === "string") {
      revisions.add(data.subject_revision);
    }
    for (const row of data.discoveries ?? []) {
      if (typeof row.subject_revision === "string") {
        revisions.add(row.subject_revision);
      }
    }
  }
  return [...revisions];
}

function commitExists(revision) {
  return (
    spawnSync("git", ["cat-file", "-e", `${revision}^{commit}`], {
      cwd: repositoryRoot,
    }).status === 0
  );
}

for (const revision of collectSubjectRevisions()) {
  if (commitExists(revision)) {
    continue;
  }
  execFileSync("git", ["fetch", "origin", revision], {
    cwd: repositoryRoot,
    stdio: "inherit",
  });
  if (!commitExists(revision)) {
    throw new Error(
      `VOC-112 capture commit ${revision} is unavailable after fetch`,
    );
  }
}
