import type { Metadata } from "next";

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
    <main className="grid min-h-screen place-items-center bg-neutral-50 p-6">
      <div
        role="status"
        aria-live="polite"
        className="w-full max-w-[36rem] space-y-[var(--spacing-md)] rounded-xl border border-red-300 bg-white p-[var(--spacing-lg)] text-base text-neutral-900 shadow-sm"
      >
        <div>
          <h1 className="text-2xl font-semibold text-red-900">
            Your account has been deactivated.
          </h1>
          {formattedPurgeAfter ? (
            <p className="mt-[var(--spacing-xs)]">
              We&apos;ll permanently anonymize your saved words, reviews, and
              practice history on{" "}
              <span className="font-medium">{formattedPurgeAfter}</span>. After
              that, your data is gone for good.
            </p>
          ) : null}
          <p className="mt-[var(--spacing-sm)]">
            You can sign in again before then to reactivate, and your data will
            be restored.
          </p>
        </div>
        <a
          href="/signin"
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Go to sign in
        </a>
      </div>
    </main>
  );
}
