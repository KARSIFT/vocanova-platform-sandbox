"use client";

import { useState } from "react";

import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getAppOrigin } from "@/lib/env";

interface MagicLinkFormProps {
  returnTo: string;
}

export function MagicLinkForm({ returnTo }: MagicLinkFormProps) {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<{
    type: "idle" | "loading" | "success" | "error";
    message: string;
  }>({ type: "idle", message: "" });

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus({ type: "loading", message: "Sending sign-in link..." });

    const client = createApiClient();
    try {
      await client.requestMagicLink({ email, returnTo });
      setStatus({
        type: "success",
        message: `If ${email} is valid, a sign-in link has been sent. Check your email and return to ${returnTo}.`,
      });
    } catch (error) {
      const message =
        error instanceof ApiResponseError
          ? error.message
          : "Unable to send sign-in link. Please try again.";
      setStatus({ type: "error", message });
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-[var(--spacing-md)]">
      <div>
        <label
          htmlFor="email"
          className="block text-base font-medium text-neutral-900"
        >
          Email address
        </label>
        <input
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="mt-[var(--spacing-xs)] block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
        />
      </div>
      {status.message ? (
        <p
          role="alert"
          aria-live="polite"
          className={`text-base ${status.type === "success" ? "text-green-700" : status.type === "error" ? "text-red-700" : "text-neutral-700"}`}
        >
          {status.message}
        </p>
      ) : null}
      <button
        type="submit"
        disabled={status.type === "loading"}
        aria-busy={status.type === "loading"}
        className="min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {status.type === "loading" ? "Sending..." : "Send sign-in link"}
      </button>
    </form>
  );
}

interface OAuthButtonProps {
  returnTo: string;
}

export function OAuthButton({ returnTo }: OAuthButtonProps) {
  const [status, setStatus] = useState<{
    type: "idle" | "loading" | "error";
    message: string;
  }>({
    type: "idle",
    message: "",
  });

  async function handleClick() {
    setStatus({ type: "loading", message: "Redirecting to Google..." });
    const client = createApiClient();
    try {
      const redirectUri = `${getAppOrigin()}${returnTo}`;
      const { data } = await client.startOAuth({ redirectUri });
      window.location.href = data.url;
    } catch (error) {
      const message =
        error instanceof ApiResponseError
          ? error.message
          : "Unable to start Google sign-in. Please try again.";
      setStatus({ type: "error", message });
    }
  }

  return (
    <div className="space-y-[var(--spacing-sm)]">
      <button
        type="button"
        onClick={handleClick}
        disabled={status.type === "loading"}
        aria-busy={status.type === "loading"}
        className="min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] w-full rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {status.type === "loading" ? "Redirecting..." : "Continue with Google"}
      </button>
      {status.message ? (
        <p role="alert" aria-live="polite" className="text-base text-red-700">
          {status.message}
        </p>
      ) : null}
    </div>
  );
}
