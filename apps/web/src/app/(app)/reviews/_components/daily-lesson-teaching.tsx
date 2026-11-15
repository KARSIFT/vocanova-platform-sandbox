"use client";

import { useEffect, useState } from "react";
import type {
  DailyLesson,
  MeaningImage,
  WordDetail,
} from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getPrimaryDefinition } from "@/lib/word-definition";
import { MeaningPicture } from "@/ui/meaning-picture";

type CheckedFeedback = NonNullable<DailyLesson["feedback"]>;
type Step = NonNullable<DailyLesson["currentStep"]>;
interface Teaching {
  identity: string;
  word?: WordDetail;
  image?: MeaningImage;
  unavailable?: boolean;
}

/** Optional teaching never changes the frozen answer or owns the Continue action. */
export function DailyLessonTeaching({
  sessionId,
  step,
  feedback,
}: {
  sessionId: string;
  step: Step;
  feedback: CheckedFeedback;
}) {
  const identity = `${sessionId}:${step.id}:${feedback.meaningId}`;
  const [read, setRead] = useState<Teaching>();
  const wordSlug = feedback.wordSlug ?? step.wordSlug;
  const frozenRevision =
    feedback.meaningRevision ??
    (step.pronunciationMeaningId === feedback.meaningId
      ? step.pronunciationRevision
      : undefined);
  useEffect(() => {
    const controller = new AbortController();
    let active = true;
    const client = createApiClient();
    // Start independent optional reads together. A failed image cannot hide
    // canonical alternate meanings, and neither read blocks the action rail.
    if (wordSlug)
      void client
        .getCanonicalWord(wordSlug, {
          cache: "no-store",
          signal: controller.signal,
        })
        .then(({ data }) => {
          if (!active) return;
          const word =
            data.word.slug === wordSlug &&
            data.word.text.toLocaleLowerCase() ===
              feedback.wordText.toLocaleLowerCase()
              ? data.word
              : undefined;
          setRead((prior) => ({
            ...(prior?.identity === identity ? prior : {}),
            identity,
            word,
            unavailable: !word,
          }));
        })
        .catch(() => {
          if (active)
            setRead((prior) => ({
              ...(prior?.identity === identity ? prior : {}),
              identity,
              unavailable: true,
            }));
        });
    if (frozenRevision) {
      void client
        .getMeaningImages([feedback.meaningId], {
          cache: "no-store",
          signal: controller.signal,
        })
        .then(({ data }) => {
          if (active)
            setRead((prior) => ({
              ...(prior?.identity === identity ? prior : {}),
              identity,
              image: data.items.find(
                (image) =>
                  image.meaningId === feedback.meaningId &&
                  image.meaningRevision === frozenRevision,
              ),
            }));
        })
        .catch(() => {
          /* Optional media failure leaves readable feedback. */
        });
    }
    return () => {
      active = false;
      controller.abort();
    };
  }, [
    identity,
    wordSlug,
    feedback.wordText,
    feedback.meaningId,
    frozenRevision,
  ]);
  const current = read?.identity === identity ? read : undefined;
  const canonicalImage = current?.word?.meanings.find(
    (meaning) => meaning.id === feedback.meaningId,
  )?.image;
  const otherMeanings = current?.word?.meanings.filter(
    (meaning) => meaning.id !== feedback.meaningId,
  );
  return (
    <section
      aria-label={`Learn more about ${feedback.wordText}`}
      className="mt-4"
    >
      {frozenRevision && (
        <MeaningPicture
          key={identity}
          meaningId={feedback.meaningId}
          meaningRevision={frozenRevision}
          image={current?.image ?? canonicalImage}
        />
      )}
      {otherMeanings && otherMeanings.length > 0 ? (
        <details className="mt-4 rounded-xl border border-neutral-200 p-4">
          <summary className="min-h-11 cursor-pointer content-center font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
            Other meanings of {feedback.wordText}
          </summary>
          <ul className="mt-3 space-y-4">
            {otherMeanings.map((meaning) => (
              <li key={meaning.id} className="border-t border-neutral-200 pt-3">
                <p className="text-sm text-neutral-600">
                  {meaning.partOfSpeech}
                </p>
                <p className="mt-1 text-base text-neutral-900">
                  {getPrimaryDefinition(
                    meaning.shortDefinition,
                    meaning.learnerDefinition,
                  )}
                </p>
                {meaning.meaningRevision && (
                  <MeaningPicture
                    meaningId={meaning.id}
                    meaningRevision={meaning.meaningRevision}
                    image={meaning.image}
                  />
                )}
              </li>
            ))}
          </ul>
        </details>
      ) : otherMeanings ? (
        <p className="mt-3 text-sm text-neutral-600">
          This is the only meaning available for {feedback.wordText}.
        </p>
      ) : current?.unavailable ? (
        <p className="mt-3 text-sm text-neutral-600">
          Other meanings could not load. You can continue.
        </p>
      ) : null}
    </section>
  );
}
