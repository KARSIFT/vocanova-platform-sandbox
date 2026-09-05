"use client";

import { useState } from "react";

import { SentenceFeedbackResult } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { handleApiError } from "@/lib/session";

import {
  acceptSentenceEdit,
  countSentenceCharacters,
  MAX_SENTENCE_CHARACTERS,
} from "./sentence-feedback-input";

interface SentenceFeedbackProps {
  targetWord: string;
  attemptId: string;
  source: "word_detail" | "review" | "daily_mission" | "free_practice";
  shortDefinition?: string;
  onFeedbackSubmitted?: (result: SentenceFeedbackResult) => void;
}

const AI_LIMITATION_COPY =
  "AI feedback can make mistakes. Use your own judgment and your teacher's guidance when learning.";

const RETRY_MESSAGE =
  "Vocanova could not check this sentence right now. Your sentence is still here, so you can try again.";

const REPORT_REASONS = [
  ["already_correct", "Already correct"],
  ["correction_changed_meaning", "Correction changed my meaning"],
  ["explanation_unclear", "Explanation was unclear"],
  ["inappropriate", "Inappropriate"],
  ["something_else", "Something else"],
] as const;

export function SentenceFeedback({
  targetWord,
  attemptId,
  source,
  shortDefinition,
  onFeedbackSubmitted,
}: SentenceFeedbackProps) {
  const [sentence, setSentence] = useState("");
  const [result, setResult] = useState<SentenceFeedbackResult | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [reported, setReported] = useState(false);
  const [reportStatus, setReportStatus] = useState<
    "idle" | "loading" | "error"
  >("idle");
  const [showReportReasons, setShowReportReasons] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setErrorMessage("Session is not ready. Please refresh the page.");
      return;
    }

    setIsLoading(true);
    setErrorMessage(null);
    setReported(false);
    setReportStatus("idle");

    const client = createApiClient();
    try {
      const { data } = await client.submitSentenceFeedback(
        { sentenceText: sentence, source, attemptId },
        generateIdempotencyKey(),
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      setResult(data);
      if (data.errorCode) {
        setErrorMessage(
          data.errorMessage || getDefaultErrorMessage(data, targetWord),
        );
      } else {
        setErrorMessage(null);
      }
      onFeedbackSubmitted?.(data);
    } catch (error) {
      setResult(null);
      // T06: a 401 here means the session expired mid-sentence-submission.
      // Never lose the learner's sentence — the textarea stays populated
      // (controlled by component state) and we route to re-auth. The
      // learner can copy their text and resume after sign-in.
      setErrorMessage(
        handleApiError(
          error,
          "Unable to check this sentence right now. Please try again.",
        ),
      );
    } finally {
      setIsLoading(false);
    }
  }

  async function handleReport(reason: (typeof REPORT_REASONS)[number][0]) {
    if (!result?.attemptId) {
      return;
    }

    const csrfToken = getCookieValue(CSRF_COOKIE_NAME);
    if (!csrfToken) {
      setReportStatus("error");
      return;
    }

    setReportStatus("loading");

    const client = createApiClient();
    try {
      await client.reportSentenceFeedback(
        result.attemptId,
        { reason },
        generateIdempotencyKey(),
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      setReported(true);
      setShowReportReasons(false);
      setReportStatus("idle");
    } catch (error) {
      // T06: a 401 on a report submission routes the learner to
      // re-auth. The text is in component state and the feedback
      // result is still visible — nothing is lost.
      if (
        error &&
        typeof error === "object" &&
        "status" in error &&
        (error as { status: number }).status === 401
      ) {
        handleApiError(error, "Unable to report. Please try again.");
        return;
      }
      setReportStatus("error");
    }
  }

  const hasResult = result !== null;
  const hasSuccessResult = hasResult && !result.errorCode;
  const statusLabel = result ? getStatusLabel(result.status) : null;

  return (
    <section
      aria-labelledby={`sentence-feedback-heading-${attemptId}`}
      className="mt-[var(--spacing-md)] rounded-md border border-neutral-200 bg-white p-[var(--spacing-md)] shadow-sm"
    >
      <h3
        id={`sentence-feedback-heading-${attemptId}`}
        className="text-base font-semibold text-neutral-900"
      >
        Practice with {targetWord}
        {shortDefinition ? ` — ${shortDefinition}` : null}
      </h3>
      <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
        Write a sentence using the word{" "}
        <span className="font-medium text-neutral-900">{targetWord}</span>.
      </p>

      <form
        onSubmit={handleSubmit}
        className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]"
      >
        <div>
          <label
            htmlFor={`sentence-input-${attemptId}`}
            className="sr-only"
          >{`Write a sentence using ${targetWord}`}</label>
          <textarea
            id={`sentence-input-${attemptId}`}
            name="sentence"
            value={sentence}
            onChange={(event) =>
              setSentence((previous) =>
                acceptSentenceEdit(previous, event.target.value),
              )
            }
            disabled={isLoading}
            rows={3}
            placeholder={`Type a sentence using "${targetWord}"...`}
            className="w-full rounded-md border border-neutral-300 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base text-neutral-900 placeholder:text-neutral-500 focus:border-primary-500 focus:outline focus:outline-2 focus:outline-primary-500/20 disabled:cursor-not-allowed disabled:opacity-60"
          />
          <p className="mt-[var(--spacing-xs)] text-right text-sm text-neutral-600">
            {countSentenceCharacters(sentence)}/{MAX_SENTENCE_CHARACTERS}
          </p>
        </div>

        <button
          type="submit"
          disabled={isLoading || sentence.trim().length === 0}
          aria-busy={isLoading}
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isLoading ? "Checking..." : "Check my sentence"}
        </button>
      </form>

      {errorMessage && !hasResult ? (
        <p
          role="alert"
          aria-live="polite"
          className="mt-[var(--spacing-md)] rounded-md bg-red-50 p-[var(--spacing-sm)] text-base text-red-700"
        >
          {errorMessage}
        </p>
      ) : null}

      {hasResult ? (
        <div className="mt-[var(--spacing-md)] space-y-[var(--spacing-md)]">
          {statusLabel ? (
            <div
              role="status"
              aria-label={`Feedback result: ${statusLabel}`}
              className={`rounded-md p-[var(--spacing-md)] ${getStatusClasses(result.status)}`}
            >
              <p className="font-semibold">{statusLabel}</p>
              {result.explanation ? (
                <p className="mt-[var(--spacing-xs)] text-base">
                  {result.explanation}
                </p>
              ) : null}
            </div>
          ) : null}

          {result.errorCode && !result.crisisResourceMessage ? (
            <div
              role="alert"
              aria-live="polite"
              className="rounded-md bg-red-50 p-[var(--spacing-md)] text-base text-red-700"
            >
              {errorMessage}
            </div>
          ) : null}

          {result.crisisResourceMessage ? (
            <div
              role="alert"
              aria-live="assertive"
              className="rounded-md bg-amber-50 p-[var(--spacing-md)] text-base text-amber-900"
            >
              <p className="font-semibold">We are here to help</p>
              <p className="mt-[var(--spacing-xs)]">
                {result.crisisResourceMessage}
              </p>
            </div>
          ) : null}

          {result.correctedSentence ? (
            <div className="rounded-md bg-neutral-50 p-[var(--spacing-md)]">
              <p className="text-sm font-medium text-neutral-700">
                Corrected sentence
              </p>
              <p className="mt-[var(--spacing-xs)] text-base text-neutral-900">
                {result.correctedSentence}
              </p>
            </div>
          ) : null}

          {result.improvementTip ? (
            <div className="rounded-md bg-neutral-50 p-[var(--spacing-md)]">
              <p className="text-sm font-medium text-neutral-700">Tip</p>
              <p className="mt-[var(--spacing-xs)] text-base text-neutral-900">
                {result.improvementTip}
              </p>
            </div>
          ) : null}

          {hasSuccessResult ? (
            <div className="rounded-md border border-neutral-200 p-[var(--spacing-md)]">
              <p className="text-sm text-neutral-600">{AI_LIMITATION_COPY}</p>
              <div className="mt-[var(--spacing-sm)] flex items-center gap-[var(--spacing-md)]">
                {reported ? (
                  <span className="text-sm text-neutral-600">Reported</span>
                ) : showReportReasons ? (
                  <fieldset className="space-y-[var(--spacing-xs)]">
                    <legend className="text-sm font-medium text-neutral-800">
                      What was the problem?
                    </legend>
                    {REPORT_REASONS.map(([reason, label]) => (
                      <button
                        key={reason}
                        type="button"
                        onClick={() => handleReport(reason)}
                        disabled={reportStatus === "loading"}
                        className="block text-left text-sm text-neutral-600 underline transition-colors hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {label}
                      </button>
                    ))}
                  </fieldset>
                ) : (
                  <button
                    type="button"
                    onClick={() => setShowReportReasons(true)}
                    disabled={reportStatus === "loading"}
                    aria-busy={reportStatus === "loading"}
                    className="text-sm font-medium text-neutral-600 underline transition-colors hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {reportStatus === "loading"
                      ? "Reporting..."
                      : "Report a problem"}
                  </button>
                )}
                {reportStatus === "error" ? (
                  <span className="text-sm text-red-700">
                    Unable to report. Try again.
                  </span>
                ) : null}
              </div>
            </div>
          ) : null}

          <p className="text-sm text-neutral-600">
            Mission completed: {result.missionCompleted ? "Yes" : "Not yet"}
          </p>
        </div>
      ) : null}
    </section>
  );
}

function getStatusLabel(status?: string): string | null {
  switch (status) {
    case "correct":
      return "Correct";
    case "needs_improvement":
      return "Needs improvement";
    case "incorrect":
      return "Incorrect";
    default:
      return null;
  }
}

function getStatusClasses(status?: string): string {
  switch (status) {
    case "correct":
      return "bg-green-50 text-green-900";
    case "needs_improvement":
      return "bg-yellow-50 text-yellow-900";
    case "incorrect":
      return "bg-red-50 text-red-900";
    default:
      return "bg-neutral-50 text-neutral-900";
  }
}

function getDefaultErrorMessage(
  result: SentenceFeedbackResult,
  targetWord: string,
): string {
  switch (result.errorCode) {
    case "too_short":
      return "Your sentence is too short. Write at least 3 words.";
    case "too_long":
      return "Your sentence is too long. Keep it under 300 characters.";
    case "missing_target":
      return `Your sentence is missing the target word "${targetWord}".`;
    case "unsupported_language":
      return "Please write your sentence in English.";
    case "invalid_input":
      return "Please check your sentence and try again.";
    case "attempt_not_eligible":
      return "This practice is not available for the selected item.";
    case "AI_FEEDBACK_RATE_LIMITED":
      return "You have reached the limit for now. Try again later.";
    case "SAFETY_BLOCKED":
      return "This sentence cannot be checked. Please try a different sentence.";
    case "SAFETY_SELF_HARM":
      return "We are here to help.";
    case "SAFETY_MODERATION_UNAVAILABLE":
      return "We cannot check this right now. Please try again later.";
    case "AI_FEEDBACK_TEMPORARY_FAILURE":
    default:
      return RETRY_MESSAGE;
  }
}

function generateIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}
