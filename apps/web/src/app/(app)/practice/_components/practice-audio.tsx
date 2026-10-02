"use client";

import Link from "next/link";
import { useEffect, useId, useRef, useState } from "react";

import { deviceSpeech, type PronunciationState } from "@/lib/speech";

import { secondaryAction, textLink } from "./practice-styles";

/** The authored speech text is used only for playback, never in DOM labels. */
export function PracticeAudio({ text }: { text: string }) {
  const owner = useRef(Symbol("practice-audio"));
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
    <div className="my-5" role="group" aria-label="Practice audio">
      <div className="flex flex-wrap gap-3">
        <button
          type="button"
          className={secondaryAction}
          aria-describedby={descriptionId}
          onClick={() => {
            if (active) deviceSpeech.stop(owner.current);
            else deviceSpeech.play(owner.current, text, slow, setState);
          }}
        >
          {active ? "Stop audio" : "Play audio"}
        </button>
        <button
          type="button"
          className={`${secondaryAction} aria-pressed:border-primary-600 aria-pressed:bg-primary-50 aria-pressed:text-primary-800`}
          aria-pressed={slow}
          disabled={active}
          onClick={() => setSlow((previous) => !previous)}
        >
          Slower audio
        </button>
      </div>
      <p
        id={descriptionId}
        role={state.message ? "status" : undefined}
        className="mt-2 text-sm text-neutral-600"
      >
        {state.message ??
          "Device pronunciation. Play it again whenever you need."}
      </p>
      {(state.status === "unavailable" || state.status === "error") && (
        <div className="mt-3 rounded-xl bg-secondary-50 p-4 text-neutral-800">
          <p>
            You can try the audio again or choose typed recall instead. Your
            saved session will still be here.
          </p>
          <Link href="/practice" className={`${textLink} mt-2`}>
            Choose another practice mode
          </Link>
        </div>
      )}
    </div>
  );
}
