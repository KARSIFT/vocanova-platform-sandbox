import assert from "node:assert/strict";
import test from "node:test";

import {
  formatSavedWordStatus,
  getSavedWordsView,
} from "../../src/app/(app)/words/_components/saved-word-view";

test("selects saved vocabulary list, first-page empty, and exhausted states", () => {
  assert.equal(getSavedWordsView(2, false), "list");
  assert.equal(getSavedWordsView(1, true), "list");
  assert.equal(getSavedWordsView(0, false), "empty");
  assert.equal(getSavedWordsView(0, true), "exhausted");
});

test("formats backend-owned saved-word lifecycle states", () => {
  assert.equal(formatSavedWordStatus("new"), "New");
  assert.equal(formatSavedWordStatus("learning"), "Learning");
  assert.equal(formatSavedWordStatus("reviewing"), "Reviewing");
  assert.equal(formatSavedWordStatus("mastered"), "Mastered");
  assert.equal(formatSavedWordStatus("ignored"), "Ignored");
  assert.equal(formatSavedWordStatus("archived"), "Archived");
  assert.equal(formatSavedWordStatus("unexpected"), null);
});

test("prefers the server review stage over a legacy new saved status", () => {
  assert.equal(formatSavedWordStatus("new", "learning"), "Learning");
  assert.equal(formatSavedWordStatus("new", "reviewing"), "Reviewing");
  assert.equal(formatSavedWordStatus("new", undefined), "New");
});
