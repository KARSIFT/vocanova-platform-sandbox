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
import { MeaningPicture } from "@/ui/meaning-picture";
import { SessionActions } from "@/ui/session";

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
  const [selectedChoice, setSelectedChoice] = useState("");
  const [tokenIds, setTokenIds] = useState<string[]>(() =>
    initialSession.feedback?.stepId === initialSession.currentStep?.id
      ? (initialSession.feedback?.selectedTokenIds ?? [])
      : [],
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [needsRetry, setNeedsRetry] = useState(false);
  const [needsRefresh, setNeedsRefresh] = useState(false);
  const inFlight = useRef(false);
  const pending = useRef<PracticeAction | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const feedback = useRef<HTMLDivElement>(null);
  const errorRegion = useRef<HTMLDivElement>(null);
  const focusAfterResponse = useRef(false);
  const priorStep = useRef(initialSession.currentStep?.id);
  const answerId = useId();
  const answerFormId = useId();
  const helpId = useId();
  const step = session.currentStep;
  const checked =
    session.feedback?.stepId === step?.id ? session.feedback : null;
  const displayedChoice = checked?.correct
    ? (checked.correctChoiceId ?? selectedChoice)
    : selectedChoice;
  const locked = busy || needsRetry || needsRefresh;
  const journey = session.mode === "journey";
  const teaching = step?.kind === "teach";
  const assembly = step?.kind === "sentence_build";
  const shownTokens = checked?.correct
    ? (checked.selectedTokenIds ?? tokenIds)
    : tokenIds;
  const answerLocked =
    locked || Boolean(checked?.correct || (journey && session.canContinue));

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

  useEffect(() => {
    if (error) errorRegion.current?.focus();
  }, [error]);

  function accept(next: PracticeSession) {
    // A response can acknowledge a replay after another request advanced the
    // session. Only the returned authoritative revision changes the UI.
    if (next.id !== session.id || next.revision < session.revision) return;
    if (next.currentStep?.id !== session.currentStep?.id) {
      setAnswer("");
      setSelectedChoice("");
      setTokenIds([]);
    } else if (next.feedback && !next.feedback.correct) {
      setSelectedChoice("");
    }
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
        ...(action === "answer" && step.kind === "sentence_build"
          ? { tokenIds: [...tokenIds] }
          : {}),
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

  function moveToken(index: number, direction: -1 | 1) {
    if (answerLocked) return;
    const destination = index + direction;
    if (destination < 0 || destination >= tokenIds.length) return;
    setTokenIds((previous) => {
      const next = [...previous];
      [next[index], next[destination]] = [next[destination]!, next[index]!];
      return next;
    });
  }

  return (
    <div className="mx-auto max-w-[42rem] pb-[9rem]">
      <div className="mb-5 mt-3">
        <p className="text-sm font-semibold text-primary-700">
          {practiceModes[session.mode].title}
        </p>
        <p className="mt-2 text-sm text-neutral-700">
          {session.completedSteps} of {session.totalSteps} steps completed
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
            {journey
              ? "Your next step is ready on your learning path."
              : "Come back for another try whenever you like."}
          </p>
          <Link href="/practice" className={`${primaryAction} mt-5`}>
            {journey ? "Back to your path" : "Choose another practice"}
          </Link>
        </Surface>
      ) : step ? (
        <Surface key={step.id}>
          <h1
            ref={heading}
            tabIndex={-1}
            className="text-2xl font-bold leading-snug text-neutral-900"
          >
            {teaching ? step.wordText : step.prompt}
          </h1>
          {step.partOfSpeech && (
            <p className="mt-2 text-sm text-neutral-600">{step.partOfSpeech}</p>
          )}
          {teaching ? (
            <div className="mt-5">
              <p className="text-lg text-neutral-700">{step.prompt}</p>
              {step.meaningId && (
                <MeaningPicture
                  meaningId={step.meaningId}
                  image={step.image}
                  meaningRevision={step.pronunciationRevision}
                  resolveMissing
                />
              )}
              {step.context && (
                <p className="mt-4 text-lg leading-relaxed text-neutral-700">
                  {step.context}
                </p>
              )}
              {step.speechText && (
                <PracticeAudio
                  text={step.speechText}
                  meaningId={step.pronunciationMeaningId}
                  meaningRevision={step.pronunciationRevision}
                />
              )}
              <p className="mt-4 text-sm text-neutral-600">
                A new word for your practice path.
              </p>
            </div>
          ) : assembly ? (
            <form
              id={answerFormId}
              className="mt-5"
              onSubmit={(event) => {
                event.preventDefault();
                if (tokenIds.length) void submit("answer");
              }}
            >
              {step.speechText && (
                <PracticeAudio
                  text={step.speechText}
                  meaningId={step.pronunciationMeaningId}
                  meaningRevision={step.pronunciationRevision}
                />
              )}
              <p id={helpId} className="text-sm text-neutral-600">
                Tap words in order. Tap a chosen word to remove it; use arrow
                keys to move it.
              </p>
              <div
                role="group"
                aria-label="Your sentence"
                aria-describedby={helpId}
                className="mt-3 flex min-h-24 flex-wrap content-start gap-2 rounded-xl border-2 border-neutral-300 p-3"
              >
                {!shownTokens.length && (
                  <span className="py-2 text-neutral-500">
                    Build your sentence here
                  </span>
                )}
                {shownTokens.map((id, index) => {
                  const token = step.choices.find((choice) => choice.id === id);
                  return (
                    <button
                      key={id}
                      type="button"
                      disabled={answerLocked}
                      className={secondaryAction}
                      aria-label={`${token?.text ?? "Word"}, position ${index + 1}. Remove word; arrow keys move it.`}
                      onClick={() =>
                        setTokenIds((previous) =>
                          previous.filter((item) => item !== id),
                        )
                      }
                      onKeyDown={(event) => {
                        if (
                          event.key === "ArrowLeft" ||
                          event.key === "ArrowRight"
                        ) {
                          event.preventDefault();
                          moveToken(index, event.key === "ArrowLeft" ? -1 : 1);
                        }
                      }}
                    >
                      {token?.text}
                    </button>
                  );
                })}
              </div>
              <div
                className="mt-4 flex flex-wrap gap-2"
                role="group"
                aria-label="Available words"
              >
                {step.choices.map((token) => (
                  <button
                    key={token.id}
                    type="button"
                    disabled={answerLocked || shownTokens.includes(token.id)}
                    onClick={() =>
                      setTokenIds((previous) => [...previous, token.id])
                    }
                    className={`${secondaryAction} disabled:opacity-40`}
                  >
                    {token.text}
                  </button>
                ))}
              </div>
              <div className="mt-3 flex gap-4">
                <button
                  type="button"
                  disabled={answerLocked || !tokenIds.length}
                  onClick={() =>
                    setTokenIds((previous) => previous.slice(0, -1))
                  }
                  className={textLink}
                >
                  Undo last word
                </button>
                <button
                  type="button"
                  disabled={answerLocked || !tokenIds.length}
                  onClick={() => setTokenIds([])}
                  className={textLink}
                >
                  Start over
                </button>
              </div>
            </form>
          ) : step.kind === "typed_recall" ? (
            <form
              id={answerFormId}
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
                disabled={answerLocked}
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
            </form>
          ) : (
            <div className="mt-4">
              {step.wordText && (
                <p className="mb-3 text-3xl font-bold text-neutral-900">
                  {step.wordText}
                </p>
              )}
              {step.context && (
                <p className="mb-4 text-lg leading-relaxed text-neutral-700">
                  {step.context}
                </p>
              )}
              {step.speechText ? (
                <PracticeAudio
                  text={step.speechText}
                  meaningId={step.pronunciationMeaningId}
                  meaningRevision={step.pronunciationRevision}
                />
              ) : !journey ? (
                <p
                  role="status"
                  className="rounded-xl bg-secondary-50 p-4 text-neutral-800"
                >
                  The audio for this question is unavailable. You can show the
                  answer or return to Practice for another activity.
                </p>
              ) : null}
              <form
                id={answerFormId}
                onSubmit={(event) => {
                  event.preventDefault();
                  if (selectedChoice) void submit("answer", selectedChoice);
                }}
              >
                <fieldset disabled={answerLocked} className="grid gap-3">
                  <legend className="sr-only">Answer choices</legend>
                  {step.choices.map((choice) => (
                    <label
                      key={choice.id}
                      className={`flex min-h-12 cursor-pointer items-center gap-3 rounded-xl border p-4 text-neutral-900 focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700 ${displayedChoice === choice.id ? "border-primary-600 bg-primary-50" : "border-neutral-300 bg-white"} ${locked || checked?.correct ? "cursor-default" : ""}`}
                    >
                      <input
                        type="radio"
                        name={answerFormId}
                        aria-label={choice.text}
                        value={choice.id}
                        checked={displayedChoice === choice.id}
                        onChange={() => setSelectedChoice(choice.id)}
                        className="h-5 w-5 shrink-0 accent-primary-700"
                      />
                      <span className="grow">{choice.text}</span>
                      {displayedChoice === choice.id && (
                        <span
                          aria-hidden="true"
                          className="text-sm font-semibold text-primary-800"
                        >
                          Selected
                        </span>
                      )}
                    </label>
                  ))}
                </fieldset>
                <p className="mt-3 text-sm text-neutral-600">
                  Choose an answer, then check it.
                </p>
              </form>
            </div>
          )}

          {!checked && !journey && (
            <p className="mt-5 text-sm text-neutral-600">
              Showing the answer counts as practice with help.
            </p>
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
                  : checked.assisted && (!journey || session.canContinue)
                    ? "Answer shown"
                    : "Keep practising"}
              </h2>
              <p className="mt-2 text-xl font-bold text-neutral-900">
                {checked.answer}
              </p>
              <p className="mt-2 text-neutral-700">{checked.explanation}</p>
              {!checked.correct && !(journey && session.canContinue) && (
                <p className="mt-2 text-sm text-neutral-700">
                  {journey
                    ? "Try again, or continue with help. Another try counts as practice with help."
                    : "Try again with the answer in view, or continue. Another try here counts as practice with help."}
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
          <SessionActions>
            {(!checked || (journey && !session.canContinue)) && !teaching && (
              <button
                type="button"
                disabled={locked}
                onClick={() => void submit("reveal")}
                className={secondaryAction}
              >
                {checked && journey
                  ? "Continue with help"
                  : journey && (step.kind === "listening_choice" || assembly)
                    ? "Can’t listen now"
                    : "Show answer"}
              </button>
            )}
            {!teaching &&
              !checked?.correct &&
              !(journey && session.canContinue) && (
                <button
                  type="submit"
                  form={answerFormId}
                  disabled={
                    locked ||
                    (step.kind === "typed_recall"
                      ? !answer.trim()
                      : assembly
                        ? !tokenIds.length
                        : !selectedChoice)
                  }
                  className={`${session.canContinue ? secondaryAction : primaryAction} grow`}
                >
                  {busy ? "Checking…" : "Check answer"}
                </button>
              )}
            {session.canContinue && (
              <button
                type="button"
                disabled={locked}
                className={`${primaryAction} grow`}
                onClick={() => void submit("continue")}
              >
                {session.completedSteps + 1 === session.totalSteps
                  ? "Finish practice"
                  : "Continue"}
              </button>
            )}
          </SessionActions>
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
          ref={errorRegion}
          tabIndex={-1}
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
