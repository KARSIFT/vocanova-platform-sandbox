import Link from "next/link";

import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

import {
  PasswordResetForm,
  PasswordResetRequestForm,
} from "../_components/password-forms";

interface PasswordResetPageProps {
  searchParams: Promise<{ token?: string }>;
}

export default async function PasswordResetPage({
  searchParams,
}: PasswordResetPageProps) {
  const { token } = await searchParams;
  const { passwordEnabled } = await getSignInAuthCapabilities();
  const hasToken = Boolean(token);
  const passwordUnavailable = !hasToken && !passwordEnabled;

  return (
    <AuthShell>
      <Surface className="w-full space-y-[var(--spacing-lg)] border-neutral-200 bg-white shadow-[0_12px_28px_rgb(15_23_42_/_0.07)]">
        <div className="space-y-[var(--spacing-sm)]">
          <h1 className="text-2xl font-semibold text-neutral-900">
            {hasToken
              ? "Choose a new password"
              : passwordUnavailable
                ? "Password sign-in is unavailable"
                : "Reset or add a password"}
          </h1>
          <p className="text-base text-neutral-700">
            {hasToken
              ? "Set a password for your VocaNova account."
              : passwordUnavailable
                ? "Use one of the available sign-in methods, then try again later."
                : "We&apos;ll send instructions to your verified email. This also lets magic-link and Google users add a password."}
          </p>
        </div>
        {hasToken ? (
          <PasswordResetForm token={token ?? ""} />
        ) : passwordEnabled ? (
          <PasswordResetRequestForm />
        ) : null}
        <Link
          href="/login"
          className="font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Back to sign in
        </Link>
      </Surface>
    </AuthShell>
  );
}
