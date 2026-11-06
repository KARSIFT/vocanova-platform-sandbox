import assert from "node:assert/strict";
import test from "node:test";

import { distinctFeedbackHeadline } from "../../src/app/(app)/_components/sentence-feedback-headline";

test("feedback shows its outcome once despite headline capitalization or whitespace", () => {
  for (const headline of [
    "Needs improvement",
    "Needs Improvement",
    " needs  improvement ",
    "NEEDS\nIMPROVEMENT",
  ]) {
    assert.equal(distinctFeedbackHeadline(headline, "Needs improvement"), null);
  }
  assert.equal(distinctFeedbackHeadline("correct", "Correct"), null);
  assert.equal(distinctFeedbackHeadline("INCORRECT", "Incorrect"), null);
});

test("feedback preserves a distinct teaching headline", () => {
  assert.equal(
    distinctFeedbackHeadline(
      "Use the past tense with yesterday",
      "Needs improvement",
    ),
    "Use the past tense with yesterday",
  );
  assert.equal(
    distinctFeedbackHeadline("Correct word, different tense", "Correct"),
    "Correct word, different tense",
  );
});

test("feedback omits missing or empty headlines", () => {
  for (const headline of [undefined, "", " \n "]) {
    assert.equal(distinctFeedbackHeadline(headline, "Correct"), null);
  }
});
