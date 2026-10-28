"use client";

import { useEffect, useId, useRef, useState } from "react";

import {
  availableEnglishVoices,
  DEFAULT_SPEECH_PREFERENCES,
  deviceSpeech,
  pronunciationCaption,
  readSpeechPreferences,
  saveSpeechPreferences,
  subscribeSpeechPreferences,
  type PronunciationState,
  type SpeechPreferences,
} from "@/lib/speech";

/** Preferences contain voice/speed only, never the authored or learner text. */
export function useSpeechPreferences() {
  const [preferences, setPreferences] = useState<SpeechPreferences>(
    DEFAULT_SPEECH_PREFERENCES,
  );
  useEffect(() => {
    const refresh = () => setPreferences(readSpeechPreferences());
    refresh();
    return subscribeSpeechPreferences(refresh);
  }, []);
  function update(patch: Partial<SpeechPreferences>) {
    setPreferences(saveSpeechPreferences(patch));
  }
  return { preferences, update };
}

export function SpeechSettings({ disabled = false }: { disabled?: boolean }) {
  const { preferences, update } = useSpeechPreferences();
  const [voices, setVoices] = useState<SpeechSynthesisVoice[]>([]);
  const id = useId();
  useEffect(() => {
    const refresh = () => setVoices(availableEnglishVoices());
    refresh();
    window.speechSynthesis?.addEventListener?.("voiceschanged", refresh);
    return () =>
      window.speechSynthesis?.removeEventListener?.("voiceschanged", refresh);
  }, []);

  return (
    <details className="min-w-0 max-w-full text-sm text-neutral-700">
      <summary className="min-h-11 cursor-pointer content-center rounded-full px-3 font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
        Voice &amp; speed
      </summary>
      <div className="space-y-3 rounded-xl border border-neutral-200 p-3">
        <div>
          <label htmlFor={`${id}-speed`} className="mb-1 block font-medium">
            Playback speed
          </label>
          <select
            id={`${id}-speed`}
            value={preferences.rate}
            disabled={disabled}
            onChange={(event) => update({ rate: Number(event.target.value) })}
            className="min-h-11 w-full rounded-md border border-neutral-300 bg-white px-2 text-neutral-900"
          >
            <option value={0.7}>Extra slow</option>
            <option value={0.85}>Learning pace</option>
            <option value={1}>Normal pace</option>
          </select>
        </div>
        <div>
          <label htmlFor={`${id}-voice`} className="mb-1 block font-medium">
            English voice
          </label>
          <select
            id={`${id}-voice`}
            value={preferences.voiceURI}
            disabled={
              disabled || (voices.length === 0 && !preferences.voiceURI)
            }
            onChange={(event) => update({ voiceURI: event.target.value })}
            className="min-h-11 w-full rounded-md border border-neutral-300 bg-white px-2 text-neutral-900"
          >
            <option value="">Natural voice when available</option>
            {preferences.voiceURI &&
            !voices.some((voice) => voice.voiceURI === preferences.voiceURI) ? (
              <option value={preferences.voiceURI}>
                Previous voice unavailable
              </option>
            ) : null}
            {voices.map((voice) => (
              <option
                key={`${voice.voiceURI}-${voice.lang}`}
                value={voice.voiceURI}
              >
                {voice.name} ({voice.lang})
              </option>
            ))}
          </select>
        </div>
        <p className="text-xs text-neutral-600">
          Meaning-specific audio uses AI pronunciation. Speed applies to all
          audio.
        </p>
      </div>
    </details>
  );
}

interface ListenButtonProps {
  /** Authored vocabulary or examples only; never pass a learner's draft. */
  text: string;
  meaningId?: string;
  meaningRevision?: string;
  label?: string;
  showCaption?: boolean;
  autoPlay?: boolean;
}

export function ListenButton({
  text,
  meaningId,
  meaningRevision,
  label = text,
  showCaption = true,
  autoPlay = false,
}: ListenButtonProps) {
  const owner = useRef(Symbol("pronunciation"));
  const [state, setState] = useState<PronunciationState>({ status: "idle" });
  const { preferences, update } = useSpeechPreferences();
  const slow = preferences.rate === 0.7;
  const descriptionId = useId();
  const active = state.status === "starting" || state.status === "playing";

  useEffect(() => {
    const currentOwner = owner.current;
    deviceSpeech.prepare();
    setState({ status: "idle" });
    if (autoPlay)
      deviceSpeech.play(
        currentOwner,
        text,
        false,
        setState,
        meaningId,
        meaningRevision,
      );
    return () => deviceSpeech.release(currentOwner);
  }, [text, meaningId, meaningRevision, autoPlay]);

  return (
    <div
      className="mt-[var(--spacing-xs)]"
      role="group"
      aria-label={`Pronunciation of ${label}`}
    >
      {meaningId && <span className="sr-only">AI-generated pronunciation</span>}
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
          className="inline-flex min-h-11 min-w-11 items-center gap-[var(--spacing-xs)] rounded-full border border-primary-200 bg-primary-50 px-[var(--spacing-md)] text-sm font-semibold text-primary-800 hover:border-primary-300 hover:bg-primary-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
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
          onClick={() => update({ rate: slow ? 0.85 : 0.7 })}
          className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded-full border border-neutral-200 px-[var(--spacing-md)] text-sm ${slow ? "text-primary-800" : "text-neutral-700"} hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 aria-pressed:border-primary-600 aria-pressed:bg-primary-50 disabled:cursor-default disabled:opacity-60`}
        >
          Slow
        </button>
        <SpeechSettings disabled={active} />
      </div>
      {showCaption || state.message ? (
        <p
          id={descriptionId}
          role={state.message ? "status" : undefined}
          className="text-xs text-neutral-600"
        >
          {state.message ?? pronunciationCaption(state)}
        </p>
      ) : null}
    </div>
  );
}
