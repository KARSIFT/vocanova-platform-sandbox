"use client";

import Link from "@/ui/link";
import { useEffect, useId, useRef, useState } from "react";

import {
  deviceSpeech,
  pronunciationCaption,
  type PronunciationState,
} from "@/lib/speech";
import { SpeechSettings, useSpeechPreferences } from "@/ui/pronunciation";

import { secondaryAction, textLink } from "./practice-styles";

/** The authored speech text is used only for playback, never in DOM labels. */
export function PracticeAudio({
  text,
  meaningId,
  meaningRevision,
}: {
  text: string;
  meaningId?: string;
  meaningRevision?: string;
}) {
  const owner = useRef(Symbol("practice-audio"));
  const [state, setState] = useState<PronunciationState>({ status: "idle" });
  const { preferences, update } = useSpeechPreferences();
  const slow = preferences.rate === 0.7;
  const descriptionId = useId();
  const active = state.status === "starting" || state.status === "playing";

  useEffect(() => {
    const currentOwner = owner.current;
    deviceSpeech.prepare();
    setState({ status: "idle" });
    return () => deviceSpeech.release(currentOwner);
  }, [text, meaningId, meaningRevision]);

  return (
    <div className="my-5" role="group" aria-label="Practice audio">
      {meaningId && <span className="sr-only">AI-generated pronunciation</span>}
      <div className="flex flex-wrap gap-3">
        <button
          type="button"
          className={secondaryAction}
          aria-describedby={descriptionId}
          onClick={() => {
            if (active) deviceSpeech.stop(owner.current);
            else
              deviceSpeech.play(
                owner.current,
                text,
                false,
                setState,
                meaningId,
                meaningRevision,
              );
          }}
        >
          {active ? "Stop audio" : "Play audio"}
        </button>
        <button
          type="button"
          className={`${secondaryAction} aria-pressed:border-primary-600 aria-pressed:bg-primary-50 aria-pressed:text-primary-800`}
          aria-pressed={slow}
          disabled={active}
          onClick={() => update({ rate: slow ? 0.85 : 0.7 })}
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
          `${pronunciationCaption(state)}. Play it again whenever you need.`}
      </p>
      <SpeechSettings disabled={active} />
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
