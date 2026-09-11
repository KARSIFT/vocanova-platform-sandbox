import Link from "next/link";
import type { Metadata } from "next";

import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";

import { getReviewsView } from "./_components/reviews-view";
import { ReviewSession } from "./_components/review-session";
import { ReviewSentenceDraftRecovery } from "./_components/review-sentence-draft-recovery";
import {
  getDueRequestLimit,
  getRemainingReviewTarget,
} from "./_components/review-session-limit";

export const metadata: Metadata = {
  title: "Review — Vocanova",
  description: "Review your saved vocabulary with Vocanova.",
};

export default async function ReviewsPage() {
  const client = await createServerApiClient();
  let dailyMissionResponse: Awaited<ReturnType<typeof client.getDailyMission>>;
  let currentUserResponse: Awaited<ReturnType<typeof client.getCurrentUser>>;
  try {
    [dailyMissionResponse, currentUserResponse] = await Promise.all([
      client.getDailyMission(),
      client.getCurrentUser(),
    ]);
  } catch (error) {
    requireAuthRedirect(error, "/reviews");
  }

  const remainingReviewTarget = getRemainingReviewTarget(
    dailyMissionResponse.data.reviewTarget,
    dailyMissionResponse.data.reviewsCompleted,
  );

  if (remainingReviewTarget === 0) {
    return <MissionTargetComplete userId={currentUserResponse.data.id} />;
  }

  let dueResponse: Awaited<ReturnType<typeof client.listDueWords>>;
  try {
    dueResponse = await client.listDueWords({
      limit: getDueRequestLimit(remainingReviewTarget),
    });
  } catch (error) {
    requireAuthRedirect(error, "/reviews");
  }

  const { items: dueWords, totalCount } = dueResponse.data;

  return (
    <PageContainer className="max-w-[40rem]">
      <div className="mb-[var(--spacing-md)] flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-neutral-900">Review</h1>
        <Link
          href="/home"
          className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          Back to Home
        </Link>
      </div>

      {getReviewsView(dueWords.length) === "empty" ? (
        <div className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
          <h2 className="text-xl font-semibold text-neutral-900">
            You&apos;re all caught up
          </h2>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            Nice work — no words are due for review right now. Come back later,
            or head to Discover to learn something new in the meantime.
          </p>
          <div className="mt-[var(--spacing-lg)] flex flex-wrap justify-center gap-[var(--spacing-md)]">
            <Link
              href="/discover"
              className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Discover new words
            </Link>
            <Link
              href="/home"
              className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
            >
              Back to Home
            </Link>
          </div>
        </div>
      ) : (
        <ReviewSession
          initialDueWords={dueWords}
          initialTotalCount={Math.min(totalCount, remainingReviewTarget)}
          reviewSessionLimit={remainingReviewTarget}
          userId={currentUserResponse.data.id}
        />
      )}
      <ReviewSentenceDraftRecovery userId={currentUserResponse.data.id} />
    </PageContainer>
  );
}

function MissionTargetComplete({ userId }: Readonly<{ userId?: string }>) {
  return (
    <PageContainer className="max-w-[40rem]">
      <div className="mb-[var(--spacing-md)] flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-neutral-900">Review</h1>
        <Link
          href="/home"
          className="text-base font-semibold text-primary-700 hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          Back to Home
        </Link>
      </div>
      <div className="flex flex-col items-center justify-center py-[var(--spacing-2xl)] text-center">
        <h2 className="text-xl font-semibold text-neutral-900">
          Today&apos;s review target is complete
        </h2>
        <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
          Nice work. Come back tomorrow for your next review session.
        </p>
        <Link
          href="/home"
          className="mt-[var(--spacing-lg)] inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Back to Home
        </Link>
        <ReviewSentenceDraftRecovery userId={userId} />
      </div>
    </PageContainer>
  );
}
