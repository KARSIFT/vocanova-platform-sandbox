import Link from "@/ui/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";

import { SettingsForm } from "./_components/settings-form";
import { ThemePreferenceControl } from "@/app/_components/theme-preference";
import { BuildIdentity } from "@/app/_components/build-identity";
import { LearnIcon } from "../_components/learn-icons";
import { CalendarReminder } from "./_components/calendar-reminder";

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
      <Link
        href="/home"
        className="-ml-2 inline-flex min-h-11 items-center gap-1.5 rounded-full px-2 text-sm font-semibold text-primary-700 hover:bg-primary-50 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
      >
        <span aria-hidden="true">←</span>
        Back to Home
      </Link>
      <div className="app-page-heading mb-5 mt-1">
        <h1 className="page-title">Settings</h1>
      </div>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <section aria-labelledby="learning-preferences-heading">
          <h2 id="learning-preferences-heading" className="section-title">
            Learning preferences
          </h2>
          <SettingsForm initialSettings={response.data} />
          <CalendarReminder />
        </section>
        <aside className="space-y-5">
          <nav aria-labelledby="account-section-heading">
            <h2 id="account-section-heading" className="section-title mb-2">
              Account
            </h2>
            <div className="grid gap-2">
              <Link href="/settings/profile" className="learn-nav-link">
                <LearnIcon name="profile" />
                <span className="min-w-0 flex-1">Profile</span>
                <LearnIcon name="arrow" />
              </Link>
              <Link href="/settings/account" className="learn-nav-link">
                <LearnIcon name="security" />
                <span className="min-w-0 flex-1">Account security</span>
                <LearnIcon name="arrow" />
              </Link>
            </div>
          </nav>
          <ThemePreferenceControl />
          <Link href="/help" className="vn-text-link">
            Help and support
          </Link>
        </aside>
      </div>
      <details className="mt-6 border-t border-neutral-200 pt-2">
        <summary className="min-h-11 cursor-pointer content-center text-sm font-semibold text-neutral-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700">
          Version and app information
        </summary>
        <BuildIdentity />
      </details>
    </PageContainer>
  );
}
