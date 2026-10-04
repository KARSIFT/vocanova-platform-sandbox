"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";
import { ApiResponseError, type WordMeaning } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError, isSessionExpiredError } from "@/lib/session";
import { formatWordReviewState } from "../../discover/[situation]/[word]/_components/word-review-state";

const actionStyle =
  "inline-flex min-h-12 items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 py-3 font-semibold text-primary-700 hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-60";
const linkStyle =
  "inline-flex min-h-11 items-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

function safeError(cause: unknown, fallback: string) {
  return isSessionExpiredError(cause)
    ? handleApiError(cause, fallback)
    : fallback;
}

export function LessonReviewSave({
  meaningId,
  wordSlug,
  wordText,
}: {
  meaningId: string;
  wordSlug: string;
  wordText: string;
}) {
  const [meaning, setMeaning] = useState<WordMeaning | null>(null);
  const [busy, setBusy] = useState(true);
  const [needsRetry, setNeedsRetry] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const mounted = useRef(false);
  const inFlight = useRef(false);
  const currentRead = useRef<AbortController | null>(null);
  const pending = useRef<{
    body: { meaningId: string; source: "journey" };
    key: string;
  } | null>(null);

  const readCurrentMeaning = useCallback(async () => {
    const controller = new AbortController();
    currentRead.current = controller;
    const { data } = await createApiClient().getCanonicalWord(wordSlug, {
      signal: controller.signal,
      cache: "no-store",
    });
    const current = data.word.meanings.find((item) => item.id === meaningId);
    if (
      data.word.slug !== wordSlug ||
      !current ||
      (current.saved && !current.userWordId)
    ) {
      throw new Error("Current meaning state unavailable");
    }
    return current;
  }, [meaningId, wordSlug]);

  const checkSavedState = useCallback(async () => {
    // Returning from another tab may refresh a confirmed state. It must not
    // replace the identity of a save whose response is still uncertain.
    if (inFlight.current || pending.current) return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const current = await readCurrentMeaning();
      if (mounted.current) setMeaning(current);
    } catch (cause) {
      if (mounted.current) {
        setMeaning(null);
        setError(
          safeError(
            cause,
            "We could not check whether this meaning is saved. Check its status before saving.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      if (mounted.current) setBusy(false);
    }
  }, [readCurrentMeaning]);

  useEffect(() => {
    mounted.current = true;
    void checkSavedState();
    const refresh = () => void checkSavedState();
    const onVisible = () => {
      if (document.visibilityState === "visible") refresh();
    };
    window.addEventListener("focus", refresh);
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      mounted.current = false;
      currentRead.current?.abort();
      window.removeEventListener("focus", refresh);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, [checkSavedState]);

  async function save() {
    if (inFlight.current || !meaning || meaning.saved) return;
    // Capture the explicit meaning and retry key before session preparation.
    // A new key is created only by a new learner action after confirmation.
    pending.current ??= {
      body: { meaningId, source: "journey" },
      key: crypto.randomUUID(),
    };
    const request = pending.current;
    inFlight.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    let responseReceived = false;
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("Session not ready");
      const { data } = await createApiClient().saveUserWord(
        request.body,
        request.key,
        { headers: { "X-CSRF-Token": token } },
      );
      responseReceived = true;
      pending.current = null;
      if (data.meaningId !== request.body.meaningId) {
        throw new Error("Save response identity unavailable");
      }
      // Reading the canonical overlay keeps known status and scheduling honest.
      // A removed saved-word replay is handled as a definitive 404 below.
      const current = await readCurrentMeaning();
      if (!mounted.current) return;
      setMeaning(current);
      setNeedsRetry(false);
      if (!current.saved) {
        setNotice(
          "This meaning is not saved. Your earlier save was not reapplied. Choose Save for review if you want to add it again.",
        );
      }
    } catch (cause) {
      if (!mounted.current) return;
      const conflict =
        cause instanceof ApiResponseError && cause.status === 409;
      if (cause instanceof ApiResponseError && cause.status === 404) {
        pending.current = null;
        setNeedsRetry(false);
        try {
          const current = await readCurrentMeaning();
          if (!mounted.current) return;
          setMeaning(current);
          setNotice(
            current.saved
              ? "The current saved meaning is up to date."
              : "This meaning is not saved. Your earlier save was not reapplied. Choose Save for review if you want to add it again.",
          );
        } catch (readError) {
          if (!mounted.current) return;
          setMeaning(null);
          setError(
            safeError(
              readError,
              "We could not confirm the current saved status. Check it before choosing whether to save again.",
            ),
          );
        }
      } else if (responseReceived || conflict) {
        pending.current = null;
        setMeaning(null);
        setNeedsRetry(false);
        setError(
          safeError(
            cause,
            "We could not confirm the current saved status. Check it before choosing whether to save again.",
          ),
        );
      } else {
        setNeedsRetry(true);
        setError(
          safeError(
            cause,
            "We could not confirm this save. Retry safely to check the same request.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      if (mounted.current) setBusy(false);
    }
  }

  const reviewState = meaning?.saved
    ? formatWordReviewState(meaning.reviewState, meaning.due)
    : null;

  return (
    <div className="mt-3 space-y-2">
      {meaning?.selfReportedKnown && (
        <p className="text-sm text-neutral-600">
          Marked as already known. Saving for review is optional.
        </p>
      )}
      {meaning?.saved ? (
        <div>
          <p className="font-semibold text-neutral-900">
            Saved in your vocabulary{reviewState ? ` · ${reviewState}` : ""}
          </p>
          <Link
            href={`/words/${encodeURIComponent(meaning.userWordId!)}`}
            className={linkStyle}
            aria-label={`Open saved ${wordText}`}
          >
            Open saved word
          </Link>
        </div>
      ) : meaning ? (
        <button
          type="button"
          className={actionStyle}
          disabled={busy || needsRetry}
          aria-busy={busy}
          aria-label={`Save ${wordText} for review`}
          onClick={() => void save()}
        >
          {busy && pending.current ? "Saving…" : "Save for review"}
        </button>
      ) : null}
      {busy && !pending.current && (
        <p role="status" className="text-sm text-neutral-600">
          Checking saved status…
        </p>
      )}
      {notice && (
        <p role="status" className="text-sm text-neutral-700">
          {notice}
        </p>
      )}
      {error && (
        <div
          role="alert"
          className="rounded-xl border border-secondary-300 bg-secondary-50 p-3 text-neutral-900"
        >
          <p className="text-sm">{error}</p>
          <button
            type="button"
            className={`${actionStyle} mt-2`}
            disabled={busy}
            onClick={() => void (needsRetry ? save() : checkSavedState())}
          >
            {needsRetry ? "Retry save" : "Check saved status"}
          </button>
        </div>
      )}
    </div>
  );
}
