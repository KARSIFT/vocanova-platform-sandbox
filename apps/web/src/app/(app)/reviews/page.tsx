import Link from "next/link";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";

import { ReviewSession } from "./_components/review-session";

export default async function ReviewsPage() {
  const client = await createServerApiClient();
  let dueResponse: Awaited<ReturnType<typeof client.listDueWords>>;
  try {
    dueResponse = await client.listDueWords({ limit: 50 });
  } catch (error) {
    requireAuthRedirect(error, "/reviews");
  }

  const { items: dueWords, totalCount } = dueResponse.data;

  return (
    <div className="p-[var(--spacing-lg)]">
      <div className="mb-[var(--spacing-md)] flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-neutral-900">Review</h1>
        <Link
          href="/home"
          className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          Back to Home
        </Link>
      </div>

      {dueWords.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            You&apos;re all caught up
          </h2>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            No words are due for review right now.
          </p>
          <Link
            href="/home"
            className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            Back to Home
          </Link>
        </div>
      ) : (
        <ReviewSession
          initialDueWords={dueWords}
          initialTotalCount={totalCount}
        />
      )}
    </div>
  );
}
