import Link from "@/ui/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, PageHeader, PageSection, Surface } from "@/ui/surface";

import { ProfileForm } from "./_components/profile-form";

export const metadata = {
  title: "Profile — Vocanova",
  description: "Update your VocaNova profile.",
};

function initials(name: string, email: string): string {
  const source = name.trim() || email.trim();
  return (
    source
      .split(/[\s@._-]+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => Array.from(part)[0]?.toUpperCase())
      .join("") || "V"
  );
}

export default async function ProfilePage() {
  const client = await createServerApiClient();
  let meResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    meResponse = await client.getCurrentUser();
  } catch (error) {
    requireAuthRedirect(error, "/settings/profile");
  }
  const user = meResponse.data;
  const email = user.email ?? "Not available";
  const hasVerifiedEmail = Boolean(user.email && user.emailVerifiedAt);
  const displayName = user.displayName ?? "";
  return (
    <PageContainer>
      <PageHeader
        title="Profile"
        back={{ href: "/settings", label: "Back to Settings" }}
        lede="Manage the details VocaNova uses to personalize your learning."
      />
      <Surface aria-labelledby="profile-details-heading">
        <div className="flex flex-wrap items-center gap-[var(--spacing-md)]">
          <span
            aria-hidden="true"
            className="grid size-14 shrink-0 place-items-center rounded-full bg-primary-100 text-lg font-bold text-primary-800"
          >
            {initials(displayName, email)}
          </span>
          <div className="min-w-0 flex-1 basis-[12rem]">
            <h2 id="profile-details-heading" className="section-title">
              Your profile
            </h2>
            <p className="text-sm text-neutral-700">
              Your initials are used as your avatar.
            </p>
          </div>
        </div>
        <ProfileForm
          key={user.id}
          userId={user.id}
          initialDisplayName={displayName}
        />
      </Surface>
      <Surface
        aria-labelledby="profile-email-heading"
        tone="flat"
        className="mt-4"
      >
        <h2
          id="profile-email-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          {hasVerifiedEmail ? "Verified email" : "Email"}
        </h2>
        <p className="mt-[var(--spacing-xs)] break-all text-base text-neutral-900">
          {email}
        </p>
        <p className="mt-[var(--spacing-xs)] text-sm text-neutral-700">
          {hasVerifiedEmail
            ? "Your email is verified and can be changed from Account security."
            : "This email has not been verified yet. Check your inbox for a verification link."}
        </p>
      </Surface>
      <PageSection
        title="More settings"
        titleId="profile-links-heading"
        className="mt-8"
      >
        <div className="flex flex-wrap gap-x-5 gap-y-1">
          <Link href="/settings/account" className="vn-text-link">
            Account security
          </Link>
          <Link href="/settings" className="vn-text-link">
            Learning preferences
          </Link>
        </div>
      </PageSection>
    </PageContainer>
  );
}
