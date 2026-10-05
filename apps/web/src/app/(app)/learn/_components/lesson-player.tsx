"use client";

import Link from "next/link";
import { useEffect, useId, useRef, useState } from "react";
import {
  ApiResponseError,
  type LessonAction,
  type LessonSession,
  type LessonSummary,
} from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { MeaningPicture } from "@/ui/meaning-picture";
import { ListenButton } from "@/ui/pronunciation";
import { Surface } from "@/ui/surface";
import { SessionActions } from "@/ui/session";
import { LessonAudio } from "./lesson-audio";
import { LessonReviewSave } from "./lesson-review-save";

const primary =
  "inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-50";
const secondary =
  "inline-flex min-h-12 items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 py-3 font-semibold text-neutral-900 hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-50";

export function LessonPlayer({
  lesson,
  initialSession,
}: {
  lesson: LessonSummary;
  initialSession: LessonSession | null;
}) {
  const [session, setSession] = useState(initialSession);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [typedAnswer, setTypedAnswer] = useState("");
  const [selectedChoice, setSelectedChoice] = useState("");
  const answerFormId = useId();
  const [needsRetry, setNeedsRetry] = useState(false);
  const [needsRefresh, setNeedsRefresh] = useState(false);
  const inFlight = useRef(false);
  const pendingAction = useRef<LessonAction | null>(null);
  const startKey = useRef<string | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const feedback = useRef<HTMLDivElement>(null);
  const focusAfterResponse = useRef(false);
  const priorStep = useRef(initialSession?.currentStep?.id);

  useEffect(() => {
    if (!focusAfterResponse.current) return;
    focusAfterResponse.current = false;
    if (
      session?.currentStep?.id !== priorStep.current ||
      session?.status === "completed"
    )
      heading.current?.focus();
    else if (session?.feedback) feedback.current?.focus();
    priorStep.current = session?.currentStep?.id;
  }, [session]);

  function accept(next: LessonSession) {
    focusAfterResponse.current = true;
    if (next.currentStep?.id !== session?.currentStep?.id) {
      setTypedAnswer("");
      setSelectedChoice("");
    } else if (next.feedback && !next.feedback.correct) {
      setSelectedChoice("");
    }
    setSession((current) =>
      current && current.id === next.id && current.revision > next.revision
        ? current
        : next,
    );
  }

  async function refreshProgress() {
    if (!session || inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    try {
      accept((await createApiClient().getLessonSession(session.id)).data);
      pendingAction.current = null;
      setNeedsRetry(false);
      setNeedsRefresh(false);
      setError("");
      setNotice("Your saved lesson progress is up to date.");
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "We could not load your saved progress. Try again when you are connected.",
        ),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  async function submit(
    action?: "answer" | "continue",
    choiceId?: string,
    typed?: string,
  ) {
    if (inFlight.current || needsRefresh) return;
    // Preserve the exact intent before session/CSRF preparation can fail. A
    // retry must have the same action even when no mutation reached the API.
    if (session && !pendingAction.current) {
      if (!session.currentStep || !action) return;
      pendingAction.current = {
        stepId: session.currentStep.id,
        expectedRevision: session.revision,
        clientActionId: crypto.randomUUID(),
        action,
        ...(choiceId ? { choiceId } : {}),
        ...(typed !== undefined ? { typedAnswer: typed } : {}),
      };
    }
    inFlight.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const init = { headers: { "X-CSRF-Token": token } };
      const client = createApiClient();
      let next: LessonSession;
      if (!session) {
        startKey.current ??= crypto.randomUUID();
        next = (await client.startLesson(lesson.key, startKey.current, init))
          .data;
      } else {
        const body = pendingAction.current;
        if (!body) return;
        next = (
          await client.submitLessonAction(
            session.id,
            body,
            body.clientActionId,
            init,
          )
        ).data;
      }
      pendingAction.current = null;
      setNeedsRetry(false);
      accept(next);
    } catch (cause) {
      if (
        cause instanceof ApiResponseError &&
        (cause.status === 400 || cause.status === 422)
      ) {
        pendingAction.current = null;
        setNeedsRetry(false);
        if (session?.currentStep?.kind !== "typed_recall")
          setSelectedChoice("");
        setError(
          session?.currentStep?.kind === "typed_recall"
            ? "That answer could not be accepted. Check your text and try again; your draft is still here."
            : "That answer could not be accepted. Choose again, then check your answer.",
        );
      } else if (
        cause instanceof ApiResponseError &&
        cause.status === 409 &&
        session
      ) {
        setNeedsRefresh(true);
        setNeedsRetry(false);
        setError(
          "This lesson changed in another request. Load your saved progress to continue.",
        );
      } else {
        setNeedsRetry(true);
        setError(
          handleApiError(
            cause,
            "We could not confirm that step. Retry to check and save it safely, or come back later.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  const step = session?.currentStep;
  const displayedChoice =
    session?.feedback?.correct && session.feedback.stepId === step?.id
      ? session.feedback.correctChoiceId
      : selectedChoice;
  const locked = busy || needsRetry || needsRefresh;

  return (
    <div className="pb-[7rem]">
      {session && (
        <div className="mb-6">
          <div className="mb-2 flex justify-between gap-3 text-sm text-neutral-700">
            <span>{session.title}</span>
            <span className="shrink-0">
              {session.completedSteps} of {session.totalSteps} steps
            </span>
          </div>
          <progress
            aria-label="Lesson progress"
            value={session.completedSteps}
            max={session.totalSteps}
            className="h-2 w-full accent-primary-700"
          />
        </div>
      )}

      {!session ? (
        <Surface>
          <p className="text-sm font-semibold text-primary-700">
            Learn, remember, use
          </p>
          <h1
            ref={heading}
            tabIndex={-1}
            className="mt-2 text-3xl font-bold tracking-tight text-neutral-900"
          >
            {lesson.title}
          </h1>
          <p className="mt-3 text-lg text-neutral-700">{lesson.description}</p>
          <p className="my-6 text-neutral-700">
            {lesson.wordCount} words, then a little practice.
          </p>
          <button
            type="button"
            disabled={locked}
            className={`${primary} w-full`}
            onClick={() => void submit()}
          >
            {busy ? "Opening lesson…" : "Start lesson"}
          </button>
          <p className="mt-3 text-sm text-neutral-600">
            Your progress saves as you go. You can leave and return.
          </p>
        </Surface>
      ) : session.status === "completed" ? (
        <Surface>
          <p className="text-sm font-semibold text-primary-700">
            {lesson.situationTitle}
          </p>
          <h1
            ref={heading}
            tabIndex={-1}
            className="mt-2 text-3xl font-bold text-neutral-900"
          >
            Lesson complete
          </h1>
          <p className="mt-3 text-lg text-neutral-700">
            You practised {session.words.length} words in context.
          </p>
          <p className="mt-2 text-neutral-600">
            {session.firstAnswersCorrect} of {session.questionsAnswered}{" "}
            questions correct on your first try.
          </p>
          <p className="mt-3 text-sm text-neutral-600">
            Want to remember these? Save the meanings you want to review.
          </p>
          <ul className="my-6 divide-y divide-neutral-200">
            {session.words.map((word) => (
              <li key={word.meaningId} className="py-4">
                <Link
                  className="inline-flex min-h-11 items-center text-lg font-semibold text-primary-700"
                  href={`/vocabulary/${encodeURIComponent(word.wordSlug)}`}
                >
                  {word.wordText}
                </Link>
                <p className="text-neutral-700">{word.definition}</p>
                <LessonReviewSave
                  meaningId={word.meaningId}
                  wordSlug={word.wordSlug}
                  wordText={word.wordText}
                />
              </li>
            ))}
          </ul>
          <div className="flex flex-wrap gap-3">
            <Link
              href={`/practice?lesson=${encodeURIComponent(session.lessonKey)}`}
              className={secondary}
            >
              Practise these words
            </Link>
            <Link href="/discover" className={primary}>
              Choose your next lesson
            </Link>
            <Link href="/review" className={secondary}>
              Review saved words
            </Link>
          </div>
        </Surface>
      ) : step ? (
        <Surface key={step.id} className="flex min-h-[32rem] flex-col">
          <div className="flex-1">
            <p className="text-sm font-semibold text-primary-700">
              {step.kind === "teach"
                ? "Meet a word"
                : step.kind === "recall"
                  ? "Remember the meaning"
                  : step.kind === "typed_recall"
                    ? "Remember the word"
                    : step.kind === "listening_choice"
                      ? "Listen and understand"
                      : "Use it in context"}
            </p>
            {step.kind === "teach" ? (
              <div className="mt-2 flex flex-wrap items-center justify-between gap-3">
                <h1
                  ref={heading}
                  tabIndex={-1}
                  className="text-3xl font-bold leading-snug text-neutral-900"
                >
                  {step.word.wordText}
                </h1>
                <ListenButton text={step.word.wordText} />
              </div>
            ) : (
              <h1
                ref={heading}
                tabIndex={-1}
                className="mt-2 text-2xl font-bold leading-snug text-neutral-900"
              >
                {step.prompt}
              </h1>
            )}
            {step.kind === "teach" ? (
              <div className="mt-3">
                <p className="mt-1 text-sm text-neutral-600">
                  {step.word.partOfSpeech}
                </p>
                <p className="mt-4 text-xl leading-relaxed text-neutral-900">
                  {step.word.definition}
                </p>
                <MeaningPicture meaningId={step.word.meaningId} />
                <blockquote className="mt-5 border-l-4 border-primary-300 pl-4 text-lg text-neutral-700">
                  {step.word.example}
                </blockquote>
                <div className="mt-2">
                  <ListenButton
                    text={step.word.example}
                    label="Listen to example"
                    showCaption={false}
                  />
                </div>
                {step.word.usageNote && (
                  <details className="mt-4 border-t border-neutral-200">
                    <summary className="min-h-11 cursor-pointer py-3 font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
                      Usage tip
                    </summary>
                    <p className="rounded-xl bg-secondary-50 p-4 text-neutral-800">
                      {step.word.usageNote}
                    </p>
                  </details>
                )}
              </div>
            ) : (
              <div className="mt-5">
                {step.context && (
                  <p className="mb-5 rounded-xl bg-secondary-50 p-4 text-lg leading-relaxed text-neutral-900">
                    {step.context}
                  </p>
                )}
                {step.kind === "listening_choice" && step.speechText && (
                  <LessonAudio text={step.speechText} />
                )}
                {step.kind === "typed_recall" ? (
                  <form
                    id={answerFormId}
                    onSubmit={(event) => {
                      event.preventDefault();
                      if (typedAnswer.trim())
                        void submit("answer", undefined, typedAnswer);
                    }}
                    className="space-y-3"
                  >
                    <label
                      htmlFor="lesson-typed-answer"
                      className="block font-semibold text-neutral-900"
                    >
                      Your word or phrase
                    </label>
                    <input
                      id="lesson-typed-answer"
                      type="text"
                      autoComplete="off"
                      autoCapitalize="none"
                      spellCheck={false}
                      maxLength={200}
                      value={typedAnswer}
                      onChange={(event) => setTypedAnswer(event.target.value)}
                      disabled={locked || session.canContinue}
                      className="min-h-12 w-full rounded-xl border border-neutral-300 bg-white px-4 py-3 text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
                    />
                    <p className="text-sm text-neutral-600">
                      Try to remember the word. Capital letters and extra spaces
                      do not matter.
                    </p>
                  </form>
                ) : (
                  <form
                    id={answerFormId}
                    onSubmit={(event) => {
                      event.preventDefault();
                      if (selectedChoice) void submit("answer", selectedChoice);
                    }}
                  >
                    <fieldset
                      disabled={locked || session.canContinue}
                      className="grid gap-3"
                    >
                      <legend className="sr-only">Answer choices</legend>
                      {step.choices.map((choice) => (
                        <label
                          key={choice.id}
                          className={`flex min-h-12 cursor-pointer items-center gap-3 rounded-xl border p-4 text-neutral-900 focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700 ${displayedChoice === choice.id ? "border-primary-600 bg-primary-50" : "border-neutral-300 bg-white"} ${locked || session.canContinue ? "cursor-default" : ""}`}
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
                )}
              </div>
            )}
            {session.feedback && (
              <div
                ref={feedback}
                tabIndex={-1}
                role="status"
                className={`mt-5 rounded-xl border p-4 ${session.feedback.correct ? "border-primary-200 bg-primary-50" : "border-secondary-200 bg-secondary-50"}`}
              >
                <h2 className="font-semibold text-neutral-900">
                  {session.feedback.correct
                    ? "That’s right"
                    : "Let’s look again"}
                </h2>
                <p className="mt-1 text-neutral-700">
                  {session.feedback.explanation}
                </p>
                {!session.feedback.correct && (
                  <p className="mt-2 text-sm text-neutral-700">
                    {step.kind === "typed_recall"
                      ? `Try typing the answer again${session.feedback.answer ? `: ${session.feedback.answer}` : "."}`
                      : "Choose an answer to try again."}
                  </p>
                )}
              </div>
            )}
          </div>
          <SessionActions>
            {session.canContinue ? (
              <button
                type="button"
                disabled={locked}
                onClick={() => void submit("continue")}
                className={`${primary} w-full`}
              >
                {busy
                  ? "Saving progress…"
                  : session.completedSteps + 1 === session.totalSteps
                    ? "Finish lesson"
                    : "Continue"}
              </button>
            ) : step.kind !== "teach" ? (
              <button
                type="submit"
                form={answerFormId}
                className={`${primary} w-full`}
                disabled={
                  locked ||
                  (step.kind === "typed_recall"
                    ? !typedAnswer.trim()
                    : !selectedChoice)
                }
              >
                {busy ? "Checking…" : "Check answer"}
              </button>
            ) : null}
          </SessionActions>
        </Surface>
      ) : null}

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
                className={secondary}
              >
                Retry step
              </button>
            )}
            {session && (
              <button
                type="button"
                disabled={busy}
                onClick={() => void refreshProgress()}
                className={secondary}
              >
                Load saved progress
              </button>
            )}
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
          Saving your progress…
        </p>
      )}
    </div>
  );
}
