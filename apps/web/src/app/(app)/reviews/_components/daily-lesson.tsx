"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import {
  ApiResponseError,
  type DailyLesson,
  type DailyLessonAction,
} from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { deviceSpeech } from "@/lib/speech";
import { ListenButton } from "@/ui/pronunciation";
import { SessionActions } from "@/ui/session";
import { VocabularySources } from "@/ui/vocabulary-sources";

const primary =
  "inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-6 py-3 text-base font-bold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-default disabled:bg-neutral-200 disabled:text-neutral-600";
const secondary =
  "inline-flex min-h-11 items-center justify-center rounded-lg border border-neutral-300 px-4 py-2 text-base font-semibold text-neutral-800 hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const NO_DUE_DETAIL = "No words are due. Discover a word to start learning.";

function automaticReviewDateLabel(instant: string, timeZone?: string) {
  if (!timeZone || !Number.isFinite(Date.parse(instant))) return "Automatic";
  try {
    return `Automatic · ${new Intl.DateTimeFormat("en", { month: "short", day: "numeric", timeZone }).format(new Date(instant))}`;
  } catch {
    return "Automatic";
  }
}

interface CompletionCheck {
  sessionId: string;
  revision: number;
  status: "loading" | "ready" | "error";
  hasSession?: boolean;
}

function savedChoice(session: DailyLesson | null) {
  const feedback = session?.feedback;
  return feedback?.stepId === session?.currentStep?.id
    ? (feedback?.selectedChoiceId ??
        (feedback?.correct ? feedback.correctChoiceId : "") ??
        "")
    : "";
}

/** Every question, result, schedule and progress count comes from the server. */
export function DailyLessonPlayer({
  initialSession,
  initialDueCount,
}: {
  initialSession: DailyLesson | null;
  initialDueCount: number;
}) {
  const [session, setSession] = useState(initialSession);
  const [dueCount, setDueCount] = useState(initialDueCount);
  const [choice, setChoice] = useState(() => savedChoice(initialSession));
  const [nextReview, setNextReview] =
    useState<NonNullable<DailyLessonAction["nextReview"]>>("automatic");
  const [busy, setBusy] = useState(false);
  const [audioEnabled, setAudioEnabled] = useState(true);
  // Server and first client render intentionally share the undated label.
  const [reviewTimeZone, setReviewTimeZone] = useState<string>();
  const [error, setError] = useState("");
  const [uncertain, setUncertain] = useState(false);
  const [conflict, setConflict] = useState(false);
  const [completionCheck, setCompletionCheck] =
    useState<CompletionCheck | null>(null);
  const completionGeneration = useRef(0);
  const inFlight = useRef(false);
  const startKey = useRef<string | null>(null);
  const pending = useRef<DailyLessonAction | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const feedbackRegion = useRef<HTMLDivElement>(null);
  const errorRegion = useRef<HTMLDivElement>(null);
  const focusResponse = useRef(false);
  const priorStep = useRef(initialSession?.currentStep?.id);
  const step = session?.currentStep;
  const feedback =
    session?.feedback?.stepId === step?.id ? session?.feedback : null;
  const locked = busy || uncertain || conflict || Boolean(feedback);
  // A new completed session/revision is unconfirmed on its first render,
  // before effects run. The preceding lesson's due count cannot enable Next.
  const completionStatus =
    completionCheck?.sessionId === session?.id &&
    completionCheck?.revision === session?.revision
      ? completionCheck?.status
      : "loading";
  const completionCanContinue =
    dueCount > 0 || Boolean(completionCheck?.hasSession);

  useEffect(() => {
    try {
      setReviewTimeZone(new Intl.DateTimeFormat().resolvedOptions().timeZone);
    } catch {
      // Without a usable device zone, keep the honest undated default.
    }
  }, []);

  useEffect(() => {
    if (session?.id)
      window.history.replaceState(
        window.history.state,
        "",
        `/review/session?lesson=${encodeURIComponent(session.id)}`,
      );
  }, [session?.id]);

  useEffect(() => {
    if (!error) return;
    errorRegion.current?.focus({ preventScroll: true });
    errorRegion.current?.scrollIntoView({
      block: "center",
      behavior: "instant",
    });
  }, [error]);

  useEffect(() => {
    if (!focusResponse.current) return;
    focusResponse.current = false;
    if (priorStep.current !== step?.id || session?.status === "completed")
      heading.current?.focus();
    else if (feedback) feedbackRegion.current?.focus();
    priorStep.current = step?.id;
  }, [session, step?.id, feedback]);

  useEffect(() => {
    if (session?.status !== "completed") return;
    let active = true;
    const generation = ++completionGeneration.current;
    const identity = { sessionId: session.id, revision: session.revision };
    setCompletionCheck({ ...identity, status: "loading" });
    void createApiClient()
      .getCurrentDailyLesson()
      .then(({ data }) => {
        if (!active || generation !== completionGeneration.current) return;
        setDueCount(data.dueCount);
        setCompletionCheck({
          ...identity,
          status: "ready",
          hasSession: Boolean(data.session),
        });
      })
      .catch((cause: unknown) => {
        if (!active || generation !== completionGeneration.current) return;
        setCompletionCheck({ ...identity, status: "error" });
        setError(
          handleApiError(
            cause,
            "Refresh to check whether more words are ready.",
          ),
        );
      });
    return () => {
      active = false;
    };
  }, [session?.id, session?.status, session?.revision]);

  function accept(next: DailyLesson) {
    if (session && next.id === session.id && next.revision < session.revision)
      return;
    completionGeneration.current += 1;
    if (savedChoice(next)) setChoice(savedChoice(next));
    else if (next.id !== session?.id || next.currentStep?.id !== step?.id)
      setChoice("");
    if (next.id !== session?.id || next.currentStep?.id !== step?.id)
      setNextReview("automatic");
    focusResponse.current = true;
    setSession(next);
  }

  async function start(loadSavedProgress = false) {
    if (
      inFlight.current ||
      (session?.status === "completed" &&
        completionStatus !== "ready" &&
        !startKey.current)
    )
      return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    deviceSpeech.prepare();
    startKey.current ??= crypto.randomUUID();
    let startAcknowledged = false;
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().startDailyLesson(
        startKey.current,
        {
          headers: { "X-CSRF-Token": token },
        },
      );
      // The POST acknowledges this key, but a replayed receipt is an immutable
      // snapshot. Saved-progress recovery must read the current lesson too.
      startKey.current = null;
      startAcknowledged = true;
      let confirmed = data;
      if (loadSavedProgress) {
        const [current, saved] = await Promise.all([
          createApiClient().getCurrentDailyLesson(),
          createApiClient().getDailyLesson(data.id),
        ]);
        const active = current.data.session;
        confirmed =
          active &&
          (active.id !== saved.data.id || active.revision > saved.data.revision)
            ? active
            : saved.data;
        setDueCount(current.data.dueCount);
      }
      accept(confirmed);
      window.history.replaceState(
        window.history.state,
        "",
        `/review/session?lesson=${encodeURIComponent(confirmed.id)}`,
      );
      setUncertain(false);
      setConflict(false);
    } catch (cause) {
      if (
        !startAcknowledged &&
        cause instanceof ApiResponseError &&
        cause.status === 409 &&
        cause.body?.detail === NO_DUE_DETAIL
      ) {
        // This exact rejection rolls back without consuming the start key.
        // Reconcile once; an ambiguous Start error must still replay its key.
        try {
          const { data } = await createApiClient().getCurrentDailyLesson();
          setDueCount(data.dueCount);
          if (data.session) accept(data.session);
          else if (session?.status === "completed")
            setCompletionCheck({
              sessionId: session.id,
              revision: session.revision,
              status: "ready",
              hasSession: false,
            });
          else setSession(null);
          startKey.current = null;
          setUncertain(false);
          setConflict(false);
          setError("");
        } catch (recoveryFailure) {
          if (session?.status === "completed")
            setCompletionCheck({
              sessionId: session.id,
              revision: session.revision,
              status: "error",
            });
          setError(
            handleApiError(
              recoveryFailure,
              "We could not check your remaining words. Load saved progress to try again.",
            ),
          );
        }
      } else {
        setError(
          handleApiError(
            cause,
            startAcknowledged
              ? "Your lesson was saved, but we could not load your progress. Load saved progress to retry."
              : "We could not start your lesson. Try again safely.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  async function submit(action?: DailyLessonAction["action"]) {
    if (inFlight.current || conflict || !session || !step) return;
    if (!pending.current) {
      if (!action || (action === "answer" && !choice)) return;
      pending.current = {
        stepId: step.id,
        expectedRevision: session.revision,
        clientActionId: crypto.randomUUID(),
        action,
        ...(action === "answer" ? { choiceId: choice } : {}),
        ...(action === "continue" && nextReview !== "automatic"
          ? { nextReview }
          : {}),
      };
    }
    const request = pending.current;
    inFlight.current = true;
    setBusy(true);
    setError("");
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      accept(
        (
          await createApiClient().submitDailyLessonAction(
            session.id,
            request,
            request.clientActionId,
            { headers: { "X-CSRF-Token": token } },
          )
        ).data,
      );
      pending.current = null;
      setUncertain(false);
    } catch (cause) {
      if (cause instanceof ApiResponseError && cause.status === 409) {
        setConflict(true);
        setError("Your lesson changed. Load your saved progress to continue.");
      } else {
        setUncertain(true);
        setError(
          handleApiError(
            cause,
            "We could not confirm that answer. Retry the same answer safely.",
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
    // A read cannot acknowledge an ambiguous Start. Confirm its same receipt
    // before loading progress, so a later Next cannot replay the old lesson.
    if (startKey.current) {
      await start(true);
      return;
    }
    inFlight.current = true;
    setBusy(true);
    if (session?.status === "completed")
      setCompletionCheck({
        sessionId: session.id,
        revision: session.revision,
        status: "loading",
      });
    try {
      const { data } = await createApiClient().getCurrentDailyLesson();
      if (data.session) accept(data.session);
      else if (session?.status !== "completed") setSession(null);
      setDueCount(data.dueCount);
      if (!data.session && session?.status === "completed")
        setCompletionCheck({
          sessionId: session.id,
          revision: session.revision,
          status: "ready",
          hasSession: false,
        });
      pending.current = null;
      setUncertain(false);
      setConflict(false);
      setNextReview("automatic");
      setError("");
    } catch (cause) {
      if (session?.status === "completed")
        setCompletionCheck({
          sessionId: session.id,
          revision: session.revision,
          status: "error",
        });
      setError(
        handleApiError(cause, "We could not load your lesson. Try again."),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto max-w-[38rem] pb-[10rem]">
      <div className="mb-3 flex items-center justify-between gap-4">
        <p className="text-base font-semibold text-neutral-700">Review words</p>
        {session && (
          <span className="text-sm font-semibold tabular-nums text-neutral-600">
            {session.completedWords} / {session.totalWords}
          </span>
        )}
      </div>
      {session && (
        <progress
          aria-label="Review progress"
          className="mb-5 h-2 w-full accent-primary-700"
          max={Math.max(1, session.totalWords)}
          value={session.completedWords}
        />
      )}

      {!session ? (
        <section>
          <h1
            ref={heading}
            tabIndex={-1}
            className="text-3xl font-bold tracking-tight text-neutral-900"
          >
            {dueCount > 0
              ? "Keep your words fresh"
              : "You’re ready for new words"}
          </h1>
          <p className="mt-4 text-lg text-neutral-700">
            {dueCount > 0
              ? `${dueCount} ${dueCount === 1 ? "word is" : "words are"} ready. Choose the meaning, one word at a time.`
              : "Discover a word and choose whether to learn it or mark it as known."}
          </p>
          <div className="mt-7 flex flex-wrap gap-3">
            {dueCount > 0 ? (
              <button
                type="button"
                onClick={() => void start()}
                disabled={busy}
                className={primary}
              >
                {busy ? "Starting…" : "Start review"}
              </button>
            ) : (
              <Link href="/discovery" className={primary}>
                Discover words
              </Link>
            )}
          </div>
        </section>
      ) : session.status === "completed" ? (
        <section>
          <h1
            ref={heading}
            tabIndex={-1}
            className="text-3xl font-bold tracking-tight text-neutral-900"
          >
            {completionStatus !== "ready"
              ? "Review saved"
              : completionCanContinue
                ? "Review set complete"
                : "You’ve finished your review"}
          </h1>
          <p className="mt-4 text-xl text-neutral-900">
            {session.creditedToday}{" "}
            {session.creditedToday === 1 ? "word" : "words"} remembered for
            today.
          </p>
          <p className="mt-3 text-base text-neutral-600">
            {session.firstAnswersCorrect} correct on the first try.
          </p>
          {completionStatus === "loading" && (
            <p role="status" className="mt-3 text-neutral-700">
              Checking for more words…
            </p>
          )}
          {completionStatus === "ready" && dueCount > 0 && (
            <p className="mt-3 text-neutral-700">
              {dueCount} {dueCount === 1 ? "word is" : "words are"} still ready.
            </p>
          )}
          <div className="mt-7 flex flex-wrap gap-3">
            {completionStatus === "ready" &&
              (completionCanContinue ? (
                <button
                  type="button"
                  className={primary}
                  disabled={busy}
                  onClick={() => void start()}
                >
                  Keep reviewing
                </button>
              ) : (
                <Link href="/discovery" className={primary}>
                  Discover words
                </Link>
              ))}
            <Link href="/home" className={secondary}>
              Back to Home
            </Link>
            <Link href="/practice" className={secondary}>
              Today’s practice
            </Link>
          </div>
        </section>
      ) : step ? (
        <form
          onSubmit={(event) => {
            event.preventDefault();
            if (!locked) void submit("answer");
          }}
        >
          <p className="text-base text-neutral-600">
            Choose the meaning · {step.partOfSpeech}
          </p>
          <h1
            ref={heading}
            tabIndex={-1}
            className="mt-2 break-words text-3xl font-bold tracking-tight text-neutral-900 sm:text-[2.75rem]"
          >
            {step.wordText}
          </h1>
          <div className="mt-2 flex items-start justify-between gap-2">
            <div className="min-w-0 flex-1">
              <ListenButton
                key={step.id}
                text={step.wordText}
                meaningId={step.pronunciationMeaningId}
                meaningRevision={step.pronunciationRevision}
                showCaption={false}
                autoPlay={audioEnabled}
              />
            </div>
            <button
              type="button"
              aria-label={`Automatic audio ${audioEnabled ? "on" : "off"}`}
              title={`Automatic audio ${audioEnabled ? "on" : "off"}`}
              aria-pressed={audioEnabled}
              onClick={() => setAudioEnabled((previous) => !previous)}
              className="inline-flex min-h-11 min-w-11 shrink-0 items-center justify-center rounded-lg text-neutral-600 hover:bg-neutral-100"
            >
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                className="size-5 fill-none stroke-current stroke-2"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="m11 5-5 4H3v6h3l5 4V5Z"
                />
                {audioEnabled ? (
                  <path
                    strokeLinecap="round"
                    d="M15 9a5 5 0 0 1 0 6m3-9a9 9 0 0 1 0 12"
                  />
                ) : (
                  <path strokeLinecap="round" d="m16 9 6 6m0-6-6 6" />
                )}
              </svg>
            </button>
          </div>
          <fieldset disabled={locked} className="mt-4 grid gap-2">
            <legend className="sr-only">Meaning of {step.wordText}</legend>
            {step.choices.map((option) => (
              <label
                key={option.id}
                className={`flex min-h-14 cursor-pointer items-center gap-3 rounded-xl border-2 px-4 py-3 text-base leading-relaxed text-neutral-900 focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700 ${choice === option.id ? "border-primary-700 bg-primary-50" : "border-neutral-200 bg-white hover:border-primary-300"}`}
              >
                <input
                  type="radio"
                  name="meaning"
                  value={option.id}
                  checked={choice === option.id}
                  onChange={() => setChoice(option.id)}
                  onFocus={(event) => {
                    // Native focus scrolling cannot account for a fixed action
                    // rail. Keep keyboard-focused choices clear of both bars.
                    const input = event.currentTarget;
                    if (!input.matches(":focus-visible")) return;
                    const label = input.closest("label");
                    const bounds = label?.getBoundingClientRect();
                    const rail = document
                      .getElementById("daily-lesson-actions")
                      ?.getBoundingClientRect();
                    const top =
                      document.querySelector("header")?.getBoundingClientRect()
                        .bottom ?? 0;
                    if (
                      bounds &&
                      rail &&
                      (bounds.bottom > rail.top - 8 || bounds.top < top + 8)
                    )
                      label?.scrollIntoView({
                        block: "center",
                        behavior: "instant",
                      });
                  }}
                  className="size-5 shrink-0 accent-primary-700"
                />
                <span>{option.text}</span>
                {feedback?.correctChoiceId === option.id && (
                  <span
                    className="ml-auto font-bold"
                    aria-label="Correct meaning"
                  >
                    ✓
                  </span>
                )}
              </label>
            ))}
          </fieldset>
          {feedback && (
            <div
              ref={feedbackRegion}
              tabIndex={-1}
              role="status"
              className="mt-5 rounded-xl border border-primary-200 bg-primary-50 p-4 text-neutral-900"
            >
              <p className="text-lg font-bold">
                {feedback.skipped
                  ? feedback.reason === "already_known"
                    ? "Already marked as known"
                    : feedback.reason === "already_reviewed"
                      ? "Already reviewed elsewhere"
                      : feedback.reason === "removed"
                        ? "This word has changed"
                        : "Skipped for now"
                  : feedback.correct
                    ? "Correct"
                    : "Let’s try this one again later"}
              </p>
              <p className="mt-2 text-base">{feedback.answer}</p>
              {feedback.explanation && (
                <p className="mt-2 text-sm">{feedback.explanation}</p>
              )}
              {!feedback.skipped && feedback.reviewStep !== undefined && (
                <p className="mt-2 text-sm font-semibold">
                  Level {feedback.reviewStep + 1} of 8
                </p>
              )}
            </div>
          )}
          <SessionActions
            id="daily-lesson-actions"
            role="region"
            aria-label="Session actions"
          >
            {!feedback ? (
              <button
                type="button"
                className="min-h-11 rounded-lg px-3 text-base font-semibold text-neutral-600 hover:text-neutral-900"
                disabled={locked}
                onClick={() => void submit("skip")}
              >
                Skip for now
              </button>
            ) : !feedback.skipped &&
              feedback.nextReviewAt &&
              feedback.reviewStep !== undefined ? (
              <label className="min-w-0 flex-1 text-sm font-semibold text-neutral-700 sm:max-w-[20rem]">
                Show this word again
                <select
                  value={nextReview}
                  disabled={busy || uncertain || conflict}
                  onChange={(event) =>
                    setNextReview(
                      event.target.value as NonNullable<
                        DailyLessonAction["nextReview"]
                      >,
                    )
                  }
                  className="mt-1 block min-h-11 w-full rounded-lg border border-neutral-300 bg-white px-3 py-2 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
                >
                  <option value="automatic">
                    {feedback.correct
                      ? automaticReviewDateLabel(
                          feedback.nextReviewAt,
                          reviewTimeZone,
                        )
                      : "Automatic · try again in this review"}
                  </option>
                  <option value="1d">In 1 day</option>
                  <option value="3d">In 3 days</option>
                  <option value="7d">In 7 days</option>
                  <option value="14d">In 14 days</option>
                  <option value="30d">In 30 days</option>
                  <option value="60d">In 60 days</option>
                  <option value="120d">In 120 days</option>
                  <option value="known">I know this word</option>
                </select>
              </label>
            ) : (
              <span />
            )}
            {feedback ? (
              <button
                type="button"
                className={primary}
                disabled={busy || uncertain || conflict}
                onClick={() => void submit("continue")}
              >
                Continue
              </button>
            ) : (
              <button
                type="submit"
                className={primary}
                disabled={locked || !choice}
              >
                {busy ? "Checking…" : "Check answer"}
              </button>
            )}
          </SessionActions>
        </form>
      ) : (
        <p role="status">Load your saved progress to continue.</p>
      )}

      {error && (
        <div
          ref={errorRegion}
          tabIndex={-1}
          role="alert"
          className="mt-5 rounded-xl border border-neutral-300 bg-white p-4 text-neutral-900"
        >
          <p>{error}</p>
          <div className="mt-3 flex flex-wrap gap-3">
            {uncertain && (
              <button
                type="button"
                className={primary}
                disabled={busy}
                onClick={() => void submit()}
              >
                Retry same answer
              </button>
            )}
            <button
              type="button"
              className={secondary}
              disabled={busy}
              onClick={() => void refresh()}
            >
              Load saved progress
            </button>
          </div>
        </div>
      )}
      {session && <VocabularySources attribution={session.attribution} />}
    </div>
  );
}
