"use client";

import Link from "next/link";
import { useEffect, useId, useRef, useState } from "react";
import {
  ApiResponseError,
  type PracticeAction,
  type PracticeSession,
} from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { Surface } from "@/ui/surface";

import { PracticeAudio } from "./practice-audio";
import { practiceModes } from "./practice-modes";
import { primaryAction, secondaryAction, textLink } from "./practice-styles";

export function PracticePlayer({
  initialSession,
}: {
  initialSession: PracticeSession;
}) {
  const [session, setSession] = useState(initialSession);
  const [answer, setAnswer] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [needsRetry, setNeedsRetry] = useState(false);
  const [needsRefresh, setNeedsRefresh] = useState(false);
  const inFlight = useRef(false);
  const pending = useRef<PracticeAction | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const feedback = useRef<HTMLDivElement>(null);
  const focusAfterResponse = useRef(false);
  const priorStep = useRef(initialSession.currentStep?.id);
  const answerId = useId();
  const helpId = useId();
  const step = session.currentStep;
  const checked =
    session.feedback?.stepId === step?.id ? session.feedback : null;
  const locked = busy || needsRetry || needsRefresh;

  useEffect(() => {
    if (!focusAfterResponse.current) return;
    focusAfterResponse.current = false;
    if (
      session.currentStep?.id !== priorStep.current ||
      session.status === "completed"
    )
      heading.current?.focus();
    else if (session.feedback) feedback.current?.focus();
    priorStep.current = session.currentStep?.id;
  }, [session]);

  function accept(next: PracticeSession) {
    // A response can acknowledge a replay after another request advanced the
    // session. Only the returned authoritative revision changes the UI.
    if (next.id !== session.id || next.revision < session.revision) return;
    if (next.currentStep?.id !== session.currentStep?.id) setAnswer("");
    focusAfterResponse.current = true;
    setSession(next);
  }

  async function submit(action?: PracticeAction["action"], choiceId?: string) {
    if (inFlight.current || needsRefresh) return;
    if (!pending.current) {
      if (!step || !action) return;
      if (action === "answer" && step.kind === "typed_recall" && !answer.trim())
        return;
      // Capture the exact answer before CSRF/network work. Retry reuses both
      // identities and the expected revision even after a lost response.
      pending.current = {
        stepId: step.id,
        expectedRevision: session.revision,
        clientActionId: crypto.randomUUID(),
        action,
        ...(action === "answer" && step.kind === "typed_recall"
          ? { typedAnswer: answer }
          : {}),
        ...(action === "answer" && choiceId ? { choiceId } : {}),
      };
    }
    const request = pending.current;
    inFlight.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().submitPracticeAction(
        session.id,
        request,
        request.clientActionId,
        { headers: { "X-CSRF-Token": token } },
      );
      accept(data);
      pending.current = null;
      setNeedsRetry(false);
    } catch (cause) {
      if (cause instanceof ApiResponseError && cause.status === 409) {
        setNeedsRefresh(true);
        setNeedsRetry(false);
        setError(
          "This practice changed in another request. Load your saved progress to continue.",
        );
      } else {
        setNeedsRetry(true);
        setError(
          handleApiError(
            cause,
            "We could not confirm your answer. Retry safely or load your saved progress.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  async function refreshProgress() {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    try {
      accept((await createApiClient().getPracticeSession(session.id)).data);
      pending.current = null;
      setNeedsRetry(false);
      setNeedsRefresh(false);
      setError("");
      setNotice("Your saved practice is up to date.");
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "We could not load your saved progress. Please try again.",
        ),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  return (
    <div>
      <Link href="/practice" className={textLink}>
        Back to Practice
      </Link>
      <div className="mb-5 mt-3">
        <p className="text-sm font-semibold text-primary-700">
          {practiceModes[session.mode].title}
        </p>
        <p className="mt-2 text-sm text-neutral-700">
          {session.completedSteps} of {session.totalSteps} questions completed
        </p>
        <progress
          className="mt-2 h-2 w-full accent-primary-700"
          aria-label="Practice progress"
          value={session.completedSteps}
          max={session.totalSteps}
        />
      </div>

      {session.status === "completed" ? (
        <Surface>
          <h1
            ref={heading}
            tabIndex={-1}
            className="text-3xl font-bold tracking-tight text-neutral-900"
          >
            Practice complete
          </h1>
          <p className="mt-3 text-lg text-neutral-900">
            {session.firstAnswersCorrect} of {session.questionsAnswered} correct
            on your first try without help.
          </p>
          <p className="mt-2 text-neutral-700">
            You took time to remember and use these words. Come back for another
            try whenever you like.
          </p>
          <Link href="/practice" className={`${primaryAction} mt-5`}>
            Choose another practice
          </Link>
        </Surface>
      ) : step ? (
        <Surface key={step.id}>
          <h1
            ref={heading}
            tabIndex={-1}
            className="text-2xl font-bold leading-snug text-neutral-900"
          >
            {step.prompt}
          </h1>
          {step.kind === "typed_recall" ? (
            <form
              className="mt-5"
              onSubmit={(event) => {
                event.preventDefault();
                void submit("answer");
              }}
            >
              <label
                htmlFor={answerId}
                className="block font-semibold text-neutral-900"
              >
                Your answer
              </label>
              <input
                id={answerId}
                type="text"
                value={answer}
                onChange={(event) => setAnswer(event.target.value)}
                disabled={locked || Boolean(checked?.correct)}
                maxLength={200}
                autoComplete="off"
                autoCapitalize="none"
                spellCheck={false}
                aria-describedby={helpId}
                className="mt-2 min-h-12 w-full rounded-xl border border-neutral-300 bg-white px-3 py-3 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-60"
              />
              <p id={helpId} className="mt-2 text-sm text-neutral-600">
                Use the word or phrase you learned. Different words can be valid
                English, but this practice checks that particular vocabulary.
              </p>
              {!checked?.correct && (
                <button
                  type="submit"
                  disabled={locked || !answer.trim()}
                  className={`${primaryAction} mt-4`}
                >
                  Check answer
                </button>
              )}
            </form>
          ) : (
            <div className="mt-4">
              {step.speechText ? (
                <PracticeAudio text={step.speechText} />
              ) : (
                <p
                  role="status"
                  className="rounded-xl bg-secondary-50 p-4 text-neutral-800"
                >
                  The audio for this question is unavailable. You can show the
                  answer or return to Practice for another activity.
                </p>
              )}
              <div
                role="group"
                aria-label="Answer choices"
                className="grid gap-3"
              >
                {step.choices.map((choice) => (
                  <button
                    type="button"
                    key={choice.id}
                    disabled={locked || Boolean(checked?.correct)}
                    className={`${secondaryAction} justify-start text-left`}
                    onClick={() => void submit("answer", choice.id)}
                  >
                    {choice.text}
                  </button>
                ))}
              </div>
            </div>
          )}

          {!checked && (
            <div className="mt-5 border-t border-neutral-200 pt-4">
              <button
                type="button"
                disabled={locked}
                onClick={() => void submit("reveal")}
                className={secondaryAction}
              >
                Show answer
              </button>
              <p className="mt-2 text-sm text-neutral-600">
                Showing the answer counts as practice with help.
              </p>
            </div>
          )}
          {checked && (
            <div
              ref={feedback}
              tabIndex={-1}
              role="status"
              className={`mt-5 rounded-xl border p-4 ${checked.correct ? "border-primary-200 bg-primary-50" : "border-secondary-200 bg-secondary-50"}`}
            >
              <h2 className="font-semibold text-neutral-900">
                {checked.correct
                  ? checked.assisted
                    ? "Practised with help"
                    : "You remembered it"
                  : checked.assisted
                    ? "Answer shown"
                    : "Keep practising"}
              </h2>
              <p className="mt-2 text-xl font-bold text-neutral-900">
                {checked.answer}
              </p>
              <p className="mt-2 text-neutral-700">{checked.explanation}</p>
              {!checked.correct && (
                <p className="mt-2 text-sm text-neutral-700">
                  Try again with the answer in view, or continue. Another try
                  here counts as practice with help.
                </p>
              )}
              <Link
                href={`/vocabulary/${encodeURIComponent(checked.wordSlug)}`}
                className={`${textLink} mt-2`}
              >
                Explore {checked.wordText}
              </Link>
            </div>
          )}
          {session.canContinue && (
            <button
              type="button"
              disabled={locked}
              className={`${primaryAction} mt-5 w-full`}
              onClick={() => void submit("continue")}
            >
              {session.completedSteps + 1 === session.totalSteps
                ? "Finish practice"
                : "Continue"}
            </button>
          )}
        </Surface>
      ) : (
        <Surface>
          <h1
            ref={heading}
            tabIndex={-1}
            className="text-2xl font-bold text-neutral-900"
          >
            Load your practice again
          </h1>
          <p className="mt-3 text-neutral-700">
            We could not display the next question.
          </p>
          <button
            type="button"
            disabled={busy}
            className={`${secondaryAction} mt-4`}
            onClick={() => void refreshProgress()}
          >
            Load saved progress
          </button>
        </Surface>
      )}

      {error && (
        <div
          role="alert"
          className="mt-4 rounded-xl border border-secondary-300 bg-secondary-50 p-4 text-neutral-900"
        >
          <p>{error}</p>
          <div className="mt-3 flex flex-wrap gap-3">
            {needsRetry && (
              <button
                type="button"
                disabled={busy}
                onClick={() => void submit()}
                className={secondaryAction}
              >
                Retry answer
              </button>
            )}
            <button
              type="button"
              disabled={busy}
              onClick={() => void refreshProgress()}
              className={secondaryAction}
            >
              Load saved progress
            </button>
          </div>
        </div>
      )}
      {notice && (
        <p role="status" className="mt-4 text-neutral-700">
          {notice}
        </p>
      )}
      {busy && (
        <p role="status" className="mt-3 text-sm text-neutral-600">
          Saving your practice…
        </p>
      )}
    </div>
  );
}
