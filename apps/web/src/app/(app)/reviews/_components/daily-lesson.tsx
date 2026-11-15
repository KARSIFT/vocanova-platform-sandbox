"use client";

import { StatusMessage } from "@/ui/status-message";
import { ActionLink, Button } from "@/ui/controls";
import { useEffect, useRef, useState } from "react";
import {
  ApiResponseError,
  type DailyLesson,
  type DailyLessonAction,
  type DailyReviewGoal,
} from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { deviceSpeech } from "@/lib/speech";
import { LearningNote } from "@/ui/learning-guide";
import { ListenButton } from "@/ui/pronunciation";
import { SessionActions } from "@/ui/session";
import { DailyLessonTeaching } from "./daily-lesson-teaching";
import { DailyLessonChoices } from "./daily-lesson-choices";
import {
  dailyReviewView,
  relativeReviewDate,
  reviewGoalDayChanged,
} from "./daily-review-view";
import { ReviewSchedule } from "./review-schedule";

const primary =
  "vn-btn vn-btn-primary text-base focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const secondary =
  "vn-btn vn-btn-secondary text-base focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
// Exact known no-due responses only: a generic 409 may be an uncertain start.
const NO_DUE_DETAILS = new Set([
  "No words are due. Discover a word to start learning.",
  "No required reviews are available. Check today’s goal or choose additional review.",
]);

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
  initialGoal,
}: {
  initialSession: DailyLesson | null;
  initialDueCount: number;
  initialGoal?: DailyReviewGoal;
}) {
  const [session, setSession] = useState(initialSession);
  const [dueCount, setDueCount] = useState(initialDueCount);
  const [goal, setGoal] = useState(initialGoal ?? initialSession?.goal);
  const [choice, setChoice] = useState(() => savedChoice(initialSession));
  const [nextReview, setNextReview] =
    useState<NonNullable<DailyLessonAction["nextReview"]>>("automatic");
  const [busy, setBusy] = useState(false);
  const [audioEnabled, setAudioEnabled] = useState(false);
  // Legacy responses have no stored zone; use the device only as a fallback.
  const [deviceTimeZone, setDeviceTimeZone] = useState<string>();
  const [error, setError] = useState("");
  const [uncertain, setUncertain] = useState(false);
  const [conflict, setConflict] = useState(false);
  const [completionCheck, setCompletionCheck] =
    useState<CompletionCheck | null>(null);
  const completionGeneration = useRef(0);
  const inFlight = useRef(false);
  const startKey = useRef<string | null>(null);
  const startAdditional = useRef(false);
  const pending = useRef<DailyLessonAction | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const feedbackRegion = useRef<HTMLDivElement>(null);
  const errorRegion = useRef<HTMLDivElement>(null);
  const focusResponse = useRef(false);
  const priorStep = useRef(initialSession?.currentStep?.id);
  const continueButton = useRef<HTMLButtonElement>(null);
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
  const view = dailyReviewView({ goal, dueCount, session });
  const completionCanContinue =
    dueCount > 0 || Boolean(completionCheck?.hasSession);
  const nextReviewDate = feedback?.nextReviewAt
    ? relativeReviewDate(
        feedback.nextReviewAt,
        goal?.timeZone ?? deviceTimeZone,
      )
    : undefined;
  const goalLocalDate = goal?.localDate;
  const goalTimeZone = goal?.timeZone;
  const nextOpportunity = goal?.nextReviewAt
    ? relativeReviewDate(goal.nextReviewAt, goal.timeZone ?? deviceTimeZone)
    : undefined;

  useEffect(() => {
    try {
      setDeviceTimeZone(new Intl.DateTimeFormat().resolvedOptions().timeZone);
    } catch {
      // Without a usable device zone, keep the honest undated default.
    }
  }, []);

  useEffect(() => {
    if (!goalTimeZone || !goalLocalDate) return;
    let active = true;
    let checking = false;
    async function checkDay() {
      if (
        !active ||
        checking ||
        document.visibilityState === "hidden" ||
        inFlight.current ||
        pending.current ||
        startKey.current ||
        !reviewGoalDayChanged({
          localDate: goalLocalDate!,
          timeZone: goalTimeZone!,
        })
      )
        return;
      checking = true;
      try {
        const { data } = await createApiClient().getCurrentDailyLesson();
        // A read started before an answer cannot overwrite its confirmation.
        // Keep the current step, checked feedback and unresolved action intact.
        if (!active || inFlight.current || pending.current || startKey.current)
          return;
        if (data.goal) setGoal(data.goal);
        setDueCount(data.dueCount);
      } catch {
        // A day-boundary read is optional; explicit actions still have recovery.
      } finally {
        checking = false;
      }
    }
    const timer = window.setInterval(() => void checkDay(), 30_000);
    const visible = () => void checkDay();
    document.addEventListener("visibilitychange", visible);
    window.addEventListener("focus", visible);
    void checkDay();
    return () => {
      active = false;
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", visible);
      window.removeEventListener("focus", visible);
    };
  }, [goalLocalDate, goalTimeZone]);

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
    else if (feedback) {
      // Native focus scrolling ignores the fixed action rail. Center the
      // checked meaning so its text stays clear of both persistent bars.
      feedbackRegion.current?.focus({ preventScroll: true });
      feedbackRegion.current?.scrollIntoView({
        block: "center",
        behavior: "instant",
      });
    }
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
        if (data.goal) setGoal(data.goal);
        setCompletionCheck({
          ...identity,
          status: "ready",
          hasSession: data.session?.status === "in_progress",
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
    if (next.goal) setGoal(next.goal);
    if (savedChoice(next)) setChoice(savedChoice(next));
    else if (next.id !== session?.id || next.currentStep?.id !== step?.id)
      setChoice("");
    if (next.id !== session?.id || next.currentStep?.id !== step?.id)
      setNextReview("automatic");
    focusResponse.current = true;
    setSession(next);
  }

  async function start(loadSavedProgress = false, additional = false) {
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
    if (!startKey.current) {
      startKey.current = crypto.randomUUID();
      startAdditional.current = additional;
    }
    let startAcknowledged = false;
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().startDailyLesson(
        startKey.current,
        {
          headers: { "X-CSRF-Token": token },
        },
        { additional: startAdditional.current },
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
        if (current.data.goal) setGoal(current.data.goal);
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
        NO_DUE_DETAILS.has(cause.body?.detail ?? "")
      ) {
        // This definitive rejection did not commit. Clear its key even if the
        // following read fails, so Load saved progress remains a read and a
        // fresh Start gets a new key. Ambiguous failures still replay theirs.
        startKey.current = null;
        try {
          const { data } = await createApiClient().getCurrentDailyLesson();
          setDueCount(data.dueCount);
          if (data.goal) setGoal(data.goal);
          if (data.session?.status === "in_progress") accept(data.session);
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
      if (data.session?.status === "in_progress") accept(data.session);
      else if (session?.status !== "completed") setSession(null);
      setDueCount(data.dueCount);
      if (data.goal) setGoal(data.goal);
      if (
        data.session?.status !== "in_progress" &&
        session?.status === "completed"
      )
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
        <p className="text-base font-semibold text-neutral-700">
          {view.complete && view.active ? "Optional review" : "Review words"}
        </p>
        {session && (
          <span className="text-sm font-semibold tabular-nums text-neutral-600">
            {session.completedWords} / {session.totalWords}
          </span>
        )}
      </div>
      {goal && session?.status === "in_progress" && (
        <p className="mb-3 text-sm font-semibold text-neutral-700">
          {view.complete ? "✓ " : ""}
          {view.reviewed} of {view.target} words reviewed today
        </p>
      )}
      {session && (
        <progress
          aria-label="Review progress"
          className="mb-5 h-2 w-full accent-primary-700"
          max={Math.max(1, session.totalWords)}
          value={session.completedWords}
        />
      )}

      {!session ? (
        <section className="learning-surface p-5 sm:p-7">
          <h1 ref={heading} tabIndex={-1} className="learning-title">
            {view.complete
              ? "Today’s review complete"
              : goal && view.allAvailableFinished && goal.reviewed > 0
                ? "All available reviews finished"
                : dueCount > 0
                  ? "Keep your words fresh"
                  : "You’re ready for new words"}
          </h1>
          {view.complete && (
            <p
              className="mt-4 text-3xl font-bold text-primary-700"
              aria-label="Daily review goal completed"
            >
              ✓
            </p>
          )}
          <p className="learning-description">
            {view.complete
              ? `${view.reviewed} of ${view.target} words reviewed today. Discover a new word or try today’s practice.`
              : goal && view.allAvailableFinished && goal.reviewed > 0
                ? `${view.reviewed} of ${view.target} words reviewed today. No words are due right now.`
                : dueCount > 0
                  ? `${dueCount} ${dueCount === 1 ? "word is" : "words are"} ready. Choose the meaning, one word at a time.`
                  : "Discover a word and choose whether to learn it or mark it as known."}
          </p>
          <LearningNote>
            {dueCount > 0
              ? "Take your time. Choose an answer, then check it before moving on."
              : "Come back to see which words are ready."}
          </LearningNote>
          <div className="mt-7 flex flex-wrap gap-3">
            {dueCount > 0 && !view.complete ? (
              <Button
                variant="primary"
                type="button"
                onClick={() => void start()}
                disabled={busy}
                className={primary}
              >
                {busy ? "Starting…" : "Start review"}
              </Button>
            ) : (
              <ActionLink
                variant="primary"
                href="/discovery"
                className={primary}
              >
                Discover words
              </ActionLink>
            )}
            {view.canReviewExtra && (
              <Button
                variant="secondary"
                type="button"
                disabled={busy}
                onClick={() => void start(false, true)}
              >
                Review more words
              </Button>
            )}
          </div>
        </section>
      ) : session.status === "completed" ? (
        <section className="learning-surface p-5 sm:p-7">
          <h1 ref={heading} tabIndex={-1} className="learning-title">
            {view.complete
              ? "Today’s review complete"
              : completionStatus !== "ready"
                ? "Review saved"
                : !completionCanContinue
                  ? goal
                    ? "All available reviews finished"
                    : "You’ve finished your review"
                  : "Review set complete"}
          </h1>
          {view.complete && (
            <p
              className="mt-4 text-3xl font-bold text-primary-700"
              aria-label="Daily review goal completed"
            >
              ✓
            </p>
          )}
          <p className="mt-4 text-xl text-neutral-900">
            {goal
              ? `${view.reviewed} of ${view.target} words reviewed today`
              : `${session.creditedToday} ${session.creditedToday === 1 ? "word" : "words"} remembered for today.`}
          </p>
          <p className="mt-3 text-base text-neutral-600">
            {session.firstAnswersCorrect} correct on the first try.
          </p>
          <LearningNote>
            <p>Your next review follows each word’s saved schedule.</p>
            {nextOpportunity ? (
              <p>
                Next review:{" "}
                <time
                  dateTime={goal!.nextReviewAt}
                  title={nextOpportunity.exact}
                  aria-label={`${nextOpportunity.label}. ${nextOpportunity.exact}`}
                >
                  {nextOpportunity.label}
                </time>
                .
              </p>
            ) : (
              <p>Come back to see which words are ready.</p>
            )}
          </LearningNote>
          {completionStatus === "loading" && (
            <StatusMessage
              as="p"
              role="status"
              className="mt-3 text-neutral-700"
            >
              Checking for more words…
            </StatusMessage>
          )}
          {completionStatus === "ready" && dueCount > 0 && (
            <p className="mt-3 text-neutral-700">
              {dueCount} {dueCount === 1 ? "word is" : "words are"} still ready.
            </p>
          )}
          <div className="mt-7 flex flex-wrap gap-3">
            {completionStatus === "ready" &&
              (completionCanContinue && !view.complete ? (
                <Button
                  variant="primary"
                  type="button"
                  className={primary}
                  disabled={busy}
                  onClick={() => void start()}
                >
                  Keep reviewing
                </Button>
              ) : (
                <ActionLink
                  variant="primary"
                  href="/discovery"
                  className={primary}
                >
                  Discover words
                </ActionLink>
              ))}
            {completionStatus === "ready" && view.canReviewExtra && (
              <Button
                variant="secondary"
                type="button"
                disabled={busy}
                onClick={() => void start(false, true)}
              >
                Review more words
              </Button>
            )}
            <ActionLink variant="secondary" href="/home" className={secondary}>
              Back to Home
            </ActionLink>
            <ActionLink
              variant="secondary"
              href="/practice"
              className={secondary}
            >
              Today’s practice
            </ActionLink>
          </div>
        </section>
      ) : step ? (
        <form
          className="learning-surface p-4 sm:p-7"
          onSubmit={(event) => {
            event.preventDefault();
            if (!locked) void submit("answer");
          }}
        >
          <p className="text-base text-neutral-600">
            {step.kind === "gap_choice"
              ? "Choose the missing word"
              : step.kind === "picture_choice"
                ? "Choose the picture"
                : "Choose the meaning"}{" "}
            · {step.partOfSpeech}
          </p>
          <h1
            ref={heading}
            tabIndex={-1}
            className="headword mt-2 text-4xl sm:text-[2.75rem]"
          >
            {step.kind === "gap_choice"
              ? "Complete the sentence"
              : step.wordText}
          </h1>
          <p className="learning-description">
            {step.kind === "gap_choice"
              ? step.context
              : step.kind === "picture_choice"
                ? "Choose the picture that fits this word."
                : "Choose the meaning that fits this word. You can listen if it helps."}
          </p>
          {step.kind !== "gap_choice" && (
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
          )}
          <DailyLessonChoices
            key={step.id}
            step={step}
            choice={choice}
            locked={locked}
            correctChoiceId={feedback?.correctChoiceId}
            onChange={setChoice}
          />
          {feedback && (
            <StatusMessage
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
                    : "Not quite"}
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
              {!feedback.skipped && !feedback.correct && (
                <p className="mt-3 text-base">
                  Read the meaning, then continue. Your first answer stays
                  recorded.
                </p>
              )}
            </StatusMessage>
          )}
          {feedback && !feedback.skipped && (
            <DailyLessonTeaching
              key={`${session.id}:${step.id}:${feedback.meaningId}`}
              sessionId={session.id}
              step={step}
              feedback={feedback}
            />
          )}
          <SessionActions
            id="daily-lesson-actions"
            role="region"
            aria-label="Session actions"
          >
            <div className="review-action-rail">
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
                <ReviewSchedule
                  value={nextReview}
                  automaticLabel={
                    nextReviewDate
                      ? `Automatic · ${nextReviewDate.label}`
                      : "Automatic"
                  }
                  automaticExact={nextReviewDate?.exact}
                  timeZone={goal?.timeZone}
                  disabled={busy || uncertain || conflict}
                  onChange={setNextReview}
                  nextAction={continueButton}
                />
              ) : (
                <span />
              )}
              {feedback ? (
                <Button
                  variant="primary"
                  ref={continueButton}
                  type="button"
                  className={`${primary} review-action-primary`}
                  disabled={busy || uncertain || conflict}
                  onClick={() => void submit("continue")}
                >
                  Continue
                </Button>
              ) : (
                <Button
                  variant="primary"
                  type="submit"
                  className={`${primary} review-action-primary`}
                  disabled={locked || !choice}
                >
                  {busy ? "Checking…" : "Check answer"}
                </Button>
              )}
            </div>
          </SessionActions>
        </form>
      ) : (
        <StatusMessage as="p" role="status">
          Load your saved progress to continue.
        </StatusMessage>
      )}

      {error && (
        <StatusMessage
          tone="error"
          ref={errorRegion}
          tabIndex={-1}
          role="alert"
          className="mt-5 rounded-xl border border-neutral-300 bg-white p-4 text-neutral-900"
        >
          <p>{error}</p>
          <div className="mt-3 flex flex-wrap gap-3">
            {uncertain && (
              <Button
                variant="primary"
                type="button"
                className={primary}
                disabled={busy}
                onClick={() => void submit()}
              >
                Retry same answer
              </Button>
            )}
            <Button
              variant="secondary"
              type="button"
              className={secondary}
              disabled={busy}
              onClick={() => void refresh()}
            >
              Load saved progress
            </Button>
          </div>
        </StatusMessage>
      )}
    </div>
  );
}
