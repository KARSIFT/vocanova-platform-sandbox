import assert from "node:assert/strict";
import test from "node:test";

import { formatWordReviewState } from "../../src/app/(app)/discover/[situation]/[word]/_components/word-review-state";

test("Word Detail displays the server-authoritative due state before a stored review state", () => {
  assert.equal(formatWordReviewState("learning", true), "Due today");
});

test("Word Detail maps only documented persisted review states to learner labels", () => {
  assert.equal(formatWordReviewState("new", false), "New");
  assert.equal(formatWordReviewState("learning", false), "Learning");
  assert.equal(formatWordReviewState("reviewing", false), "Learning");
  assert.equal(formatWordReviewState("mastered", false), "Mastered");
  assert.equal(formatWordReviewState("ignored", false), "Ignored");
  assert.equal(formatWordReviewState("archived", false), "Archived");
  assert.equal(formatWordReviewState("unknown", false), null);
  assert.equal(formatWordReviewState(undefined, false), null);
});
