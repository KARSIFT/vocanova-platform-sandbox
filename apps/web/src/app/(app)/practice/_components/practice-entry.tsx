"use client";

import Link from "next/link";
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

import { practiceModes } from "./practice-modes";
import { secondaryAction, textLink } from "./practice-styles";

export function PracticeEntry({
  initialSessions,
  lessons,
  initialLessonKey,
  initialLists,
  initialListId,
}: {
  initialSessions: PracticeSessionsResponse | null;
  lessons: LessonSummary[];
  initialLessonKey: string;
  initialLists: WordListSummary[] | null;
  initialListId: string;
}) {
  const router = useRouter();
  const [sessions, setSessions] = useState(initialSessions);
  const [selection, setSelection] = useState(
    initialListId ? `list:${initialListId}` : initialLessonKey,
  );
  const [lists, setLists] = useState(initialLists);
  const selectedListId = selection.startsWith("list:")
    ? selection.slice(5)
    : "";
  const selectedList = lists?.find((item) => item.id === selectedListId);
  const hasVocabularySelection = Boolean(
    lessons.length > 0 || lists?.length || selectedListId,
  );
  const unavailableSelection = Boolean(
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

  async function start(mode?: PracticeMode) {
    if (inFlight.current || needsRefresh || expired.current) return;
    if (!pending.current) {
      if (!mode || (mode !== "mistakes" && unavailableSelection)) return;
      pending.current = {
        body: {
          mode,
          ...(mode !== "mistakes" && selectedList
            ? { listId: selectedList.id, listRevision: selectedList.revision }
            : mode !== "mistakes" && selection && !selectedListId
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
      const [currentSessions, currentLists] = await Promise.all([
        client.listPracticeSessions(),
        client.listWordLists(),
      ]);
      setSessions(currentSessions.data);
      setLists(currentLists.data.items);
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
      {hasVocabularySelection ? (
        <div className="mt-4 max-w-[40rem]">
          <div className="flex items-center justify-between gap-3">
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
          <select
            id="practice-vocabulary"
            value={selection}
            disabled={locked}
            aria-describedby="practice-vocabulary-help"
            onChange={(event) => setSelection(event.target.value)}
            className="mt-2 min-h-12 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            <option value="">Mix from the full course</option>
            {selectedListId && !selectedList && (
              <option value={selection}>Selected list is unavailable</option>
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
          </select>
          <details className="mt-1">
            <summary className="min-h-11 cursor-pointer content-center text-sm text-neutral-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
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
      {selectedList && (
        <p className="mt-3 text-neutral-700">
          Selected list: {selectedList.name}. {selectedList.usableMemberCount}{" "}
          of {selectedList.memberCount} words are available for practice.
        </p>
      )}
      {unavailableSelection && (
        <p role="status" className="mt-3 text-neutral-700">
          {selectedList
            ? "This list has no meanings available for focused practice. Add a supported meaning or choose another vocabulary source."
            : "Your selected list could not be loaded. Load saved sessions to refresh your lists, or explicitly choose another vocabulary source."}
        </p>
      )}
      {!lists && (
        <div className="mt-3">
          <p role="status" className="text-neutral-700">
            Personal lists are unavailable. Refresh to try again.
          </p>
          {!hasVocabularySelection && (
            <button
              type="button"
              aria-label="Refresh practice sources"
              disabled={busy || expired.current}
              className={`${textLink} text-sm`}
              onClick={() => void loadSessions()}
            >
              Refresh
            </button>
          )}
        </div>
      )}
      <div className="mt-4 divide-y divide-neutral-200 border-y border-neutral-200">
        {(Object.keys(practiceModes) as PracticeMode[]).map((mode) => {
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
            <span className="min-w-0">
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
          return noMistakes ? (
            <div key={mode} className="px-3 py-4">
              {label}
            </div>
          ) : (
            <button
              key={mode}
              type="button"
              aria-label={content.startLabel}
              aria-describedby={`practice-mode-${mode}-description`}
              disabled={locked || (mode !== "mistakes" && unavailableSelection)}
              className="flex min-h-20 w-full items-center justify-between gap-4 rounded-lg px-3 py-4 text-left hover:bg-primary-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
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
          );
        })}
      </div>

      <Link href="/lists" className={`${textLink} mt-2 text-sm`}>
        Manage personal lists
      </Link>

      {busy && (
        <p role="status" className="mt-3 text-neutral-700">
          {openedSession ? "Opening your practice…" : "Loading your practice…"}
        </p>
      )}
      {openedSession && (
        <Link
          href={`/practice/session/${encodeURIComponent(openedSession)}`}
          className={`${textLink} mt-2`}
        >
          Open practice
        </Link>
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
                className={secondaryAction}
                onClick={() => void start()}
              >
                Retry start
              </button>
            )}
            <button
              type="button"
              disabled={busy}
              className={secondaryAction}
              onClick={() => void loadSessions()}
            >
              Load saved sessions
            </button>
          </div>
        </div>
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

      {sessions && sessions.items.length > 0 && (
        <details
          open={sessions.items.some(
            (session) => session.status !== "completed",
          )}
          className="mt-4"
        >
          <summary className="min-h-11 cursor-pointer content-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
            Recent practice
          </summary>
          <ul className="mt-2 divide-y divide-neutral-200">
            {sessions.items.map((session) => (
              <li
                key={session.id}
                className="flex flex-wrap items-center justify-between gap-3 py-3"
              >
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
