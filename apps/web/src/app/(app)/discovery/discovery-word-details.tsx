"use client";

import { useEffect, useRef, useState } from "react";
import type { WordDetail } from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { handleApiError, isSessionExpiredError } from "@/lib/session";
import { getAdditionalDefinition } from "@/lib/word-definition";
import styles from "../knowledge-map/workspace.module.css";

const WORD_DETAILS_DEADLINE_MS = 8_000;

/** Teaching is optional and read-only; no canonical request until opened. */
export function DiscoveryWordDetails({
  wordSlug,
  meaningId,
  meaningRevision,
}: {
  wordSlug: string;
  meaningId: string;
  meaningRevision?: string;
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
      if (
        result.data.word.slug !== wordSlug ||
        !result.data.word.meanings.some((meaning) => meaning.id === meaningId)
      )
        throw new Error("The current meaning is unavailable");
      const currentMeaning = result.data.word.meanings.find(
        (meaning) => meaning.id === meaningId,
      )!;
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

  const meanings = word
    ? [...word.meanings].sort(
        (a, b) => Number(b.id === meaningId) - Number(a.id === meaningId),
      )
    : [];
  return (
    <details
      className={styles.wordDetails}
      onToggle={(event) => {
        if (event.currentTarget.open && !word && !error) void load();
      }}
    >
      <summary>Word details &amp; other meanings</summary>
      {loading && <p role="status">Loading word details…</p>}
      {error && (
        <div role="alert">
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
        </div>
      )}
      {word && (
        <ol className={styles.meanings}>
          {meanings.map((meaning) => {
            const additional = getAdditionalDefinition(
              meaning.shortDefinition,
              meaning.learnerDefinition,
            );
            return (
              <li key={meaning.id}>
                <p className="pos">
                  {meaning.partOfSpeech}
                  {meaning.id === meaningId ? " · This meaning" : ""}
                </p>
                <p className={styles.meaningDefinition}>
                  {meaning.shortDefinition}
                </p>
                {additional && <p>{additional}</p>}
                {meaning.examples.map((example) => (
                  <p key={example.id} className={styles.example}>
                    {example.exampleText}
                  </p>
                ))}
                {meaning.usageNotes.map((note) => (
                  <p key={note.id}>{note.noteText}</p>
                ))}
              </li>
            );
          })}
        </ol>
      )}
    </details>
  );
}
