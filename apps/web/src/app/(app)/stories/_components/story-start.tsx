"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { primaryAction } from "../../practice/_components/practice-styles";

export function StoryStart({
  storyKey,
  repeat = false,
}: {
  storyKey: string;
  repeat?: boolean;
}) {
  const router = useRouter();
  const key = useRef<string | null>(null);
  const inFlight = useRef(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function start() {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    key.current ??= crypto.randomUUID();
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().startStorySession(
        { storyKey },
        key.current,
        { headers: { "X-CSRF-Token": token } },
      );
      router.push(`/stories/session/${encodeURIComponent(data.id)}`);
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "We could not confirm your story. Try again to safely resume the same request.",
        ),
      );
      inFlight.current = false;
      setBusy(false);
    }
  }
  return (
    <div>
      <button
        type="button"
        className={primaryAction}
        disabled={busy}
        onClick={() => void start()}
      >
        {busy
          ? "Opening story…"
          : error
            ? "Retry opening story"
            : repeat
              ? "Practise again"
              : "Start story"}
      </button>
      {error && (
        <p role="alert" className="mt-3 text-neutral-900">
          {error}
        </p>
      )}
    </div>
  );
}
