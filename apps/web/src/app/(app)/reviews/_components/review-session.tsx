"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { DueWord, SubmitReviewBody } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";
import { SentenceFeedback } from "../../_components/sentence-feedback";

type Rating = "again" | "hard" | "good" | "easy";

type PromptPhase = "prompt" | "feedback" | "rate";

const RATING_LABELS: Record<Rating, string> = {
  again: "Again",
  hard: "Hard",
  good: "Good",
  easy: "Easy",
};

const RATING_ORDER: Rating[] = ["again", "hard", "good", "easy"];

interface ReviewOption {
  meaningId: string;
  label: string;
}

interface ReviewSessionProps {
  initialDueWords: DueWord[];
  initialTotalCount: number;
}

export function ReviewSession({
  initialDueWords,
  initialTotalCount,
}: ReviewSessionProps) {
  const [dueWords, setDueWords] = useState<DueWord[]>(initialDueWords);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [remainingCount, setRemainingCount] = useState(initialTotalCount);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isRefetching, setIsRefetching] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [completed, setCompleted] = useState(false);
  const [phase, setPhase] = useState<PromptPhase>("prompt");
  const [selectedOption, setSelectedOption] = useState<string | null>(null);
  const [startTime, setStartTime] = useState<number>(Date.now());
  const [lastReviewedCard, setLastReviewedCard] = useState<DueWord | null>(
    null,
  );
  const [lastReviewAttemptId, setLastReviewAttemptId] = useState<string | null>(
    null,
  );

  const currentCard = dueWords[currentIndex];

  const promptType = currentCard
    ? determinePromptType(dueWords, currentIndex)
    : null;
  const options =
    currentCard && promptType === "multiple_choice"
      ? buildMultipleChoiceOptions(dueWords, currentIndex)
      : null;

  useEffect(() => {
    setPhase("prompt");
    setSelectedOption(null);
    setErrorMessage(null);
    setStartTime(Date.now());
  }, [currentIndex, dueWords]);

  const advance = () => {
    if (currentIndex + 1 < dueWords.length) {
      setCurrentIndex((index) => index + 1);
      return;
    }

    setIsRefetching(true);
    const client = createApiClient();
    client
      .listDueWords({ limit: 50 })
      .then(({ data }) => {
        if (data.items.length > 0) {
          setDueWords(data.items);
          setRemainingCount(data.totalCount);
          setCurrentIndex(0);
        } else {
          setCompleted(true);
        }
      })
      .catch((error) => {
        // T06: a 401 here means the session expired mid-review-session;
        // route the learner to re-auth instead of leaving them looking at
        // an error on a frozen card.
        setErrorMessage(
          handleApiError(error, "Unable to load more words. Please try again."),
        );
      })
      .finally(() => {
        setIsRefetching(false);
      });
  };

  const submitAttempt = async ({
    result,
    rating,
    selectedOptionMeaningId,
  }: {
    result: "correct" | "incorrect";
    rating: Rating;
    selectedOptionMeaningId?: string;
  }) => {
    if (!currentCard) {
      return;
    }

    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setErrorMessage("Session is not ready. Please refresh the page.");
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    const client = createApiClient();
    const clientAttemptId = generateClientAttemptId();
    const body: SubmitReviewBody = {
      userWordId: currentCard.userWordId,
      meaningId: currentCard.meaningId,
      attemptType: "review",
      promptType:
        promptType === "multiple_choice" ? "multiple_choice" : "self_check",
      result,
      rating,
      answeredAt: new Date().toISOString(),
      responseTimeMs: Math.max(0, Date.now() - startTime),
      selectedOptionMeaningId,
      wasHintUsed: false,
      source: "review_session",
      clientAttemptId,
    };

    try {
      const { data } = await client.submitReview(body, clientAttemptId, {
        headers: { "X-CSRF-Token": csrfToken },
      });
      setLastReviewedCard(currentCard);
      setLastReviewAttemptId(data.attemptId);
      setRemainingCount((count) => Math.max(0, count - 1));
      advance();
    } catch (error) {
      // T06: a 401 mid-review-session is the documented
      // session-expiry mid-flow case — never claim a card was
      // reviewed when the server rejected it. handleApiError routes
      // the learner to re-auth instead.
      setErrorMessage(
        handleApiError(
          error,
          "Unable to submit your answer. Please try again.",
        ),
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const isLoading = isSubmitting || isRefetching;

  if (dueWords.length === 0 || completed) {
    return (
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
        {lastReviewedCard && lastReviewAttemptId ? (
          <div className="mt-[var(--spacing-lg)] w-full max-w-md text-left">
            <SentenceFeedback
              targetWord={lastReviewedCard.wordText}
              attemptId={lastReviewAttemptId}
              source="review"
              shortDefinition={lastReviewedCard.shortDefinition}
            />
          </div>
        ) : null}
      </div>
    );
  }

  if (!currentCard) {
    return null;
  }

  const isMultipleChoiceCorrect =
    promptType === "multiple_choice" &&
    selectedOption === currentCard.meaningId;
  const isMultipleChoiceIncorrect =
    promptType === "multiple_choice" &&
    selectedOption !== null &&
    selectedOption !== currentCard.meaningId;

  return (
    <div className="flex flex-col">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-neutral-600">
          {remainingCount} word{remainingCount === 1 ? "" : "s"} remaining
        </p>
        <p className="text-sm text-neutral-500">
          Card {currentIndex + 1} of {dueWords.length}
        </p>
      </div>

      <div className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm">
        <div className="mb-[var(--spacing-lg)] text-center">
          <span className="inline-block rounded-full bg-neutral-100 px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-sm text-neutral-700">
            {currentCard.partOfSpeech}
          </span>
          <h2 className="mt-[var(--spacing-sm)] text-3xl font-semibold text-neutral-900">
            {currentCard.wordText}
          </h2>
          {promptType === "self_check" && phase !== "rate" ? (
            <p className="mt-[var(--spacing-sm)] text-base text-neutral-600">
              Think of the meaning, then reveal the answer.
            </p>
          ) : null}
          {promptType === "multiple_choice" ? (
            <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
              Select the matching meaning.
            </p>
          ) : null}
        </div>

        {promptType === "self_check" && phase === "rate" ? (
          <div className="mb-[var(--spacing-lg)] rounded-md bg-primary-50 p-[var(--spacing-md)] text-center">
            <p className="text-sm font-medium text-primary-900">Answer</p>
            <p className="mt-[var(--spacing-xs)] text-lg text-primary-900">
              {currentCard.shortDefinition}
            </p>
          </div>
        ) : null}

        {promptType === "multiple_choice" && options ? (
          <fieldset className="mb-[var(--spacing-lg)]">
            <legend className="sr-only">
              Choose the meaning for {currentCard.wordText}
            </legend>
            <div className="space-y-[var(--spacing-sm)]">
              {options.map((option) => {
                const isSelected = selectedOption === option.meaningId;
                const isCorrect = option.meaningId === currentCard.meaningId;
                const showCorrectness = phase === "feedback";
                const isDisabled = isLoading || phase === "feedback";
                return (
                  <button
                    key={option.meaningId}
                    type="button"
                    aria-pressed={isSelected}
                    disabled={isDisabled}
                    onClick={() => {
                      if (phase === "prompt") {
                        setSelectedOption(option.meaningId);
                        setPhase("feedback");
                      }
                    }}
                    className={`w-full rounded-md border px-[var(--spacing-md)] py-[var(--spacing-sm)] text-left text-base transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-60 ${
                      showCorrectness && isCorrect
                        ? "border-primary-500 bg-primary-50 text-primary-900"
                        : showCorrectness && isSelected && !isCorrect
                          ? "border-red-300 bg-red-50 text-red-900"
                          : "border-neutral-200 bg-neutral-50 text-neutral-900 hover:bg-neutral-100"
                    }`}
                  >
                    <span className="font-medium">{option.label}</span>
                    {showCorrectness && isCorrect ? (
                      <span className="ml-[var(--spacing-sm)] text-sm">
                        (correct)
                      </span>
                    ) : null}
                  </button>
                );
              })}
            </div>
          </fieldset>
        ) : null}

        {promptType === "self_check" && phase === "prompt" ? (
          <button
            type="button"
            onClick={() => setPhase("rate")}
            disabled={isLoading}
            className="w-full rounded-md border border-neutral-200 bg-neutral-50 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Show answer
          </button>
        ) : null}

        {phase === "feedback" && isMultipleChoiceIncorrect ? (
          <div className="mb-[var(--spacing-lg)] rounded-md border border-red-200 bg-red-50 p-[var(--spacing-md)]">
            <p className="font-medium text-red-900">Not quite</p>
            <p className="mt-[var(--spacing-xs)] text-base text-red-800">
              The correct answer was: {currentCard.shortDefinition}
            </p>
            <button
              type="button"
              onClick={() =>
                submitAttempt({
                  result: "incorrect",
                  rating: "again",
                  selectedOptionMeaningId: selectedOption ?? undefined,
                })
              }
              disabled={isLoading}
              className="mt-[var(--spacing-md)] w-full rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Continue
            </button>
          </div>
        ) : null}

        {(phase === "feedback" && isMultipleChoiceCorrect) ||
        (promptType === "self_check" && phase === "rate") ? (
          <div className="mb-[var(--spacing-lg)]">
            {isMultipleChoiceCorrect ? (
              <p className="mb-[var(--spacing-md)] text-center text-lg font-medium text-primary-900">
                Correct
              </p>
            ) : null}
            <fieldset>
              <legend className="sr-only">
                {isMultipleChoiceCorrect
                  ? "How well did you know this word?"
                  : "How well did you know this word?"}
              </legend>
              <div
                className={`grid gap-[var(--spacing-sm)] ${
                  isMultipleChoiceCorrect ? "grid-cols-3" : "grid-cols-2"
                }`}
              >
                {(isMultipleChoiceCorrect
                  ? RATING_ORDER.filter((rating) => rating !== "again")
                  : RATING_ORDER
                ).map((rating) => (
                  <button
                    key={rating}
                    type="button"
                    onClick={() =>
                      submitAttempt({
                        result: rating === "again" ? "incorrect" : "correct",
                        rating,
                        selectedOptionMeaningId:
                          promptType === "multiple_choice"
                            ? currentCard.meaningId
                            : undefined,
                      })
                    }
                    disabled={isLoading}
                    className="rounded-md border border-neutral-200 bg-neutral-50 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-900 transition-colors hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {RATING_LABELS[rating]}
                  </button>
                ))}
              </div>
            </fieldset>
          </div>
        ) : null}

        {errorMessage ? (
          <p
            role="alert"
            aria-live="polite"
            className="mt-[var(--spacing-md)] rounded-md bg-red-50 p-[var(--spacing-sm)] text-sm text-red-700"
          >
            {errorMessage}
          </p>
        ) : null}
      </div>
    </div>
  );
}

function determinePromptType(
  dueWords: DueWord[],
  currentIndex: number,
): "multiple_choice" | "self_check" {
  const options = buildMultipleChoiceOptions(dueWords, currentIndex);
  // Build a mix of both prompt types when possible: even-indexed cards use
  // multiple-choice if enough distractors exist, otherwise fall back to self-check.
  if (options.length >= 4 && currentIndex % 2 === 0) {
    return "multiple_choice";
  }
  return "self_check";
}

function buildMultipleChoiceOptions(
  dueWords: DueWord[],
  currentIndex: number,
): ReviewOption[] {
  const current = dueWords[currentIndex];
  if (!current) {
    return [];
  }
  const distractors = dueWords
    .filter((_, index) => index !== currentIndex)
    .slice(0, 3)
    .map((dueWord) => ({
      meaningId: dueWord.meaningId,
      label: `${dueWord.partOfSpeech} — ${dueWord.shortDefinition}`,
    }));
  const all = [
    {
      meaningId: current.meaningId,
      label: `${current.partOfSpeech} — ${current.shortDefinition}`,
    },
    ...distractors,
  ];
  return shuffleArray(all);
}

function shuffleArray<T>(items: readonly T[]): T[] {
  const result = [...items];
  for (let index = result.length - 1; index > 0; index--) {
    const swapIndex = Math.floor(Math.random() * (index + 1));
    const temp = result[index]!;
    result[index] = result[swapIndex]!;
    result[swapIndex] = temp;
  }
  return result;
}

function generateClientAttemptId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}
