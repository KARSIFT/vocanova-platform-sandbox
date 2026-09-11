import { spawnSync } from "node:child_process";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";

const staging = readFileSync(".github/workflows/deploy-staging.yml", "utf8");
const production = readFileSync(
  ".github/workflows/deploy-production.yml",
  "utf8",
);

test("environment-specific images cannot overwrite another environment's commit tag", () => {
  for (const image of ["api", "web"]) {
    const pattern = new RegExp(
      `ghcr\\.io/karsift/vocanova-${image}:([^\\n]+short_sha[^\\n]+)`,
    );
    const stagingTag = staging.match(pattern)?.[1];
    const productionTag = production.match(pattern)?.[1];
    assert.ok(stagingTag, "staging publishes a commit-addressed image");
    assert.ok(productionTag, "production publishes a commit-addressed image");
    assert.notEqual(stagingTag, productionTag);
    const publishedPrefix = productionTag.split("${{")[0];
    assert.ok(
      production.includes(
        `PRODUCTION_IMAGE_TAG="${publishedPrefix}\${PRODUCTION_IMAGE_SHA_TAG}"`,
      ),
      "production must pull the same tag prefix it publishes",
    );
  }
});

test("release verification survives bot filtering and rejects mixed builds", () => {
  const result = spawnSync(
    "python3",
    [
      "-m",
      "unittest",
      "discover",
      "-s",
      "infra/scripts",
      "-p",
      "test_verify_deployed_release.py",
    ],
    { encoding: "utf8" },
  );
  assert.equal(result.status, 0, result.stderr || result.error?.message);
});
