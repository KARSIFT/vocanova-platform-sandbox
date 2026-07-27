"use client";

import { useState } from "react";

import { Settings, UpdateSettingsBody } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

const DAILY_REVIEW_TARGETS = [5, 10, 15, 20, 30, 50, 75, 100];

const REVIEW_INTERVAL_PRESETS = [
  {
    value: "vocanova_default",
    label: "Vocanova default",
    helper: "Spaced repetition tuned for everyday vocabulary.",
  },
  {
    value: "wordup_like",
    label: "Faster reminders",
    helper: "Words come back sooner — useful before an exam or trip.",
  },
  {
    value: "custom",
    label: "Custom",
    helper: "A later custom preset. We will add it here when it ships.",
  },
] as const;

type SaveStatus =
  | { type: "idle" }
  | { type: "saving" }
  | { type: "saved" }
  | { type: "error"; message: string };

interface SettingsFormProps {
  initialSettings: Settings;
}

type FormState = Omit<Settings, "displayName" | "reviewIntervalPreset"> & {
  displayName: string;
  reviewIntervalPreset: Settings["reviewIntervalPreset"];
};

export function SettingsForm({ initialSettings }: SettingsFormProps) {
  const [state, setState] = useState<FormState>({
    dailyReviewTarget: initialSettings.dailyReviewTarget,
    reviewIntervalPreset: initialSettings.reviewIntervalPreset,
    appLanguage: initialSettings.appLanguage,
    notificationsEnabled: initialSettings.notificationsEnabled,
    marketingEmailsEnabled: initialSettings.marketingEmailsEnabled,
    displayName: initialSettings.displayName,
  });
  const [status, setStatus] = useState<SaveStatus>({ type: "idle" });

  function patch<K extends keyof FormState>(key: K, value: FormState[K]) {
    setState((current) => ({ ...current, [key]: value }));
    if (status.type === "saved" || status.type === "error") {
      setStatus({ type: "idle" });
    }
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setStatus({
        type: "error",
        message:
          "Your session is missing a security token. Please refresh the page and try again.",
      });
      return;
    }

    const body = buildUpdateBody(state, initialSettings);
    if (Object.keys(body).length === 0) {
      setStatus({ type: "saved" });
      return;
    }

    setStatus({ type: "saving" });
    const client = createApiClient();
    try {
      const { data } = await client.updateSettings(body, {
        headers: { "X-CSRF-Token": csrfToken },
      });
      setState({
        dailyReviewTarget: data.dailyReviewTarget,
        reviewIntervalPreset: data.reviewIntervalPreset,
        appLanguage: data.appLanguage,
        notificationsEnabled: data.notificationsEnabled,
        marketingEmailsEnabled: data.marketingEmailsEnabled,
        displayName: data.displayName,
      });
      setStatus({ type: "saved" });
    } catch (error) {
      // T06: a 401 mid-settings-write is the documented
      // session-expiry mid-flow case. handleApiError routes the
      // learner to re-auth; the form state is preserved on the
      // controlled inputs so the learner does not need to retype
      // their changes after re-authentication.
      const message = handleApiError(
        error,
        "We couldn't save your settings. Please try again.",
      );
      setStatus({ type: "error", message });
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      aria-label="Practice settings"
      className="mt-[var(--spacing-lg)] space-y-[var(--spacing-lg)]"
    >
      <fieldset className="space-y-[var(--spacing-md)]">
        <legend className="text-lg font-semibold text-neutral-900">
          Daily review target
        </legend>
        <p className="text-base text-neutral-700">
          How many saved words would you like to review each day? Changes take
          effect from your next local day.
        </p>
        <ul
          role="radiogroup"
          aria-label="Daily review target"
          className="grid grid-cols-2 gap-[var(--spacing-sm)] sm:grid-cols-4"
        >
          {DAILY_REVIEW_TARGETS.map((target) => {
            const checked = state.dailyReviewTarget === target;
            return (
              <li key={target}>
                <label
                  className={`flex min-h-[var(--spacing-2xl)] cursor-pointer items-center justify-center rounded-md border p-[var(--spacing-sm)] text-base font-medium transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700 ${
                    checked
                      ? "border-primary-600 bg-primary-50 text-primary-800"
                      : "border-neutral-200 bg-white text-neutral-900 hover:border-primary-300"
                  }`}
                >
                  <input
                    type="radio"
                    name="dailyReviewTarget"
                    value={target}
                    checked={checked}
                    onChange={() => patch("dailyReviewTarget", target)}
                    className="sr-only"
                  />
                  {target}
                </label>
              </li>
            );
          })}
        </ul>
      </fieldset>

      <fieldset className="space-y-[var(--spacing-md)]">
        <legend className="text-lg font-semibold text-neutral-900">
          Review rhythm
        </legend>
        <p className="text-base text-neutral-700">
          Pick the reminder rhythm that fits your schedule.
        </p>
        <div
          role="radiogroup"
          aria-label="Review rhythm preset"
          className="space-y-[var(--spacing-sm)]"
        >
          {REVIEW_INTERVAL_PRESETS.map((preset) => {
            const checked = state.reviewIntervalPreset === preset.value;
            return (
              <label
                key={preset.value}
                className={`flex min-h-[var(--spacing-2xl)] cursor-pointer items-start gap-[var(--spacing-sm)] rounded-md border p-[var(--spacing-md)] transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] ${
                  checked
                    ? "border-primary-600 bg-primary-50"
                    : "border-neutral-200 bg-white hover:border-primary-300"
                } focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700`}
              >
                <input
                  type="radio"
                  name="reviewIntervalPreset"
                  value={preset.value}
                  checked={checked}
                  onChange={() => patch("reviewIntervalPreset", preset.value)}
                  className="mt-[var(--spacing-xs)] size-4 accent-primary-600"
                />
                <span className="flex flex-col">
                  <span className="text-base font-medium text-neutral-900">
                    {preset.label}
                  </span>
                  <span className="text-sm text-neutral-700">
                    {preset.helper}
                  </span>
                </span>
              </label>
            );
          })}
        </div>
      </fieldset>

      <fieldset className="space-y-[var(--spacing-md)]">
        <legend className="text-lg font-semibold text-neutral-900">
          App language
        </legend>
        <p className="text-base text-neutral-700">
          English is the only language Vocanova is offered in today. We will add
          more languages as we translate the app.
        </p>
        <label className="flex min-h-[var(--spacing-2xl)] cursor-pointer items-start gap-[var(--spacing-sm)] rounded-md border border-primary-600 bg-primary-50 p-[var(--spacing-md)] focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700">
          <input
            type="radio"
            name="appLanguage"
            value="en"
            checked={state.appLanguage === "en"}
            onChange={() => patch("appLanguage", "en")}
            className="mt-[var(--spacing-xs)] size-4 accent-primary-600"
          />
          <span className="flex flex-col">
            <span className="text-base font-medium text-neutral-900">
              English
            </span>
            <span className="text-sm text-neutral-700">
              Selected — Vocanova&apos;s only supported language right now.
            </span>
          </span>
        </label>
      </fieldset>

      <fieldset className="space-y-[var(--spacing-md)]">
        <legend className="text-lg font-semibold text-neutral-900">
          Notifications and emails
        </legend>
        <div className="space-y-[var(--spacing-sm)]">
          <ToggleRow
            id="notifications-enabled"
            label="Daily review reminders"
            description="We will send a gentle reminder when your reviews are waiting."
            checked={state.notificationsEnabled}
            onChange={(value) => patch("notificationsEnabled", value)}
          />
          <ToggleRow
            id="marketing-emails-enabled"
            label="Product news and tips"
            description="Occasional updates about new features and learning tips. You can opt out any time."
            checked={state.marketingEmailsEnabled}
            onChange={(value) => patch("marketingEmailsEnabled", value)}
          />
        </div>
      </fieldset>

      <fieldset className="space-y-[var(--spacing-md)]">
        <legend className="text-lg font-semibold text-neutral-900">
          Display name
        </legend>
        <p className="text-base text-neutral-700">
          This is how Vocanova will greet you. Leave blank to use your email.
        </p>
        <label className="sr-only" htmlFor="displayName">
          Display name
        </label>
        <input
          id="displayName"
          name="displayName"
          type="text"
          maxLength={80}
          autoComplete="nickname"
          value={state.displayName}
          onChange={(event) => patch("displayName", event.target.value)}
          className="block w-full rounded-md border border-neutral-300 px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base text-neutral-900 focus:border-primary-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-600"
        />
      </fieldset>

      {status.type === "error" ? (
        <p
          role="alert"
          aria-live="assertive"
          className="rounded-md border border-red-300 bg-red-50 p-[var(--spacing-sm)] text-base text-red-800"
        >
          {status.message}
        </p>
      ) : null}

      {status.type === "saved" ? (
        <p
          role="status"
          aria-live="polite"
          className="rounded-md border border-green-300 bg-green-50 p-[var(--spacing-sm)] text-base text-green-800"
        >
          Your settings have been saved.
        </p>
      ) : null}

      <div className="flex flex-wrap items-center justify-end gap-[var(--spacing-sm)]">
        <button
          type="submit"
          disabled={status.type === "saving"}
          aria-busy={status.type === "saving"}
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {status.type === "saving" ? "Saving..." : "Save settings"}
        </button>
      </div>
    </form>
  );
}

function buildUpdateBody(
  next: FormState,
  baseline: Settings,
): UpdateSettingsBody {
  const body: UpdateSettingsBody = {};
  if (next.dailyReviewTarget !== baseline.dailyReviewTarget) {
    body.dailyReviewTarget = next.dailyReviewTarget;
  }
  if (next.reviewIntervalPreset !== baseline.reviewIntervalPreset) {
    body.reviewIntervalPreset = next.reviewIntervalPreset;
  }
  if (next.appLanguage !== baseline.appLanguage) {
    body.appLanguage = next.appLanguage;
  }
  if (next.notificationsEnabled !== baseline.notificationsEnabled) {
    body.notificationsEnabled = next.notificationsEnabled;
  }
  if (next.marketingEmailsEnabled !== baseline.marketingEmailsEnabled) {
    body.marketingEmailsEnabled = next.marketingEmailsEnabled;
  }
  if (next.displayName !== baseline.displayName) {
    body.displayName = next.displayName;
  }
  return body;
}

interface ToggleRowProps {
  id: string;
  label: string;
  description: string;
  checked: boolean;
  onChange: (next: boolean) => void;
}

function ToggleRow({
  id,
  label,
  description,
  checked,
  onChange,
}: ToggleRowProps) {
  return (
    <div className="flex items-start justify-between gap-[var(--spacing-md)] rounded-md border border-neutral-200 bg-white p-[var(--spacing-md)]">
      <div>
        <label htmlFor={id} className="text-base font-medium text-neutral-900">
          {label}
        </label>
        <p className="mt-[var(--spacing-xs)] text-sm text-neutral-700">
          {description}
        </p>
      </div>
      <div className="flex flex-col items-end gap-[var(--spacing-xs)]">
        <span className="sr-only">{`Toggle ${label}`}</span>
        <button
          id={id}
          type="button"
          role="switch"
          aria-checked={checked}
          onClick={() => onChange(!checked)}
          className={`relative inline-flex h-[var(--spacing-xl)] w-[var(--spacing-2xl)] items-center rounded-full transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 ${
            checked ? "bg-primary-600" : "bg-neutral-300"
          }`}
        >
          <span
            aria-hidden="true"
            className={`inline-block h-[var(--spacing-md)] w-[var(--spacing-md)] transform rounded-full bg-white transition-transform duration-[var(--duration-fast)] ease-[var(--ease-out)] ${
              checked
                ? "translate-x-[var(--spacing-xl)]"
                : "translate-x-[var(--spacing-xs)]"
            }`}
          />
        </button>
        <span className="text-sm text-neutral-700" aria-hidden="true">
          {checked ? "On" : "Off"}
        </span>
      </div>
    </div>
  );
}
