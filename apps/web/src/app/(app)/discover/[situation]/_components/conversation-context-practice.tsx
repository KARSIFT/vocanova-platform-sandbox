"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import type { ConversationPracticeCase } from "./conversation-context-content";

interface ConversationContextPracticeProps {
  cases: ConversationPracticeCase[];
  situationSlug: string;
}

const primaryButton =
  "inline-flex min-h-11 items-center justify-center rounded-md bg-primary-600 px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-white hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";
const secondaryButton =
  "inline-flex min-h-11 items-center justify-center rounded-md border border-neutral-300 bg-white px-[var(--spacing-md)] py-[var(--spacing-sm)] text-base font-semibold text-neutral-900 hover:bg-neutral-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

export function ConversationContextPractice({
  cases,
  situationSlug,
}: ConversationContextPracticeProps) {
  const [stage, setStage] = useState<"intro" | "practice" | "complete">(
    "intro",
  );
  const [caseIndex, setCaseIndex] = useState(0);
  const [selectedMeaningId, setSelectedMeaningId] = useState<string | null>(
    null,
  );
  const promptHeading = useRef<HTMLHeadingElement>(null);
  const firstChoice = useRef<HTMLButtonElement>(null);
  const completionHeading = useRef<HTMLHeadingElement>(null);
  const nextFocus = useRef<"prompt" | "choice" | "complete" | null>(null);

  useEffect(() => {
    if (nextFocus.current === "prompt") promptHeading.current?.focus();
    if (nextFocus.current === "choice") firstChoice.current?.focus();
    if (nextFocus.current === "complete") completionHeading.current?.focus();
    nextFocus.current = null;
  }, [stage, caseIndex, selectedMeaningId]);

  if (cases.length === 0) return null;
  const currentCase = cases[caseIndex];
  const answeredCorrectly = selectedMeaningId === currentCase?.correctMeaningId;

  function startPractice() {
    nextFocus.current = "prompt";
    setCaseIndex(0);
    setSelectedMeaningId(null);
    setStage("practice");
  }

  function exitToWords() {
    nextFocus.current = null;
    setStage("intro");
    setCaseIndex(0);
    setSelectedMeaningId(null);
    document.getElementById("situation-words")?.focus();
  }

  function nextExample() {
    if (!answeredCorrectly) return;
    setSelectedMeaningId(null);
    if (caseIndex === cases.length - 1) {
      nextFocus.current = "complete";
      setStage("complete");
    } else {
      nextFocus.current = "prompt";
      setCaseIndex(caseIndex + 1);
    }
  }

  return (
    <section
      aria-labelledby="conversation-practice-heading"
      className="mt-[var(--spacing-md)] rounded-[var(--radius-lg)] border border-neutral-200 bg-white p-[var(--spacing-md)]"
    >
      <h2
        id="conversation-practice-heading"
        className="text-lg font-semibold text-neutral-900"
      >
        Choose the word for the situation
      </h2>
      {stage === "intro" ? (
        <>
          <p className="mt-[var(--spacing-xs)] text-base text-neutral-700">
            Try three short examples about making and changing plans.
          </p>
          <button
            type="button"
            onClick={startPractice}
            className={`${secondaryButton} mt-[var(--spacing-sm)]`}
          >
            Start context practice
          </button>
        </>
      ) : stage === "practice" && currentCase ? (
        <div key={currentCase.id} className="mt-[var(--spacing-md)]">
          <p className="text-sm text-neutral-600">
            Example {caseIndex + 1} of {cases.length}
          </p>
          <h3
            ref={promptHeading}
            tabIndex={-1}
            className="mt-[var(--spacing-xs)] text-lg font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            {currentCase.title}
          </h3>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            {currentCase.context}
          </p>
          <p
            id="conversation-practice-prompt"
            className="mt-[var(--spacing-sm)] text-base font-medium text-neutral-900"
          >
            {currentCase.prompt}
          </p>
          <div
            role="group"
            aria-labelledby="conversation-practice-prompt"
            className="mt-[var(--spacing-sm)] flex flex-wrap gap-[var(--spacing-sm)]"
          >
            {currentCase.choices.map((choice, index) => (
              <button
                key={choice.meaningId}
                ref={index === 0 ? firstChoice : undefined}
                type="button"
                aria-pressed={selectedMeaningId === choice.meaningId}
                disabled={selectedMeaningId !== null}
                onClick={() => {
                  if (selectedMeaningId === null)
                    setSelectedMeaningId(choice.meaningId);
                }}
                className={`${secondaryButton} disabled:cursor-default aria-pressed:border-primary-600 aria-pressed:bg-primary-50`}
              >
                {choice.wordText}
              </button>
            ))}
          </div>
          <div
            role="status"
            aria-live="polite"
            aria-atomic="true"
            className={
              selectedMeaningId !== null
                ? "mt-[var(--spacing-md)] rounded-md bg-secondary-50 p-[var(--spacing-md)] text-secondary-900"
                : undefined
            }
          >
            {selectedMeaningId !== null ? (
              <>
                <p className="font-semibold">
                  {answeredCorrectly
                    ? "That fits this situation."
                    : "Not quite. Compare the two choices."}
                </p>
                <dl className="mt-[var(--spacing-sm)] space-y-[var(--spacing-sm)] text-base">
                  {currentCase.choices.map((choice) => (
                    <div key={choice.meaningId}>
                      <dt className="font-semibold">{choice.wordText}</dt>
                      <dd>{choice.explanation}</dd>
                    </div>
                  ))}
                </dl>
              </>
            ) : null}
          </div>
          {selectedMeaningId !== null ? (
            <>
              <div className="mt-[var(--spacing-sm)]">
                {answeredCorrectly ? (
                  <button
                    type="button"
                    onClick={nextExample}
                    className={primaryButton}
                  >
                    {caseIndex === cases.length - 1
                      ? "Finish practice"
                      : "Next example"}
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => {
                      nextFocus.current = "choice";
                      setSelectedMeaningId(null);
                    }}
                    className={primaryButton}
                  >
                    Try again
                  </button>
                )}
              </div>
            </>
          ) : null}
        </div>
      ) : stage === "complete" ? (
        <div className="mt-[var(--spacing-md)]">
          <h3
            ref={completionHeading}
            tabIndex={-1}
            className="text-lg font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
          >
            You’ve explored three situations
          </h3>
          <p className="mt-[var(--spacing-sm)] text-base text-neutral-700">
            Try one of these words in your own sentence. Open a word, then save
            it to unlock sentence practice. If it is already saved, you can
            start writing.
          </p>
          <div className="mt-[var(--spacing-sm)] flex flex-wrap gap-[var(--spacing-sm)]">
            {cases.map((example) => {
              const word = example.choices.find(
                (choice) => choice.meaningId === example.correctMeaningId,
              );
              return word ? (
                <Link
                  key={word.meaningId}
                  href={`/discover/${encodeURIComponent(situationSlug)}/${encodeURIComponent(word.wordSlug)}`}
                  className={secondaryButton}
                >
                  Practice with {word.wordText}
                </Link>
              ) : null;
            })}
          </div>
          <button
            type="button"
            onClick={startPractice}
            className={`${secondaryButton} mt-[var(--spacing-md)]`}
          >
            Restart practice
          </button>
        </div>
      ) : null}
      {stage !== "intro" ? (
        <button
          type="button"
          onClick={exitToWords}
          className="mt-[var(--spacing-sm)] inline-flex min-h-11 items-center px-[var(--spacing-sm)] text-base font-semibold text-primary-700 underline hover:text-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700"
        >
          Exit to words
        </button>
      ) : null}
    </section>
  );
}
