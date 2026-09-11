import assert from "node:assert/strict";
import { after, afterEach, describe, it } from "node:test";

import {
  clearSentenceFeedbackDraft,
  clearSentenceFeedbackDrafts,
  hasSentenceFeedbackDrafts,
  readSentenceFeedbackDraft,
  readSentenceFeedbackDraftIntent,
  readReviewCompletionContext,
  saveSentenceFeedbackDraft,
  saveReviewCompletionContext,
} from "../../src/app/(app)/_components/sentence-feedback-drafts";

class MemoryStorage {
  private readonly values = new Map<string, string>();

  get length(): number {
    return this.values.size;
  }

  clear(): void {
    this.values.clear();
  }

  getItem(key: string): string | null {
    return this.values.get(key) ?? null;
  }

  key(index: number): string | null {
    return [...this.values.keys()][index] ?? null;
  }

  removeItem(key: string): void {
    this.values.delete(key);
  }

  setItem(key: string, value: string): void {
    this.values.set(key, value);
  }
}

const memoryStorage = new MemoryStorage();
const originalWindow = globalThis.window;

Object.defineProperty(globalThis, "window", {
  configurable: true,
  value: { sessionStorage: memoryStorage },
});

describe("sentence feedback drafts", () => {
  it("restores a draft only for the matching learner, source, and attempt", () => {
    saveSentenceFeedbackDraft({
      userId: "learner-a",
      source: "word_detail",
      attemptId: "saved-word-1",
      sentence: "I pour coffee every morning.",
    });

    assert.equal(
      readSentenceFeedbackDraft({
        userId: "learner-a",
        source: "word_detail",
        attemptId: "saved-word-1",
      }),
      "I pour coffee every morning.",
    );
    assert.equal(
      readSentenceFeedbackDraft({
        userId: "learner-b",
        source: "word_detail",
        attemptId: "saved-word-1",
      }),
      null,
    );
    assert.equal(
      readSentenceFeedbackDraft({
        userId: "learner-a",
        source: "review",
        attemptId: "saved-word-1",
      }),
      null,
    );
  });

  it("clears a submitted draft and clears all tab drafts on an explicit account action", () => {
    saveSentenceFeedbackDraft({
      userId: "learner-a",
      source: "word_detail",
      attemptId: "saved-word-1",
      sentence: "I pour coffee every morning.",
    });
    saveSentenceFeedbackDraft({
      userId: "learner-a",
      source: "review",
      attemptId: "review-1",
      sentence: "I pour water carefully.",
    });

    clearSentenceFeedbackDraft({
      userId: "learner-a",
      source: "word_detail",
      attemptId: "saved-word-1",
    });
    assert.equal(hasSentenceFeedbackDrafts(), true);

    clearSentenceFeedbackDrafts();
    assert.equal(hasSentenceFeedbackDrafts(), false);
  });

  it("does not store an over-limit sentence", () => {
    saveSentenceFeedbackDraft({
      userId: "learner-a",
      source: "word_detail",
      attemptId: "saved-word-1",
      sentence: "a".repeat(301),
    });

    assert.equal(hasSentenceFeedbackDrafts(), false);
  });

  it("keeps an unresolved idempotency intent and review context user-scoped", () => {
    saveSentenceFeedbackDraft({
      userId: "learner-a",
      source: "review",
      attemptId: "review-attempt-1",
      sentence: "I pour water carefully.",
      idempotencyKey: "sentence-request-1",
    });
    saveReviewCompletionContext({
      userId: "learner-a",
      attemptId: "review-attempt-1",
      targetWord: "pour",
      shortDefinition: "make liquid flow",
    });

    assert.deepEqual(
      readSentenceFeedbackDraftIntent({
        userId: "learner-a",
        source: "review",
        attemptId: "review-attempt-1",
      }),
      {
        sentence: "I pour water carefully.",
        idempotencyKey: "sentence-request-1",
      },
    );
    assert.equal(readReviewCompletionContext("learner-b"), null);
    const context = readReviewCompletionContext("learner-a");
    assert.ok(context);
    assert.equal(context.attemptId, "review-attempt-1");
    assert.equal(context.targetWord, "pour");
    assert.equal(context.shortDefinition, "make liquid flow");
    assert.equal(typeof context.savedAt, "number");
  });
});

afterEach(() => {
  memoryStorage.clear();
});

after(() => {
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: originalWindow,
  });
});
