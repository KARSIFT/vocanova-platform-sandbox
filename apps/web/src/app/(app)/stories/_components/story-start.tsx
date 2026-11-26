"use client";

import { StatusMessage } from "@/ui/status-message";
import { Button } from "@/ui/controls";
import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import {
  primaryAction,
  secondaryAction,
  textLink,
} from "../../practice/_components/practice-styles";

export function StoryStart({
  storyKey,
  repeat = false,
  secondary = false,
}: {
  storyKey: string;
  repeat?: boolean;
  secondary?: boolean;
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
    setBusy(true);
    setError("");
    key.current ??= crypto.randomUUID();
    let opened = false;
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const { data } = await createApiClient().startStorySession(
        { storyKey },
        key.current,
        { headers: { "X-CSRF-Token": token } },
      );
      if (typeof data?.id !== "string" || !data.id)
        throw new Error("missing story session id");
      const href = `/stories/session/${encodeURIComponent(data.id)}`;
      // Confirmed: stay locked so no second start is sent. If the client
      // navigation stalls, the native link below opens the same session.
      opened = true;
      setOpenedHref(href);
      router.push(href);
    } catch (cause) {
      if (opened) return;
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
      <Button
        variant={secondary ? "secondary" : "primary"}
        type="button"
        className={secondary ? secondaryAction : primaryAction}
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
      </Button>
      {/* Recover independently if the automatic client navigation stalls. */}
      {openedHref && (
        <p className="mt-2">
          <a href={openedHref} className={textLink}>
            Open story
          </a>
        </p>
      )}
      {error && (
        <StatusMessage
          as="p"
          tone="error"
          role="alert"
          className="mt-3 text-neutral-900"
        >
          {error}
        </StatusMessage>
      )}
    </div>
  );
}
