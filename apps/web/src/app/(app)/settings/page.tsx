import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";

import { SettingsForm } from "./_components/settings-form";

export const metadata = {
  title: "Settings — Vocanova",
  description: "Update your practice preferences.",
};

export default async function SettingsPage() {
  const client = await createServerApiClient();
  let response: Awaited<ReturnType<typeof client.getSettings>>;
  try {
    response = await client.getSettings();
  } catch (error) {
    requireAuthRedirect(error, "/settings");
  }

  return (
    <div className="p-[var(--spacing-lg)]">
      <div className="mb-[var(--spacing-md)] flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-neutral-900">Settings</h1>
        <Link
          href="/home"
          className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          Back to Home
        </Link>
      </div>

      <p className="text-base text-neutral-700">
        Update your daily review target, review rhythm, and other practice
        preferences.
      </p>

      <SettingsForm initialSettings={response.data} />

      <section
        aria-labelledby="account-section-heading"
        className="mt-[var(--spacing-lg)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h2
          id="account-section-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          Account
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          Change your sign-in email, display name, or delete your account.
        </p>
        <Link
          href="/settings/account"
          className="mt-[var(--spacing-md)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Manage account
        </Link>
      </section>
    </div>
  );
}
