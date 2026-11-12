import type {
  EnglishLevel,
  LearningGoal,
  MainUseCase,
  UpdateSettingsBody,
} from "@vocanova/api-client";

import { isDisplayNameWithinLimit } from "./display-name";

const PREFIX = "vocanova:preferences-draft:";
const TTL_MS = 2 * 60 * 60 * 1000;
const MAX_BYTES = 4096;
const TARGETS = [5, 10, 15, 20, 30, 50, 75, 100];
const departedDocuments = new WeakSet<object>();

/** Successful explicit account exits end this document's recovery lifetime. */
export function isPreferencesDraftDocumentActive(): boolean {
  return typeof window !== "undefined" && !departedDocuments.has(window);
}

export interface OnboardingDraftValues {
  englishLevel: EnglishLevel | null;
  nativeLanguage: string;
  learningGoal: LearningGoal | null;
  mainUseCase: MainUseCase | null;
  dailyReviewTarget: number;
}

interface OnboardingDraft {
  values: OnboardingDraftValues;
  step: number;
}

type Kind = "onboarding" | "settings" | "profile" | "email-change";

function storageFor(userId?: string): Storage | null {
  if (!userId || !isPreferencesDraftDocumentActive()) return null;
  try {
    const storage = window.sessionStorage;
    // A tab changing accounts must not retain the previous learner's drafts.
    const ownPrefix = `${PREFIX}${encodeURIComponent(userId)}:`;
    for (let index = storage.length - 1; index >= 0; index--) {
      const key = storage.key(index);
      if (key?.startsWith(PREFIX) && !key.startsWith(ownPrefix)) {
        storage.removeItem(key);
      }
    }
    return storage;
  } catch {
    return null;
  }
}

function keyFor(userId: string, kind: Kind): string {
  return `${PREFIX}${encodeURIComponent(userId)}:${kind}`;
}

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function boundedText(value: unknown): value is string {
  return typeof value === "string" && isDisplayNameWithinLimit(value);
}

function onboardingValues(value: unknown): OnboardingDraftValues | null {
  if (!record(value)) return null;
  if (
    ![null, "a1", "a2", "b1", "b2", "unknown"].includes(
      value.englishLevel as string | null,
    ) ||
    !boundedText(value.nativeLanguage) ||
    ![
      null,
      "general",
      "work",
      "travel",
      "study",
      "conversation",
      "exam",
    ].includes(value.learningGoal as string | null) ||
    ![null, "daily_life", "work", "travel", "study", "social"].includes(
      value.mainUseCase as string | null,
    ) ||
    !TARGETS.includes(value.dailyReviewTarget as number)
  )
    return null;
  // Explicit allowlist: no arbitrary fields, identity or auth material.
  return {
    englishLevel: value.englishLevel as EnglishLevel | null,
    nativeLanguage: value.nativeLanguage,
    learningGoal: value.learningGoal as LearningGoal | null,
    mainUseCase: value.mainUseCase as MainUseCase | null,
    dailyReviewTarget: value.dailyReviewTarget as number,
  };
}

function settingsChanges(value: unknown): UpdateSettingsBody | null {
  if (!record(value)) return null;
  const changes: UpdateSettingsBody = {};
  if ("dailyReviewTarget" in value) {
    if (!TARGETS.includes(value.dailyReviewTarget as number)) return null;
    changes.dailyReviewTarget = value.dailyReviewTarget as number;
  }
  if ("reviewIntervalPreset" in value) {
    if (
      !["vocanova_default", "wordup_like", "custom"].includes(
        value.reviewIntervalPreset as string,
      )
    )
      return null;
    changes.reviewIntervalPreset =
      value.reviewIntervalPreset as UpdateSettingsBody["reviewIntervalPreset"];
  }
  if ("appLanguage" in value) {
    if (value.appLanguage !== "en") return null;
    changes.appLanguage = "en";
  }
  for (const field of [
    "notificationsEnabled",
    "marketingEmailsEnabled",
  ] as const) {
    if (field in value) {
      if (typeof value[field] !== "boolean") return null;
      changes[field] = value[field];
    }
  }
  if ("displayName" in value) {
    if (!boundedText(value.displayName)) return null;
    changes.displayName = value.displayName;
  }
  return Object.keys(changes).length > 0 ? changes : null;
}

function read(
  userId: string | undefined,
  kind: Kind,
): Record<string, unknown> | null {
  const storage = storageFor(userId);
  if (!storage || !userId) return null;
  const key = keyFor(userId, kind);
  try {
    const raw = storage.getItem(key);
    if (!raw) return null;
    const envelope: unknown = raw.length <= MAX_BYTES ? JSON.parse(raw) : null;
    if (
      record(envelope) &&
      envelope.version === 1 &&
      typeof envelope.savedAt === "number" &&
      Number.isFinite(envelope.savedAt) &&
      envelope.savedAt <= Date.now() &&
      Date.now() - envelope.savedAt <= TTL_MS
    )
      return envelope;
    storage.removeItem(key);
  } catch {
    clear(userId, kind);
  }
  return null;
}

function save(
  userId: string | undefined,
  kind: Kind,
  payload: object | null,
): void {
  const storage = storageFor(userId);
  if (!storage || !userId) return;
  try {
    if (!payload) {
      storage.removeItem(keyFor(userId, kind));
      return;
    }
    storage.setItem(
      keyFor(userId, kind),
      JSON.stringify({ version: 1, savedAt: Date.now(), ...payload }),
    );
  } catch {
    // Recovery is optional: blocked or full tab storage must not block forms.
  }
}

function clear(userId: string | undefined, kind: Kind): void {
  const storage = storageFor(userId);
  if (!storage || !userId) return;
  try {
    storage.removeItem(keyFor(userId, kind));
  } catch {
    /* Optional recovery. */
  }
}

export function readOnboardingDraft(userId?: string): OnboardingDraft | null {
  const envelope = read(userId, "onboarding");
  if (!envelope) return null;
  const values = onboardingValues(envelope.values);
  if (
    !values ||
    !Number.isInteger(envelope.step) ||
    (envelope.step as number) < 0 ||
    (envelope.step as number) > 4
  ) {
    clearOnboardingDraft(userId);
    return null;
  }
  return { values, step: envelope.step as number };
}

export function saveOnboardingDraft({
  userId,
  values,
  step,
}: {
  userId?: string;
  values: OnboardingDraftValues;
  step: number;
}): void {
  const safeValues = onboardingValues(values);
  save(
    userId,
    "onboarding",
    safeValues && Number.isInteger(step) && step >= 0 && step <= 4
      ? { values: safeValues, step }
      : null,
  );
}

export function readSettingsDraft(userId?: string): UpdateSettingsBody | null {
  const envelope = read(userId, "settings");
  if (!envelope) return null;
  const changes = settingsChanges(envelope.changes);
  if (!changes) clearSettingsDraft(userId);
  return changes;
}

export function saveSettingsDraft({
  userId,
  changes,
}: {
  userId?: string;
  changes: UpdateSettingsBody;
}): void {
  const safeChanges = settingsChanges(changes);
  save(userId, "settings", safeChanges ? { changes: safeChanges } : null);
}

export function clearOnboardingDraft(userId?: string): void {
  clear(userId, "onboarding");
}
export function clearSettingsDraft(userId?: string): void {
  clear(userId, "settings");
}

export function readProfileDraft(userId?: string): string | null {
  const envelope = read(userId, "profile");
  if (!envelope) return null;
  if (boundedText(envelope.displayName)) return envelope.displayName;
  clearProfileDraft(userId);
  return null;
}

export function saveProfileDraft({
  userId,
  displayName,
}: {
  userId?: string;
  displayName: string;
}): void {
  save(userId, "profile", boundedText(displayName) ? { displayName } : null);
}

export function clearProfileDraft(userId?: string): void {
  clear(userId, "profile");
}

function emailDraftText(value: unknown): value is string {
  // An unfinished address may not be valid yet. Bound recovery text without
  // claiming it is ready to submit; native and server validation still apply.
  return typeof value === "string" && value.length <= 320;
}

export function readEmailChangeDraft(
  userId: string | undefined,
  currentEmail: string,
): string | null {
  const envelope = read(userId, "email-change");
  if (!envelope) return null;
  if (
    envelope.currentEmail === currentEmail &&
    emailDraftText(envelope.newEmail) &&
    envelope.newEmail.trim()
  )
    return envelope.newEmail;
  clearEmailChangeDraft(userId);
  return null;
}

export function saveEmailChangeDraft({
  userId,
  currentEmail,
  newEmail,
}: {
  userId?: string;
  currentEmail: string;
  newEmail: string;
}): void {
  save(
    userId,
    "email-change",
    emailDraftText(currentEmail) && emailDraftText(newEmail) && newEmail.trim()
      ? { currentEmail, newEmail }
      : null,
  );
}

export function clearEmailChangeDraft(userId?: string): void {
  clear(userId, "email-change");
}

/** Called only after successful explicit sign-out or account deletion. */
export function clearPreferencesDrafts(): void {
  if (typeof window === "undefined") return;
  // Mark departure before accessing optional storage: pending callbacks must
  // stay inactive even if cleanup throws or replacement navigation fails.
  // Authentication resumes through full navigation in a fresh document.
  departedDocuments.add(window);
  try {
    const storage = window.sessionStorage;
    for (let index = storage.length - 1; index >= 0; index--) {
      const key = storage.key(index);
      if (key?.startsWith(PREFIX)) storage.removeItem(key);
    }
  } catch {
    /* Optional recovery. */
  }
}
