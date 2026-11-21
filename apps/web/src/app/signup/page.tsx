import Link from "@/ui/link";

import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
import { normalizeReturnTo } from "@/lib/return-to";
import { AuthDivider, AuthShell } from "@/ui/auth-shell";

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
  const capabilities = await getSignInAuthCapabilities();
  const { passwordEnabled, oauthEnabled } = capabilities;
  const unavailable = !passwordEnabled && !oauthEnabled;
  const description = passwordEnabled
    ? oauthEnabled
      ? "Use Google, or we'll email a verification link before your password can be used."
      : "We'll email a verification link before your password can be used."
    : oauthEnabled
      ? "Continue with Google to sign in or create your account."
      : "Please try again later or use an available sign-in method.";

  return (
    <AuthShell
      title={
        unavailable ? "Account creation is unavailable" : "Create your account"
      }
      description={description}
      showExample
    >
      {oauthEnabled ? (
        <OAuthButton
          returnTo={safeReturnTo}
          capabilities={capabilities}
          intent="signup"
        />
      ) : null}

      {passwordEnabled && oauthEnabled ? (
        <AuthDivider>or use email and password</AuthDivider>
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
    </AuthShell>
  );
}
