import Link from "@/ui/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader, PageSection } from "@/ui/surface";

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
  let user: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    [response, user] = await Promise.all([
      client.getSettings(),
      client.getCurrentUser(),
    ]);
  } catch (error) {
    requireAuthRedirect(error, "/settings");
  }

  return (
    <PageContainer width="wide">
      <PageHeader
        title="Settings"
        back={{ href: "/home", label: "Back to Home" }}
      />

      <div className="grid items-start gap-x-10 gap-y-8 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <PageSection
          title="Learning preferences"
          titleId="learning-preferences-heading"
        >
          <SettingsForm
            key={user.data.id}
            initialSettings={response.data}
            userId={user.data.id}
          />
          <CalendarReminder />
        </PageSection>
        <aside className="min-w-0 space-y-6">
          <nav aria-labelledby="account-section-heading">
            <h2 id="account-section-heading" className="section-title mb-3">
              Account
            </h2>
            <div className="vn-list">
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
      <details className="mt-10 border-t border-neutral-200 pt-2 text-sm">
        <summary className="vn-summary text-neutral-600">
          Version and app information
        </summary>
        <BuildIdentity />
      </details>
    </PageContainer>
  );
}
