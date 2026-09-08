import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { reviewSeedActions } from "../staging-e2e/review-seed";

describe("persistent staging review seed", () => {
  it("saves an initially unsaved word once", () => {
    assert.deepEqual(reviewSeedActions(false), ["save"]);
  });

  it("removes and restores an already-saved word", () => {
    assert.deepEqual(reviewSeedActions(true), ["remove", "save"]);
  });
});
