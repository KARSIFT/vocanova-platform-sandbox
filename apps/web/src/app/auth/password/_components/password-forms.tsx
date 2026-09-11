"use client";

import Link from "next/link";
import { useState } from "react";

import { createApiClient } from "@/lib/api";
import { getAuthErrorMessage } from "@/lib/auth-feedback";
import { normalizeReturnTo } from "@/lib/return-to";

const PASSWORD_MIN_LENGTH = 15;
const PASSWORD_MAX_LENGTH = 128;

const inputClassName =
  "mt-[var(--spacing-xs)] block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600";

const primaryButtonClassName =
  "inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50";

function passwordError(password: string): string | null {
  const length = Array.from(password).length;
  if (length < PASSWORD_MIN_LENGTH || length > PASSWORD_MAX_LENGTH) {
    return `Use a password between ${PASSWORD_MIN_LENGTH} and ${PASSWORD_MAX_LENGTH} characters.`;
  }
  return null;
}

function PasswordInput({
  id,
  autoComplete,
  value,
  onChange,
  disabled,
}: {
  id: string;
  autoComplete: "current-password" | "new-password";
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}) {
  const [visible, setVisible] = useState(false);
  return (
    <div>
      <label
        htmlFor={id}
        className="block text-base font-medium text-neutral-900"
      >
        Password
      </label>
      <div className="relative">
        <input
          id={id}
          name={id}
          type={visible ? "text" : "password"}
          autoComplete={autoComplete}
          required
          value={value}
          onChange={(event) => onChange(event.target.value)}
          disabled={disabled}
          className={`${inputClassName} pr-20`}
        />
        <button
          type="button"
          onClick={() => setVisible((current) => !current)}
          aria-label={visible ? "Hide password" : "Show password"}
          className="absolute right-[var(--spacing-xs)] top-[var(--spacing-xs)] min-h-[var(--spacing-2xl)] px-[var(--spacing-sm)] text-sm font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          {visible ? "Hide" : "Show"}
        </button>
      </div>
    </div>
  );
}

export function PasswordLoginForm({ returnTo }: { returnTo: string }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [state, setState] = useState<"idle" | "loading" | "error">("idle");
  const [message, setMessage] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setState("loading");
    setMessage("");
    try {
      await createApiClient().loginWithPassword({
        email: email.trim(),
        password,
      });
      window.location.assign(normalizeReturnTo(returnTo));
    } catch (error) {
      setState("error");
      setMessage(getAuthErrorMessage(error, "password-login"));
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-[var(--spacing-md)]">
      <div>
        <label
          htmlFor="signin-email"
          className="block text-base font-medium text-neutral-900"
        >
          Email address
        </label>
        <input
          id="signin-email"
          name="email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          disabled={state === "loading"}
          className={inputClassName}
        />
      </div>
      <PasswordInput
        id="signin-password"
        autoComplete="current-password"
        value={password}
        onChange={setPassword}
        disabled={state === "loading"}
      />
      {state === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="text-base text-red-700"
        >
          {message}
        </p>
      ) : null}
      <button
        type="submit"
        disabled={state === "loading"}
        aria-busy={state === "loading"}
        className={`${primaryButtonClassName} w-full`}
      >
        {state === "loading" ? "Signing in..." : "Sign in"}
      </button>
      <div className="flex flex-wrap justify-between gap-[var(--spacing-sm)] text-base">
        <Link
          href={`/signup?${new URLSearchParams({ returnTo: normalizeReturnTo(returnTo) })}`}
          className="font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Create account
        </Link>
        <Link
          href="/auth/password/reset"
          className="font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Forgot password?
        </Link>
      </div>
    </form>
  );
}

export function PasswordSignupForm() {
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [state, setState] = useState<"idle" | "loading" | "sent" | "error">(
    "idle",
  );
  const [message, setMessage] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = passwordError(password);
    if (validation) {
      setState("error");
      setMessage(validation);
      return;
    }
    setState("loading");
    setMessage("");
    try {
      await createApiClient().requestPasswordSignup({
        email: email.trim(),
        password,
        ...(displayName.trim() ? { displayName: displayName.trim() } : {}),
      });
      setState("sent");
    } catch (error) {
      setState("error");
      setMessage(getAuthErrorMessage(error, "password-signup"));
    }
  }

  if (state === "sent") {
    return (
      <p
        role="status"
        aria-live="polite"
        className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-md)] text-base text-primary-900"
      >
        If this address can create an account, we&apos;ve sent a verification
        link. Open it to finish setting up your password.
      </p>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-[var(--spacing-md)]">
      <div>
        <label
          htmlFor="signup-email"
          className="block text-base font-medium text-neutral-900"
        >
          Email address
        </label>
        <input
          id="signup-email"
          name="email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          disabled={state === "loading"}
          className={inputClassName}
        />
      </div>
      <div>
        <label
          htmlFor="signup-display-name"
          className="block text-base font-medium text-neutral-900"
        >
          Display name{" "}
          <span className="font-normal text-neutral-700">(optional)</span>
        </label>
        <input
          id="signup-display-name"
          name="name"
          type="text"
          autoComplete="name"
          maxLength={80}
          value={displayName}
          onChange={(event) => setDisplayName(event.target.value)}
          disabled={state === "loading"}
          className={inputClassName}
        />
      </div>
      <PasswordInput
        id="signup-password"
        autoComplete="new-password"
        value={password}
        onChange={setPassword}
        disabled={state === "loading"}
      />
      <p className="text-sm text-neutral-700">Use 15 to 128 characters.</p>
      {state === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="text-base text-red-700"
        >
          {message}
        </p>
      ) : null}
      <button
        type="submit"
        disabled={state === "loading"}
        aria-busy={state === "loading"}
        className={`${primaryButtonClassName} w-full`}
      >
        {state === "loading" ? "Creating account..." : "Create account"}
      </button>
    </form>
  );
}

export function PasswordResetRequestForm({
  initialEmail = "",
  actionLabel = "Send password reset link",
}: {
  initialEmail?: string;
  actionLabel?: string;
}) {
  const [email, setEmail] = useState(initialEmail);
  const [state, setState] = useState<"idle" | "loading" | "sent" | "error">(
    "idle",
  );
  const [message, setMessage] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setState("loading");
    setMessage("");
    try {
      await createApiClient().requestPasswordReset({ email: email.trim() });
      setState("sent");
    } catch (error) {
      setState("error");
      setMessage(getAuthErrorMessage(error, "password-reset-request"));
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-[var(--spacing-md)]">
      <div>
        <label
          htmlFor="reset-email"
          className="block text-base font-medium text-neutral-900"
        >
          Email address
        </label>
        <input
          id="reset-email"
          name="email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          disabled={state === "loading" || state === "sent"}
          className={inputClassName}
        />
      </div>
      <p className="text-sm text-neutral-700">
        We&apos;ll email instructions if this address can use a password. You
        can also use this to add a password after signing in with a magic link
        or Google.
      </p>
      {state === "sent" ? (
        <p
          role="status"
          aria-live="polite"
          className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-sm)] text-base text-primary-900"
        >
          If this address can use a password, we&apos;ve sent instructions to
          it.
        </p>
      ) : null}
      {state === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="text-base text-red-700"
        >
          {message}
        </p>
      ) : null}
      {state !== "sent" ? (
        <button
          type="submit"
          disabled={state === "loading"}
          aria-busy={state === "loading"}
          className={primaryButtonClassName}
        >
          {state === "loading" ? "Sending..." : actionLabel}
        </button>
      ) : null}
    </form>
  );
}

export function PasswordResetForm({ token }: { token: string }) {
  const [password, setPassword] = useState("");
  const [state, setState] = useState<"idle" | "loading" | "done" | "error">(
    "idle",
  );
  const [message, setMessage] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = passwordError(password);
    if (validation) {
      setState("error");
      setMessage(validation);
      return;
    }
    setState("loading");
    setMessage("");
    try {
      await createApiClient().resetPassword({ token, password });
      window.history.replaceState(
        window.history.state,
        "",
        "/auth/password/reset",
      );
      setState("done");
    } catch (error) {
      setState("error");
      setMessage(getAuthErrorMessage(error, "password-reset"));
    }
  }

  if (state === "done")
    return (
      <div className="space-y-[var(--spacing-md)]">
        <p
          role="status"
          aria-live="polite"
          className="rounded-md border border-primary-200 bg-primary-50 p-[var(--spacing-md)] text-base text-primary-900"
        >
          Your password is ready. Sign in with your email and new password.
        </p>
        <Link href="/login" className={primaryButtonClassName}>
          Sign in
        </Link>
      </div>
    );
  return (
    <form onSubmit={handleSubmit} className="space-y-[var(--spacing-md)]">
      <PasswordInput
        id="reset-password"
        autoComplete="new-password"
        value={password}
        onChange={setPassword}
        disabled={state === "loading"}
      />
      <p className="text-sm text-neutral-700">Use 15 to 128 characters.</p>
      {state === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="text-base text-red-700"
        >
          {message}
        </p>
      ) : null}
      <button
        type="submit"
        disabled={state === "loading"}
        aria-busy={state === "loading"}
        className={primaryButtonClassName}
      >
        {state === "loading" ? "Saving password..." : "Save new password"}
      </button>
    </form>
  );
}
