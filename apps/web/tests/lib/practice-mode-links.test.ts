import assert from "node:assert/strict";
import { test } from "node:test";

import {
  practiceModeHref,
  readPracticeMode,
} from "../../src/app/(app)/practice/_components/practice-modes";

test("only explicit supported modes select a Practice activity", () => {
  for (const mode of [
    "journey",
    "mistakes",
    "typed_recall",
    "listening_choice",
  ]) {
    assert.equal(readPracticeMode(mode), mode);
  }
  for (const value of [
    undefined,
    "constructor",
    "__proto__",
    "unknown",
    ["mistakes"],
  ]) {
    assert.equal(readPracticeMode(value), undefined);
  }
});

test("changing exercise format preserves the vocabulary source without starting a session", () => {
  assert.equal(
    practiceModeHref("typed_recall", { listId: "a/b" }),
    "/practice?mode=typed_recall&list=a%2Fb#practice-mode-typed_recall",
  );
  assert.equal(
    practiceModeHref("listening_choice", { lessonKey: "work-meeting" }),
    "/practice?mode=listening_choice&lesson=work-meeting#practice-mode-listening_choice",
  );
  const source = { listId: "my-list", lessonKey: "old-lesson" };
  assert.equal(
    practiceModeHref("mistakes", source),
    "/practice?mode=mistakes#practice-mode-mistakes",
  );
  assert.equal(
    practiceModeHref("journey", source),
    "/practice?mode=journey#practice-journey",
  );
});
