"use client";

import { useEffect, useId, useRef, useState } from "react";
import { deviceSpeech, type PronunciationState } from "@/lib/speech";
import { secondaryAction } from "../../practice/_components/practice-styles";

export function StoryAudio({ text, label }: { text: string; label: string }) {
  const owner = useRef(Symbol("story-audio"));
  const [state, setState] = useState<PronunciationState>({ status: "idle" });
  const description = useId();
  const active = state.status === "starting" || state.status === "playing";
  useEffect(() => {
    const id = owner.current;
    deviceSpeech.prepare();
    return () => deviceSpeech.release(id);
  }, [text]);
  return (
    <div className="mt-2">
      <button
        type="button"
        className={secondaryAction}
        aria-label={active ? `Stop audio: ${label}` : `Play audio: ${label}`}
        aria-describedby={description}
        onClick={() =>
          active
            ? deviceSpeech.stop(owner.current)
            : deviceSpeech.play(owner.current, text, false, setState)
        }
      >
        {active ? "Stop audio" : "Play line"}
      </button>
      <p
        id={description}
        role={state.message ? "status" : undefined}
        className="mt-1 text-sm text-neutral-600"
      >
        {state.message ?? "Optional device audio. You can read the line above."}
      </p>
    </div>
  );
}
