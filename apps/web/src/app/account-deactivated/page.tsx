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
          {formattedPurgeAfter ? (
            <p className="page-lede">
              We&apos;ll permanently anonymize your saved words, reviews, and
              practice history on{" "}
              <span className="font-medium">{formattedPurgeAfter}</span>. After
              that, your data is gone for good.
            </p>
          ) : null}
          <p className="page-lede">
            You can sign in again before then to reactivate, and your data will
            be restored.
          </p>
        </div>
        <a href="/login" className={actionClasses("primary")}>
          Go to sign in
        </a>
      </StatusMessage>
    </AuthShell>
  );
}
