"use client";

import { PasswordResetRequestForm } from "@/app/auth/password/_components/password-forms";

export function PasswordSecurity({
  email,
  hasPassword,
}: {
  email: string;
  hasPassword: boolean;
}) {
  return (
    <section
      aria-labelledby="password-security-heading"
      className="mt-[var(--spacing-lg)] rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm sm:p-[var(--spacing-lg)]"
    >
      <h2
        id="password-security-heading"
        className="text-lg font-semibold text-neutral-900"
      >
        Password
      </h2>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        {hasPassword
          ? "A password is set for this account. We’ll email a link so you can choose a new one."
          : "No password is set. Add one to sign in with your email and password."}
      </p>
      <div className="mt-[var(--spacing-md)]">
        <PasswordResetRequestForm
          initialEmail={email}
          actionLabel={
            hasPassword
              ? "Email password reset link"
              : "Email link to add a password"
          }
        />
      </div>
    </section>
  );
}
