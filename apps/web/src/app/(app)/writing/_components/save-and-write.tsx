"use client";

import { useEffect, useRef, useState } from "react";
import { ApiResponseError, type WordMeaning } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError, isSessionExpiredError } from "@/lib/session";
import { SentenceFeedback } from "../../_components/sentence-feedback";

const actionStyle =
  "inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-600 px-5 py-3 font-semibold text-white hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-60";

export function SaveAndWrite({
  meaningId,
  wordSlug,
  wordText,
  shortDefinition,
  initialMeaning,
  userId,
}: {
  meaningId: string;
  wordSlug: string;
  wordText: string;
  shortDefinition?: string;
  initialMeaning: WordMeaning;
  /** Draft recovery stays disabled when the API omits its stable user id. */
  userId?: string;
}) {
  const [meaning, setMeaning] = useState<WordMeaning | null>(
    initialMeaning.saved && !initialMeaning.userWordId ? null : initialMeaning,
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const mounted = useRef(false);
  const inFlight = useRef(false);
  const pending = useRef<{
    key: string;
    body: { meaningId: string; source: "journey" };
  } | null>(null);
  const readController = useRef<AbortController | null>(null);
  const focusAfterSave = useRef(false);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      readController.current?.abort();
    };
  }, []);

  useEffect(() => {
    if (
      focusAfterSave.current &&
      meaning?.saved &&
      meaning.userWordId &&
      !busy
    ) {
      focusAfterSave.current = false;
      document.getElementById(`sentence-input-${meaning.userWordId}`)?.focus();
    }
  }, [meaning, busy]);

  function safeError(cause: unknown, fallback: string) {
    return isSessionExpiredError(cause)
      ? handleApiError(cause, fallback)
      : fallback;
  }

  async function readMeaning() {
    const controller = new AbortController();
    readController.current = controller;
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
      throw new Error("Saved meaning could not be confirmed");
    }
    return current;
  }

  async function checkSavedState() {
    if (inFlight.current || pending.current) return;
    inFlight.current = true;
    setBusy(true);
    setError("");
    try {
      const current = await readMeaning();
      if (mounted.current) setMeaning(current);
    } catch (cause) {
      if (mounted.current)
        setError(
          safeError(
            cause,
            "We could not confirm the saved meaning. Check its status again.",
          ),
        );
    } finally {
      inFlight.current = false;
      if (mounted.current) setBusy(false);
    }
  }

  async function save() {
    if (inFlight.current || !meaning || meaning.saved) return;
    focusAfterSave.current = true;
    // An uncertain response is retried with the same body and request identity.
    // There is no save on mount, topic selection, or feedback submission.
    pending.current ??= {
      key: crypto.randomUUID(),
      body: { meaningId, source: "journey" },
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
        {
          headers: { "X-CSRF-Token": token },
        },
      );
      responseReceived = true;
      pending.current = null;
      if (data.meaningId !== meaningId)
        throw new Error("Save response did not match");
      const current = await readMeaning();
      if (!mounted.current) return;
      setMeaning(current);
      if (!current.saved)
        setNotice(
          "This meaning is no longer saved. Choose Save and write to add it again.",
        );
    } catch (cause) {
      if (!mounted.current) return;
      if (
        cause instanceof ApiResponseError &&
        (cause.status === 404 || cause.status === 409)
      ) {
        pending.current = null;
        try {
          const current = await readMeaning();
          if (!mounted.current) return;
          setMeaning(current);
          if (!current.saved)
            setNotice(
              "This meaning is not saved. Your earlier request was not reapplied. Choose Save and write to add it.",
            );
        } catch (readError) {
          if (!mounted.current) return;
          setMeaning(null);
          setError(
            safeError(
              readError,
              "We could not confirm the saved meaning. Check its status again.",
            ),
          );
        }
      } else if (responseReceived) {
        setMeaning(null);
        setError(
          safeError(
            cause,
            "Your save responded, but we could not confirm its current status. Check saved status to continue.",
          ),
        );
      } else {
        setError(
          safeError(
            cause,
            "We could not confirm this save. Retry safely with the same request.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      if (mounted.current) setBusy(false);
    }
  }

  if (meaning?.saved && meaning.userWordId) {
    return (
      <SentenceFeedback
        targetWord={wordText}
        attemptId={meaning.userWordId}
        source="word_detail"
        userId={userId}
        shortDefinition={shortDefinition}
      />
    );
  }
  return (
    <div className="mt-4 rounded-xl bg-neutral-50 p-4">
      <p className="max-w-prose text-neutral-700">
        Save this meaning to write with it here. It also joins your saved
        vocabulary for review.
      </p>
      {meaning && !pending.current && (
        <button
          type="button"
          onClick={() => void save()}
          disabled={busy}
          aria-busy={busy}
          className={`${actionStyle} mt-4`}
        >
          {busy ? "Saving…" : "Save and write"}
        </button>
      )}
      {busy && (
        <p role="status" className="mt-3 text-sm text-neutral-600">
          Confirming saved meaning…
        </p>
      )}
      {notice && (
        <p role="status" className="mt-3 text-sm text-neutral-700">
          {notice}
        </p>
      )}
      {error && (
        <div role="alert" className="mt-3 space-y-3 text-sm text-red-700">
          <p>{error}</p>
          <button
            type="button"
            disabled={busy}
            onClick={() => void (pending.current ? save() : checkSavedState())}
            className={actionStyle}
          >
            {pending.current ? "Retry save" : "Check saved status"}
          </button>
        </div>
      )}
      {!meaning && !error && (
        <button
          type="button"
          disabled={busy}
          onClick={() => void checkSavedState()}
          className={`${actionStyle} mt-3`}
        >
          Check saved status
        </button>
      )}
    </div>
  );
}
