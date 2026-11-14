"use client";

import { StatusMessage } from "@/ui/status-message";
import { useEffect, useRef, useState } from "react";
import type { WordDetail } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { handleApiError, isSessionExpiredError } from "@/lib/session";
import {
  getAdditionalDefinition,
  getPrimaryDefinition,
} from "@/lib/word-definition";
import styles from "../knowledge-map/workspace.module.css";

const WORD_DETAILS_DEADLINE_MS = 8_000;

function MeaningDetails({
  meaning,
  current = false,
  exampleSentence,
  displayedDefinition,
}: {
  meaning: WordDetail["meanings"][number];
  current?: boolean;
  exampleSentence?: string;
  displayedDefinition?: string;
}) {
  const primaryDefinition = getPrimaryDefinition(
    meaning.shortDefinition,
    meaning.learnerDefinition,
  );
  const currentExplanation = current
    ? getAdditionalDefinition(
        displayedDefinition ?? meaning.shortDefinition,
        primaryDefinition,
      )
    : null;
  const referenceDefinition = getAdditionalDefinition(
    current ? (displayedDefinition ?? primaryDefinition) : primaryDefinition,
    meaning.shortDefinition,
  );
  const examples = current
    ? meaning.examples.filter(
        (example) => example.exampleText !== exampleSentence,
      )
    : meaning.examples;

  if (
    current &&
    !currentExplanation &&
    !referenceDefinition &&
    !examples.length &&
    !meaning.usageNotes.length
  )
    return null;

  return (
    <div className={current ? "my-3 space-y-2" : undefined}>
      <p className="pos">
        {meaning.partOfSpeech}
        {current ? " · This meaning" : ""}
      </p>
      {!current && (
        <p className={styles.meaningDefinition}>
          {getPrimaryDefinition(
            meaning.shortDefinition,
            meaning.learnerDefinition,
          )}
        </p>
      )}
      {currentExplanation && <p>{currentExplanation}</p>}
      {referenceDefinition && (
        <p>
          <span className="font-medium">Reference definition: </span>
          {referenceDefinition}
        </p>
      )}
      {examples.map((example) => (
        <p key={example.id} className={styles.example}>
          {example.exampleText}
        </p>
      ))}
      {meaning.usageNotes.map((note) => (
        <p key={note.id}>{note.noteText}</p>
      ))}
    </div>
  );
}

/** Teaching is optional and read-only; no canonical request until opened. */
export function DiscoveryWordDetails({
  wordSlug,
  meaningId,
  meaningRevision,
  exampleSentence,
  displayedDefinition,
}: {
  wordSlug: string;
  meaningId: string;
  meaningRevision?: string;
  exampleSentence?: string;
  displayedDefinition?: string;
}) {
  const [word, setWord] = useState<WordDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [revisionChanged, setRevisionChanged] = useState(false);
  const request = useRef<AbortController | null>(null);
  useEffect(() => () => request.current?.abort(), []);

  async function load() {
    if (request.current || word) return;
    const controller = new AbortController();
    const deadline = new AbortController();
    const timeout = setTimeout(
      () => deadline.abort(),
      WORD_DETAILS_DEADLINE_MS,
    );
    const clearDeadline = () => clearTimeout(timeout);
    controller.signal.addEventListener("abort", clearDeadline, { once: true });
    request.current = controller;
    setLoading(true);
    setError("");
    setRevisionChanged(false);
    try {
      const result = await createApiClient().getCanonicalWord(wordSlug, {
        // Bound headers and body while keeping timeout recovery distinct from
        // unmount cancellation, which must not update the abandoned component.
        signal: AbortSignal.any([controller.signal, deadline.signal]),
      });
      if (controller.signal.aborted) return;
      // A response must contain the exact displayed sense, never substitute a
      // similarly spelt word or an unrelated first dictionary meaning.
      const currentMeaning = result.data.word.meanings.find(
        (meaning) => meaning.id === meaningId,
      );
      if (result.data.word.slug !== wordSlug || !currentMeaning)
        throw new Error("The current meaning is unavailable");
      if (
        meaningRevision &&
        currentMeaning.meaningRevision !== meaningRevision
      ) {
        setRevisionChanged(true);
        setError(
          "This word has changed. Reload the page to see its current details.",
        );
        return;
      }
      setWord(result.data.word);
    } catch (failure) {
      if (!controller.signal.aborted)
        setError(
          isSessionExpiredError(failure)
            ? handleApiError(failure, "")
            : "Word details could not load. Try again.",
        );
    } finally {
      clearDeadline();
      controller.signal.removeEventListener("abort", clearDeadline);
      if (!controller.signal.aborted) {
        setLoading(false);
        request.current = null;
      }
    }
  }

  const currentMeaning = word?.meanings.find(
    (meaning) => meaning.id === meaningId,
  );
  const otherMeanings =
    word?.meanings.filter((meaning) => meaning.id !== meaningId) ?? [];
  return (
    <details
      className={styles.wordDetails}
      onToggle={(event) => {
        if (event.currentTarget.open && !word && !error) void load();
      }}
    >
      <summary>Word details &amp; other meanings</summary>
      {loading && (
        <StatusMessage as="p" role="status">
          Loading word details…
        </StatusMessage>
      )}
      {error && (
        <StatusMessage tone="error" role="alert">
          <p>{error}</p>
          {revisionChanged && (
            <button
              type="button"
              className={styles.secondary}
              onClick={() => window.location.reload()}
            >
              Reload word
            </button>
          )}
          {!revisionChanged && (
            <button
              type="button"
              className={styles.secondary}
              onClick={() => void load()}
            >
              Retry word details
            </button>
          )}
        </StatusMessage>
      )}
      {currentMeaning && (
        <>
          {otherMeanings.length ? (
            <>
              <h2 className="font-semibold text-neutral-900">Other meanings</h2>
              <ol className={styles.meanings}>
                {otherMeanings.map((meaning) => (
                  <li key={meaning.id}>
                    <MeaningDetails meaning={meaning} />
                  </li>
                ))}
              </ol>
            </>
          ) : (
            <p>This is the only meaning in Vocanova.</p>
          )}
          <MeaningDetails
            current
            exampleSentence={exampleSentence}
            displayedDefinition={displayedDefinition}
            meaning={currentMeaning}
          />
        </>
      )}
    </details>
  );
}
