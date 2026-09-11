import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer, Surface } from "@/ui/surface";

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
      <div className="mb-[var(--spacing-md)] flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-neutral-900">Profile</h1>
        <Link
          href="/settings"
          className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          Back to Settings
        </Link>
      </div>
      <p className="text-base text-neutral-700">
        Manage the details VocaNova uses to personalize your learning.
      </p>
      <Surface
        aria-labelledby="profile-details-heading"
        className="mt-[var(--spacing-lg)]"
      >
        <div className="flex items-center gap-[var(--spacing-md)]">
          <span
            aria-hidden="true"
            className="grid size-14 shrink-0 place-items-center rounded-full bg-primary-100 text-lg font-bold text-primary-800"
          >
            {initials(displayName, email)}
          </span>
          <div>
            <h2
              id="profile-details-heading"
              className="text-lg font-semibold text-neutral-900"
            >
              Your profile
            </h2>
            <p className="text-sm text-neutral-700">
              Your initials are used as your avatar.
            </p>
          </div>
        </div>
        <ProfileForm initialDisplayName={displayName} />
      </Surface>
      <Surface
        aria-labelledby="profile-email-heading"
        className="mt-[var(--spacing-lg)]"
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
      <Surface
        aria-labelledby="profile-links-heading"
        className="mt-[var(--spacing-lg)]"
      >
        <h2
          id="profile-links-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          More settings
        </h2>
        <div className="mt-[var(--spacing-md)] flex flex-wrap gap-[var(--spacing-md)]">
          <Link
            href="/settings/account"
            className="inline-flex min-h-[var(--spacing-2xl)] items-center font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Account security
          </Link>
          <Link
            href="/settings"
            className="inline-flex min-h-[var(--spacing-2xl)] items-center font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Learning preferences
          </Link>
        </div>
      </Surface>
    </PageContainer>
  );
}
