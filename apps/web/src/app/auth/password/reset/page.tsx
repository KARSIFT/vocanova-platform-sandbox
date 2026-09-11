import Link from "next/link";

import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
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
    <main className="grid min-h-screen place-items-center bg-neutral-100 p-6">
      <Surface className="w-full max-w-[28rem] space-y-[var(--spacing-lg)]">
        <div className="space-y-[var(--spacing-sm)]">
          <p className="text-sm font-bold tracking-wide text-primary-700">
            VOCANOVA
          </p>
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
    </main>
  );
}
