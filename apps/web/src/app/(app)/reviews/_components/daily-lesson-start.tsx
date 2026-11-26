"use client";

import { StatusMessage } from "@/ui/status-message";
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
  additional = false,
}: {
  className?: string;
  label?: string;
  additional?: boolean;
}) {
  const router = useRouter();
  const key = useRef<string | null>(null);
  const inFlight = useRef(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  // Exact session address once the server confirmed the start.
  const [openedHref, setOpenedHref] = useState<string | null>(null);

  async function start() {
    if (inFlight.current) return;
    inFlight.current = true;
    key.current ??= crypto.randomUUID();
    setBusy(true);
    setError("");
    deviceSpeech.prepare();
    let opened = false;
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().startDailyLesson(
        key.current,
        {
          headers: { "X-CSRF-Token": token },
        },
        { additional },
      );
      if (typeof data?.id !== "string" || !data.id)
        throw new Error("missing lesson id");
      const href = `/review/session?lesson=${encodeURIComponent(data.id)}`;
      // Confirmed: never POST again from this control. If the client
      // navigation stalls, the native link below opens the same session.
      opened = true;
      setOpenedHref(href);
      router.push(href);
    } catch (cause) {
      if (opened) return;
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
      if (!opened) {
        inFlight.current = false;
        setBusy(false);
      }
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
        {openedHref ? "Opening review…" : busy ? "Starting…" : label}
      </button>
      {/* Recover independently if the automatic client navigation stalls. */}
      {openedHref && (
        <p className="mt-2">
          <a href={openedHref} className="vn-text-link">
            Open review
          </a>
        </p>
      )}
      {error && (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          className="mt-3 text-base text-neutral-700"
        >
          {error}
        </StatusMessage>
      )}
    </div>
  );
}
