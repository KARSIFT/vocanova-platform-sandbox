import { StatusMessage } from "@/ui/status-message";
import type { Metadata } from "next";
import { AuthShell } from "@/ui/auth-shell";
import { actionClasses } from "@/ui/controls";

export const metadata: Metadata = {
  title: "Account deactivated — Vocanova",
  description: "Confirmation that your Vocanova account has been deactivated.",
};

interface AccountDeactivatedPageProps {
  searchParams: Promise<{ purgeAfter?: string }>;
}

function formatPurgeAfter(purgeAfter: string | undefined): string | null {
  if (!purgeAfter) {
    return null;
  }

  const date = new Date(purgeAfter);
  if (Number.isNaN(date.getTime())) {
    return null;
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "long",
    timeZone: "UTC",
  }).format(date);
}

export default async function AccountDeactivatedPage({
  searchParams,
}: AccountDeactivatedPageProps) {
  const { purgeAfter } = await searchParams;
  const formattedPurgeAfter = formatPurgeAfter(purgeAfter);

  return (
    <AuthShell linkBrand={false}>
      <StatusMessage
        role="status"
        aria-live="polite"
        className="space-y-[var(--spacing-lg)]"
      >
        <div>
          <h1 className="page-title">Your account has been deactivated.</h1>
          <p className="page-lede">
            Access to your account has ended, and you can no longer sign in with
            it.
          </p>
          {formattedPurgeAfter ? (
            <p className="page-lede">
              Your saved words, reviews, and practice history are scheduled to
              be permanently anonymized on{" "}
              <span className="font-medium">{formattedPurgeAfter}</span>.
            </p>
          ) : (
            <p className="page-lede">
              Your saved words, reviews, and practice history are scheduled to
              be permanently anonymized.
            </p>
          )}
          <p className="page-lede">
            If you have questions about your account, our support team can help.
          </p>
        </div>
        <a href="/help" className={actionClasses("primary")}>
          Get help and support
        </a>
      </StatusMessage>
    </AuthShell>
  );
}
