export type PromptPhase = "prompt" | "feedback" | "rate";

/**
 * Background queue refetch (`isRefetching`) must not disable multiple-choice
 * options on a prompt-ready card — only an in-flight submission or the
 * post-selection feedback phase should block re-selection (VOC-076-T01).
 */
export function isMultipleChoiceOptionDisabled(
  phase: PromptPhase,
  isSubmitting: boolean,
): boolean {
  return isSubmitting || phase === "feedback";
}

/** Learner actions (show answer, rate, continue) block only during submit. */
export function isReviewActionDisabled(isSubmitting: boolean): boolean {
  return isSubmitting;
}
