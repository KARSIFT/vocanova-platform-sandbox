"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { isDisplayNameWithinLimit } from "@/lib/display-name";
import {
  clearProfileDraft,
  isPreferencesDraftDocumentActive,
  readProfileDraft,
  saveProfileDraft,
} from "@/lib/preferences-draft";

export function ProfileForm({
  initialDisplayName,
  userId,
}: {
  initialDisplayName: string;
  userId?: string;
}) {
  const router = useRouter();
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [state, setState] = useState<
    "idle" | "saving" | "saved" | "newer_changes" | "error"
  >("idle");
  const [message, setMessage] = useState("");
  const currentDisplayName = useRef(initialDisplayName);
  const isSaving = useRef(false);
  const mounted = useRef(false);
  const recoveryScope = useRef<{ userId?: string } | null>(null);
  const [recoveredScope, setRecoveredScope] = useState<{
    userId?: string;
  } | null>(null);
  const baseline = useRef(initialDisplayName);
  const [restored, setRestored] = useState(false);
  const ready = recoveredScope !== null && recoveredScope.userId === userId;

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(() => {
    if (recoveryScope.current?.userId === userId && recoveryScope.current)
      return;
    const scope = { userId };
    recoveryScope.current = scope;
    baseline.current = initialDisplayName;
    const draft = readProfileDraft(userId);
    const next = draft ?? initialDisplayName;
    currentDisplayName.current = next;
    setDisplayName(next);
    const dirty = next !== initialDisplayName;
    if (!dirty) clearProfileDraft(userId);
    setRestored(dirty);
    setState("idle");
    setMessage("");
    setRecoveredScope(scope);
  }, [initialDisplayName, userId]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!ready) return;
    if (isSaving.current || !isPreferencesDraftDocumentActive()) return;

    const submittedDisplayName = currentDisplayName.current;
    const scope = recoveryScope.current;
    isSaving.current = true;
    setState("saving");
    setMessage("");
    try {
      const csrfToken = await getOrRefreshCSRFToken();
      if (
        !isPreferencesDraftDocumentActive() ||
        !mounted.current ||
        recoveryScope.current !== scope
      )
        return;
      if (!csrfToken) {
        setState("error");
        setMessage(
          "We couldn't prepare these changes securely. Please try again.",
        );
        return;
      }
      const { data } = await createApiClient().updateSettings(
        { displayName: submittedDisplayName.trim() },
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      if (
        !isPreferencesDraftDocumentActive() ||
        !mounted.current ||
        recoveryScope.current !== scope
      )
        return;
      const savedDisplayName = data.displayName ?? submittedDisplayName.trim();
      baseline.current = savedDisplayName;
      // Only replace the submitted text with the server's normalized value.
      // Typing during the request creates a newer draft, not another save.
      if (currentDisplayName.current === submittedDisplayName) {
        currentDisplayName.current = savedDisplayName;
        setDisplayName(savedDisplayName);
      }
      const hasNewerChanges =
        currentDisplayName.current.trim() !== savedDisplayName;
      if (hasNewerChanges) {
        saveProfileDraft({ userId, displayName: currentDisplayName.current });
      } else {
        clearProfileDraft(userId);
      }
      setRestored(false);
      setState(!hasNewerChanges ? "saved" : "newer_changes");
      router.refresh();
    } catch (error) {
      if (
        !isPreferencesDraftDocumentActive() ||
        !mounted.current ||
        recoveryScope.current !== scope
      )
        return;
      // A failed response may follow a committed write. Keep the latest
      // input, including an intentional reversion, for an explicit retry.
      saveProfileDraft({ userId, displayName: currentDisplayName.current });
      setState("error");
      setMessage(
        handleApiError(
          error,
          "We couldn't save your profile. Please try again.",
        ),
      );
    } finally {
      isSaving.current = false;
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      aria-busy={!ready || state === "saving"}
      className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]"
    >
      {!ready ? (
        <p role="status" className="text-sm text-neutral-700">
          Loading your profile…
        </p>
      ) : null}
      {restored ? (
        <p role="status" className="text-sm text-neutral-700">
          Your unsaved profile changes were restored in this tab. Save when
          you&apos;re ready.
        </p>
      ) : null}
      <div>
        <label
          htmlFor="profile-display-name"
          className="block text-base font-medium text-neutral-900"
        >
          Display name
        </label>
        <input
          id="profile-display-name"
          name="displayName"
          type="text"
          autoComplete="nickname"
          aria-describedby="profile-display-name-limit"
          value={displayName}
          disabled={!ready}
          onChange={(event) => {
            if (!isDisplayNameWithinLimit(event.target.value)) return;
            currentDisplayName.current = event.target.value;
            setDisplayName(event.target.value);
            if (isSaving.current || event.target.value !== baseline.current) {
              saveProfileDraft({ userId, displayName: event.target.value });
            } else {
              clearProfileDraft(userId);
              setRestored(false);
            }
            if (!isSaving.current) {
              setState("idle");
              setMessage("");
            }
          }}
          className="mt-[var(--spacing-xs)] block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
        />
        <p
          id="profile-display-name-limit"
          className="mt-[var(--spacing-xs)] text-sm text-neutral-700"
        >
          Up to 80 characters.
        </p>
      </div>
      {state === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="text-base text-red-700"
        >
          {message}
        </p>
      ) : null}
      {state === "saved" ? (
        <p
          role="status"
          aria-live="polite"
          className="text-base text-green-800"
        >
          Your profile has been saved.
        </p>
      ) : null}
      {state === "newer_changes" ? (
        <p
          role="status"
          aria-live="polite"
          className="text-base text-neutral-700"
        >
          Your earlier profile changes were saved. You have newer changes to
          save.
        </p>
      ) : null}
      <button
        type="submit"
        disabled={!ready || state === "saving"}
        aria-busy={state === "saving"}
        className="vn-btn vn-btn-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
      >
        {state === "saving" ? "Saving..." : "Save profile"}
      </button>
    </form>
  );
}
