"use client";

import Image from "next/image";
import { useState } from "react";
import type { DailyLesson } from "@vocanova/api-client";
import { getExternalMeaningPicture } from "@/lib/meaning-pictures";

type Step = NonNullable<DailyLesson["currentStep"]>;

/** Every option keeps its frozen ID when pictures fail or text is requested. */
export function DailyLessonChoices({
  step,
  choice,
  locked,
  correctChoiceId,
  onChange,
}: {
  step: Step;
  choice: string;
  locked: boolean;
  correctChoiceId?: string;
  onChange: (choice: string) => void;
}) {
  const [useText, setUseText] = useState(false);
  const [failed, setFailed] = useState(false);
  const pictures = step.choices.map((option) =>
    option.image
      ? getExternalMeaningPicture(
          option.image.meaningId,
          option.image,
          option.image.meaningRevision,
        )
      : undefined,
  );
  const pictureMode =
    step.kind === "picture_choice" &&
    !useText &&
    step.choices.length === 4 &&
    pictures.every(Boolean);
  return (
    <>
      {pictureMode && (
        <div className="mt-4 flex flex-wrap items-center justify-between gap-2">
          <p className="text-sm text-neutral-600">
            {failed
              ? "A picture could not load. Use text options to continue."
              : "Choose a picture, or use text options."}
          </p>
          <button
            type="button"
            className="vn-btn vn-btn-secondary min-h-11"
            onClick={() => setUseText(true)}
          >
            Use text options
          </button>
        </div>
      )}
      <fieldset
        disabled={locked}
        className={`mt-4 grid min-w-0 gap-2 ${pictureMode ? "grid-cols-2" : ""}`}
      >
        <legend className="sr-only">
          {step.kind === "gap_choice"
            ? "Choose the missing word"
            : `Meaning of ${step.wordText}`}
        </legend>
        {step.choices.map((option, index) => (
          <label
            key={option.id}
            className={`flex min-h-14 min-w-0 cursor-pointer flex-wrap items-center gap-2 rounded-xl border-2 px-3 py-3 text-base leading-relaxed text-neutral-900 focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-700 ${choice === option.id ? "border-primary-700 bg-primary-50" : "border-neutral-200 bg-white hover:border-primary-300"}`}
          >
            {pictureMode && pictures[index] && (
              <Image
                src={pictures[index]!.src}
                alt={pictures[index]!.alt}
                width={pictures[index]!.width ?? 960}
                height={pictures[index]!.height ?? 640}
                unoptimized
                sizes="(max-width: 640px) 45vw, 270px"
                className="aspect-[3/2] w-full rounded-lg object-contain"
                onError={() => setFailed(true)}
              />
            )}
            <input
              type="radio"
              name="meaning"
              value={option.id}
              checked={choice === option.id}
              onChange={() => onChange(option.id)}
              aria-label={pictureMode ? option.text : undefined}
              onFocus={(event) => {
                const input = event.currentTarget;
                if (!input.matches(":focus-visible")) return;
                const label = input.closest("label");
                const bounds = label?.getBoundingClientRect();
                const rail = document
                  .getElementById("daily-lesson-actions")
                  ?.getBoundingClientRect();
                const top =
                  document.querySelector("header")?.getBoundingClientRect()
                    .bottom ?? 0;
                if (
                  bounds &&
                  rail &&
                  (bounds.bottom > rail.top - 8 || bounds.top < top + 8)
                )
                  label?.scrollIntoView({
                    block: "center",
                    behavior: "instant",
                  });
              }}
              className="size-5 shrink-0 accent-primary-700"
            />
            <span
              className={
                pictureMode
                  ? "sr-only"
                  : "min-w-0 flex-1 [overflow-wrap:anywhere]"
              }
            >
              {option.text}
            </span>
            {correctChoiceId === option.id && (
              <span
                className={
                  pictureMode
                    ? "ml-auto font-bold"
                    : "basis-full text-right font-bold"
                }
                aria-label="Correct answer"
              >
                ✓
              </span>
            )}
          </label>
        ))}
      </fieldset>
    </>
  );
}
