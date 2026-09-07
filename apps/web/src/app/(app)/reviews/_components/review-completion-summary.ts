/**
 * Keep the learner-facing completion summary tied only to submissions the
 * backend confirmed. Ratings, correctness, rewards, and mission state are
 * intentionally outside this UI-level count.
 */
export function getCompletedReviewCountAfterSubmission(
  completedReviewCount: number,
  submissionWasConfirmed: boolean,
): number {
  return submissionWasConfirmed
    ? completedReviewCount + 1
    : completedReviewCount;
}

export function getReviewCompletionSummary(
  completedReviewCount: number,
): string | null {
  if (completedReviewCount === 0) {
    return null;
  }

  return `You reviewed ${completedReviewCount} word${
    completedReviewCount === 1 ? "" : "s"
  }.`;
}
