"use client";

import { StatusMessage } from "@/ui/status-message";
import { Button, Select } from "@/ui/controls";
import Link from "@/ui/link";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import {
  ApiResponseError,
  type PracticeMode,
  type PracticeSessionsResponse,
  type PracticeStartRequest,
  type LessonSummary,
  type WordListSummary,
} from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError, isSessionExpiredError } from "@/lib/session";

import { PracticeJourney } from "./practice-journey";
import { practiceModes } from "./practice-modes";
import { secondaryAction, textLink } from "./practice-styles";

export function PracticeEntry({
  initialSessions,
  lessons: initialLessons,
  initialLessonKey,
  requestedMode,
  initialLists,
  initialListId,
  explicitSelection = false,
  lessonsUnavailable: initialLessonsUnavailable = false,
}: {
  initialSessions: PracticeSessionsResponse | null;
  lessons: LessonSummary[];
  initialLessonKey: string;
  requestedMode?: PracticeMode;
  initialLists: WordListSummary[] | null;
  initialListId: string;
  explicitSelection?: boolean;
  lessonsUnavailable?: boolean;
}) {
  const router = useRouter();
  const [sessions, setSessions] = useState(initialSessions);
  const [selection, setSelection] = useState(
    initialListId ? `list:${initialListId}` : initialLessonKey,
  );
  const [lists, setLists] = useState(initialLists);
  const [lessons, setLessons] = useState(initialLessons);
  const [lessonsUnavailable, setLessonsUnavailable] = useState(
    initialLessonsUnavailable,
  );
  const selectedListId = selection.startsWith("list:")
    ? selection.slice(5)
    : "";
  const selectedLesson = lessons.find((lesson) => lesson.key === selection);
  const selectedList = lists?.find((item) => item.id === selectedListId);
  // An explicitly requested lesson that is unknown, or whose catalog could not
  // load, stays selected. It never silently becomes a full-course mix.
  const unavailableLesson = Boolean(
    selection && !selectedListId && !selectedLesson,
  );
  const hasVocabularySelection = Boolean(
    lessons.length > 0 || lists?.length || selectedListId || unavailableLesson,
  );
  const unavailableSelection =
    unavailableLesson ||
    Boolean(
      selectedListId && (!selectedList || selectedList.usableMemberCount === 0),
    );
  const expired = useRef(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [needsRetry, setNeedsRetry] = useState(false);
  const [needsRefresh, setNeedsRefresh] = useState(false);
  const [openedSession, setOpenedSession] = useState<string | null>(null);
  const pending = useRef<{ body: PracticeStartRequest; key: string } | null>(
    null,
  );
  const inFlight = useRef(false);
  const locked = busy || needsRetry || needsRefresh || expired.current;
  const manualSessions =
    sessions?.items.filter((item) => item.mode !== "journey") ?? [];

  const focusedMode = requestedMode === "journey" ? undefined : requestedMode;
  const modes: PracticeMode[] = [
    "typed_recall",
    "listening_choice",
    "mistakes",
  ];
  const orderedModes = focusedMode
    ? [focusedMode, ...modes.filter((mode) => mode !== focusedMode)]
    : modes;
  const resume = manualSessions.find(
    (item) =>
      item.mode === focusedMode &&
      item.status === "in_progress" &&
      (item.mode === "mistakes" ||
        ((item.listId ?? "") === selectedListId &&
          (item.lessonKey ?? "") === (selectedListId ? "" : selection))),
  );
  const recentSessions = manualSessions.filter(
    (item) => item.id !== resume?.id,
  );
  const personalPath = (
    <div id="practice-journey" className="scroll-mt-24">
      <PracticeJourney
        journey={sessions?.journey ?? null}
        busy={locked}
        onStart={() => void start("journey")}
        onRefresh={() => void loadSessions()}
      />
    </div>
  );

  async function start(mode?: PracticeMode) {
    if (inFlight.current || needsRefresh || expired.current) return;
    if (!pending.current) {
      if (
        !mode ||
        (mode !== "mistakes" && mode !== "journey" && unavailableSelection)
      )
        return;
      pending.current = {
        body: {
          mode,
          ...(mode !== "mistakes" && mode !== "journey" && selectedList
            ? { listId: selectedList.id, listRevision: selectedList.revision }
            : mode !== "mistakes" &&
                mode !== "journey" &&
                selection &&
                !selectedListId
              ? { lessonKey: selection }
              : {}),
        },
        key: crypto.randomUUID(),
      };
    }
    const request = pending.current;
    inFlight.current = true;
    setBusy(true);
    setError("");
    let opened = false;
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().startPracticeSession(
        request.body,
        request.key,
        { headers: { "X-CSRF-Token": token } },
      );
      setOpenedSession(data.id);
      opened = true;
      router.push(`/practice/session/${encodeURIComponent(data.id)}`);
    } catch (cause) {
      if (isSessionExpiredError(cause)) expired.current = true;
      if (cause instanceof ApiResponseError && cause.status === 409) {
        setNeedsRefresh(true);
        setNeedsRetry(false);
        setError(
          "Your available practice has changed. Load saved sessions to see what is ready now.",
        );
      } else {
        setNeedsRetry(!isSessionExpiredError(cause));
        setError(
          handleApiError(
            cause,
            "We could not confirm that practice started. Retry safely, or load your saved sessions.",
          ),
        );
      }
    } finally {
      if (!opened) {
        inFlight.current = false;
        setBusy(false);
      }
    }
  }

  async function loadSessions() {
    if (inFlight.current || expired.current) return;
    inFlight.current = true;
    setBusy(true);
    try {
      const client = createApiClient();
      const [currentSessions, currentLists, currentLessons] =
        await Promise.allSettled([
          client.listPracticeSessions(),
          client.listWordLists(),
          client.listLessons({ signal: AbortSignal.timeout(5000) }),
        ]);
      for (const result of [currentSessions, currentLists, currentLessons]) {
        if (
          result.status === "rejected" &&
          isSessionExpiredError(result.reason)
        )
          throw result.reason;
      }
      // The lesson catalog recovers independently. A failed read keeps the
      // previous catalog and the learner's selection, and only reports that
      // lesson choices are unavailable.
      if (currentLessons.status === "fulfilled") {
        setLessons(currentLessons.value.data.items);
        setLessonsUnavailable(false);
      } else {
        setLessonsUnavailable(true);
      }
      if (currentSessions.status === "rejected") throw currentSessions.reason;
      if (currentLists.status === "rejected") throw currentLists.reason;
      setSessions(currentSessions.value.data);
      setLists(currentLists.value.data.items);
      pending.current = null;
      setNeedsRetry(false);
      setNeedsRefresh(false);
      setError("");
    } catch (cause) {
      if (isSessionExpiredError(cause)) expired.current = true;
      setError(
        handleApiError(
          cause,
          "We could not load your saved sessions. Please try again.",
        ),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="focused-practice-heading" className="mb-6">
      <h2 id="focused-practice-heading" className="sr-only">
        Remember your words
      </h2>
      {!focusedMode && personalPath}
      <details
        open={explicitSelection || Boolean(focusedMode)}
        className="mt-4 rounded-2xl border border-neutral-200 bg-white px-4 sm:px-6"
      >
        <summary className="vn-summary py-3">More practice</summary>
        {lessonsUnavailable && (
          <StatusMessage
            as="p"
            role="status"
            className="mt-3 text-sm text-neutral-600"
          >
            Situation lesson choices are unavailable. Your personal practice
            path is still available.
          </StatusMessage>
        )}
        {hasVocabularySelection ? (
          <div className="mt-4 max-w-[40rem]">
            <div className="flex flex-wrap items-center justify-between gap-x-3">
              <label
                htmlFor="practice-vocabulary"
                className="font-semibold text-neutral-900"
              >
                Practice vocabulary
              </label>
              <button
                type="button"
                aria-label="Refresh practice sources"
                disabled={busy || expired.current}
                className={`${textLink} text-sm`}
                onClick={() => void loadSessions()}
              >
                Refresh
              </button>
            </div>
            <Select
              id="practice-vocabulary"
              value={selection}
              disabled={locked}
              aria-describedby="practice-vocabulary-help"
              onChange={(event) => setSelection(event.target.value)}
              className="mt-2 w-full"
            >
              <option value="">Mix from the full course</option>
              {selectedListId && !selectedList && (
                <option value={selection}>Selected list is unavailable</option>
              )}
              {unavailableLesson && (
                <option value={selection}>
                  Selected lesson is unavailable
                </option>
              )}
              {lists && lists.length > 0 && (
                <optgroup label="Personal lists">
                  {lists.map((list) => (
                    <option key={list.id} value={`list:${list.id}`}>
                      {list.name} ({list.usableMemberCount} practice meanings)
                    </option>
                  ))}
                </optgroup>
              )}
              {lessons.map((lesson) => (
                <option key={lesson.key} value={lesson.key}>
                  {lesson.situationTitle}: {lesson.title}
                </option>
              ))}
            </Select>
            <details className="mt-1">
              <summary className="vn-summary text-sm">
                About this selection
              </summary>
              <p
                id="practice-vocabulary-help"
                className="pb-3 text-sm text-neutral-600"
              >
                {selection
                  ? "Practise these words by typing or listening."
                  : "A full-course mix can include words you have not studied yet."}{" "}
                Mistake practice uses your past answers.
              </p>
            </details>
          </div>
        ) : null}
        {selectedLesson && (
          <div className="mt-4 rounded-xl border border-neutral-200 p-4">
            <h3 className="font-semibold text-neutral-900">
              {selectedLesson.title}
            </h3>
            <p className="mt-1 text-sm text-neutral-700">
              {selectedLesson.description}
            </p>
            <p className="mt-2 text-sm text-neutral-600">
              {selectedLesson.wordCount}{" "}
              {selectedLesson.wordCount === 1 ? "word" : "words"} ·{" "}
              {selectedLesson.status === "completed"
                ? "Lesson complete"
                : selectedLesson.status === "in_progress"
                  ? `${selectedLesson.completedSteps} of ${selectedLesson.stepCount} steps completed`
                  : `${selectedLesson.stepCount} lesson steps`}
            </p>
            <Link
              href={`/learn/${encodeURIComponent(selectedLesson.key)}`}
              className={textLink}
            >
              {selectedLesson.status === "in_progress"
                ? "Resume situation lesson"
                : "Open situation lesson"}
            </Link>
          </div>
        )}
        {selectedList && (
          <p className="mt-3 text-neutral-700">
            Selected list: {selectedList.name}. {selectedList.usableMemberCount}{" "}
            of {selectedList.memberCount} words are available for practice.
          </p>
        )}
        {unavailableSelection && (
          <StatusMessage as="p" role="status" className="mt-3 text-neutral-700">
            {unavailableLesson
              ? lessonsUnavailable
                ? "Your selected lesson could not be loaded. Refresh to try again, or explicitly choose another vocabulary source."
                : "This lesson is not available for focused practice. Choose another vocabulary source."
              : selectedList
                ? "This list has no meanings available for focused practice. Add a supported meaning or choose another vocabulary source."
                : "Your selected list could not be loaded. Load saved sessions to refresh your lists, or explicitly choose another vocabulary source."}
          </StatusMessage>
        )}
        {(!lists || lessonsUnavailable) && !hasVocabularySelection ? (
          <div className="mt-3">
            {!lists && (
              <StatusMessage as="p" role="status" className="text-neutral-700">
                Personal lists are unavailable. Refresh to try again.
              </StatusMessage>
            )}
            <button
              type="button"
              aria-label="Refresh practice sources"
              disabled={busy || expired.current}
              className={`${textLink} text-sm`}
              onClick={() => void loadSessions()}
            >
              Refresh
            </button>
          </div>
        ) : (
          !lists && (
            <StatusMessage
              as="p"
              role="status"
              className="mt-3 text-neutral-700"
            >
              Personal lists are unavailable. Refresh to try again.
            </StatusMessage>
          )
        )}
        <ul className="vn-list mt-4">
          {orderedModes.map((mode) => {
            const content = practiceModes[mode];
            const noMistakes =
              mode === "mistakes" && sessions?.availableMistakes === 0;
            const description =
              mode === "mistakes" && sessions
                ? noMistakes
                  ? "No past mistakes to revisit right now."
                  : `${sessions.availableMistakes} ${sessions.availableMistakes === 1 ? "word is" : "words are"} ready for another try.`
                : content.description;
            const label = (
              <span className="min-w-0 flex-1 [overflow-wrap:anywhere]">
                <span className="block text-lg font-semibold text-neutral-900">
                  {noMistakes ? content.title : content.startLabel}
                </span>
                <span
                  id={`practice-mode-${mode}-description`}
                  className="mt-1 block text-sm font-normal text-neutral-600"
                >
                  {description}
                </span>
              </span>
            );
            if (resume && resume.mode === mode) {
              return (
                <li
                  key={mode}
                  id={`practice-mode-${mode}`}
                  className="scroll-mt-24"
                >
                  <Link
                    href={`/practice/session/${encodeURIComponent(resume.id)}`}
                    className="vn-list-row min-h-12"
                  >
                    <span>
                      <span className="block font-semibold text-neutral-900">
                        {mode === "mistakes"
                          ? "Resume mistake practice"
                          : mode === "typed_recall"
                            ? "Resume typed recall"
                            : "Resume listening practice"}
                      </span>
                      <span className="block text-sm text-neutral-600">
                        {resume.completedSteps} of {resume.totalSteps} questions
                        completed
                      </span>
                    </span>
                  </Link>
                </li>
              );
            }
            return noMistakes ? (
              <li
                key={mode}
                id={`practice-mode-${mode}`}
                className="vn-list-row scroll-mt-24"
              >
                {label}
              </li>
            ) : (
              <li
                key={mode}
                id={`practice-mode-${mode}`}
                className="scroll-mt-24"
              >
                <button
                  type="button"
                  aria-label={content.startLabel}
                  aria-describedby={`practice-mode-${mode}-description`}
                  disabled={
                    locked || (mode !== "mistakes" && unavailableSelection)
                  }
                  className="vn-list-row min-h-12 w-full text-left hover:enabled:bg-primary-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-3px] focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
                  onClick={() => void start(mode)}
                >
                  {label}
                  <svg
                    aria-hidden="true"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className="size-5 shrink-0 text-primary-700"
                  >
                    <path d="m9 5 7 7-7 7" />
                  </svg>
                </button>
              </li>
            );
          })}
        </ul>

        <Link href="/lists" className={`${textLink} mt-2 text-sm`}>
          Manage personal lists
        </Link>
      </details>
      {focusedMode && personalPath}
      {busy && (
        <StatusMessage as="p" role="status" className="mt-3 text-neutral-700">
          {openedSession ? "Opening your practice…" : "Loading your practice…"}
        </StatusMessage>
      )}
      {/* Recover independently if the automatic client navigation stalls. */}
      {openedSession && (
        <a
          href={`/practice/session/${encodeURIComponent(openedSession)}`}
          className={`${textLink} mt-2`}
        >
          Open practice
        </a>
      )}
      {error && (
        <StatusMessage
          tone="error"
          role="alert"
          className="mt-4 rounded-xl border border-secondary-300 bg-secondary-50 p-4 text-neutral-900"
        >
          <p>{error}</p>
          <div className="mt-3 flex flex-wrap gap-3">
            {needsRetry && (
              <Button
                variant="secondary"
                type="button"
                disabled={busy}
                className={secondaryAction}
                onClick={() => void start()}
              >
                Retry start
              </Button>
            )}
            <Button
              variant="secondary"
              type="button"
              disabled={busy}
              className={secondaryAction}
              onClick={() => void loadSessions()}
            >
              Load saved sessions
            </Button>
          </div>
        </StatusMessage>
      )}
      {!sessions && !error && (
        <div className="mt-4 text-neutral-700">
          <p>
            We could not load your recent practice. You can try starting a
            session or load your saved sessions again.
          </p>
          <button
            type="button"
            disabled={busy}
            className={`${textLink} mt-2`}
            onClick={() => void loadSessions()}
          >
            Load saved sessions
          </button>
        </div>
      )}

      {sessions && recentSessions.length > 0 && (
        <details
          open={recentSessions.some(
            (session) => session.status !== "completed",
          )}
          className="mt-4"
        >
          <summary className="vn-summary">Recent practice</summary>
          <ul className="vn-list mt-2">
            {recentSessions.map((session) => (
              <li key={session.id} className="vn-list-row">
                <div>
                  <p className="font-semibold text-neutral-900">
                    {practiceModes[session.mode].title}
                  </p>
                  {session.listName && (
                    <p className="text-sm text-neutral-600">
                      Personal list: {session.listName}
                    </p>
                  )}
                  <p className="text-sm text-neutral-600">
                    {session.status === "completed"
                      ? `${session.firstAnswersCorrect} of ${session.questionsAnswered} correct on your first try`
                      : `${session.completedSteps} of ${session.totalSteps} questions completed`}
                  </p>
                </div>
                <Link
                  href={`/practice/session/${encodeURIComponent(session.id)}`}
                  className={textLink}
                >
                  {session.status === "completed"
                    ? "View result"
                    : "Resume practice"}
                  <span className="sr-only">
                    : {practiceModes[session.mode].title}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </details>
      )}
    </section>
  );
}
