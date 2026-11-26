import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
import { PageContainer, PageHeader, Surface } from "@/ui/surface";

import { AccountDeletionForm } from "./_components/account-deletion-form";
import { PersonalDataExport } from "./_components/personal-data-export";
import { EmailChangeForm } from "./_components/email-change-form";
import { PasswordSecurity } from "./_components/password-security";

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
  const { passwordEnabled, emailChangeEnabled } =
    await getSignInAuthCapabilities();

  return (
    <PageContainer>
      <PageHeader
        title="Account"
        back={{ href: "/settings", label: "Back to Settings" }}
        lede="Manage your sign-in and personal data."
      />

      <Surface aria-labelledby="email-change-heading">
        <h2 id="email-change-heading" className="section-title">
          Sign-in email
        </h2>
        {emailChangeEnabled ? (
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            We&apos;ll send a confirmation link to your new address. Your
            current sign-in stays active until you confirm.
          </p>
        ) : null}
        <p className="mt-[var(--spacing-sm)] text-base text-neutral-900">
          Current address:{" "}
          <span className="font-medium [overflow-wrap:anywhere]">
            {currentEmail ? currentEmail : "Not available"}
          </span>
        </p>
        {emailChangeEnabled ? (
          <EmailChangeForm
            key={meResponse.data.id}
            userId={meResponse.data.id}
            currentEmail={currentEmail}
          />
        ) : (
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            Changing your sign-in email is currently unavailable. Your current
            address stays in use.
          </p>
        )}
      </Surface>

      <PersonalDataExport />

      {passwordEnabled ? (
        <PasswordSecurity
          email={currentEmail}
          hasPassword={meResponse.data.hasPassword === true}
        />
      ) : null}

      <section
        aria-labelledby="account-deletion-heading"
        className="mt-[var(--spacing-lg)] rounded-[1.25rem] border border-red-200 bg-red-50 p-[var(--spacing-md)] sm:p-[var(--spacing-lg)]"
      >
        <h2
          id="account-deletion-heading"
          className="text-lg font-semibold text-red-900"
        >
          Delete your account
        </h2>
        <p className="mt-[var(--spacing-xs)] text-base text-red-900">
          Your account closes immediately. We&apos;ll schedule your data for
          permanent anonymization after 30 days.
        </p>
        <AccountDeletionForm />
      </section>
    </PageContainer>
  );
}
