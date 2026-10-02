"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import {
  ApiResponseError,
  type VocabularySearchItem,
} from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { ListenButton } from "@/ui/pronunciation";
import { Surface } from "@/ui/surface";

const primaryAction =
  "inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-3 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-60 sm:px-5";
const secondaryAction =
  "inline-flex min-h-12 items-center justify-center rounded-xl border border-neutral-300 bg-white px-3 py-3 font-semibold text-primary-700 hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-60 sm:px-5";
const textLink =
  "inline-flex min-h-11 items-center font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-60";
type Choice = "known" | "learn";
type PendingChoice = { meaningId: string; choice: Choice; key: string };

export function VocabularySelfCheck({
  items,
}: {
  items: VocabularySearchItem[] | null;
}) {
  const [index, setIndex] = useState(0);
  const [counts, setCounts] = useState({ known: 0, learn: 0, skipped: 0 });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [needsRetry, setNeedsRetry] = useState(false);
  const [needsReload, setNeedsReload] = useState(false);
  const pending = useRef<PendingChoice | null>(null);
  const inFlight = useRef(false);
  const heading = useRef<HTMLHeadingElement>(null);
  const item = items?.[index];
  const locked = busy || needsRetry || needsReload;

  useEffect(() => {
    if (index > 0) heading.current?.focus();
  }, [index]);

  function advance(choice: Choice | "skipped") {
    pending.current = null;
    setNeedsRetry(false);
    setError("");
    setCounts((previous) => ({ ...previous, [choice]: previous[choice] + 1 }));
    setIndex((previous) => previous + 1);
  }

  async function choose(choice?: Choice) {
    if (inFlight.current || needsReload || !item) return;
    if (!pending.current) {
      if (!choice) return;
      // Capture identity before CSRF recovery or any asynchronous work.
      pending.current = {
        meaningId: item.meaningId,
        choice,
        key: crypto.randomUUID(),
      };
    }
    const request = pending.current;
    inFlight.current = true;
    setBusy(true);
    setError("");
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("session not ready");
      const client = createApiClient();
      const init = { headers: { "X-CSRF-Token": token } };
      const confirmed =
        request.choice === "known"
          ? (
              await client.setMeaningKnown(
                request.meaningId,
                true,
                request.key,
                init,
              )
            ).data
          : (
              await client.saveUserWord(
                { meaningId: request.meaningId, source: "search" },
                request.key,
                init,
              )
            ).data;
      if (
        confirmed.meaningId !== request.meaningId ||
        ("selfReportedKnown" in confirmed
          ? !confirmed.selfReportedKnown
          : !confirmed.saved)
      ) {
        setNeedsReload(true);
        setNeedsRetry(false);
        setError(
          "This word changed in another request. Reload the check to see your current choices.",
        );
        return;
      }
      advance(request.choice);
    } catch (cause) {
      if (cause instanceof ApiResponseError && cause.status === 409) {
        setNeedsReload(true);
        setNeedsRetry(false);
        setError(
          "This choice could not be confirmed. Reload the check to see your current words.",
        );
      } else {
        setNeedsRetry(true);
        setError(
          handleApiError(
            cause,
            "We could not confirm your choice. Retry it safely before moving on.",
          ),
        );
      }
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  if (!items)
    return (
      <Surface aria-labelledby="check-unavailable-heading">
        <h2
          id="check-unavailable-heading"
          className="text-xl font-bold text-neutral-900"
        >
          Your starting words are unavailable
        </h2>
        <p role="status" className="mt-3 text-neutral-700">
          We could not load your words. Try again to start the check.
        </p>
        <button
          type="button"
          className={`${primaryAction} mt-4`}
          onClick={() => window.location.reload()}
        >
          Reload the check
        </button>
      </Surface>
    );

  if (!item)
    return (
      <Surface aria-labelledby="check-complete-heading" tone="primary">
        <h2
          id="check-complete-heading"
          tabIndex={-1}
          ref={heading}
          className="text-2xl font-bold text-neutral-900"
        >
          {items.length
            ? "Your starting words are ready"
            : "You have explored these words"}
        </h2>
        <p className="mt-3 text-neutral-700">
          {items.length
            ? `In this check, you marked ${counts.known} ${counts.known === 1 ? "meaning" : "meanings"} as already known and saved ${counts.learn} to learn.`
            : "There are no unexplored meanings available right now. Return to your saved words or choose a lesson."}
        </p>
        {counts.skipped > 0 && (
          <p className="mt-2 text-sm text-neutral-600">
            You skipped {counts.skipped} for now. Skipped words can appear next
            time.
          </p>
        )}
        <p className="mt-2 text-sm text-neutral-600">
          Your choices are a starting point, not a test score or an estimate of
          your English level.
        </p>
        <div className="mt-5 flex flex-wrap gap-3">
          <Link href="/plan" className={primaryAction}>
            See your learning plan
          </Link>
          <Link href="/words" className={secondaryAction}>
            Open saved vocabulary
          </Link>
          <Link href="/vocabulary?knowledge=known" className={textLink}>
            Review your known words
          </Link>
        </div>
      </Surface>
    );

  return (
    <Surface aria-labelledby="check-word-heading">
      <p className="text-sm font-semibold text-primary-700">
        Meaning {index + 1} of {items.length}
      </p>
      <h2
        id="check-word-heading"
        ref={heading}
        tabIndex={-1}
        className="mt-3 break-words text-3xl font-bold text-neutral-900"
      >
        {item.wordText}
      </h2>
      <p className="mt-1 text-sm text-neutral-600">{item.partOfSpeech}</p>
      <p className="mt-4 text-lg text-neutral-900">{item.shortDefinition}</p>
      <ListenButton key={item.meaningId} text={item.wordText} />
      <div className="mt-4 grid grid-cols-2 gap-3">
        <button
          type="button"
          disabled={locked}
          className={`${secondaryAction} min-w-0 text-sm sm:text-base`}
          onClick={() => void choose("known")}
        >
          Already know
        </button>
        <button
          type="button"
          disabled={locked}
          className={`${primaryAction} min-w-0 text-sm sm:text-base`}
          onClick={() => void choose("learn")}
        >
          Want to learn
        </button>
        <button
          type="button"
          disabled={locked}
          className={`${textLink} col-span-2 justify-self-center`}
          onClick={() => advance("skipped")}
        >
          Skip for now
        </button>
      </div>
      <p className="mt-3 text-sm text-neutral-600">
        “Already know” records your own assessment. “Want to learn” saves this
        meaning for practice. Skipping changes nothing.
      </p>
      {busy && (
        <p role="status" className="mt-3 text-neutral-700">
          Saving your choice…
        </p>
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
                onClick={() => void choose()}
              >
                Retry choice
              </button>
            )}
            <button
              type="button"
              disabled={busy}
              onClick={() => window.location.reload()}
              className={textLink}
            >
              Reload the check
            </button>
          </div>
        </div>
      )}
    </Surface>
  );
}
