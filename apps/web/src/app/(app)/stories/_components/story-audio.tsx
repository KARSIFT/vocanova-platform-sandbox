"use client";

import { StatusMessage } from "@/ui/status-message";
import { Button } from "@/ui/controls";
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
      <Button
        variant="secondary"
        type="button"
        className={secondaryAction}
        aria-label={active ? `Stop audio: ${label}` : `Play audio: ${label}`}
        aria-describedby={state.message ? description : undefined}
        onClick={() =>
          active
            ? deviceSpeech.stop(owner.current)
            : deviceSpeech.play(owner.current, text, false, setState)
        }
      >
        {active ? "Stop audio" : "Play line"}
      </Button>
      {state.message && (
        <StatusMessage
          as="p"
          id={description}
          role="status"
          className="mt-1 text-sm text-neutral-600"
        >
          {state.message}
        </StatusMessage>
      )}
    </div>
  );
}
