"use client";

import Link from "@/ui/link";
import { useEffect, useRef, useState } from "react";

import { createApiClient } from "@/lib/api";
import { getAuthErrorMessage } from "@/lib/auth-feedback";
import { normalizeReturnTo } from "@/lib/return-to";
import { isDisplayNameWithinLimit } from "@/lib/display-name";

const PASSWORD_MIN_LENGTH = 15;
const PASSWORD_MAX_LENGTH = 128;

const inputClassName = "auth-input mt-[var(--spacing-xs)]";

const primaryButtonClassName = "auth-primary";

function useResendCooldown() {
  const [availableAt, setAvailableAt] = useState(0);
  const [now, setNow] = useState(0);
  useEffect(() => {
    if (availableAt <= now) return;
    const interval = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(interval);
  }, [availableAt, now]);
  return {
    seconds: Math.max(0, Math.ceil((availableAt - now) / 1000)),
    isBlocked: () => Date.now() < availableAt,
    start: () => {
      const startedAt = Date.now();
      setNow(startedAt);
      setAvailableAt(startedAt + 60_000);
    },
    reset: () => setAvailableAt(0),
  };
}

function CheckPasswordEmail({
  email,
  signup = false,
  sending,
  seconds,
  message,
  onResend,
  onDifferentEmail,
}: {
  email: string;
  signup?: boolean;
  sending: boolean;
  seconds: number;
  message: string;
  onResend: () => void;
  onDifferentEmail: () => void;
}) {
  return (
    <div className="space-y-[var(--spacing-md)]">
      <p
        role="status"
        aria-live="polite"
        className="rounded-xl border border-primary-200 bg-primary-50 p-[var(--spacing-md)] text-base text-primary-900"
      >
        {signup
          ? "If this address can create an account, we've sent a verification link. Open it to finish setting up your password."
          : "If this address can use a password, we've sent instructions to it."}
        <span className="mt-[var(--spacing-xs)] block font-medium [overflow-wrap:anywhere]">
          {email}
        </span>
      </p>
      <p className="text-sm text-neutral-700">
        Check your spam folder too. Links expire after 15 minutes.
      </p>
      <div className="flex flex-wrap gap-[var(--spacing-sm)]">
        <button
          type="button"
          onClick={onResend}
          disabled={sending || seconds > 0}
          aria-busy={sending}
          className="auth-primary"
        >
          {sending
            ? "Sending..."
            : seconds > 0
              ? `Resend available in ${seconds}s`
              : signup
                ? "Resend verification link"
                : "Resend password reset link"}
        </button>
        <button
          type="button"
          onClick={onDifferentEmail}
          disabled={sending}
          className="auth-secondary"
        >
          Use a different email
        </button>
      </div>
      {message ? (
        <p
          role="alert"
          aria-live="assertive"
          className="text-base text-red-700"
        >
          {message}
        </p>
      ) : null}
    </div>
  );
}

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
      <div className="relative mt-[var(--spacing-xs)]">
        <input
          id={id}
          name={id}
          type={visible ? "text" : "password"}
          autoComplete={autoComplete}
          required
          value={value}
          onChange={(event) => onChange(event.target.value)}
          disabled={disabled}
          className="auth-input pr-20"
        />
        <button
          type="button"
          onClick={() => setVisible((current) => !current)}
          aria-label={visible ? "Hide password" : "Show password"}
          className="absolute inset-y-0 right-1 my-auto inline-flex h-11 min-w-11 items-center justify-center rounded-lg px-[var(--spacing-sm)] text-sm font-semibold text-primary-700 hover:bg-primary-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
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
  const [hasSent, setHasSent] = useState(false);
  const requesting = useRef(false);
  const emailInput = useRef<HTMLInputElement>(null);
  const focusEmailOnReturn = useRef(false);
  const cooldown = useResendCooldown();

  useEffect(() => {
    if (!hasSent && focusEmailOnReturn.current) {
      emailInput.current?.focus();
      focusEmailOnReturn.current = false;
    }
  }, [hasSent]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await requestLink();
  }

  async function requestLink() {
    if (requesting.current || (hasSent && cooldown.isBlocked())) return;
    const validation = passwordError(password);
    if (validation) {
      setState("error");
      setMessage(validation);
      return;
    }
    requesting.current = true;
    setState("loading");
    setMessage("");
    try {
      await createApiClient().requestPasswordSignup({
        email: email.trim(),
        password,
        ...(displayName.trim() ? { displayName: displayName.trim() } : {}),
      });
      setHasSent(true);
      setState("sent");
      cooldown.start();
    } catch (error) {
      setState("error");
      setMessage(getAuthErrorMessage(error, "password-signup"));
    } finally {
      requesting.current = false;
    }
  }

  if (hasSent) {
    return (
      <CheckPasswordEmail
        email={email.trim()}
        signup
        sending={state === "loading"}
        seconds={cooldown.seconds}
        message={message}
        onResend={() => void requestLink()}
        onDifferentEmail={() => {
          focusEmailOnReturn.current = true;
          setPassword("");
          setHasSent(false);
          setState("idle");
          setMessage("");
          cooldown.reset();
        }}
      />
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
          ref={emailInput}
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
          aria-describedby="signup-display-name-limit"
          value={displayName}
          onChange={(event) => {
            if (isDisplayNameWithinLimit(event.target.value)) {
              setDisplayName(event.target.value);
            }
          }}
          disabled={state === "loading"}
          className={inputClassName}
        />
        <p
          id="signup-display-name-limit"
          className="mt-[var(--spacing-xs)] text-sm text-neutral-700"
        >
          Up to 80 characters.
        </p>
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
  const [hasSent, setHasSent] = useState(false);
  const requesting = useRef(false);
  const emailInput = useRef<HTMLInputElement>(null);
  const focusEmailOnReturn = useRef(false);
  const cooldown = useResendCooldown();

  useEffect(() => {
    if (!hasSent && focusEmailOnReturn.current) {
      emailInput.current?.focus();
      focusEmailOnReturn.current = false;
    }
  }, [hasSent]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await requestLink();
  }

  async function requestLink() {
    if (requesting.current || (hasSent && cooldown.isBlocked())) return;
    requesting.current = true;
    setState("loading");
    setMessage("");
    try {
      await createApiClient().requestPasswordReset({ email: email.trim() });
      setHasSent(true);
      setState("sent");
      cooldown.start();
    } catch (error) {
      setState("error");
      setMessage(getAuthErrorMessage(error, "password-reset-request"));
    } finally {
      requesting.current = false;
    }
  }

  if (hasSent) {
    return (
      <CheckPasswordEmail
        email={email.trim()}
        sending={state === "loading"}
        seconds={cooldown.seconds}
        message={message}
        onResend={() => void requestLink()}
        onDifferentEmail={() => {
          focusEmailOnReturn.current = true;
          setHasSent(false);
          setState("idle");
          setMessage("");
          cooldown.reset();
        }}
      />
    );
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
          ref={emailInput}
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
      <p className="text-sm text-neutral-700">
        We&apos;ll email instructions if this address can use a password. You
        can also use this to add a password after signing in with a magic link
        or Google.
      </p>
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
        {state === "loading" ? "Sending..." : actionLabel}
      </button>
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
          className="rounded-xl border border-primary-200 bg-primary-50 p-[var(--spacing-md)] text-base text-primary-900"
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
