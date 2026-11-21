import { getSignInAuthCapabilities } from "@/lib/auth-capabilities";
import { AuthShell } from "@/ui/auth-shell";
import { ActionLink } from "@/ui/controls";

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
  const title = hasToken
    ? "Choose a new password"
    : passwordUnavailable
      ? "Password sign-in is unavailable"
      : "Reset or add a password";
  const description = hasToken
    ? "Set a password for your VocaNova account."
    : passwordUnavailable
      ? "Use one of the available sign-in methods, then try again later."
      : "We'll send instructions to your verified email. This also lets magic-link and Google users add a password.";

  return (
    <AuthShell title={title} description={description}>
      {hasToken ? (
        <PasswordResetForm token={token ?? ""} />
      ) : passwordEnabled ? (
        <PasswordResetRequestForm />
      ) : null}
      <ActionLink href="/login" variant="text" className="justify-self-start">
        Back to sign in
      </ActionLink>
    </AuthShell>
  );
}
