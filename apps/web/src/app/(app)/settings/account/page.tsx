import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";

import { AccountDeletionForm } from "./_components/account-deletion-form";
import { PersonalDataExport } from "./_components/personal-data-export";
import { EmailChangeForm } from "./_components/email-change-form";

export const metadata = {
  title: "Account — Vocanova",
  description: "Manage your Vocanova sign-in and account.",
};

export default async function SettingsAccountPage() {
  const client = await createServerApiClient();
  let meResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    meResponse = await client.getCurrentUser();
  } catch (error) {
    requireAuthRedirect(error, "/settings/account");
  }

  const currentEmail = meResponse.data.email ?? "";

  return (
    <div className="p-[var(--spacing-lg)]">
      <div className="mb-[var(--spacing-md)] flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-neutral-900">Account</h1>
        <Link
          href="/settings"
          className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          Back to Settings
        </Link>
      </div>

      <p className="text-base text-neutral-700">
        Update your sign-in email, then finish with a confirmation link. You can
        also delete your account from here.
      </p>

      <section
        aria-labelledby="email-change-heading"
        className="mt-[var(--spacing-lg)] rounded-md border border-neutral-200 bg-neutral-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h2
          id="email-change-heading"
          className="text-lg font-semibold text-neutral-900"
        >
          Sign-in email
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
          We&apos;ll send a confirmation link to your new address. Your current
          sign-in stays active until you confirm.
        </p>
        <p className="mt-[var(--spacing-sm)] text-base text-neutral-900">
          Current address:{" "}
          <span className="font-medium">
            {currentEmail ? currentEmail : "Not available"}
          </span>
        </p>
        <EmailChangeForm currentEmail={currentEmail} />
      </section>

      <PersonalDataExport />

      <section
        aria-labelledby="account-deletion-heading"
        className="mt-[var(--spacing-lg)] rounded-md border border-red-200 bg-red-50 p-[var(--spacing-md)] shadow-sm"
      >
        <h2
          id="account-deletion-heading"
          className="text-lg font-semibold text-red-900"
        >
          Delete your account
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-red-900">
          We&apos;ll deactivate your account right away, then permanently
          anonymize your data after 30 days. After 30 days, your data is gone
          for good.
        </p>
        <AccountDeletionForm />
      </section>
    </div>
  );
}
