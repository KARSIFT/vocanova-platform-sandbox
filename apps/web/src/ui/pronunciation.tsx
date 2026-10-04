"use client";

import { useEffect, useId, useRef, useState } from "react";

import { deviceSpeech, type PronunciationState } from "@/lib/speech";

interface ListenButtonProps {
  /** Authored vocabulary or examples only; never pass a learner's draft. */
  text: string;
  label?: string;
  showCaption?: boolean;
}

export function ListenButton({
  text,
  label = text,
  showCaption = true,
}: ListenButtonProps) {
  const owner = useRef(Symbol("pronunciation"));
  const [state, setState] = useState<PronunciationState>({ status: "idle" });
  const [slow, setSlow] = useState(false);
  const descriptionId = useId();
  const active = state.status === "starting" || state.status === "playing";

  useEffect(() => {
    const currentOwner = owner.current;
    deviceSpeech.prepare();
    setState({ status: "idle" });
    return () => deviceSpeech.release(currentOwner);
  }, [text]);

  return (
    <div
      className="mt-[var(--spacing-xs)]"
      role="group"
      aria-label={`Pronunciation of ${label}`}
    >
      <div className="flex flex-wrap items-center gap-[var(--spacing-xs)]">
        <button
          type="button"
          aria-label={
            active ? `Stop listening to ${label}` : `Listen to ${label}`
          }
          aria-describedby={
            showCaption || state.message ? descriptionId : undefined
          }
          onClick={() => {
            if (active) deviceSpeech.stop(owner.current);
            else deviceSpeech.play(owner.current, text, slow, setState);
          }}
          className="inline-flex min-h-11 min-w-11 items-center gap-[var(--spacing-xs)] rounded-md px-[var(--spacing-sm)] text-sm font-semibold text-primary-700 hover:bg-primary-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          <svg
            aria-hidden="true"
            viewBox="0 0 24 24"
            className="h-4 w-4 fill-none stroke-current stroke-2"
          >
            {active ? (
              <path d="M7 7h10v10H7z" />
            ) : (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="m11 5-5 4H3v6h3l5 4V5Zm4 4a5 5 0 0 1 0 6m3-9a9 9 0 0 1 0 12"
              />
            )}
          </svg>
          {active ? "Stop" : "Listen"}
        </button>
        <button
          type="button"
          aria-label={`Slow pronunciation of ${label}`}
          aria-pressed={slow}
          disabled={active}
          onClick={() => setSlow((previous) => !previous)}
          className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded-md border border-neutral-200 px-[var(--spacing-sm)] text-sm ${slow ? "text-primary-800" : "text-neutral-700"} hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 aria-pressed:border-primary-600 aria-pressed:bg-primary-50 disabled:cursor-default disabled:opacity-60`}
        >
          Slow
        </button>
      </div>
      {showCaption || state.message ? (
        <p
          id={descriptionId}
          role={state.message ? "status" : undefined}
          className="text-xs text-neutral-600"
        >
          {state.message ?? "Device pronunciation"}
        </p>
      ) : null}
    </div>
  );
}
