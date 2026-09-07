import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  getCompletedReviewCountAfterSubmission,
  getReviewCompletionSummary,
} from "../../src/app/(app)/reviews/_components/review-completion-summary";

describe("review completion summary", () => {
  it("does not make an initial empty queue look like a completed session", () => {
    assert.equal(getReviewCompletionSummary(0), null);
  });

  it("reports the exact number of server-confirmed reviews", () => {
    let completedReviewCount = 0;
    completedReviewCount = getCompletedReviewCountAfterSubmission(
      completedReviewCount,
      true,
    );
    completedReviewCount = getCompletedReviewCountAfterSubmission(
      completedReviewCount,
      true,
    );

    assert.equal(
      getReviewCompletionSummary(completedReviewCount),
      "You reviewed 2 words.",
    );
  });

  it("keeps every confirmed review when the session spans a batch refill", () => {
    let completedReviewCount = 0;
    for (let review = 0; review < 51; review += 1) {
      completedReviewCount = getCompletedReviewCountAfterSubmission(
        completedReviewCount,
        true,
      );
    }

    assert.equal(
      getReviewCompletionSummary(completedReviewCount),
      "You reviewed 51 words.",
    );
  });

  it("uses singular wording for one confirmed review", () => {
    assert.equal(getReviewCompletionSummary(1), "You reviewed 1 word.");
  });

  it("does not count failed, rejected, or ambiguous submissions", () => {
    let completedReviewCount = 1;
    for (const outcome of ["failed", "rejected", "ambiguous"]) {
      completedReviewCount = getCompletedReviewCountAfterSubmission(
        completedReviewCount,
        false,
      );
      assert.equal(completedReviewCount, 1, outcome);
    }
  });

  it("counts a successful retry once after an unconfirmed submission", () => {
    let completedReviewCount = getCompletedReviewCountAfterSubmission(0, false);
    completedReviewCount = getCompletedReviewCountAfterSubmission(
      completedReviewCount,
      true,
    );

    assert.equal(
      getReviewCompletionSummary(completedReviewCount),
      "You reviewed 1 word.",
    );
  });
});
