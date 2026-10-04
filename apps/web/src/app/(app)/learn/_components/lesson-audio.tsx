"use client";

import { ListenButton } from "@/ui/pronunciation";

/** Playback never starts automatically. A transcript keeps the lesson usable without audio. */
export function LessonAudio({ text }: { text: string }) {
  return (
    <div className="mb-5 rounded-xl border border-neutral-200 p-4">
      <ListenButton text={text} label="lesson audio" />
      <details className="mt-3 text-neutral-700">
        <summary className="min-h-11 cursor-pointer py-3 font-semibold focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-700">
          Show audio text
        </summary>
        <p className="mt-2 text-lg">{text}</p>
        <p className="mt-2 text-sm">
          Use this help if you cannot play or hear the audio. This lesson
          records first answers, not unaided listening ability.
        </p>
      </details>
    </div>
  );
}
