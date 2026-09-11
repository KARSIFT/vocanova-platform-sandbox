import Link from "next/link";

import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
import { normalizeReturnTo } from "@/lib/return-to";
import { AuthShell } from "@/ui/auth-shell";
import { Surface } from "@/ui/surface";

import { PasswordSignupForm } from "../auth/password/_components/password-forms";
import { OAuthButton } from "../signin/_components/auth-forms";

export const metadata = {
  title: "Create account — Vocanova",
  description:
    "Create a VocaNova account with Google or your email and password.",
};

interface SignupPageProps {
  searchParams: Promise<{ returnTo?: string }>;
}

export default async function SignupPage({ searchParams }: SignupPageProps) {
  const { returnTo } = await searchParams;
  const safeReturnTo = normalizeReturnTo(returnTo);
  const { passwordEnabled, oauthEnabled } = await getSignInAuthCapabilities();
  const unavailable = !passwordEnabled && !oauthEnabled;

  return (
    <AuthShell>
      <Surface className="w-full space-y-[var(--spacing-lg)] border-neutral-200 bg-white shadow-[0_12px_28px_rgb(15_23_42_/_0.07)]">
        <div className="space-y-[var(--spacing-sm)]">
          <h1 className="text-xl font-semibold text-neutral-900">
            {unavailable
              ? "Account creation is unavailable"
              : "Create your account"}
          </h1>
          <p className="text-base text-neutral-700">
            {passwordEnabled
              ? "Use Google, or we'll email a verification link before your password can be used."
              : oauthEnabled
                ? "Continue with Google to sign in or create your account."
                : "Please try again later or use an available sign-in method."}
          </p>
        </div>

        {oauthEnabled ? <OAuthButton returnTo={safeReturnTo} /> : null}

        {passwordEnabled && oauthEnabled ? (
          <div className="relative flex items-center gap-[var(--spacing-sm)]">
            <div className="h-px flex-1 bg-neutral-200" />
            <span className="text-sm text-neutral-500">
              or use email and password
            </span>
            <div className="h-px flex-1 bg-neutral-200" />
          </div>
        ) : null}

        {passwordEnabled ? <PasswordSignupForm /> : null}

        <p className="text-base text-neutral-700">
          Already have an account?{" "}
          <Link
            href={`/login?${new URLSearchParams({ returnTo: safeReturnTo })}`}
            className="font-semibold text-primary-700 underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Sign in
          </Link>
        </p>
      </Surface>
    </AuthShell>
  );
}
