"use client";

import { useState } from "react";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";

export function ProfileForm({
  initialDisplayName,
}: {
  initialDisplayName: string;
}) {
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [state, setState] = useState<"idle" | "saving" | "saved" | "error">(
    "idle",
  );
  const [message, setMessage] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
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
        { displayName: displayName.trim() },
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      setDisplayName(data.displayName ?? displayName.trim());
      setState("saved");
    } catch (error) {
      setState("error");
      setMessage(
        handleApiError(
          error,
          "We couldn't save your profile. Please try again.",
        ),
      );
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
            setDisplayName(event.target.value);
            if (state !== "idle") setState("idle");
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
