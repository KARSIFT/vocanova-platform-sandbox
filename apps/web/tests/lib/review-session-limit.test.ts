import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  getDueRequestLimit,
  getRemainingReviewTarget,
  hasReachedReviewSessionLimit,
} from "../../src/app/(app)/reviews/_components/review-session-limit";

describe("review session daily-target limit", () => {
  it("uses only the remaining mission target instead of the due backlog", () => {
    assert.equal(getRemainingReviewTarget(20, 0), 20);
    assert.equal(getRemainingReviewTarget(20, 7), 13);
  });

  it("does not start another card after today's target is complete", () => {
    assert.equal(getRemainingReviewTarget(20, 20), 0);
    assert.equal(getRemainingReviewTarget(20, 24), 0);
  });

  it("caps a queue request at the session allowance and API boundary", () => {
    assert.equal(getDueRequestLimit(13), 13);
    assert.equal(getDueRequestLimit(100), 50);
  });

  it("ends exactly on the confirmed review that reaches the target", () => {
    assert.equal(hasReachedReviewSessionLimit(4, 5), false);
    assert.equal(hasReachedReviewSessionLimit(5, 5), true);
  });
});
