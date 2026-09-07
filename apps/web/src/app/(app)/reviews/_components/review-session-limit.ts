const MAX_DUE_WORDS_PER_REQUEST = 50;

/** Returns the number of reviews left in today's authoritative mission. */
export function getRemainingReviewTarget(
  reviewTarget: number,
  reviewsCompleted: number,
): number {
  return Math.max(0, reviewTarget - reviewsCompleted);
}

/** Keeps each due-queue request inside both the session and API limits. */
export function getDueRequestLimit(remainingSessionReviews: number): number {
  return Math.min(MAX_DUE_WORDS_PER_REQUEST, remainingSessionReviews);
}

/** A session ends as soon as it reaches its confirmed-review allowance. */
export function hasReachedReviewSessionLimit(
  confirmedReviewCount: number,
  reviewSessionLimit: number,
): boolean {
  return confirmedReviewCount >= reviewSessionLimit;
}
