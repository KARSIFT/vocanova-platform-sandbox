"use client";

import { useRef, useState } from "react";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";

export function ProfileForm({
  initialDisplayName,
}: {
  initialDisplayName: string;
}) {
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [state, setState] = useState<
    "idle" | "saving" | "saved" | "newer_changes" | "error"
  >("idle");
  const [message, setMessage] = useState("");
  const currentDisplayName = useRef(initialDisplayName);
  const isSaving = useRef(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isSaving.current) return;

    const submittedDisplayName = currentDisplayName.current;
    isSaving.current = true;
    setState("saving");
    setMessage("");
    try {
      const csrfToken = await getOrRefreshCSRFToken();
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
      const savedDisplayName = data.displayName ?? submittedDisplayName.trim();
      // Only replace the submitted text with the server's normalized value.
      // Typing during the request creates a newer draft, not another save.
      if (currentDisplayName.current === submittedDisplayName) {
        currentDisplayName.current = savedDisplayName;
        setDisplayName(savedDisplayName);
      }
      setState(
        currentDisplayName.current.trim() === savedDisplayName
          ? "saved"
          : "newer_changes",
      );
    } catch (error) {
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
      className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]"
    >
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
          maxLength={80}
          value={displayName}
          onChange={(event) => {
            currentDisplayName.current = event.target.value;
            setDisplayName(event.target.value);
            if (!isSaving.current) {
              setState("idle");
              setMessage("");
            }
          }}
          className="mt-[var(--spacing-xs)] block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
        />
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
        disabled={state === "saving"}
        aria-busy={state === "saving"}
        className="inline-flex min-h-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {state === "saving" ? "Saving..." : "Save profile"}
      </button>
    </form>
  );
}
