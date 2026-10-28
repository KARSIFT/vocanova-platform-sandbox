"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import type { PracticeSessionsResponse } from "@vocanova/api-client";
import { primaryAction, secondaryAction } from "./practice-styles";

type Journey = NonNullable<PracticeSessionsResponse["journey"]>;

/** A small window onto actual durable progress, not a made-up course length. */
export function PracticeJourney({
  journey,
  compact = false,
  busy: parentBusy = false,
  onStart,
  onRefresh,
}: {
  journey: Journey | null;
  compact?: boolean;
  busy?: boolean;
  onStart?: () => void;
  onRefresh?: () => void;
}) {
  const router = useRouter();
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState("");
  const [openedId, setOpenedId] = useState("");
  const pendingKey = useRef<string | null>(null);
  const inFlight = useRef(false);
  const busy = parentBusy || starting;
  async function start() {
    if (onStart) {
      onStart();
      return;
    }
    if (inFlight.current || openedId) return;
    inFlight.current = true;
    setStarting(true);
    setError("");
    pendingKey.current ??= crypto.randomUUID();
    try {
      const csrf = await getOrRefreshCSRFToken();
      if (!csrf) throw new Error("Session not ready");
      const { data } = await createApiClient().startPracticeSession(
        { mode: "journey" },
        pendingKey.current,
        { headers: { "X-CSRF-Token": csrf } },
      );
      setOpenedId(data.id);
      router.push(`/practice/session/${encodeURIComponent(data.id)}`);
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "We could not confirm that practice started. Try again safely.",
        ),
      );
      inFlight.current = false;
      setStarting(false);
    }
  }
  function refresh() {
    if (onRefresh) onRefresh();
    else router.refresh();
  }
  if (!journey)
    return (
      <section
        aria-label="Your learning path"
        className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-[var(--vn-shadow-card)]"
      >
        <h2 className="text-xl font-bold text-neutral-900">
          {compact ? "Today’s practice" : "Your learning path"}
        </h2>
        <p role="status" className="mt-2 text-neutral-600">
          We could not load your learning path.
        </p>
        <button
          type="button"
          disabled={busy}
          onClick={refresh}
          className={`${secondaryAction} mt-4`}
        >
          Try again
        </button>
      </section>
    );
  const current = journey.currentSession;
  const lesson = journey.nextLesson;
  const today = Math.min(journey.todayCompletedLessons, journey.dailyGoal);
  const first = Math.max(1, lesson - (compact ? 1 : 2));
  const nodes = Array.from(
    { length: compact ? 3 : lesson - first + 3 },
    (_, index) => first + index,
  );
  return (
    <section
      aria-label="Your learning path"
      className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-[var(--vn-shadow-card)] sm:p-6"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-xl font-bold tracking-tight text-neutral-900">
            {compact ? "Today’s practice" : "Your path"}
          </h2>
          {compact && (
            <p className="mt-1 text-neutral-600">
              Listen. Complete. Put it together.
            </p>
          )}
        </div>
        <span className="shrink-0 whitespace-nowrap rounded-full border border-primary-200 bg-primary-50 px-3 py-1.5 text-sm font-semibold tabular-nums text-primary-800">
          {journey.todayCompletedLessons} / {journey.dailyGoal}
          {compact ? "" : " today"}
        </span>
      </div>
      <progress
        aria-label="Practice lessons today"
        value={today}
        max={journey.dailyGoal}
        className="mt-5 h-2.5 w-full"
      />
      <p className="mt-2 text-sm text-neutral-600">
        {journey.completedLessons}{" "}
        {journey.completedLessons === 1 ? "lesson" : "lessons"} completed
      </p>
      {compact ? (
        <>
          <ol
            aria-label="Practice path"
            className="relative mt-5 grid grid-cols-3 gap-2 before:absolute before:inset-x-[16.5%] before:top-6 before:h-0.5 before:rounded-full before:bg-neutral-200 before:content-['']"
          >
            {nodes.map((number) => {
              const done = number < lesson;
              const active = number === lesson;
              return (
                <li
                  key={number}
                  aria-current={active ? "step" : undefined}
                  className="flex min-w-0 flex-col items-center text-center"
                >
                  <span
                    aria-hidden="true"
                    className={`relative flex size-12 items-center justify-center rounded-full border-b-4 font-bold ${done ? "border-primary-300 bg-primary-100 text-primary-800" : active ? "border-primary-900 bg-primary-700 text-white ring-4 ring-primary-100" : "border-neutral-300 bg-neutral-100 text-neutral-600"}`}
                  >
                    {done ? "✓" : active ? "▶" : number}
                  </span>
                  <span className="mt-2 text-sm font-semibold text-neutral-900">
                    Lesson {number}
                  </span>
                  <span className="text-xs text-neutral-600">
                    {done ? "Completed" : active ? "Current" : "Coming next"}
                  </span>
                </li>
              );
            })}
          </ol>
          {current ? (
            <Link
              href={`/practice/session/${encodeURIComponent(current.id)}`}
              className={`${primaryAction} mt-5 w-full`}
            >
              Continue practice
            </Link>
          ) : (
            <button
              type="button"
              disabled={busy}
              onClick={() => void start()}
              className={`${primaryAction} mt-5 w-full`}
            >
              Start practice lesson
            </button>
          )}
        </>
      ) : (
        <ol
          aria-label="Practice path"
          className="mx-auto mt-4 max-w-[24rem] space-y-3"
        >
          {nodes.map((number) => {
            const done = number < lesson;
            const active = number === lesson;
            return (
              <li
                key={number}
                aria-current={active ? "step" : undefined}
                className={`relative flex items-center gap-4 ${number % 2 ? "pr-7" : "pl-7"}`}
              >
                <span
                  aria-hidden="true"
                  className={`flex size-14 shrink-0 items-center justify-center rounded-full border-b-4 text-xl font-bold ${done ? "border-primary-300 bg-primary-100 text-primary-800" : active ? "border-primary-900 bg-primary-700 text-white ring-4 ring-primary-100" : "border-neutral-300 bg-neutral-100 text-neutral-500"}`}
                >
                  {done ? "✓" : active ? "▶" : number}
                </span>
                <div className="min-w-0 py-2">
                  <p
                    className={`font-bold ${active ? "text-neutral-900" : "text-neutral-600"}`}
                  >
                    Lesson {number}
                  </p>
                  <p className="text-sm text-neutral-600">
                    {done
                      ? "Completed"
                      : active
                        ? current
                          ? "Ready to continue"
                          : "Your next step"
                        : "Coming next"}
                  </p>
                  {active &&
                    (current ? (
                      <Link
                        href={`/practice/session/${encodeURIComponent(current.id)}`}
                        className={`${primaryAction} mt-3`}
                      >
                        Continue practice
                      </Link>
                    ) : (
                      <button
                        type="button"
                        disabled={busy}
                        onClick={() => void start()}
                        className={`${primaryAction} mt-3`}
                      >
                        Start practice lesson
                      </button>
                    ))}
                </div>
              </li>
            );
          })}
        </ol>
      )}
      {error && (
        <div
          role="alert"
          className="mt-5 rounded-xl border border-secondary-300 bg-secondary-50 p-4 text-neutral-900"
        >
          <p>{error}</p>
          <button
            type="button"
            disabled={busy}
            onClick={() => void start()}
            className={`${secondaryAction} mt-3`}
          >
            Retry start
          </button>
        </div>
      )}
      {starting && (
        <p role="status" className="mt-3 text-sm text-neutral-600">
          Opening your practice…
        </p>
      )}
      {openedId && (
        <Link
          href={`/practice/session/${encodeURIComponent(openedId)}`}
          className={`${secondaryAction} mt-3`}
        >
          Open practice
        </Link>
      )}
      {journey.supportedKnownWords === 0 && (
        <p className="mt-5 text-sm text-neutral-600">
          Start with a few useful words, then practise them along your path.
        </p>
      )}
      {journey.totalKnownWords > journey.supportedKnownWords && (
        <details className="mt-4 text-sm text-neutral-600">
          <summary className="min-h-11 cursor-pointer content-center">
            Words in this path
          </summary>
          <p>
            {journey.supportedKnownWords} of your {journey.totalKnownWords}{" "}
            words currently have supported practice exercises. More words remain
            available in your map.
          </p>
        </details>
      )}
    </section>
  );
}
