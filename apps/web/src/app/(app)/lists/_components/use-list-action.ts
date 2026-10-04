"use client";

import { useRef, useState } from "react";
import { ApiResponseError } from "@vocanova/api-client";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError, isSessionExpiredError } from "@/lib/session";

export async function listWriteHeaders() {
  const token = await getOrRefreshCSRFToken();
  if (!token) throw new Error("Session is not ready");
  return { headers: { "X-CSRF-Token": token } };
}

// Keep the exact closure (body, revision and identity) until confirmed or a
// successful current-state reload. Never replay a changed edit under its key.
export function useListAction() {
  const inFlight = useRef(false);
  const expired = useRef(false);
  const pending = useRef<(() => Promise<string>) | null>(null);
  const [busy, setBusy] = useState(false);
  const [needsRetry, setNeedsRetry] = useState(false);
  const [needsRefresh, setNeedsRefresh] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  function failed(cause: unknown, writing: boolean) {
    if (isSessionExpiredError(cause)) expired.current = true;
    const status = cause instanceof ApiResponseError ? cause.status : 0;
    if (writing && (status === 409 || status === 404)) {
      setNeedsRefresh(true);
      setNeedsRetry(false);
      setError(
        status === 404
          ? "This list may have been deleted. Load its current status before making another change."
          : "This list changed elsewhere. Load the latest list, then choose your change again.",
      );
      return;
    }
    const rejected = status === 400 || status === 422 || status === 403;
    if (writing && rejected) pending.current = null;
    if (writing) setNeedsRetry(!rejected && !expired.current);
    setError(
      handleApiError(
        cause,
        writing
          ? "We could not confirm your change. Retry safely with the same request, or load current status."
          : "We could not load your lists. Please try again.",
      ),
    );
  }
  async function run(task?: () => Promise<string>) {
    if (inFlight.current || expired.current || needsRefresh) return;
    pending.current ??= task ?? null;
    if (!pending.current) return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const result = await pending.current();
      pending.current = null;
      setNeedsRetry(false);
      setNeedsRefresh(false);
      setMessage(result);
    } catch (cause) {
      failed(cause, true);
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }
  async function reload(task: () => Promise<void>) {
    if (inFlight.current || expired.current) return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await task();
      pending.current = null;
      setNeedsRetry(false);
      setNeedsRefresh(false);
    } catch (cause) {
      failed(cause, false);
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }
  return {
    busy,
    error,
    message,
    needsRetry,
    needsRefresh,
    locked: busy || needsRetry || needsRefresh || expired.current,
    run,
    reload,
  };
}
