import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, Surface } from "@/ui/surface";

import { SettingsForm } from "./_components/settings-form";
import { ThemePreferenceControl } from "@/app/_components/theme-preference";
import { BuildIdentity } from "@/app/_components/build-identity";

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
    <PageContainer className="lg:max-w-[72rem]">
      <div className="app-page-heading mb-[var(--spacing-md)] flex items-center justify-between gap-[var(--spacing-md)]">
        <h1 className="text-3xl font-semibold tracking-[-0.04em] text-neutral-900">
          Settings
        </h1>
        <Link
          href="/home"
          className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          Back to Home
        </Link>
      </div>

      <p className="app-page-heading text-base leading-7 text-neutral-700">
        Update your daily review target, review rhythm, and other practice
        preferences.
      </p>

      <div className="mt-[var(--spacing-lg)] grid gap-[var(--spacing-lg)] lg:grid-cols-[minmax(0,1fr)_22rem] lg:items-start">
        <aside className="space-y-[var(--spacing-md)] lg:order-2">
          <Surface
            aria-labelledby="account-section-heading"
            className="border-secondary-200 bg-secondary-50"
          >
            <h2
              id="account-section-heading"
              className="text-lg font-semibold text-neutral-900"
            >
              Account
            </h2>
            <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
              Manage your profile, sign-in email, password, or account deletion.
            </p>
            <Link
              href="/settings/account"
              className="mt-[var(--spacing-md)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Account security
            </Link>
            <Link
              href="/settings/profile"
              className="mt-[var(--spacing-md)] ml-[var(--spacing-sm)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Profile
            </Link>
          </Surface>
          <ThemePreferenceControl />
          <BuildIdentity />
        </aside>
        <div className="lg:order-1">
          <h2 className="text-xl font-semibold tracking-[-0.03em] text-neutral-900">
            Learning preferences
          </h2>
          <p className="mt-[var(--spacing-xs)] text-sm leading-6 text-neutral-600">
            Set the pace and reminders that support your daily practice.
          </p>
          <SettingsForm initialSettings={response.data} />
        </div>
      </div>
    </PageContainer>
  );
}
