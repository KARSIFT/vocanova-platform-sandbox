import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { describe, it } from "node:test";

/**
 * The Reviews and Discover screens have no DOM-rendering test harness in
 * this project (no jsdom/RTL, and Node's built-in `--test` type-stripping
 * cannot transform JSX — see `review-session-prompt-readiness.test.ts` for
 * the established pattern of testing pure `.ts` logic instead).
 *
 * These checks cover the error-state render path (VOC-1179 acceptance:
 * "an error state (retry affordance, not a raw error dump)") by asserting
 * the segment `error.tsx` boundaries exist, expose a retry affordance via
 * `reset()`, and never surface the raw `error` object to the learner.
 */

const webRoot = fileURLToPath(new URL("../..", import.meta.url));

async function readSource(relativePath: string): Promise<string> {
  return readFile(`${webRoot}/${relativePath}`, "utf8");
}

describe("reviews and discover error-state render paths (VOC-1179)", () => {
  it("gives the Reviews screen a segment error boundary with a retry affordance", async () => {
    const source = await readSource("src/app/(app)/reviews/error.tsx");
    assert.match(source, /"use client"/);
    assert.match(source, /reset:\s*\(\)\s*=>\s*void/);
    assert.match(source, /onClick=\{reset\}/);
    assert.match(source, /Try again/);
    // The raw error/digest must never be interpolated into the UI copy.
    assert.doesNotMatch(source, /\{error\.message\}/);
    assert.doesNotMatch(source, /\{error\.digest\}/);
  });

  it("gives the Discover screen a segment error boundary with a retry affordance", async () => {
    const source = await readSource("src/app/(app)/discover/error.tsx");
    assert.match(source, /"use client"/);
    assert.match(source, /reset:\s*\(\)\s*=>\s*void/);
    assert.match(source, /onClick=\{reset\}/);
    assert.match(source, /Try again/);
    assert.doesNotMatch(source, /\{error\.message\}/);
    assert.doesNotMatch(source, /\{error\.digest\}/);
  });
});
