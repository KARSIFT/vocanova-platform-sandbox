"use client";

import { useState } from "react";
import { Button } from "./controls";
import { MeaningPicture } from "./meaning-picture";
import styles from "./learning-preview.module.css";

/** Original, authored examples only. This demo does not evaluate user writing. */
export function LearningPreview() {
  const [showFeedback, setShowFeedback] = useState(false);
  return (
    <section aria-label="Learning example" className={styles.window}>
      <div className={styles.toolbar}>
        <p>Try a learning example</p>
      </div>
      <div className={styles.body}>
        <div aria-live="polite" className={styles.beat}>
          {showFeedback ? (
            <div className={styles.feedback}>
              <h2 className="text-2xl font-bold tracking-tight text-neutral-900">
                Make it your own.
              </h2>
              <p className="mt-3 text-sm font-semibold text-neutral-600">
                Example sentence
              </p>
              <p className="mt-1 text-lg text-neutral-900">
                I want invite my friend.
              </p>
              <div className="mt-4 rounded-r-xl border-l-4 border-primary-300 bg-primary-50 px-4 py-3">
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
            </div>
          ) : (
            <>
              <h2 className={styles.headword}>invite</h2>
              <p className="mt-2 text-neutral-700">
                To ask someone to come to an event or do something with you.
              </p>
              <MeaningPicture
                meaningId="ee53d6ba-4303-5394-b7f9-79937ce66d09"
                priority
              />
              <p className={styles.example}>
                I want to invite you to my party.
              </p>
            </>
          )}
        </div>
        <Button
          type="button"
          variant="secondary"
          aria-pressed={showFeedback}
          onClick={() => setShowFeedback(!showFeedback)}
          className={styles.action}
        >
          {showFeedback ? "Back to the word" : "See example feedback"}
        </Button>
      </div>
    </section>
  );
}
