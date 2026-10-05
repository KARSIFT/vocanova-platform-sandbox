"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";

import { SentenceFeedbackResult } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { CSRF_COOKIE_NAME, getCookieValue } from "@/lib/cookies";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { compareSentences, type SentenceChange } from "./sentence-comparison";

import {
  acceptSentenceEdit,
  countSentenceCharacters,
  getSentenceCharacterCountDescription,
  getSentenceCharacterLimitStatus,
  MAX_SENTENCE_CHARACTERS,
} from "./sentence-feedback-input";
import {
  canUseSentenceFeedbackDraftStorage,
  clearSentenceFeedbackDraft,
  readSentenceFeedbackDraftIntent,
  saveSentenceFeedbackDraft,
} from "./sentence-feedback-drafts";

interface SentenceFeedbackProps {
  targetWord: string;
  attemptId: string;
  source: "word_detail" | "review" | "daily_mission" | "free_practice";
  /** The API-provided stable user id; drafts stay disabled without it. */
  userId?: string;
  shortDefinition?: string;
  onFeedbackSubmitted?: (result: SentenceFeedbackResult) => void;
}

const AI_LIMITATION_COPY =
  "AI feedback can make mistakes. Keep the meaning you intended.";

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
  userId,
  shortDefinition,
  onFeedbackSubmitted,
}: SentenceFeedbackProps) {
  const [sentence, setSentence] = useState("");
  const [result, setResult] = useState<SentenceFeedbackResult | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<{
    text: string;
    source: "local" | "result";
  } | null>(null);
  const [reported, setReported] = useState(false);
  const [reportStatus, setReportStatus] = useState<
    "idle" | "loading" | "error"
  >("idle");
  const [showReportReasons, setShowReportReasons] = useState(false);
  const [submittedSentence, setSubmittedSentence] = useState<string | null>(
    null,
  );
  const [previousFeedback, setPreviousFeedback] = useState<{
    result: SentenceFeedbackResult;
    sentence: string;
  } | null>(null);
  const [canRecoverDraft, setCanRecoverDraft] = useState(false);
  const pendingSubmission = useRef<{
    idempotencyKey: string;
    sentenceText: string;
  } | null>(null);
  const submittingSynchronously = useRef(false);
  // Reports belong to the displayed feedback, not a later sentence result.
  const feedbackGeneration = useRef(0);

  useEffect(() => {
    feedbackGeneration.current += 1;
    setCanRecoverDraft(canUseSentenceFeedbackDraftStorage(userId));
    const draft = readSentenceFeedbackDraftIntent({
      userId,
      source,
      attemptId,
    });
    pendingSubmission.current = draft?.idempotencyKey
      ? { idempotencyKey: draft.idempotencyKey, sentenceText: draft.sentence }
      : null;
    setSentence(draft?.sentence ?? "");
    setResult(null);
    setSubmittedSentence(null);
    setPreviousFeedback(null);
    setErrorMessage(null);
    setReported(false);
    setReportStatus("idle");
    setShowReportReasons(false);
    return () => {
      feedbackGeneration.current += 1;
    };
  }, [attemptId, source, userId]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (submittingSynchronously.current) {
      return;
    }

    // Capture and persist the exact intent before session recovery can fail.
    // A recovered cookie or same-user return must retry the same submission.
    const pending = pendingSubmission.current ?? {
      idempotencyKey: generateIdempotencyKey(),
      sentenceText: sentence,
    };
    pendingSubmission.current = pending;
    saveSentenceFeedbackDraft({
      userId,
      source,
      attemptId,
      sentence: pending.sentenceText,
      idempotencyKey: pending.idempotencyKey,
    });
    submittingSynchronously.current = true;
    feedbackGeneration.current += 1;
    setIsLoading(true);
    setErrorMessage(null);
    setReported(false);
    setReportStatus("idle");
    setShowReportReasons(false);

    const client = createApiClient();
    let submissionStarted = false;
    try {
      const csrfToken = await getOrRefreshCSRFToken();
      if (!csrfToken) throw new Error("Session is not ready");
      submissionStarted = true;
      const { data } = await client.submitSentenceFeedback(
        { sentenceText: pending.sentenceText, source, attemptId },
        pending.idempotencyKey,
        { headers: { "X-CSRF-Token": csrfToken } },
      );
      setResult(data);
      setSubmittedSentence(data.originalSentence || pending.sentenceText);
      // A deduplicated response may represent feedback that was already
      // reported in an earlier submission. The backend is authoritative for
      // that persisted state, so do not make the learner report it again.
      setReported(data.reported);
      setShowReportReasons(false);
      if (data.errorCode) {
        // The API made a definite rejection. Keep the learner's words, but
        // discard the prior request key so a corrected sentence is a new
        // submission rather than a replay of that rejection.
        pendingSubmission.current = null;
        saveSentenceFeedbackDraft({
          userId,
          source,
          attemptId,
          sentence: data.originalSentence || pending.sentenceText,
        });
        setErrorMessage({
          text: data.errorMessage || getDefaultErrorMessage(data, targetWord),
          source: "result",
        });
      } else if (data.processingStatus === "completed") {
        pendingSubmission.current = null;
        setErrorMessage(null);
        clearSentenceFeedbackDraft({ userId, source, attemptId });
      }
      onFeedbackSubmitted?.(data);
    } catch (error) {
      // Recovery did not check a new sentence. Keep earlier feedback and
      // crisis guidance visible alongside the current preparation error.
      if (submissionStarted) setResult(null);
      // The controlled textarea and its user-scoped tab draft preserve an
      // unresolved submission through recoverable failures and re-auth.
      setErrorMessage({
        text: handleApiError(
          error,
          "Unable to check this sentence right now. Please try again.",
        ),
        source: "local",
      });
    } finally {
      setIsLoading(false);
      submittingSynchronously.current = false;
    }
  }

  async function handleReport(reason: (typeof REPORT_REASONS)[number][0]) {
    if (submittingSynchronously.current || !result?.attemptId) {
      return;
    }
    const generation = feedbackGeneration.current;

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
      if (feedbackGeneration.current !== generation) {
        return;
      }
      setReported(true);
      setShowReportReasons(false);
      setReportStatus("idle");
    } catch (error) {
      if (feedbackGeneration.current !== generation) {
        return;
      }
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
  const hasSuccessResult =
    hasResult && result.processingStatus === "completed" && !result.errorCode;
  const statusLabel = result ? getStatusLabel(result.status) : null;
  const characterCount = countSentenceCharacters(sentence);
  const characterCountId = `sentence-character-count-${attemptId}`;
  const characterLimitMessageId = `sentence-character-limit-${attemptId}`;
  const characterLimitStatus = getSentenceCharacterLimitStatus(sentence);
  const correctedSentence = result?.correctedSentence;
  const comparison = useMemo(
    () =>
      submittedSentence && correctedSentence && hasSuccessResult
        ? compareSentences(submittedSentence, correctedSentence)
        : null,
    [submittedSentence, correctedSentence, hasSuccessResult],
  );

  function handleTryAnotherSentence() {
    if (submittingSynchronously.current) {
      return;
    }
    feedbackGeneration.current += 1;

    pendingSubmission.current = null;
    setResult(null);
    setSubmittedSentence(null);
    setPreviousFeedback(null);
    setSentence("");
    setErrorMessage(null);
    clearSentenceFeedbackDraft({ userId, source, attemptId });
    document.getElementById(`sentence-input-${attemptId}`)?.focus();
  }

  function handleDiscardDraft() {
    if (submittingSynchronously.current) {
      return;
    }

    pendingSubmission.current = null;
    setSentence("");
    setErrorMessage(null);
    clearSentenceFeedbackDraft({ userId, source, attemptId });
  }

  function handleReviseSentence() {
    if (submittingSynchronously.current) {
      return;
    }
    feedbackGeneration.current += 1;

    if (result && submittedSentence) {
      setPreviousFeedback({ result, sentence: submittedSentence });
      // Keep the visible draft, including edits made after this feedback.
      // The earlier result still belongs to submittedSentence. Discard the
      // completed request key so the revision receives a fresh identity.
      saveSentenceFeedbackDraft({
        userId,
        source,
        attemptId,
        sentence,
      });
    }
    pendingSubmission.current = null;
    setResult(null);
    setErrorMessage(null);
    document.getElementById(`sentence-input-${attemptId}`)?.focus();
  }

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
      </h3>
      {shortDefinition ? (
        <p className="mt-1 text-sm text-neutral-700">{shortDefinition}</p>
      ) : null}
      <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
        Write one sentence using{" "}
        <span className="font-semibold text-neutral-900">{targetWord}</span>.
      </p>
      <p className="mt-[var(--spacing-xs)] text-xs text-neutral-600">
        Avoid personal details such as contact information or passwords.
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
            aria-describedby={`${characterCountId}${
              characterLimitStatus ? ` ${characterLimitMessageId}` : ""
            }`}
            value={sentence}
            onChange={(event) => {
              const next = acceptSentenceEdit(sentence, event.target.value);
              if (next !== sentence) {
                setErrorMessage(null);
                pendingSubmission.current = null;
                saveSentenceFeedbackDraft({
                  userId,
                  source,
                  attemptId,
                  sentence: next,
                });
              }
              setSentence(next);
            }}
            disabled={isLoading}
            rows={3}
            placeholder={`Type a sentence using "${targetWord}"...`}
            className="w-full rounded-md border border-neutral-300 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base text-neutral-900 placeholder:text-neutral-500 focus:border-primary-500 focus:outline focus:outline-2 focus:outline-primary-500/20 disabled:cursor-not-allowed disabled:opacity-60"
          />
          <p
            id={characterCountId}
            aria-label={getSentenceCharacterCountDescription(sentence)}
            className="mt-[var(--spacing-xs)] text-right text-sm text-neutral-600"
          >
            {characterCount}/{MAX_SENTENCE_CHARACTERS}
          </p>
          {characterLimitStatus ? (
            <p id={characterLimitMessageId} role="status" className="sr-only">
              {characterLimitStatus}
            </p>
          ) : null}
        </div>

        <button
          type="submit"
          disabled={isLoading || sentence.trim().length === 0}
          aria-busy={isLoading}
          className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-medium text-neutral-50 transition-colors duration-[var(--duration-fast)] ease-[var(--ease-out)] hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isLoading ? "Checking..." : "Check my sentence"}
        </button>
        {canRecoverDraft ? (
          <div className="flex flex-wrap items-center gap-[var(--spacing-sm)] rounded-md bg-neutral-50 px-[var(--spacing-sm)] py-[var(--spacing-xs)] text-sm text-neutral-700">
            <p>Draft stays in this tab for up to two hours.</p>
            {!hasSuccessResult && !isLoading && sentence.trim() ? (
              <button
                type="button"
                onClick={handleDiscardDraft}
                className="inline-flex min-h-11 items-center px-[var(--spacing-xs)] font-semibold text-primary-700 underline hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
              >
                Discard draft
              </button>
            ) : null}
          </div>
        ) : null}
      </form>

      {errorMessage?.source === "local" ? (
        <p
          role="alert"
          aria-live="polite"
          className="mt-[var(--spacing-md)] rounded-md bg-red-50 p-[var(--spacing-sm)] text-base text-red-700"
        >
          {errorMessage.text}
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
              <p className="text-sm font-semibold">{statusLabel}</p>
              {result.headline && result.headline !== statusLabel ? (
                <p className="font-semibold">{result.headline}</p>
              ) : null}
              {result.explanation ? (
                <p className="mt-[var(--spacing-xs)] text-base">
                  {result.explanation}
                </p>
              ) : null}
            </div>
          ) : null}

          {submittedSentence ? (
            <section
              aria-label="Sentence comparison"
              className="rounded-xl border border-neutral-200 p-4"
            >
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="min-w-0 rounded-lg bg-neutral-50 p-3">
                  <p className="text-sm font-medium text-neutral-700">
                    {hasSuccessResult ? "Sentence checked" : "Your sentence"}
                  </p>
                  <p className="mt-2 whitespace-pre-wrap wrap-anywhere text-base leading-relaxed text-neutral-900">
                    {comparison?.hasChanges ? (
                      <ChangedSentence parts={comparison.original} />
                    ) : (
                      submittedSentence
                    )}
                  </p>
                </div>
                {comparison?.hasChanges ? (
                  <div className="min-w-0 rounded-lg bg-primary-50 p-3">
                    <p className="text-sm font-medium text-neutral-700">
                      Corrected sentence
                    </p>
                    <p className="mt-2 whitespace-pre-wrap wrap-anywhere text-base leading-relaxed text-neutral-900">
                      <ChangedSentence parts={comparison.suggested} />
                    </p>
                  </div>
                ) : null}
              </div>
              {comparison ? (
                <p className="mt-3 text-sm text-neutral-600">
                  {comparison.hasChanges
                    ? "Highlighted text shows what changed."
                    : "No wording changes suggested."}
                </p>
              ) : null}
              {hasSuccessResult ? (
                <button
                  type="button"
                  onClick={handleReviseSentence}
                  disabled={isLoading}
                  className="mt-3 inline-flex min-h-11 items-center rounded-lg bg-primary-600 px-4 py-3 font-semibold text-white hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Revise sentence
                </button>
              ) : null}
            </section>
          ) : null}

          {result.errorCode &&
          !result.crisisResourceMessage &&
          errorMessage?.source === "result" ? (
            <div
              role="alert"
              aria-live="polite"
              className="rounded-md bg-red-50 p-[var(--spacing-md)] text-base text-red-700"
            >
              {errorMessage.text}
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
              <div className="flex flex-wrap gap-[var(--spacing-sm)]">
                <button
                  type="button"
                  onClick={handleTryAnotherSentence}
                  disabled={isLoading}
                  className="inline-flex min-h-11 items-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-neutral-900 hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Try another sentence
                </button>
                <Link
                  href="/progress/sentences"
                  className="inline-flex min-h-11 items-center px-[var(--spacing-sm)] py-[var(--spacing-sm)] text-base font-semibold text-primary-700 underline hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
                >
                  View sentence history
                </Link>
              </div>
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
                        disabled={isLoading || reportStatus === "loading"}
                        className="flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center text-left text-sm text-neutral-600 underline transition-colors hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {label}
                      </button>
                    ))}
                  </fieldset>
                ) : (
                  <button
                    type="button"
                    onClick={() => {
                      if (!submittingSynchronously.current) {
                        setShowReportReasons(true);
                      }
                    }}
                    disabled={isLoading || reportStatus === "loading"}
                    aria-busy={reportStatus === "loading"}
                    className="inline-flex min-h-[var(--spacing-2xl)] min-w-[var(--spacing-2xl)] items-center justify-center text-sm font-medium text-neutral-600 underline transition-colors hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {reportStatus === "loading"
                      ? "Reporting..."
                      : "Report a problem"}
                  </button>
                )}
                <span role="alert" className="text-sm text-red-700">
                  {reportStatus === "error"
                    ? "Unable to report. Try again."
                    : ""}
                </span>
              </div>
            </div>
          ) : null}

          {result.missionCompleted || source === "daily_mission" ? (
            <p className="text-sm text-neutral-600">
              Mission completed: {result.missionCompleted ? "Yes" : "Not yet"}
            </p>
          ) : null}
        </div>
      ) : null}

      {previousFeedback ? (
        <aside
          aria-label="Previous feedback"
          className="mt-[var(--spacing-md)] rounded-md border border-secondary-200 bg-secondary-50 p-[var(--spacing-md)] text-secondary-900"
        >
          <p className="text-sm font-semibold">Previous feedback</p>
          <p className="mt-[var(--spacing-xs)] wrap-anywhere text-base">
            {previousFeedback.sentence}
          </p>
          {previousFeedback.result.correctedSentence ? (
            <p className="mt-[var(--spacing-xs)] wrap-anywhere text-sm">
              Suggested revision: {previousFeedback.result.correctedSentence}
            </p>
          ) : null}
          {previousFeedback.result.explanation ? (
            <p className="mt-[var(--spacing-xs)] text-sm">
              {previousFeedback.result.explanation}
            </p>
          ) : null}
        </aside>
      ) : null}
    </section>
  );
}

function ChangedSentence({ parts }: { parts: SentenceChange[] }) {
  return parts.map((part, index) =>
    part.changed ? (
      <mark
        key={index}
        className="rounded-sm bg-secondary-100 text-neutral-900 underline decoration-secondary-600 decoration-2 underline-offset-4"
      >
        {part.text}
      </mark>
    ) : (
      <span key={index}>{part.text}</span>
    ),
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
      return `Add "${targetWord}" to your sentence, then try again.`;
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
