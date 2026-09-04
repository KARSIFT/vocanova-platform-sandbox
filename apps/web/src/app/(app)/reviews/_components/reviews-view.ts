/**
 * Pure view-selection logic for the Reviews screen (VOC-1179).
 *
 * Kept separate from `page.tsx` so the "which state renders" decision is
 * unit-testable without a DOM/React renderer, matching the existing
 * `review-session-prompt.ts` pattern in this directory.
 */
export type ReviewsView = "empty" | "session";

/**
 * Decide whether the Reviews screen should show the "all caught up" empty
 * state or the active review session, based on how many words are due.
 */
export function getReviewsView(dueWordCount: number): ReviewsView {
  return dueWordCount > 0 ? "session" : "empty";
}
