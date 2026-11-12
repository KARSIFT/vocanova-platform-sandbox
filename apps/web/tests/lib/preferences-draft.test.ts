import assert from "node:assert/strict";
import { after, afterEach, it } from "node:test";

import {
  clearOnboardingDraft,
  clearPreferencesDrafts,
  clearSettingsDraft,
  readOnboardingDraft,
  readSettingsDraft,
  saveOnboardingDraft,
  saveSettingsDraft,
  readProfileDraft,
  saveProfileDraft,
  clearProfileDraft,
  readEmailChangeDraft,
  saveEmailChangeDraft,
  clearEmailChangeDraft,
  isPreferencesDraftDocumentActive,
} from "../../src/lib/preferences-draft";

class MemoryStorage {
  values = new Map<string, string>();
  get length() {
    return this.values.size;
  }
  key(index: number) {
    return [...this.values.keys()][index] ?? null;
  }
  getItem(key: string) {
    return this.values.get(key) ?? null;
  }
  setItem(key: string, value: string) {
    this.values.set(key, value);
  }
  removeItem(key: string) {
    this.values.delete(key);
  }
  clear() {
    this.values.clear();
  }
}

const storage = new MemoryStorage();
const originalWindow = globalThis.window;
function setStorage(value: unknown) {
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: { sessionStorage: value },
  });
}
setStorage(storage);

const answers = {
  englishLevel: "b2" as const,
  nativeLanguage: "Persian",
  learningGoal: "conversation" as const,
  mainUseCase: "social" as const,
  dailyReviewTarget: 15,
};
const settingsKey = "vocanova:preferences-draft:learner-a:settings";

it("a successful account exit permanently blocks late draft work in the departing document", () => {
  assert.equal(isPreferencesDraftDocumentActive(), true);
  saveProfileDraft({
    userId: "learner-a",
    displayName: "Private before logout",
  });
  clearPreferencesDrafts();
  assert.equal(isPreferencesDraftDocumentActive(), false);
  // These calls model pending save/error continuations running after cleanup
  // while the old document's replacement navigation has not committed yet.
  saveProfileDraft({
    userId: "learner-a",
    displayName: "Private late profile",
  });
  saveEmailChangeDraft({
    userId: "learner-a",
    currentEmail: "old@example.test",
    newEmail: "private-late@example.test",
  });
  saveSettingsDraft({
    userId: "learner-a",
    changes: { displayName: "Private late settings" },
  });
  saveOnboardingDraft({ userId: "learner-a", values: answers, step: 4 });
  assert.equal(storage.length, 0);
  assert.equal(readProfileDraft("learner-a"), null);
  assert.equal(readEmailChangeDraft("learner-a", "old@example.test"), null);
  assert.equal(readSettingsDraft("learner-a"), null);
  assert.equal(readOnboardingDraft("learner-a"), null);
  // A fresh authenticated document has its own lifetime and can edit again.
  setStorage(storage);
  assert.equal(isPreferencesDraftDocumentActive(), true);
  saveProfileDraft({ userId: "learner-a", displayName: "Fresh document" });
  assert.equal(readProfileDraft("learner-a"), "Fresh document");
});

it("marks departure before optional storage cleanup can fail", () => {
  let blocked = true;
  setStorage({
    get length() {
      if (blocked) throw new Error("storage unavailable during exit");
      return storage.length;
    },
    key: (index: number) => storage.key(index),
    getItem: (key: string) => storage.getItem(key),
    setItem: (key: string, value: string) => storage.setItem(key, value),
    removeItem: (key: string) => storage.removeItem(key),
  });
  assert.doesNotThrow(() => clearPreferencesDrafts());
  blocked = false;
  assert.equal(isPreferencesDraftDocumentActive(), false);
  saveProfileDraft({ userId: "learner-a", displayName: "Late private value" });
  assert.equal(storage.length, 0);
});

it("restores profile and unsent email drafts only in the same authenticated account", () => {
  saveProfileDraft({ userId: "learner-a", displayName: "𐐀".repeat(80) });
  saveEmailChangeDraft({
    userId: "learner-a",
    currentEmail: "old@example.test",
    newEmail: "unfinished@",
  });
  assert.equal(readProfileDraft("learner-a"), "𐐀".repeat(80));
  assert.equal(
    readEmailChangeDraft("learner-a", "old@example.test"),
    "unfinished@",
  );
  assert.equal(readProfileDraft(), null);
  assert.equal(readEmailChangeDraft("learner-a", "changed@example.test"), null);
  assert.equal(readProfileDraft("learner-b"), null);
  assert.equal(readProfileDraft("learner-a"), null);
});

it("rejects overlong account drafts, preserves intentional empty names and clears confirmed drafts", () => {
  saveProfileDraft({ userId: "learner-a", displayName: "" });
  assert.equal(readProfileDraft("learner-a"), "");
  clearProfileDraft("learner-a");
  assert.equal(readProfileDraft("learner-a"), null);
  saveProfileDraft({ userId: "learner-a", displayName: "𐐀".repeat(81) });
  assert.equal(readProfileDraft("learner-a"), null);
  saveEmailChangeDraft({
    userId: "learner-a",
    currentEmail: "old@example.test",
    newEmail: "a".repeat(321),
  });
  assert.equal(readEmailChangeDraft("learner-a", "old@example.test"), null);
  saveEmailChangeDraft({
    userId: "learner-a",
    currentEmail: "old@example.test",
    newEmail: "new@example.test",
  });
  clearEmailChangeDraft("learner-a");
  assert.equal(readEmailChangeDraft("learner-a", "old@example.test"), null);
  saveProfileDraft({ userId: "learner-a", displayName: "Unsaved" });
  saveEmailChangeDraft({
    userId: "learner-a",
    currentEmail: "old@example.test",
    newEmail: "new@example.test",
  });
  clearPreferencesDrafts();
  assert.equal(readProfileDraft("learner-a"), null);
  assert.equal(readEmailChangeDraft("learner-a", "old@example.test"), null);
});

it("restores onboarding answers and step only for the authenticated learner", () => {
  saveOnboardingDraft({ userId: "learner-a", values: answers, step: 4 });
  assert.deepEqual(readOnboardingDraft("learner-a"), {
    values: answers,
    step: 4,
  });
  assert.equal(readOnboardingDraft(), null);
  assert.equal(readOnboardingDraft("learner-b"), null);
  // Switching accounts discards the previous account's private tab drafts.
  assert.equal(readOnboardingDraft("learner-a"), null);
});

it("stores only explicit editable settings fields, never credentials or arbitrary data", () => {
  saveSettingsDraft({
    userId: "learner-a",
    changes: {
      displayName: "Learner 𐐀",
      marketingEmailsEnabled: false,
      dailyReviewTarget: 10,
      password: "must-not-be-stored",
      token: "must-not-be-stored",
      email: "must-not-be-stored",
    } as Parameters<typeof saveSettingsDraft>[0]["changes"],
  });
  assert.deepEqual(readSettingsDraft("learner-a"), {
    dailyReviewTarget: 10,
    marketingEmailsEnabled: false,
    displayName: "Learner 𐐀",
  });
  assert.equal(
    storage.getItem(settingsKey)!.includes("must-not-be-stored"),
    false,
  );
});

it("clears confirmed saves independently and clears all preference drafts on account actions", () => {
  saveOnboardingDraft({ userId: "learner-a", values: answers, step: 3 });
  saveSettingsDraft({
    userId: "learner-a",
    changes: { displayName: "Unsaved" },
  });
  clearSettingsDraft("learner-a");
  assert.equal(readSettingsDraft("learner-a"), null);
  assert.ok(readOnboardingDraft("learner-a"));
  clearOnboardingDraft("learner-a");
  assert.equal(readOnboardingDraft("learner-a"), null);
  saveSettingsDraft({
    userId: "learner-a",
    changes: { dailyReviewTarget: 20 },
  });
  storage.setItem("unrelated-key", "keep");
  clearPreferencesDrafts();
  assert.equal(readSettingsDraft("learner-a"), null);
  assert.equal(storage.getItem("unrelated-key"), "keep");
});

it("drops empty patches and invalid, expired, future, unsupported or oversized envelopes", () => {
  saveSettingsDraft({
    userId: "learner-a",
    changes: { displayName: "Unsaved" },
  });
  saveSettingsDraft({ userId: "learner-a", changes: {} });
  assert.equal(readSettingsDraft("learner-a"), null);
  for (const raw of [
    "not JSON",
    "null",
    JSON.stringify({
      version: 1,
      savedAt: Date.now() - 3 * 60 * 60 * 1000,
      changes: { displayName: "Old" },
    }),
    JSON.stringify({
      version: 1,
      savedAt: Date.now() + 60_000,
      changes: { displayName: "Future" },
    }),
    JSON.stringify({
      version: 2,
      savedAt: Date.now(),
      changes: { displayName: "Unknown" },
    }),
    JSON.stringify({
      version: 1,
      savedAt: Date.now(),
      changes: { dailyReviewTarget: 999 },
    }),
    JSON.stringify({
      version: 1,
      savedAt: Date.now(),
      changes: { displayName: "a".repeat(81) },
    }),
    "a".repeat(4097),
  ]) {
    storage.setItem(settingsKey, raw);
    assert.equal(readSettingsDraft("learner-a"), null);
    assert.equal(storage.getItem(settingsKey), null);
  }
});

it("rejects invalid onboarding values and bounds text by Unicode code points", () => {
  saveOnboardingDraft({
    userId: "learner-a",
    values: { ...answers, nativeLanguage: "a".repeat(81) },
    step: 4,
  });
  assert.equal(readOnboardingDraft("learner-a"), null);
  saveOnboardingDraft({ userId: "learner-a", values: answers, step: 5 });
  assert.equal(readOnboardingDraft("learner-a"), null);
  const name = "𐐀".repeat(80);
  saveSettingsDraft({ userId: "learner-a", changes: { displayName: name } });
  assert.equal(readSettingsDraft("learner-a")?.displayName, name);
});

it("keeps forms usable when storage is unavailable or throws", () => {
  setStorage({
    get length() {
      throw new Error("blocked");
    },
    getItem() {
      throw new Error("blocked");
    },
    setItem() {
      throw new Error("blocked");
    },
    removeItem() {
      throw new Error("blocked");
    },
  });
  assert.doesNotThrow(() =>
    saveSettingsDraft({
      userId: "learner-a",
      changes: { displayName: "Unsaved" },
    }),
  );
  assert.doesNotThrow(() =>
    saveOnboardingDraft({ userId: "learner-a", values: answers, step: 1 }),
  );
  assert.equal(readSettingsDraft("learner-a"), null);
  assert.equal(readOnboardingDraft("learner-a"), null);
  assert.doesNotThrow(() => clearPreferencesDrafts());
});

afterEach(() => {
  storage.clear();
  setStorage(storage);
});
after(() =>
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: originalWindow,
  }),
);
