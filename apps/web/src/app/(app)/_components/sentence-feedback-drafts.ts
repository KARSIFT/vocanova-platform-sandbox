import {
  countSentenceCharacters,
  MAX_SENTENCE_CHARACTERS,
} from "./sentence-feedback-input";

const DRAFT_PREFIX = "vocanova:sentence-feedback-draft:";
const REVIEW_CONTEXT_PREFIX = "vocanova:sentence-feedback-review-context:";
const DRAFT_TTL_MS = 2 * 60 * 60 * 1000;

type SentenceFeedbackSource =
  "word_detail" | "review" | "daily_mission" | "free_practice";

interface SentenceFeedbackDraft {
  attemptId: string;
  savedAt: number;
  sentence: string;
  source: SentenceFeedbackSource;
  idempotencyKey?: string;
}

export interface SentenceFeedbackDraftIntent {
  sentence: string;
  idempotencyKey?: string;
}

export interface ReviewCompletionContext {
  attemptId: string;
  savedAt: number;
  shortDefinition?: string;
  targetWord: string;
}

function getSessionStorage(): Storage | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    return window.sessionStorage;
  } catch {
    // Browser privacy settings can deny storage entirely. Draft recovery is
    // optional, so it must never interfere with learning or account actions.
    return null;
  }
}

export function canUseSentenceFeedbackDraftStorage(userId?: string): boolean {
  return Boolean(userId && getSessionStorage());
}

function getDraftKey(
  userId: string,
  source: SentenceFeedbackSource,
  attemptId: string,
): string {
  // The key is scoped to the authenticated user and this browser tab. Draft
  // content never carries an identity, session cookie, or CSRF material.
  return `${DRAFT_PREFIX}${encodeURIComponent(userId)}:${source}:${encodeURIComponent(attemptId)}`;
}

function getReviewContextKey(userId: string): string {
  return `${REVIEW_CONTEXT_PREFIX}${encodeURIComponent(userId)}`;
}

export function readSentenceFeedbackDraft({
  userId,
  source,
  attemptId,
}: {
  userId?: string;
  source: SentenceFeedbackSource;
  attemptId: string;
}): string | null {
  return (
    readSentenceFeedbackDraftIntent({ userId, source, attemptId })?.sentence ??
    null
  );
}

export function readSentenceFeedbackDraftIntent({
  userId,
  source,
  attemptId,
}: {
  userId?: string;
  source: SentenceFeedbackSource;
  attemptId: string;
}): SentenceFeedbackDraftIntent | null {
  const storage = getSessionStorage();
  if (!userId || !storage) {
    return null;
  }

  const key = getDraftKey(userId, source, attemptId);
  let rawDraft: string | null;
  try {
    rawDraft = storage.getItem(key);
  } catch {
    return null;
  }
  if (!rawDraft) {
    return null;
  }

  try {
    const draft = JSON.parse(rawDraft) as SentenceFeedbackDraft;
    if (
      draft.attemptId !== attemptId ||
      draft.source !== source ||
      typeof draft.sentence !== "string" ||
      !Number.isFinite(draft.savedAt) ||
      draft.savedAt > Date.now() ||
      Date.now() - draft.savedAt > DRAFT_TTL_MS ||
      countSentenceCharacters(draft.sentence) > MAX_SENTENCE_CHARACTERS
    ) {
      try {
        storage.removeItem(key);
      } catch {
        // Storage cleanup is best effort.
      }
      return null;
    }
    return {
      sentence: draft.sentence,
      ...(typeof draft.idempotencyKey === "string"
        ? { idempotencyKey: draft.idempotencyKey }
        : {}),
    };
  } catch {
    try {
      storage.removeItem(key);
    } catch {
      // Storage cleanup is best effort.
    }
    return null;
  }
}

export function saveSentenceFeedbackDraft({
  userId,
  source,
  attemptId,
  sentence,
  idempotencyKey,
}: {
  userId?: string;
  source: SentenceFeedbackSource;
  attemptId: string;
  sentence: string;
  idempotencyKey?: string;
}): void {
  const storage = getSessionStorage();
  if (!userId || !storage) {
    return;
  }

  const key = getDraftKey(userId, source, attemptId);
  if (
    !sentence.trim() ||
    countSentenceCharacters(sentence) > MAX_SENTENCE_CHARACTERS
  ) {
    try {
      storage.removeItem(key);
    } catch {
      // Storage is optional.
    }
    return;
  }

  const draft: SentenceFeedbackDraft = {
    attemptId,
    savedAt: Date.now(),
    sentence,
    source,
    ...(idempotencyKey ? { idempotencyKey } : {}),
  };
  try {
    storage.setItem(key, JSON.stringify(draft));
  } catch {
    // Storage quotas and privacy settings must not interrupt typing.
  }
}

export function saveReviewCompletionContext({
  userId,
  attemptId,
  targetWord,
  shortDefinition,
}: {
  userId?: string;
  attemptId: string;
  targetWord: string;
  shortDefinition?: string;
}): void {
  const storage = getSessionStorage();
  if (!userId || !storage || !attemptId || !targetWord) {
    return;
  }
  const context: ReviewCompletionContext = {
    attemptId,
    savedAt: Date.now(),
    targetWord,
    ...(shortDefinition ? { shortDefinition } : {}),
  };
  try {
    storage.setItem(getReviewContextKey(userId), JSON.stringify(context));
  } catch {
    // Recovery context is optional.
  }
}

export function readReviewCompletionContext(
  userId?: string,
): ReviewCompletionContext | null {
  const storage = getSessionStorage();
  if (!userId || !storage) {
    return null;
  }
  const key = getReviewContextKey(userId);
  try {
    const rawContext = storage.getItem(key);
    if (!rawContext) {
      return null;
    }
    const context = JSON.parse(rawContext) as ReviewCompletionContext;
    if (
      typeof context.attemptId !== "string" ||
      typeof context.targetWord !== "string" ||
      !context.attemptId ||
      !context.targetWord ||
      !Number.isFinite(context.savedAt) ||
      context.savedAt > Date.now() ||
      Date.now() - context.savedAt > DRAFT_TTL_MS
    ) {
      storage.removeItem(key);
      return null;
    }
    return context;
  } catch {
    try {
      storage.removeItem(key);
    } catch {
      // Storage cleanup is best effort.
    }
    return null;
  }
}

export function clearSentenceFeedbackDraft({
  userId,
  source,
  attemptId,
}: {
  userId?: string;
  source: SentenceFeedbackSource;
  attemptId: string;
}): void {
  const storage = getSessionStorage();
  if (!userId || !storage) {
    return;
  }
  try {
    storage.removeItem(getDraftKey(userId, source, attemptId));
  } catch {
    // Storage is optional.
  }
}

/** Called after an explicit logout or account deletion. */
export function clearSentenceFeedbackDrafts(): void {
  const storage = getSessionStorage();
  if (!storage) {
    return;
  }
  try {
    for (let index = storage.length - 1; index >= 0; index -= 1) {
      const key = storage.key(index);
      if (
        key?.startsWith(DRAFT_PREFIX) ||
        key?.startsWith(REVIEW_CONTEXT_PREFIX)
      ) {
        storage.removeItem(key);
      }
    }
  } catch {
    // Explicit account actions must still complete if storage is unavailable.
  }
}

/** Used by the explicit logout confirmation; expired drafts do not count. */
export function hasSentenceFeedbackDrafts(): boolean {
  const storage = getSessionStorage();
  if (!storage) {
    return false;
  }
  try {
    // Removing a stale key shifts Storage indexes. Sweep backwards so an
    // adjacent current draft is still inspected before the logout prompt.
    for (let index = storage.length - 1; index >= 0; index -= 1) {
      const key = storage.key(index);
      if (!key?.startsWith(DRAFT_PREFIX)) {
        continue;
      }
      const rawDraft = storage.getItem(key);
      if (!rawDraft) {
        continue;
      }
      try {
        const draft = JSON.parse(rawDraft) as SentenceFeedbackDraft;
        const isCurrent =
          typeof draft.sentence === "string" &&
          typeof draft.attemptId === "string" &&
          Number.isFinite(draft.savedAt) &&
          draft.savedAt <= Date.now() &&
          Date.now() - draft.savedAt <= DRAFT_TTL_MS &&
          countSentenceCharacters(draft.sentence) <= MAX_SENTENCE_CHARACTERS;
        if (isCurrent) {
          return true;
        }
      } catch {
        // Fall through to remove an unreadable stale value.
      }
      storage.removeItem(key);
    }
  } catch {
    return false;
  }
  return false;
}
