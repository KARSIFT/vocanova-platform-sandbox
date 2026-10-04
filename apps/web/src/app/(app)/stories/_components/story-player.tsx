"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import {
  ApiResponseError,
  type StoryAction,
  type StorySession,
} from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { Surface } from "@/ui/surface";
import {
  primaryAction,
  secondaryAction,
  textLink,
} from "../../practice/_components/practice-styles";
import { StoryAudio } from "./story-audio";
import { StoryVocabularyHelp } from "./story-vocabulary";
import { StoryStart } from "./story-start";
import { storySituationSlug } from "../../discover/[situation]/_components/unit-guide-content";

export function StoryPlayer({
  initialSession,
}: {
  initialSession: StorySession;
}) {
  const [session, setSession] = useState(initialSession);
  const [choice, setChoice] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [needsRetry, setNeedsRetry] = useState(false);
  const [needsRefresh, setNeedsRefresh] = useState(false);
  const pending = useRef<StoryAction | null>(null);
  const inFlight = useRef(false);
  const heading = useRef<HTMLHeadingElement>(null);
  const feedback = useRef<HTMLDivElement>(null);
  const focusAfterResponse = useRef(false);
  const priorStep = useRef(initialSession.currentStep?.id);
  const situationSlug = storySituationSlug(session.situation);
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
  function accept(next: StorySession) {
    if (next.id !== session.id || next.revision < session.revision) return;
    if (
      next.currentStep?.id !== session.currentStep?.id ||
      (next.feedback && !next.feedback.correct)
    )
      setChoice("");
    focusAfterResponse.current = true;
    setSession(next);
  }
  async function submit(action?: StoryAction["action"]) {
    if (inFlight.current || needsRefresh) return;
    if (!pending.current) {
      if (!step || !action || (action === "answer" && !choice)) return;
      pending.current = {
        stepId: step.id,
        expectedRevision: session.revision,
        clientActionId: crypto.randomUUID(),
        action,
        ...(action === "answer" ? { choiceId: choice } : {}),
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
      const { data } = await createApiClient().submitStoryAction(
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
          "This story changed in another request. Load your saved progress to continue.",
        );
      } else {
        setNeedsRetry(true);
        setError(
          handleApiError(
            cause,
            "We could not confirm your story progress. Retry safely or load your saved progress.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }
  async function refresh() {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    try {
      accept((await createApiClient().getStorySession(session.id)).data);
      pending.current = null;
      setNeedsRetry(false);
      setNeedsRefresh(false);
      setError("");
      setNotice("Your saved story is up to date.");
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "We could not load your saved story. Please try again.",
        ),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }
  return (
    <div>
      <Link href="/stories" className={textLink}>
        Back to Stories
      </Link>
      <p className="mb-2 mt-3 text-sm font-semibold text-primary-700">
        {session.situation} · A2–B1
      </p>
      <p className="text-sm text-neutral-700">
        {session.completedSteps} of {session.totalSteps} steps completed
      </p>
      <progress
        className="mb-5 mt-2 h-2 w-full accent-primary-700"
        aria-label="Story progress"
        value={session.completedSteps}
        max={session.totalSteps}
      />
      <Surface>
        <h1
          ref={heading}
          tabIndex={-1}
          className="text-3xl font-bold tracking-tight text-neutral-900"
        >
          {session.status === "completed" ? "Story complete" : session.title}
        </h1>
        {session.status === "completed" ? (
          <div>
            <p className="mt-3 text-lg text-neutral-900">
              {session.firstAnswersCorrect} of {session.questionsAnswered}{" "}
              checks correct on your first try.
            </p>
            <p className="mt-2 text-neutral-700">
              You followed the whole dialogue and practised its useful phrases.
              Story practice does not change scheduled word reviews or award
              points.
            </p>
            <div className="my-5 flex flex-wrap items-center gap-4">
              <StoryStart storyKey={session.storyKey} repeat />
              <Link
                href={`/stories/${encodeURIComponent(session.storyKey)}`}
                className={textLink}
              >
                Read the full story
              </Link>
              {situationSlug && (
                <Link
                  href={`/discover/${encodeURIComponent(situationSlug)}#unit-guide`}
                  className={textLink}
                >
                  Open the situation guide
                </Link>
              )}
              <Link href="/discover" className={textLink}>
                Return to Journey
              </Link>
            </div>
          </div>
        ) : (
          <p className="mt-2 text-neutral-700">
            Read each line, then continue. Your current step is saved. Audio is
            optional.
          </p>
        )}
        <ol aria-label="Story dialogue" className="mt-5 space-y-4">
          {session.visibleLines.map((line, index) => (
            <li key={line.id} className="rounded-xl bg-neutral-50 p-4">
              <p className="text-sm font-semibold text-primary-700">
                {line.speaker}
              </p>
              <p className="mt-1 text-lg leading-relaxed text-neutral-900">
                {line.text}
              </p>
              <StoryAudio
                text={line.text}
                label={`line ${index + 1} by ${line.speaker}`}
              />
            </li>
          ))}
        </ol>
        {step && step.kind !== "line" && (
          <form
            className="mt-6"
            onSubmit={(event) => {
              event.preventDefault();
              void submit("answer");
            }}
          >
            <fieldset disabled={locked || !!checked?.correct}>
              <legend className="mb-3 text-xl font-bold text-neutral-900">
                {step.prompt}
              </legend>
              {step.choices?.map((option) => (
                <label
                  key={option.id}
                  className={`mb-3 flex min-h-12 cursor-pointer items-center gap-3 rounded-xl border p-4 text-neutral-900 ${choice === option.id ? "border-primary-600 bg-primary-50" : "border-neutral-300 bg-white"}`}
                >
                  <input
                    type="radio"
                    name="story-choice"
                    value={option.id}
                    checked={choice === option.id}
                    onChange={() => setChoice(option.id)}
                    className="h-5 w-5 shrink-0 accent-primary-700"
                  />
                  <span>{option.text}</span>
                </label>
              ))}
            </fieldset>
            {!checked?.correct && (
              <button
                type="submit"
                className={primaryAction}
                disabled={locked || !choice}
              >
                {busy ? "Checking…" : checked ? "Check again" : "Check answer"}
              </button>
            )}
          </form>
        )}
        {checked && (
          <div
            ref={feedback}
            tabIndex={-1}
            role="status"
            className={`mt-4 rounded-xl border p-4 text-neutral-900 ${checked.correct ? "border-primary-300 bg-primary-50" : "border-secondary-300 bg-secondary-50"}`}
          >
            <p className="font-bold">
              {checked.correct ? "That's right." : "Let's try that again."}
            </p>
            {!checked.correct && (
              <p className="mt-2">Answer: {checked.answer}</p>
            )}
            <p className="mt-2">{checked.explanation}</p>
            {!checked.correct && (
              <p className="mt-2">
                Choose the correct answer before continuing.
              </p>
            )}
          </div>
        )}
        {session.canContinue && (
          <button
            type="button"
            className={`${primaryAction} mt-5`}
            disabled={locked}
            onClick={() => void submit("continue")}
          >
            {busy
              ? "Saving…"
              : session.completedSteps + 1 === session.totalSteps
                ? "Finish story"
                : "Continue"}
          </button>
        )}
        {error && (
          <div
            role="alert"
            className="mt-5 rounded-xl border border-neutral-300 p-4"
          >
            <p className="text-neutral-900">{error}</p>
            <div className="mt-3 flex flex-wrap gap-3">
              {needsRetry && (
                <button
                  type="button"
                  className={secondaryAction}
                  disabled={busy}
                  onClick={() => void submit()}
                >
                  Retry safely
                </button>
              )}
              <button
                type="button"
                className={secondaryAction}
                disabled={busy}
                onClick={() => void refresh()}
              >
                Load saved progress
              </button>
            </div>
          </div>
        )}
        {notice && (
          <p role="status" className="mt-3 text-neutral-700">
            {notice}
          </p>
        )}
        <StoryVocabularyHelp vocabulary={session.vocabulary} />
      </Surface>
    </div>
  );
}
