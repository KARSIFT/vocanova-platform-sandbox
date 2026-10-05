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
import { Surface } from "@/ui/surface";

import { practiceModes } from "./practice-modes";
import { primaryAction, secondaryAction, textLink } from "./practice-styles";

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
      <h2
        id="focused-practice-heading"
        className="scroll-mt-24 text-xl font-bold tracking-tight text-neutral-900"
      >
        Remember your words
      </h2>
      <p className="mt-2 text-neutral-700">
        Choose how to practise. Your progress saves as you go.
      </p>
      {lessons.length > 0 || lists?.length || selectedListId ? (
        <div className="mt-4 max-w-[40rem]">
          <label
            htmlFor="practice-vocabulary"
            className="font-semibold text-neutral-900"
          >
            Practice vocabulary
          </label>
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
          <p
            id="practice-vocabulary-help"
            className="mt-2 text-sm text-neutral-600"
          >
            {selection
              ? "Practise these words by typing or listening."
              : "A full-course mix can include words you have not studied yet."}{" "}
            Mistake practice uses your past answers.
          </p>
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
        <p role="status" className="mt-3 text-neutral-700">
          Personal lists are unavailable. Load saved sessions to try again.
        </p>
      )}
      <div className="mt-2 flex flex-wrap gap-3">
        <Link href="/lists" className={textLink}>
          Manage personal lists
        </Link>
        <button
          type="button"
          disabled={busy || expired.current}
          className={textLink}
          onClick={() => void loadSessions()}
        >
          Refresh practice sources
        </button>
      </div>
      <div className="mt-4 grid gap-3 md:grid-cols-3">
        {(Object.keys(practiceModes) as PracticeMode[]).map((mode) => {
          const content = practiceModes[mode];
          const noMistakes =
            mode === "mistakes" && sessions?.availableMistakes === 0;
          return (
            <article
              key={mode}
              className="flex flex-col border-t-2 border-primary-200 bg-white p-4"
            >
              <h3 className="text-xl font-bold text-neutral-900">
                {content.title}
              </h3>
              <p className="mt-2 grow text-neutral-700">
                {content.description}
              </p>
              {mode === "mistakes" && sessions && (
                <p className="mt-3 text-sm text-neutral-600">
                  {noMistakes
                    ? "No past mistakes to revisit right now."
                    : `${sessions.availableMistakes} ${sessions.availableMistakes === 1 ? "word is" : "words are"} ready for another try.`}
                </p>
              )}
              {!noMistakes && (
                <button
                  type="button"
                  disabled={
                    locked || (mode !== "mistakes" && unavailableSelection)
                  }
                  className={`${primaryAction} mt-4`}
                  onClick={() => void start(mode)}
                >
                  {content.startLabel}
                </button>
              )}
            </article>
          );
        })}
      </div>

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
        <Surface aria-labelledby="recent-practice-heading" className="mt-4">
          <h3
            id="recent-practice-heading"
            className="text-lg font-bold text-neutral-900"
          >
            Recent practice
          </h3>
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
        </Surface>
      )}
    </section>
  );
}
