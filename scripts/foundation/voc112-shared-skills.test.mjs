// VOC-112-T02 — shared engineering skills count, provenance, and repository adaptations.
//
// Runs via `node --test scripts/foundation/voc112-shared-skills.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import {
  FORBIDDEN_PATTERNS,
  validateAgentSkillsTree,
} from "./voc112-agent-skills.test.mjs";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const canonicalSkillsRoot = path.join(repositoryRoot, ".agents/skills");

/** Fixed VOC-112-D06 shared skill set (directory names). */
export const SHARED_SKILL_NAMES = [
  "context-mapping",
  "systematic-debugging",
  "verification-before-completion",
  "github-actions-efficiency",
  "react-next-performance",
  "playwright-browser-testing",
  "security-threat-modeling",
];

const VERCEL_REJECTED_UPSTREAM = "dd089a8c752c966dee8bf0f27cb625ba193ffd9e";

const REPOSITORY_ADAPTATION_MARKERS = [
  "docs/development.md",
  "pnpm validate",
  "Governance precedence",
  "repository sources win",
];

const SAFETY_MARKERS = [/(?:Never|Do not)/i, /\.env/, /raw CI log/i];

function readSkill(skillName) {
  return readFileSync(
    path.join(canonicalSkillsRoot, skillName, "SKILL.md"),
    "utf8",
  );
}

function readProvenance(skillName) {
  return readFileSync(
    path.join(canonicalSkillsRoot, skillName, "PROVENANCE.yaml"),
    "utf8",
  );
}

function listCanonicalSkillDirectories() {
  return readdirSync(canonicalSkillsRoot, { withFileTypes: true })
    .filter(
      (entry) =>
        entry.isDirectory() &&
        !entry.isSymbolicLink() &&
        !entry.name.startsWith("."),
    )
    .map((entry) => entry.name)
    .sort();
}

function parseProvenanceYaml(source) {
  const record = {};
  let currentList = null;
  let currentObject = null;

  for (const line of source.split(/\r?\n/)) {
    if (!line.trim() || line.trim().startsWith("#")) {
      continue;
    }
    if (line.startsWith("  - path:")) {
      currentObject = { path: line.split(":", 2)[1].trim() };
      currentList.push(currentObject);
      continue;
    }
    if (line.startsWith("    sha256:") && currentObject) {
      currentObject.sha256 = line.split(":", 2)[1].trim();
      continue;
    }
    if (line.startsWith("  - ")) {
      currentList.push(line.slice(4).trim());
      continue;
    }
    const separator = line.indexOf(":");
    if (separator < 0) {
      continue;
    }
    const key = line.slice(0, separator).trim();
    const value = line.slice(separator + 1).trim();
    if (!value) {
      currentList = [];
      record[key] = currentList;
      currentObject = null;
      continue;
    }
    record[key] = value;
    currentList = null;
    currentObject = null;
  }
  return record;
}

test("VOC-112-TEST-08: exactly seven shared skills with canonical and Claude adapters", () => {
  const canonicalSkills = listCanonicalSkillDirectories();
  const sharedPresent = SHARED_SKILL_NAMES.filter((name) =>
    canonicalSkills.includes(name),
  );

  assert.equal(
    sharedPresent.length,
    SHARED_SKILL_NAMES.length,
    `expected ${SHARED_SKILL_NAMES.length} shared skills, found: ${sharedPresent.join(", ")}`,
  );

  for (const skillName of SHARED_SKILL_NAMES) {
    const canonicalPath = path.join(canonicalSkillsRoot, skillName, "SKILL.md");
    const adapterPath = path.join(
      repositoryRoot,
      ".claude/skills",
      skillName,
      "SKILL.md",
    );
    assert.ok(
      statSync(canonicalPath).isFile(),
      `${skillName} canonical missing`,
    );
    assert.ok(statSync(adapterPath).isFile(), `${skillName} adapter missing`);
  }

  const treeErrors = validateAgentSkillsTree();
  assert.deepEqual(treeErrors, [], treeErrors.join("\n"));
});

test("VOC-112-TEST-09: shared skills include repository adaptations and provenance pins", () => {
  for (const skillName of SHARED_SKILL_NAMES) {
    const skillSource = readSkill(skillName);
    const provenance = parseProvenanceYaml(readProvenance(skillName));

    assert.equal(provenance.skill_name, skillName);
    assert.ok(
      provenance.source === "adapted" ||
        provenance.source === "repository-native",
      `${skillName}: provenance source must be adapted or repository-native`,
    );

    for (const marker of REPOSITORY_ADAPTATION_MARKERS) {
      assert.match(
        skillSource,
        new RegExp(marker.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), "i"),
        `${skillName}: missing repository adaptation marker ${marker}`,
      );
    }

    for (const marker of SAFETY_MARKERS) {
      assert.match(
        skillSource,
        marker,
        `${skillName}: missing safety marker ${marker}`,
      );
    }

    if (provenance.source === "adapted") {
      for (const field of [
        "upstream_repo",
        "upstream_commit",
        "upstream_path",
        "upstream_sha256",
        "local_sha256",
        "license",
      ]) {
        assert.ok(
          provenance[field],
          `${skillName}: missing adapted field ${field}`,
        );
      }
      assert.match(provenance.upstream_repo, /^https:\/\//);
      assert.match(provenance.upstream_commit, /^[0-9a-f]{7,40}$/);
      assert.match(provenance.upstream_sha256, /^[0-9a-f]{64}$/);
      assert.match(provenance.local_sha256, /^[0-9a-f]{64}$/);
      assert.ok(
        Array.isArray(provenance.retained_license_paths) &&
          provenance.retained_license_paths.length > 0,
        `${skillName}: retained_license_paths required`,
      );
    }

    if (skillName === "react-next-performance") {
      assert.equal(provenance.source, "repository-native");
      assert.match(
        skillSource,
        new RegExp(VERCEL_REJECTED_UPSTREAM),
        "react-next-performance must record Vercel source rejection",
      );
      const withoutRejectionSection = skillSource.replace(
        /## Rejected upstream source[\s\S]*?(?=\n## )/,
        "",
      );
      assert.doesNotMatch(
        withoutRejectionSection,
        /vercel-labs\/agent-skills/i,
        "must not vendor Vercel agent-skills prose outside the rejection note",
      );
    }

    if (skillName === "playwright-browser-testing") {
      assert.match(skillSource, /@playwright\/test/);
      assert.match(skillSource, /apps\/web/);
      assert.doesNotMatch(
        skillSource,
        /npm install -g|@latest|npx playwright install(?!.*frozen)/i,
        "must not instruct unpinned/global Playwright installs",
      );
    }

    if (skillName === "github-actions-efficiency") {
      assert.match(skillSource, /\.github\/workflows/);
      assert.doesNotMatch(
        skillSource,
        /--log-failed|paste raw CI logs/i,
        "must not instruct raw CI log ingestion",
      );
    }
  }

  const repositoryWideScan = SHARED_SKILL_NAMES.map((name) => {
    let source = readSkill(name);
    if (name === "react-next-performance") {
      source = source.replace(
        /## Rejected upstream source[\s\S]*?(?=\n## )/,
        "",
      );
    }
    return source;
  }).join("\n");
  assert.doesNotMatch(
    repositoryWideScan,
    /vercel-labs\/agent-skills/i,
    "shared skills must not include Vercel agent-skills copy",
  );

  for (const rule of FORBIDDEN_PATTERNS) {
    for (const skillName of SHARED_SKILL_NAMES) {
      const source = readSkill(skillName);
      const flags = rule.pattern.flags.includes("g")
        ? rule.pattern.flags
        : `${rule.pattern.flags}g`;
      const matcher = new RegExp(rule.pattern.source, flags);
      for (const match of source.matchAll(matcher)) {
        const prefix = source.slice(Math.max(0, match.index - 24), match.index);
        if (
          /(?:do not|don't|never|must not)\s+(?:(?:run|use|execute)\s+)?$/i.test(
            prefix,
          )
        ) {
          continue;
        }
        assert.fail(
          `${skillName}: matched forbidden pattern ${rule.id} (${rule.description})`,
        );
      }
    }
  }
});

test("VOC-112-TEST-09 supplemental: shared skills are distinct from navigator-only scope", () => {
  const allSkills = listCanonicalSkillDirectories();
  const nonShared = allSkills.filter(
    (name) => !SHARED_SKILL_NAMES.includes(name),
  );
  assert.ok(
    nonShared.includes("vocanova-repo-navigator"),
    "navigator remains outside the shared-skill count",
  );
  assert.equal(
    SHARED_SKILL_NAMES.length,
    7,
    "shared skill roster length is fixed by VOC-112-D06",
  );
});
