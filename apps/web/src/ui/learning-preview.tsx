"use client";

import { useState } from "react";
import { MeaningPicture } from "./meaning-picture";

/** Original, authored examples only. This demo does not evaluate user writing. */
export function LearningPreview() {
  const [showFeedback, setShowFeedback] = useState(false);
  return (
    <section
      aria-label="Learning example"
      className="mx-auto w-full max-w-[31rem] overflow-hidden rounded-2xl border border-neutral-200 bg-white p-5 shadow-[0_12px_35px_rgb(30_41_59_/_0.08)] sm:p-6"
    >
      <div className="flex items-center justify-between gap-3 border-b border-neutral-200 pb-3">
        <p className="text-sm font-semibold text-primary-700">
          Try a learning example
        </p>
        <span className="text-sm text-neutral-500">Daily conversation</span>
      </div>
      <div aria-live="polite" className="mt-4">
        {showFeedback ? (
          <>
            <h2 className="text-2xl font-bold tracking-tight text-neutral-900">
              Make it your own.
            </h2>
            <p className="mt-3 text-sm font-semibold text-neutral-600">
              Example sentence
            </p>
            <p className="mt-1 text-lg text-neutral-900">
              I want invite my friend.
            </p>
            <div className="mt-4 border-l-4 border-primary-300 bg-primary-50 px-4 py-3">
              <p className="text-sm font-semibold text-primary-800">
                Suggested sentence
              </p>
              <p className="mt-1 text-lg text-neutral-900">
                I want{" "}
                <strong className="underline decoration-primary-300 underline-offset-4">
                  to invite
                </strong>{" "}
                my friend.
              </p>
            </div>
            <p className="mt-4 text-neutral-700">
              Use “to” before “invite” after “want”.
            </p>
            <p className="mt-4 text-xs text-neutral-500">
              Example feedback for the sentence shown.
            </p>
          </>
        ) : (
          <>
            <h2 className="text-3xl font-bold tracking-tight text-neutral-900">
              invite
            </h2>
            <p className="mt-2 text-neutral-700">
              To ask someone to come to an event or do something with you.
            </p>
            <MeaningPicture
              meaningId="ee53d6ba-4303-5394-b7f9-79937ce66d09"
              priority
            />
            <p className="mt-4 border-l-4 border-primary-300 pl-3 text-lg text-neutral-900">
              I want to invite you to my party.
            </p>
          </>
        )}
      </div>
      <button
        type="button"
        aria-pressed={showFeedback}
        onClick={() => setShowFeedback(!showFeedback)}
        className="mt-5 inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800"
      >
        {showFeedback ? "Back to the word" : "See example feedback"}
      </button>
    </section>
  );
}
