"use client";

import { useEffect, useState } from "react";

import { SentenceFeedback } from "../../_components/sentence-feedback";
import {
  readReviewCompletionContext,
  readSentenceFeedbackDraft,
  type ReviewCompletionContext,
} from "../../_components/sentence-feedback-drafts";

export function ReviewSentenceDraftRecovery({
  userId,
}: Readonly<{ userId?: string }>) {
  const [context, setContext] = useState<ReviewCompletionContext | null>(null);

  useEffect(() => {
    setContext(null);
    const storedContext = readReviewCompletionContext(userId);
    if (
      storedContext &&
      readSentenceFeedbackDraft({
        userId,
        source: "review",
        attemptId: storedContext.attemptId,
      })
    ) {
      setContext(storedContext);
    }
  }, [userId]);

  if (!context) {
    return null;
  }

  return (
    <section
      aria-labelledby="recovered-sentence-practice-heading"
      className="mt-[var(--spacing-lg)] w-full max-w-[28rem] text-left"
    >
      <h2
        id="recovered-sentence-practice-heading"
        className="text-lg font-semibold text-neutral-900"
      >
        Continue your sentence practice
      </h2>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Your unfinished sentence is ready to continue in this tab.
      </p>
      <SentenceFeedback
        targetWord={context.targetWord}
        attemptId={context.attemptId}
        source="review"
        userId={userId}
        shortDefinition={context.shortDefinition}
      />
    </section>
  );
}
