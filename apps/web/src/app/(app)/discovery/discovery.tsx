"use client";

import {
  SESSION_LOCATION_CHANGED,
  replaceSessionLocation,
} from "../_components/session-route";
import Link from "@/ui/link";
import { MeaningPicture } from "@/ui/meaning-picture";
import { useRouter, useSearchParams } from "next/navigation";
import { DiscoveryWordDetails } from "./discovery-word-details";
import {
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
  useTransition,
} from "react";
import {
  ApiResponseError,
  type KnowledgeMapDecision,
} from "@vocanova/api-client";
import { createApiClient } from "@/lib/api";
import { getOrRefreshCSRFToken } from "@/lib/csrf";
import { getExternalMeaningPicture } from "@/lib/meaning-pictures";
import { handleApiError } from "@/lib/session";
import { deviceSpeech, type PronunciationState } from "@/lib/speech";
import {
  applyWordState,
  confirmedWordState,
  confirmsChoice,
  journeyRangeHref,
  nextDiscoveryTarget,
  wordState,
  wordStateLabel,
  type KnowledgeMapData,
} from "../knowledge-map/models";
import styles from "../knowledge-map/workspace.module.css";

interface PendingDecision {
  body: KnowledgeMapDecision;
  key: string;
}

function subscribeDiscoveryLocation(onChange: () => void) {
  window.addEventListener(SESSION_LOCATION_CHANGED, onChange);
  window.addEventListener("popstate", onChange);
  return () => {
    window.removeEventListener(SESSION_LOCATION_CHANGED, onChange);
    window.removeEventListener("popstate", onChange);
  };
}

function browserSearchSnapshot() {
  return window.location.search;
}

export function Discovery({
  initialData,
  initialRank,
  initiallyActive = false,
}: {
  initialData: KnowledgeMapData;
  initialRank: number;
  initiallyActive?: boolean;
}) {
  const [data, setData] = useState(initialData);
  const [rank, setRank] = useState(initialRank);
  const [started, setStarted] = useState(initiallyActive);
  const router = useRouter();
  const serverParams = useSearchParams();
  const serverSearch = serverParams.size ? `?${serverParams.toString()}` : "";
  const currentSearch = useSyncExternalStore(
    subscribeDiscoveryLocation,
    browserSearchSnapshot,
    () => serverSearch,
  );
  const [navigating, startNavigation] = useTransition();
  const [pickerOpen, setPickerOpen] = useState(false);
  const [selectedRank, setSelectedRank] = useState(initialRank);
  const [selectedBand, setSelectedBand] = useState(
    initialData.currentBand.number,
  );
  const [muted, setMuted] = useState(false);
  const [audioReady, setAudioReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [done, setDone] = useState(false);
  const [speech, setSpeech] = useState<PronunciationState>({ status: "idle" });
  const pending = useRef<PendingDecision | null>(null);
  const lock = useRef(false);
  const speechOwner = useRef(Symbol("discovery"));
  const lastSpoken = useRef("");
  const titleRef = useRef<HTMLHeadingElement>(null);
  const errorRef = useRef<HTMLDivElement>(null);
  const item =
    data.currentBand.items.find((word) => word.rank === rank) ??
    data.currentBand.items[0];
  const band = data.currentBand.number;
  const bandSummary = data.bands.find((value) => value.number === band);
  const displayedRank = item?.rank;
  const displayedWord = item?.wordText;
  const displayedMeaning = item?.meaningId;
  const displayedRevision = item?.meaningRevision;
  const navigationLocked =
    busy || navigating || conflict || Boolean(pending.current);
  // The likely next word: the first unseen one after this word in the range.
  const upcoming = data.currentBand.items.find(
    (word) =>
      word.rank > (displayedRank ?? 0) &&
      !word.selfReportedKnown &&
      !word.saved,
  );

  const firstRank = bandSummary?.startRank ?? 1;
  const lastRank = bandSummary?.endRank ?? data.totalWords;

  const historyWord = useRef({ rank, started, blocked: navigationLocked });
  historyWord.current = { rank, started, blocked: navigationLocked };

  // Owned writes preserve Next's opaque history state. Observe the actual
  // browser location (as DiscoveryExitLink does), because Next searchParams
  // need not update for those entries or their Back/Forward navigation.
  useEffect(() => {
    const searchParams = new URLSearchParams(currentSearch);
    const requested = Number(searchParams.get("word"));
    if (
      Number(searchParams.get("band")) === band &&
      Number.isInteger(requested) &&
      requested >= firstRank &&
      requested <= lastRank
    ) {
      const current = historyWord.current;
      const active = searchParams.get("active") === "1";
      if (
        current.blocked &&
        (requested !== current.rank || active !== current.started)
      ) {
        // Back/Forward bypass disabled controls. Restore the pending word as a
        // new entry, retaining the attempted destination underneath it so Back
        // works normally after the original decision has been confirmed.
        window.history.pushState(
          window.history.state,
          "",
          `/discovery?band=${band}&word=${current.rank}${current.started ? "&active=1" : ""}`,
        );
        window.dispatchEvent(new Event(SESSION_LOCATION_CHANGED));
        return;
      }
      setRank(requested);
      setSelectedRank(requested);
      setStarted(active);
      setDone(false);
    }
  }, [currentSearch, band, firstRank, lastRank]);

  useEffect(() => {
    setSelectedBand(band);
  }, [band]);

  function navigateToWord(nextBand: number, nextRank: number) {
    if (lock.current || navigationLocked) return;
    deviceSpeech.stop(speechOwner.current);
    setError("");
    setPickerOpen(false);
    const href = `/discovery?band=${nextBand}&word=${nextRank}&active=1`;
    if (nextBand === band) {
      window.history.pushState(window.history.state, "", href);
      window.dispatchEvent(new Event(SESSION_LOCATION_CHANGED));
      setStarted(true);
      setRank(nextRank);
      setSelectedRank(nextRank);
      if (nextRank === item?.rank) titleRef.current?.focus();
    } else startNavigation(() => router.push(href));
  }

  function previousWord() {
    if (!item || item.rank <= 1) return;
    const previousRank = item.rank - 1;
    const previousBand = data.bands.find(
      (value) =>
        previousRank >= value.startRank && previousRank <= value.endRank,
    );
    if (previousBand) navigateToWord(previousBand.number, previousRank);
  }

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
    setAudioReady(true);
    return () => deviceSpeech.release(owner);
  }, []);

  useEffect(() => {
    if (displayedRank === undefined || done) return;
    // The server may select a later unseen range before Start. Canonicalize
    // its address now so Back preserves that real range without another read.
    replaceSessionLocation(
      `/discovery?band=${band}&word=${displayedRank}${started ? "&active=1" : ""}`,
    );
    if (started) titleRef.current?.focus();
  }, [started, displayedRank, band, done]);

  useEffect(() => {
    if (
      !audioReady ||
      !started ||
      displayedRank === undefined ||
      !displayedWord ||
      done
    )
      return;
    const identity = `${displayedRank}:${displayedMeaning}`;
    if (lastSpoken.current !== identity) {
      lastSpoken.current = identity;
      if (!muted)
        deviceSpeech.play(
          speechOwner.current,
          displayedWord,
          false,
          setSpeech,
          displayedMeaning,
          displayedRevision,
        );
    }
  }, [
    audioReady,
    started,
    displayedRank,
    displayedWord,
    displayedMeaning,
    displayedRevision,
    muted,
    done,
  ]);

  // Warm the browser cache with the next word's approved picture while the
  // learner reads this one, so the following card appears complete.
  useEffect(() => {
    if (!started || !upcoming?.image) return;
    const picture = getExternalMeaningPicture(
      upcoming.meaningId,
      upcoming.image,
      upcoming.meaningRevision,
    );
    if (!picture) return;
    const preload = new Image();
    preload.decoding = "async";
    preload.src = picture.src;
  }, [started, upcoming]);

  function start() {
    if (!item) return;
    setStarted(true);
    // Keep the first playback inside the learner's gesture for browsers that require it.
    lastSpoken.current = `${item.rank}:${item.meaningId}`;
    if (!muted)
      deviceSpeech.play(
        speechOwner.current,
        item.wordText,
        false,
        setSpeech,
        item.meaningId,
        item.meaningRevision,
      );
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
    if (lock.current || navigating || conflict || !item) return;
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
      const { data: result } = await createApiClient().decideKnowledgeMapWord(
        request.body,
        request.key,
        { headers: { "X-CSRF-Token": csrf } },
      );
      // Use the meaning's state as the server reports it. When it is exactly
      // what this choice produces, continue within the range without
      // re-reading all 100 words; a later word changed elsewhere still
      // surfaces as a revision conflict when it is chosen.
      const state = confirmedWordState(result, request.body);
      if (state && confirmsChoice(state, request.body.decision)) {
        const local = applyWordState(data, request.body.meaningId, state);
        const nextInRange = local.currentBand.items.find(
          (word) => !word.selfReportedKnown && !word.saved,
        );
        if (nextInRange) {
          pending.current = null;
          setData(local);
          setRank(nextInRange.rank);
          return;
        }
      }
      // A replayed receipt that reports a different current state (changed
      // since, in another tab or by a removal), an incomplete response or the
      // end of the range: read confirmed state before choosing the next unseen
      // word. Keep the exact pending intent until that range loads.
      const client = createApiClient();
      const current = (await client.getKnowledgeMap(band)).data;
      const target = nextDiscoveryTarget(current);
      const fresh =
        target && target.band !== band
          ? (await client.getKnowledgeMap(target.band)).data
          : current;
      const nextWord = target
        ? fresh.currentBand.items.find(
            (word) => !word.selfReportedKnown && !word.saved,
          )
        : null;
      if (target && !nextWord)
        throw new Error("The next unseen word could not be confirmed");
      pending.current = null;
      setData(fresh);
      if (!nextWord) setDone(true);
      else setRank(nextWord.rank);
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
        <Link href={journeyRangeHref(band)} className={styles.secondary}>
          Back to Journey
        </Link>
      </div>
    );

  if (done)
    return (
      <div className={styles.focus}>
        <h1 className={styles.title}>You’ve explored the whole map</h1>
        <p className={styles.definition}>
          Your choices are saved. Keep the words you are learning fresh with a
          short review.
        </p>
        <div className={styles.startActions}>
          <Link className={styles.primary} href="/reviews">
            Review words
          </Link>
          <Link className={styles.secondary} href={`/discover?band=${band}`}>
            View map
          </Link>
        </div>
      </div>
    );

  const rangeName = bandSummary
    ? `Words ${bandSummary.startRank.toLocaleString()}–${bandSummary.endRank.toLocaleString()}`
    : "";

  return (
    <div className={styles.focus}>
      <div className={styles.top}>
        <span className={styles.counter}>{rangeName}</span>
        <span className={styles.counter}>
          Word {item.rank.toLocaleString()} of{" "}
          {data.totalWords.toLocaleString()}
        </span>
      </div>
      <div className={styles.discoveryNavigation}>
        <button
          type="button"
          className={styles.secondary}
          onClick={previousWord}
          disabled={navigationLocked || item.rank <= 1}
        >
          Previous word
        </button>
        <form
          className={styles.discoveryRange}
          onSubmit={(event) => {
            event.preventDefault();
            const target = data.bands.find(
              (value) => value.number === selectedBand,
            );
            if (target) navigateToWord(target.number, target.startRank);
          }}
        >
          <label htmlFor="discovery-range">Word range</label>
          <div>
            <select
              id="discovery-range"
              value={selectedBand}
              disabled={navigationLocked}
              onChange={(event) => setSelectedBand(Number(event.target.value))}
            >
              {data.bands.map((value) => (
                <option key={value.number} value={value.number}>
                  {value.startRank.toLocaleString()}–
                  {value.endRank.toLocaleString()}
                </option>
              ))}
            </select>
            <button
              type="submit"
              className={styles.secondary}
              disabled={navigationLocked}
            >
              Go to range
            </button>
          </div>
        </form>
      </div>
      <button
        type="button"
        className={styles.progressPicker}
        aria-label="Choose a word"
        aria-expanded={pickerOpen}
        aria-controls="discovery-word-picker"
        disabled={navigationLocked}
        onClick={() => {
          setSelectedRank(item.rank);
          setPickerOpen(!pickerOpen);
        }}
      >
        <span className={styles.pickerLabel}>
          Choose a word <span aria-hidden="true">⌄</span>
        </span>
        <span
          className={`${styles.strip} ${styles.discoveryStrip}`}
          aria-hidden="true"
        >
          {data.currentBand.items.map((word) => (
            <span
              key={word.meaningId}
              data-state={wordState(word)}
              data-next={word.rank === item.rank ? "true" : undefined}
            />
          ))}
        </span>
      </button>
      <div
        id="discovery-word-picker"
        hidden={!pickerOpen}
        className={styles.wordPicker}
      >
        <form
          onSubmit={(event) => {
            event.preventDefault();
            navigateToWord(band, selectedRank);
          }}
        >
          <label htmlFor="discovery-rank">Word in this range</label>
          <select
            id="discovery-rank"
            value={selectedRank}
            disabled={navigationLocked}
            onChange={(event) => setSelectedRank(Number(event.target.value))}
          >
            {data.currentBand.items.map((word) => (
              <option key={word.meaningId} value={word.rank}>
                {word.rank.toLocaleString()}. {word.wordText} —{" "}
                {wordStateLabel(word)}
              </option>
            ))}
          </select>
          <button
            type="submit"
            className={styles.secondary}
            disabled={navigationLocked}
          >
            Go to word
          </button>
        </form>
      </div>
      {navigating && (
        <p role="status" className={styles.status}>
          Loading word…
        </p>
      )}
      {!started ? (
        <section className={styles.start}>
          <h1 className={styles.title}>Discover your next words</h1>
          <p>Listen, then choose Learn or Already know.</p>
          <div className={styles.startActions}>
            <button type="button" className={styles.primary} onClick={start}>
              Start discovering
            </button>
            <div className={styles.audio}>
              <button type="button" aria-pressed={!muted} onClick={toggleSound}>
                {muted ? "Sound off" : "Sound on"}
              </button>
            </div>
          </div>
        </section>
      ) : (
        <>
          <article className={styles.card} aria-labelledby="discovery-word">
            <div className={styles.entry}>
              <h1
                id="discovery-word"
                ref={titleRef}
                tabIndex={-1}
                className={`headword ${styles.cardWord}`}
              >
                {item.wordText}
              </h1>
              <span className={styles.stateChip} data-state={wordState(item)}>
                {wordStateLabel(item)}
              </span>
            </div>
            <div className={styles.entryMeta}>
              <p className="pos">{item.partOfSpeech}</p>
            </div>
            <div className={styles.audio}>
              <span className="sr-only">AI-generated pronunciation</span>
              <button
                type="button"
                onClick={() =>
                  deviceSpeech.play(
                    speechOwner.current,
                    item.wordText,
                    false,
                    setSpeech,
                    item.meaningId,
                    item.meaningRevision,
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
            <MeaningPicture
              meaningId={item.meaningId}
              image={item.image}
              meaningRevision={item.meaningRevision}
              priority
            />
            {item.exampleSentence && (
              <p className={styles.example}>{item.exampleSentence}</p>
            )}
          </article>
          <div className={styles.actions}>
            <button
              type="button"
              className={styles.primary}
              disabled={navigationLocked}
              onClick={() => void decide("learn")}
            >
              Learn
            </button>
            <button
              type="button"
              className={styles.secondary}
              disabled={navigationLocked}
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
        </>
      )}
      {started && (
        <DiscoveryWordDetails
          key={`${item.meaningId}:${item.meaningRevision ?? ""}`}
          wordSlug={item.wordSlug}
          meaningId={item.meaningId}
          meaningRevision={item.meaningRevision}
        />
      )}
    </div>
  );
}
