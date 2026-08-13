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

/**
 * Post-submit learner actions (show answer, rate, continue) must stay locked
 * while a submission is in flight *or* while a batch-end `listDueWords`
 * refetch is replacing the queue. After `submitAttempt` succeeds, `advance()`
 * may set `isRefetching` while the same card remains on screen; clearing only
 * `isSubmitting` in `finally` would otherwise re-enable rate/continue and
 * allow a duplicate `submitAttempt` (VOC-076-T01 remediation).
 */
export function isReviewActionDisabled(
  isSubmitting: boolean,
  isRefetching: boolean,
): boolean {
  return isSubmitting || isRefetching;
}
