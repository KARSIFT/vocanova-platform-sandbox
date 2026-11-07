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
 * refresh plus `reset()`, and never surface the raw error to the learner.
 * Actual failed-render recovery is covered by frontend-error-recovery.spec.ts.
 */

const webRoot = fileURLToPath(new URL("../..", import.meta.url));

async function readSource(relativePath: string): Promise<string> {
  return readFile(`${webRoot}/${relativePath}`, "utf8");
}

describe("reviews and discover error-state render paths (VOC-1179)", () => {
  it("refreshes failed server data together with resetting the client boundary", async () => {
    const source = await readSource("src/ui/use-route-error-retry.ts");
    assert.match(source, /startTransition\(\(\) => \{\s*router\.refresh\(\);\s*reset\(\);/);
  });

  it("gives the Reviews screen a segment error boundary with a retry affordance", async () => {
    const source = await readSource("src/app/(app)/reviews/error.tsx");
    assert.match(source, /"use client"/);
    assert.match(source, /reset:\s*\(\)\s*=>\s*void/);
    assert.match(source, /useRouteErrorRetry\(reset, error\)/);
    assert.match(source, /onClick=\{retry\}/);
    assert.match(source, /disabled=\{isRetrying\}/);
    assert.match(source, /aria-busy=\{isRetrying\}/);
    assert.match(source, /Try again/);
    // The raw error/digest must never be interpolated into the UI copy.
    assert.doesNotMatch(source, /\{error\.message\}/);
    assert.doesNotMatch(source, /\{error\.digest\}/);
  });

  it("gives the Discover screen a segment error boundary with a retry affordance", async () => {
    const source = await readSource("src/app/(app)/discover/error.tsx");
    assert.match(source, /"use client"/);
    assert.match(source, /reset:\s*\(\)\s*=>\s*void/);
    assert.match(source, /useRouteErrorRetry\(reset, error\)/);
    assert.match(source, /onClick=\{retry\}/);
    assert.match(source, /disabled=\{isRetrying\}/);
    assert.match(source, /aria-busy=\{isRetrying\}/);
    assert.match(source, /Try again/);
    assert.doesNotMatch(source, /\{error\.message\}/);
    assert.doesNotMatch(source, /\{error\.digest\}/);
  });
});
