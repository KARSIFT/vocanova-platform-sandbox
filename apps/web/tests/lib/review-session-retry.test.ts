import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  matchesPendingReviewSubmission,
  type PendingReviewSubmission,
} from "../../src/app/(app)/reviews/_components/review-session-retry";

const pending: PendingReviewSubmission = {
  idempotencyKey: "attempt-1",
  body: {
    userWordId: "word-1",
    meaningId: "meaning-1",
    attemptType: "review",
    promptType: "self_check",
    result: "correct",
    rating: "good",
    answeredAt: "2026-09-05T10:00:00.000Z",
    responseTimeMs: 1234,
    wasHintUsed: false,
    source: "review_session",
    clientAttemptId: "attempt-1",
  },
};

describe("review-session retry identity", () => {
  it("allows an ambiguous retry only for the exact original answer", () => {
    assert.equal(
      matchesPendingReviewSubmission(pending, {
        userWordId: "word-1",
        meaningId: "meaning-1",
        promptType: "self_check",
        result: "correct",
        rating: "good",
      }),
      true,
    );
  });

  it("rejects a changed answer instead of reusing its idempotency key", () => {
    assert.equal(
      matchesPendingReviewSubmission(pending, {
        userWordId: "word-1",
        meaningId: "meaning-1",
        promptType: "self_check",
        result: "correct",
        rating: "easy",
      }),
      false,
    );
  });
});
