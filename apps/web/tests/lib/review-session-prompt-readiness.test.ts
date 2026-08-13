import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  isMultipleChoiceOptionDisabled,
  isReviewActionDisabled,
} from "../../src/app/(app)/reviews/_components/review-session-prompt";

describe("review-session prompt readiness (VOC-076-T01)", () => {
  it("enables multiple-choice options in prompt phase when not submitting", () => {
    assert.equal(isMultipleChoiceOptionDisabled("prompt", false), false);
  });

  it("disables multiple-choice options during feedback after a selection", () => {
    assert.equal(isMultipleChoiceOptionDisabled("feedback", false), true);
  });

  it("disables multiple-choice options while a submission is in flight", () => {
    assert.equal(isMultipleChoiceOptionDisabled("prompt", true), true);
    assert.equal(isMultipleChoiceOptionDisabled("feedback", true), true);
  });

  it("does not treat background refetch as a submit lock for learner actions", () => {
    // isRefetching is intentionally excluded from isReviewActionDisabled so a
    // slow listDueWords refetch cannot block the next card's interactions.
    assert.equal(isReviewActionDisabled(false), false);
    assert.equal(isReviewActionDisabled(true), true);
  });
});
