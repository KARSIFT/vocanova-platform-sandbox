"use client";

import { useId, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import type { MeaningKnowledgeUpdate } from "@vocanova/api-client";

import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";

const buttonStyle =
  "min-h-12 rounded-xl border border-neutral-300 px-4 py-3 font-semibold text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-50";

export function MeaningKnowledgeEditor({
  meaningId,
  initialKnown = false,
}: {
  meaningId: string;
  initialKnown?: boolean;
}) {
  const router = useRouter();
  const id = useId();
  const [open, setOpen] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [known, setKnown] = useState(initialKnown);
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const inFlight = useRef(false);
  const pending = useRef<{ body: MeaningKnowledgeUpdate; key: string } | null>(
    null,
  );

  async function load() {
    if (inFlight.current) return;
    inFlight.current = true;
    setOpen(true);
    setBusy(true);
    setError("");
    try {
      const { data } = await createApiClient().getMeaningKnowledge(meaningId);
      setKnown(data.selfReportedKnown);
      setNote(data.note);
      setLoaded(true);
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "Your personal word details could not be loaded. Try again.",
        ),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  async function save() {
    if (inFlight.current || !loaded) return;
    inFlight.current = true;
    // Retain the learner's exact change before session recovery, including on a
    // lost response. Retrying must never submit a new edit under the same key.
    pending.current ??= {
      body: { selfReportedKnown: known, note },
      key: crypto.randomUUID(),
    };
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const token = await getOrRefreshCSRFToken();
      if (!token) throw new Error("Session is not ready");
      const { data } = await createApiClient().updateMeaningKnowledge(
        meaningId,
        pending.current.body,
        pending.current.key,
        { headers: { "X-CSRF-Token": token } },
      );
      setKnown(data.selfReportedKnown);
      setNote(data.note);
      pending.current = null;
      setMessage("Your knowledge and note are saved.");
      router.refresh();
    } catch (cause) {
      setError(
        handleApiError(
          cause,
          "We could not confirm your changes. Retry to check and save the same update.",
        ),
      );
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  }

  const locked = busy || pending.current !== null;
  return (
    <section
      aria-labelledby={`${id}-heading`}
      className="mt-3 border-t border-neutral-200 pt-3"
    >
      <h3 id={`${id}-heading`} className="sr-only">
        Your knowledge and note
      </h3>
      {!open && initialKnown && (
        <p className="text-sm text-primary-700">Already known</p>
      )}
      {open && (
        <p className="mt-2 text-sm text-neutral-600">
          Keep a personal example or memory cue. Your note is private to your
          account.
        </p>
      )}
      {!open ? (
        <button
          type="button"
          className={`${buttonStyle} mt-3`}
          onClick={() => void load()}
        >
          Edit knowledge and note
        </button>
      ) : (
        <>
          {loaded && (
            <form
              className="mt-4 space-y-4"
              onSubmit={(event) => {
                event.preventDefault();
                void save();
              }}
              aria-busy={busy}
            >
              <label className="flex min-h-12 cursor-pointer items-center gap-3 rounded-xl bg-neutral-50 p-3 text-neutral-900">
                <input
                  type="checkbox"
                  checked={known}
                  disabled={locked}
                  onChange={(event) => {
                    setKnown(event.target.checked);
                    setMessage("");
                  }}
                  className="h-5 w-5 accent-primary-700"
                />
                I already know this meaning
              </label>
              <p className="text-sm text-neutral-600">
                This is your own assessment. You can still save and review the
                word; it does not award mastery or points.
              </p>
              <div>
                <label
                  htmlFor={`${id}-note`}
                  className="font-semibold text-neutral-900"
                >
                  My note
                </label>
                <textarea
                  id={`${id}-note`}
                  value={note}
                  disabled={locked}
                  onChange={(event) => {
                    setNote(event.target.value);
                    setMessage("");
                  }}
                  maxLength={2000}
                  rows={4}
                  placeholder="An example from your life, a translation, or a way to remember this meaning"
                  className="mt-2 w-full resize-y rounded-xl border border-neutral-300 bg-white p-3 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700 disabled:opacity-60"
                />
                <p className="mt-1 text-sm text-neutral-600">
                  {Array.from(note).length}/2,000 characters. Leave this blank
                  to remove your note.
                </p>
              </div>
              <button type="submit" disabled={busy} className={buttonStyle}>
                {busy
                  ? "Saving…"
                  : pending.current
                    ? "Retry saving changes"
                    : "Save my changes"}
              </button>
            </form>
          )}
          {!loaded && !busy && (
            <button
              type="button"
              onClick={() => void load()}
              className={`${buttonStyle} mt-3`}
            >
              Try loading again
            </button>
          )}
          {!loaded && busy && (
            <p role="status" className="mt-3 text-neutral-600">
              Loading your personal word details…
            </p>
          )}
        </>
      )}
      {error && (
        <p role="alert" className="mt-3 text-sm text-red-700">
          {error}
        </p>
      )}
      {message && (
        <p role="status" className="mt-3 text-sm font-medium text-primary-700">
          {message}
        </p>
      )}
    </section>
  );
}
