import type { SubmitReviewBody } from "@vocanova/api-client";

export interface ReviewSubmissionIntent {
  userWordId: string;
  meaningId: string;
  promptType: "multiple_choice" | "self_check";
  result: "correct" | "incorrect";
  rating: "again" | "hard" | "good" | "easy";
  selectedOptionMeaningId?: string;
}

export interface PendingReviewSubmission {
  body: SubmitReviewBody;
  idempotencyKey: string;
}

// A transport failure after the server commits is ambiguous. Keep the exact
// original body and key so retrying the same learner decision is safe, but
// never reuse that key for a changed answer.
export function matchesPendingReviewSubmission(
  pending: PendingReviewSubmission,
  intent: ReviewSubmissionIntent,
): boolean {
  const { body } = pending;
  return (
    body.userWordId === intent.userWordId &&
    body.meaningId === intent.meaningId &&
    body.promptType === intent.promptType &&
    body.result === intent.result &&
    body.rating === intent.rating &&
    body.selectedOptionMeaningId === intent.selectedOptionMeaningId
  );
}
