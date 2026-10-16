"use client";

import Link from "next/link";
import { MeaningPicture } from "@/ui/meaning-picture";
import { VocabularySources } from "@/ui/vocabulary-sources";
import { useEffect, useRef, useState } from "react";
import {
  ApiResponseError,
  type KnowledgeMapDecision,
} from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { handleApiError } from "@/lib/session";
import { deviceSpeech, type PronunciationState } from "@/lib/speech";
import { wordStateLabel, type KnowledgeMapData } from "../knowledge-map/models";
import styles from "../knowledge-map/workspace.module.css";

interface PendingDecision {
  body: KnowledgeMapDecision;
  key: string;
}

export function Discovery({
  initialData,
  initialRank,
}: {
  initialData: KnowledgeMapData;
  initialRank: number;
}) {
  const [data, setData] = useState(initialData);
  const [rank, setRank] = useState(initialRank);
  const [started, setStarted] = useState(false);
  const [muted, setMuted] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [done, setDone] = useState(false);
  const [speech, setSpeech] = useState<PronunciationState>({ status: "idle" });
  const pending = useRef<PendingDecision | null>(null);
  const lock = useRef(false);
  const speechOwner = useRef(Symbol("discovery"));
  const lastSpoken = useRef(0);
  const titleRef = useRef<HTMLHeadingElement>(null);
  const errorRef = useRef<HTMLDivElement>(null);
  const item =
    data.currentBand.items.find((word) => word.rank === rank) ??
    data.currentBand.items[0];
  const band = data.currentBand.number;
  const bandSummary = data.bands.find((value) => value.number === band);

  useEffect(() => {
    if (!error) return;
    errorRef.current?.focus({ preventScroll: true });
    errorRef.current?.scrollIntoView({ block: "center", behavior: "instant" });
  }, [error]);

  useEffect(() => {
    const owner = speechOwner.current;
    deviceSpeech.prepare();
    try {
      setMuted(sessionStorage.getItem("vocanova-discovery-muted") === "true");
    } catch {
      /* Device storage may be unavailable. */
    }
    return () => deviceSpeech.release(owner);
  }, []);

  useEffect(() => {
    if (!started || !item || done) return;
    window.history.replaceState(
      window.history.state,
      "",
      `/discovery?band=${band}&word=${item.rank}`,
    );
    titleRef.current?.focus();
    if (lastSpoken.current !== item.rank) {
      lastSpoken.current = item.rank;
      if (!muted)
        deviceSpeech.play(speechOwner.current, item.wordText, false, setSpeech);
    }
  }, [started, item, band, muted, done]);

  function start() {
    if (!item) return;
    setStarted(true);
    // Keep the first playback inside the learner's gesture for browsers that require it.
    lastSpoken.current = item.rank;
    if (!muted)
      deviceSpeech.play(speechOwner.current, item.wordText, false, setSpeech);
  }

  function toggleSound() {
    const next = !muted;
    setMuted(next);
    if (next) deviceSpeech.stop(speechOwner.current);
    try {
      sessionStorage.setItem("vocanova-discovery-muted", String(next));
    } catch {
      /* Optional device preference. */
    }
  }

  async function decide(decision?: "learn" | "known") {
    if (lock.current || conflict || !item) return;
    if (!pending.current) {
      if (!decision) return;
      pending.current = {
        body: {
          meaningId: item.meaningId,
          decision,
          expectedRevision: item.revision,
        },
        key: crypto.randomUUID(),
      };
    }
    const request = pending.current;
    lock.current = true;
    setBusy(true);
    setError("");
    deviceSpeech.stop(speechOwner.current);
    try {
      const csrf = await getOrRefreshCSRFToken();
      if (!csrf) throw new Error("Session not ready");
      await createApiClient().decideKnowledgeMapWord(
        request.body,
        request.key,
        { headers: { "X-CSRF-Token": csrf } },
      );
      const nextRank = rank + 1;
      // Read confirmed state before changing the card; a retry retains the exact mutation key/body.
      const nextBand =
        nextRank <= data.totalWords
          ? Math.floor((nextRank - 1) / data.bandSize) + 1
          : band;
      const confirmed = (await createApiClient().getKnowledgeMap(nextBand))
        .data;
      pending.current = null;
      setData(confirmed);
      if (nextRank > data.totalWords) setDone(true);
      else setRank(nextRank);
    } catch (failure) {
      if (failure instanceof ApiResponseError && failure.status === 409) {
        pending.current = null;
        setConflict(true);
        try {
          const fresh = (await createApiClient().getKnowledgeMap(band)).data;
          setData(fresh);
          setConflict(false);
          setError(
            "This word changed elsewhere. Its latest state is shown. Choose again.",
          );
        } catch (refreshFailure) {
          setError(
            handleApiError(
              refreshFailure,
              "The latest word could not load. Refresh this page to continue.",
            ),
          );
        }
      } else {
        setError(
          handleApiError(
            failure,
            "Your choice could not be confirmed. Retry to continue.",
          ),
        );
      }
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }

  async function reloadWord() {
    if (lock.current) return;
    lock.current = true;
    setBusy(true);
    try {
      const fresh = (await createApiClient().getKnowledgeMap(band)).data;
      setData(fresh);
      pending.current = null;
      setConflict(false);
      setError("");
    } catch (failure) {
      setError(
        handleApiError(failure, "The latest word could not load. Try again."),
      );
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }

  if (!item)
    return (
      <div className={styles.focus}>
        <p role="alert">This word could not load.</p>
        <Link href={`/knowledge-map?band=${band}`} className={styles.secondary}>
          Back to map
        </Link>
      </div>
    );

  if (done)
    return (
      <div className={styles.focus}>
        <h1 className={styles.title}>You reached the end of the map</h1>
        <p className={styles.definition}>
          Your choices are saved. Keep the words you are learning fresh with a
          short review.
        </p>
        <div className={styles.actions}>
          <Link className={styles.primary} href="/reviews">
            Review words
          </Link>
          <Link
            className={styles.secondary}
            href={`/knowledge-map?band=${band}`}
          >
            View map
          </Link>
        </div>
      </div>
    );

  return (
    <div className={styles.focus}>
      <div className={styles.top}>
        <Link
          href={`/knowledge-map?band=${band}#range-words`}
          className={styles.back}
        >
          Knowledge map
        </Link>
        <span className={styles.counter}>
          {item.rank.toLocaleString()} / {data.totalWords.toLocaleString()}
        </span>
      </div>
      <progress
        className={styles.progress}
        value={item.rank - 1}
        max={data.totalWords}
        aria-label="Position in word map"
      />
      {!started ? (
        <section className={styles.start}>
          <h1 className={styles.title}>Discover your next words</h1>
          <p>
            Hear a word, read its meaning, then choose Learn or Already know.
          </p>
          <button type="button" className={styles.primary} onClick={start}>
            Start discovering
          </button>
          <div className={styles.audio}>
            <button type="button" aria-pressed={!muted} onClick={toggleSound}>
              {muted ? "Sound off" : "Sound on"}
            </button>
          </div>
        </section>
      ) : (
        <>
          <article className={styles.card} aria-labelledby="discovery-word">
            <div className={styles.header}>
              <p className={styles.muted}>{item.partOfSpeech}</p>
              <span className={styles.state}>{wordStateLabel(item)}</span>
            </div>
            <h1
              id="discovery-word"
              ref={titleRef}
              tabIndex={-1}
              className={styles.cardWord}
            >
              {item.wordText}
            </h1>
            <div className={styles.audio}>
              <button
                type="button"
                onClick={() =>
                  deviceSpeech.play(
                    speechOwner.current,
                    item.wordText,
                    false,
                    setSpeech,
                  )
                }
                aria-label={`Listen to ${item.wordText}`}
              >
                Listen
              </button>
              <button type="button" aria-pressed={!muted} onClick={toggleSound}>
                {muted ? "Automatic sound off" : "Automatic sound on"}
              </button>
            </div>
            {speech.message && (
              <p className={styles.status} role="status">
                {speech.message} You can keep learning without sound.
              </p>
            )}
            <p className={styles.definition}>{item.shortDefinition}</p>
            <MeaningPicture meaningId={item.meaningId} />
            {item.exampleSentence && (
              <p className={styles.example}>{item.exampleSentence}</p>
            )}
            <div className={styles.actions}>
              <button
                type="button"
                className={styles.primary}
                disabled={busy || conflict || Boolean(pending.current)}
                onClick={() => void decide("learn")}
              >
                Learn
              </button>
              <button
                type="button"
                className={styles.secondary}
                disabled={busy || conflict || Boolean(pending.current)}
                onClick={() => void decide("known")}
              >
                Already know
              </button>
            </div>
            {busy && (
              <p role="status" className={styles.status}>
                Saving your choice…
              </p>
            )}
          </article>
          {error && (
            <div
              ref={errorRef}
              tabIndex={-1}
              role="alert"
              className={styles.error}
            >
              <p>{error}</p>
              {conflict && (
                <button
                  type="button"
                  disabled={busy}
                  className={styles.primary}
                  onClick={() => void reloadWord()}
                >
                  Load latest word
                </button>
              )}
              {pending.current && (
                <button
                  type="button"
                  disabled={busy}
                  className={styles.primary}
                  onClick={() => void decide()}
                >
                  Retry choice
                </button>
              )}
            </div>
          )}
          <p className={styles.status}>
            Words {bandSummary?.startRank.toLocaleString()}–
            {bandSummary?.endRank.toLocaleString()}
          </p>
          <Link
            className={styles.back}
            href={`/knowledge-map?band=${band}#range-words`}
          >
            Choose another word
          </Link>
        </>
      )}
      <VocabularySources attribution={data.attribution} />
    </div>
  );
}
