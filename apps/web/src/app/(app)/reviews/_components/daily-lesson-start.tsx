"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiResponseError } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { deviceSpeech } from "@/lib/speech";

/** Start from Home without making the learner pass a second start screen. */
export function StartDailyLessonButton({
  className,
  label = "Start today’s lesson",
}: {
  className?: string;
  label?: string;
}) {
  const router = useRouter();
  const key = useRef<string | null>(null);
  const inFlight = useRef(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function start() {
    if (inFlight.current) return;
    inFlight.current = true;
    key.current ??= crypto.randomUUID();
    setBusy(true);
    setError("");
    deviceSpeech.prepare();
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().startDailyLesson(key.current, {
        headers: { "X-CSRF-Token": token },
      });
      router.push(`/review/session?lesson=${encodeURIComponent(data.id)}`);
    } catch (cause) {
      if (cause instanceof ApiResponseError && cause.status === 409) {
        // Another tab may have started or finished the current queue.
        router.push("/review");
        router.refresh();
      } else {
        setError(
          handleApiError(
            cause,
            "Your lesson could not start. Try again safely.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  return (
    <div>
      <button
        type="button"
        disabled={busy}
        aria-busy={busy}
        className={className}
        onClick={() => void start()}
      >
        {busy ? "Starting…" : label}
      </button>
      {error && (
        <p role="alert" className="mt-3 text-base text-neutral-700">
          {error}
        </p>
      )}
    </div>
  );
}
