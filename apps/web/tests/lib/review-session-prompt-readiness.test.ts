import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  isMultipleChoiceOptionDisabled,
  isReviewActionDisabled,
  shouldShowReviewCardPrompt,
} from "../../src/app/(app)/reviews/_components/review-session-prompt";

describe("review-session prompt readiness (VOC-076-T01)", () => {
  it("enables multiple-choice options in prompt phase when not submitting", () => {
    assert.equal(isMultipleChoiceOptionDisabled("prompt", false, false), false);
  });

  it("disables multiple-choice options during feedback after a selection", () => {
    assert.equal(isMultipleChoiceOptionDisabled("feedback", false, false), true);
  });

  it("disables multiple-choice options while a submission is in flight", () => {
    assert.equal(isMultipleChoiceOptionDisabled("prompt", true, false), true);
    assert.equal(isMultipleChoiceOptionDisabled("feedback", true, false), true);
  });

  it("does not treat background refetch as a lock for multiple-choice options", () => {
    // isRefetching is intentionally excluded from isMultipleChoiceOptionDisabled
    // so a prompt-ready next card is interactable as soon as dueWords lands,
    // even if the batch-end refetch flag has not cleared yet.
    assert.equal(isMultipleChoiceOptionDisabled("prompt", false, false), false);
  });

  it("locks rate/continue/show-answer during submit and during batch-end refetch", () => {
    // After submitAttempt succeeds, advance() may set isRefetching while the
    // same card remains visible; isSubmitting clears in finally. Post-submit
    // actions must stay disabled for the whole refetch to prevent duplicate
    // submitAttempt calls (VOC-076-T01 independent-review remediation).
    assert.equal(isReviewActionDisabled(false, false, false), false);
    assert.equal(isReviewActionDisabled(true, false, false), true);
    assert.equal(isReviewActionDisabled(false, true, false), true);
    assert.equal(isReviewActionDisabled(true, true, false), true);
    assert.equal(isReviewActionDisabled(false, false, true), true);
  });

  it("hides the prior card prompt body while batch-end refetch is in flight (VOC-076-T02)", () => {
    // Leaving the completed feedback MC fieldset mounted during listDueWords
    // is the run #227 failure mode (disabled options, aria-pressed=false).
    assert.equal(shouldShowReviewCardPrompt(false), true);
    assert.equal(shouldShowReviewCardPrompt(true), false);
  });
});
