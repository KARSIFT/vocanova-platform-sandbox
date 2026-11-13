"use client";

import { getPrimaryDefinition } from "@/lib/word-definition";

import {
  SESSION_LOCATION_CHANGED,
  replaceSessionLocation,
} from "../_components/session-route";
import Link from "@/ui/link";
import { LearningIntro, LearningNote } from "@/ui/learning-guide";
import { MeaningPicture } from "@/ui/meaning-picture";
import { useRouter, useSearchParams } from "next/navigation";
import { DiscoveryFinished } from "./discovery-finished";
import { DiscoveryNavigation } from "./discovery-navigation";
import { DiscoveryWordDetails } from "./discovery-word-details";
import {
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
  useTransition,
  type FocusEvent,
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
  discoveryWordInRange,
  isUnseen,
  journeyRangeHref,
  nextDiscoveryTarget,
  wordState,
  wordStateLabel,
  type KnowledgeMapData,
  type MapWord,
} from "../knowledge-map/models";
import styles from "../knowledge-map/workspace.module.css";

interface PendingDecision {
  body: KnowledgeMapDecision;
  key: string;
  rank: number;
  word: Pick<
    MapWord,
    "wordSlug" | "wordText" | "shortDefinition" | "learnerDefinition"
  >;
}

interface SavedChoice {
  meaningId: string;
  word: PendingDecision["word"];
  // A receipt belongs to the next card reached by that choice, not every
  // later word or a restored browser-history selection.
  band: number;
  rank?: number;
}

function SavedMeaning({ choice }: { choice: SavedChoice }) {
  return (
    <section aria-label="Saved meaning" className="mt-4">
      <LearningNote>
        <p className="font-semibold">
          Saved {choice.word.wordText} for review.
        </p>
        <p className="mt-1">
          {getPrimaryDefinition(
            choice.word.shortDefinition,
            choice.word.learnerDefinition,
          )}
        </p>
        <Link
          href={`/vocabulary/${encodeURIComponent(choice.word.wordSlug)}#sentence-practice-${choice.meaningId}`}
          className="vn-text-link mt-2 inline-flex min-h-11 items-center"
        >
          Use {choice.word.wordText} in a sentence
        </Link>
      </LearningNote>
    </section>
  );
}

function revealFocusedControl(event: FocusEvent<HTMLDivElement>) {
  const target = event.target;
  if (!(target instanceof HTMLElement) || target.closest("dialog")) return;
  // Let native keyboard scrolling settle before measuring the real
  // reading area beneath the sticky header.
  requestAnimationFrame(() => {
    if (
      !target.isConnected ||
      document.activeElement !== target ||
      !target.matches(":focus-visible")
    )
      return;
    const header = document.querySelector("header.app-header");
    const safeTop = (header?.getBoundingClientRect().bottom ?? 0) + 8;
    const safeBottom = window.innerHeight - 8;
    const bounds = target.getBoundingClientRect();
    const delta =
      bounds.top < safeTop
        ? bounds.top - safeTop
        : bounds.bottom > safeBottom
          ? bounds.bottom - safeBottom
          : 0;
    if (delta) window.scrollBy({ top: delta, behavior: "instant" });
  });
}

// Next can restore any previously visited snapshot of a range on Back or Forward.
const snapshotVersions = new WeakMap<KnowledgeMapData, number>();
const rangeVersions = new Map<number, number>();

function rangeVersion(band: number) {
  return rangeVersions.get(band) ?? 0;
}

function wordHref(
  band: number,
  rank: number,
  active: boolean,
  revisit: boolean,
) {
  return `/discovery?band=${band}&word=${rank}${active ? "&active=1" : ""}${revisit ? "&revisit=1" : ""}`;
}

/** Continue from the chosen word, wrapping only when no later new word remains. */
function nextUnseenInRange(data: KnowledgeMapData, afterRank: number) {
  const unseen = (word: KnowledgeMapData["currentBand"]["items"][number]) =>
    word.rank !== afterRank && isUnseen(word);
  return (
    data.currentBand.items.find(
      (word) => word.rank > afterRank && unseen(word),
    ) ?? data.currentBand.items.find(unseen)
  );
}

async function loadDiscoveryPosition(
  client: ReturnType<typeof createApiClient>,
  current: KnowledgeMapData,
  word: MapWord | undefined,
) {
  if (word) return { data: current, word };
  const target = nextDiscoveryTarget(current);
  if (!target) return { data: current, word: undefined };
  const data = (await client.getKnowledgeMap(target.band)).data;
  const nextWord = discoveryWordInRange(data);
  if (!nextWord) throw new Error("The next unseen word could not be confirmed");
  return { data, word: nextWord };
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
  initiallyRevisiting = false,
}: {
  initialData: KnowledgeMapData;
  initialRank: number;
  initiallyActive?: boolean;
  initiallyRevisiting?: boolean;
}) {
  const [data, setData] = useState(initialData);
  const [rank, setRank] = useState(initialRank);
  const [started, setStarted] = useState(initiallyActive);
  const [revisit, setRevisit] = useState(initiallyRevisiting);
  const [restoring] = useState(() => {
    const version = snapshotVersions.get(initialData);
    return (
      version !== undefined &&
      version !== rangeVersion(initialData.currentBand.number)
    );
  });
  const router = useRouter();
  const serverParams = useSearchParams();
  const serverSearch = serverParams.size ? `?${serverParams.toString()}` : "";
  const currentSearch = useSyncExternalStore(
    subscribeDiscoveryLocation,
    browserSearchSnapshot,
    () => serverSearch,
  );
  const [navigating, startNavigation] = useTransition();
  const [muted, setMuted] = useState(true);
  const [audioReady, setAudioReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [done, setDone] = useState(false);
  const [savedChoice, setSavedChoice] = useState<SavedChoice | null>(null);
  const [speech, setSpeech] = useState<PronunciationState>({ status: "idle" });
  const pending = useRef<PendingDecision | null>(null);
  const lock = useRef(false);
  const speechOwner = useRef(Symbol("discovery"));
  const lastSpoken = useRef("");
  const titleRef = useRef<HTMLHeadingElement>(null);
  const errorRef = useRef<HTMLDivElement>(null);
  const item = revisit
    ? data.currentBand.items.find((word) => word.rank === rank)
    : discoveryWordInRange(data, rank);
  const band = data.currentBand.number;
  const bandSummary = data.bands.find((value) => value.number === band);
  const displayedRank = item?.rank;
  const displayedWord = item?.wordText;
  const displayedMeaning = item?.meaningId;
  const displayedRevision = item?.meaningRevision;
  const navigationLocked =
    busy || navigating || conflict || Boolean(pending.current);
  // Preload the same candidate a confirmed choice would display, including wrap.
  const upcoming = nextUnseenInRange(data, displayedRank ?? 0);
  const confirmedSave =
    savedChoice &&
    !navigationLocked &&
    !revisit &&
    savedChoice.band === band &&
    savedChoice.rank === displayedRank
      ? savedChoice
      : null;
  const savedMeaning = confirmedSave ? (
    <SavedMeaning choice={confirmedSave} />
  ) : null;
  const saveAnnouncement = (
    <p
      key="save-confirmation"
      role="status"
      aria-label="Save confirmation"
      aria-live="polite"
      aria-atomic="true"
      className="sr-only"
    >
      {confirmedSave ? `Saved ${confirmedSave.word.wordText} for review.` : ""}
    </p>
  );

  const firstRank = bandSummary?.startRank ?? 1;
  const lastRank = bandSummary?.endRank ?? data.totalWords;

  const historyWord = useRef({
    data,
    rank,
    started,
    revisit,
    blocked: navigationLocked,
  });
  historyWord.current = {
    data,
    rank,
    started,
    revisit,
    blocked: navigationLocked,
  };

  useEffect(() => {
    if (restoring) window.location.reload();
    else
      snapshotVersions.set(
        initialData,
        rangeVersion(initialData.currentBand.number),
      );
  }, [initialData, restoring]);

  // Owned writes preserve Next's opaque history state. Observe the actual
  // browser location (as DiscoveryExitLink does), because Next searchParams
  // need not update for those entries or their Back/Forward navigation.
  useEffect(() => {
    if (restoring) return;
    // Next commits its URL after rendering the new route. Use the live URL
    // here so a completed range cannot restore the previous route's selection.
    const searchParams = new URLSearchParams(window.location.search);
    const requested = Number(searchParams.get("word"));
    if (
      Number(searchParams.get("band")) === band &&
      Number.isInteger(requested) &&
      requested >= firstRank &&
      requested <= lastRank
    ) {
      const current = historyWord.current;
      const active = searchParams.get("active") === "1";
      const explicitlySelected = searchParams.get("revisit") === "1";
      if (
        current.blocked &&
        (requested !== current.rank ||
          active !== current.started ||
          explicitlySelected !== current.revisit)
      ) {
        // Back/Forward bypass disabled controls. Restore the pending word as a
        // new entry, retaining the attempted destination underneath it so Back
        // works normally after the original decision has been confirmed.
        window.history.pushState(
          window.history.state,
          "",
          wordHref(band, current.rank, current.started, current.revisit),
        );
        window.dispatchEvent(new Event(SESSION_LOCATION_CHANGED));
        return;
      }
      const word = explicitlySelected
        ? current.data.currentBand.items.find((item) => item.rank === requested)
        : discoveryWordInRange(current.data, requested);
      if (!word) {
        // Completion removes revisit from its own URL. Retain the confirmed
        // receipt for that same position; a different history selection clears it.
        if (
          requested !== current.rank ||
          active !== current.started ||
          explicitlySelected !== current.revisit
        )
          setSavedChoice(null);
        setDone(true);
        return;
      }
      if (
        word.rank !== current.rank ||
        active !== current.started ||
        explicitlySelected !== current.revisit
      )
        setSavedChoice(null);
      // Keep canonicalization from overwriting a history selection before its
      // card has rendered.
      historyWord.current = {
        ...current,
        rank: word.rank,
        started: active,
        revisit: explicitlySelected,
      };
      setRank(word.rank);
      setStarted(active);
      setRevisit(explicitlySelected);
      setDone(false);
      if (word.rank !== requested)
        replaceSessionLocation(
          wordHref(band, word.rank, active, explicitlySelected),
        );
    }
  }, [currentSearch, band, firstRank, lastRank, restoring]);

  function navigateToWord(nextBand: number, nextRank: number) {
    if (lock.current || navigationLocked) return;
    deviceSpeech.stop(speechOwner.current);
    setError("");
    setSavedChoice(null);
    const href = wordHref(nextBand, nextRank, true, true);
    if (nextBand !== band) {
      startNavigation(() => router.push(href));
      return;
    }
    window.history.pushState(window.history.state, "", href);
    window.dispatchEvent(new Event(SESSION_LOCATION_CHANGED));
    setStarted(true);
    setRevisit(true);
    setRank(nextRank);
    if (nextRank === item?.rank) titleRef.current?.focus();
  }

  function navigateToRange(nextBand: number) {
    if (lock.current || navigationLocked) return;
    deviceSpeech.stop(speechOwner.current);
    setError("");
    setSavedChoice(null);
    startNavigation(() => router.push(`/discovery?band=${nextBand}&active=1`));
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
      setMuted(sessionStorage.getItem("vocanova-discovery-muted") !== "false");
    } catch {
      /* Device storage may be unavailable. */
    }
    setAudioReady(true);
    return () => deviceSpeech.release(owner);
  }, []);

  useEffect(() => {
    if (done) {
      if (new URLSearchParams(window.location.search).get("revisit") === "1")
        replaceSessionLocation(wordHref(band, rank, started, false));
      titleRef.current?.focus();
      return;
    }
    if (displayedRank === undefined || restoring) return;
    const current = historyWord.current;
    if (
      current.rank !== rank ||
      current.started !== started ||
      current.revisit !== revisit
    )
      return;
    // The server may select a later unseen range before Start. Canonicalize
    // its address now so Back preserves that real range without another read.
    const href = wordHref(band, displayedRank, started, revisit);
    if (window.location.search !== href.slice(href.indexOf("?")))
      replaceSessionLocation(href);
    if (started) titleRef.current?.focus();
  }, [started, displayedRank, rank, band, done, restoring, revisit]);

  useEffect(() => {
    if (
      !audioReady ||
      !started ||
      displayedRank === undefined ||
      !displayedWord ||
      restoring ||
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
    restoring,
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
      setSavedChoice(null);
      pending.current = {
        body: {
          meaningId: item.meaningId,
          decision,
          expectedRevision: item.revision,
        },
        key: crypto.randomUUID(),
        rank: item.rank,
        word: {
          wordSlug: item.wordSlug,
          wordText: item.wordText,
          shortDefinition: item.shortDefinition,
          learnerDefinition: item.learnerDefinition,
        },
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
      rangeVersions.set(band, rangeVersion(band) + 1);
      // Use the meaning's state as the server reports it. When it is exactly
      // what this choice produces, continue within the range without
      // re-reading all 100 words; a later word changed elsewhere still
      // surfaces as a revision conflict when it is chosen.
      const state = confirmedWordState(result, request.body);
      if (state && confirmsChoice(state, request.body.decision)) {
        const local = applyWordState(data, request.body.meaningId, state);
        const nextInRange = nextUnseenInRange(local, request.rank);
        if (nextInRange) {
          pending.current = null;
          if (request.body.decision === "learn")
            setSavedChoice({
              meaningId: request.body.meaningId,
              word: request.word,
              band,
              rank: nextInRange.rank,
            });
          setData(local);
          setRank(nextInRange.rank);
          setRevisit(false);
          return;
        }
      }
      // A replayed receipt that reports a different current state (changed
      // since, in another tab or by a removal), an incomplete response or the
      // end of the range: read confirmed state before choosing the next unseen
      // word. Keep the exact pending intent until that range loads.
      const client = createApiClient();
      const current = (await client.getKnowledgeMap(band)).data;
      const currentDecision = current.currentBand.items.find(
        (word) => word.meaningId === request.body.meaningId,
      );
      // An incomplete receipt can still be confirmed by this fresh read. If
      // another tab changed/removed the word, retain the existing reconciliation
      // policy instead of treating our original intent as a confirmed choice.
      const nextAfterChoice =
        currentDecision &&
        confirmsChoice(currentDecision, request.body.decision)
          ? nextUnseenInRange(current, request.rank)
          : undefined;
      const { data: fresh, word: nextWord } = await loadDiscoveryPosition(
        client,
        current,
        nextAfterChoice ?? discoveryWordInRange(current),
      );
      pending.current = null;
      if (
        request.body.decision === "learn" &&
        currentDecision &&
        confirmsChoice(currentDecision, "learn")
      )
        setSavedChoice({
          meaningId: request.body.meaningId,
          word: request.word,
          band: fresh.currentBand.number,
          // Completion has no next rank; displayedRank is also undefined there.
          rank: nextWord?.rank,
        });
      setData(fresh);
      setRevisit(false);
      if (!nextWord) setDone(true);
      else setRank(nextWord.rank);
    } catch (failure) {
      if (failure instanceof ApiResponseError && failure.status === 409) {
        rangeVersions.set(band, rangeVersion(band) + 1);
        pending.current = null;
        setConflict(true);
        try {
          const nextWord = await refreshWord();
          setConflict(false);
          setError(
            nextWord?.rank === request.rank
              ? "This word changed elsewhere. Its latest state is shown. Choose again."
              : "This word changed elsewhere. Continue with the next new word.",
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

  async function refreshWord() {
    const client = createApiClient();
    const current = (await client.getKnowledgeMap(band)).data;
    const position = await loadDiscoveryPosition(
      client,
      current,
      revisit
        ? current.currentBand.items.find((word) => word.rank === rank)
        : discoveryWordInRange(current, rank),
    );
    setData(position.data);
    if (position.word) {
      setRank(position.word.rank);
      if (position.word.rank !== rank) setRevisit(false);
    } else setDone(true);
    return position.word;
  }

  async function reloadWord() {
    if (lock.current) return;
    setSavedChoice(null);
    lock.current = true;
    setBusy(true);
    try {
      await refreshWord();
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

  if (restoring)
    return (
      <p className={styles.focus} role="status">
        Loading word…
      </p>
    );

  if (done)
    return (
      <div className={styles.focus} onFocusCapture={revealFocusedControl}>
        {saveAnnouncement}
        <DiscoveryFinished data={data} active={started} headingRef={titleRef} />
        {savedMeaning}
      </div>
    );

  if (!item)
    return (
      <div className={styles.focus} onFocusCapture={revealFocusedControl}>
        {saveAnnouncement}
        <p role="alert">This word could not load.</p>
        <Link href={journeyRangeHref(band)} className={styles.secondary}>
          Back to Journey
        </Link>
      </div>
    );

  const rangeName = bandSummary
    ? `Words ${bandSummary.startRank.toLocaleString()}–${bandSummary.endRank.toLocaleString()}`
    : "";

  return (
    <div className={styles.focus} onFocusCapture={revealFocusedControl}>
      {saveAnnouncement}
      <div className={styles.top}>
        <span className={styles.counter}>{rangeName}</span>
        <span className={styles.counter}>
          Word {item.rank.toLocaleString()} of{" "}
          {data.totalWords.toLocaleString()}
        </span>
      </div>
      <DiscoveryNavigation
        data={data}
        rank={item.rank}
        locked={navigationLocked}
        onPrevious={previousWord}
        onRange={navigateToRange}
        onWord={(nextRank) => navigateToWord(band, nextRank)}
      />
      {navigating && (
        <p role="status" className={styles.status}>
          Loading word…
        </p>
      )}
      {!started ? (
        <section
          aria-labelledby="discovery-intro"
          className="learning-surface mt-6 p-4 sm:p-6"
        >
          <LearningIntro
            id="discovery-intro"
            headingLevel={1}
            eyebrow="Your next useful words"
            title="Discover your next words"
            description="Listen, then choose Learn or Already know."
          />
          <LearningNote className="mt-4">
            Learn saves this meaning for review. Already know records your own
            choice.
          </LearningNote>
          <div className={`${styles.startActions} mt-5`}>
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
            <p className="learning-eyebrow mt-4">This meaning</p>
            <p className={styles.definition}>
              {getPrimaryDefinition(
                item.shortDefinition,
                item.learnerDefinition,
              )}
            </p>
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
          {savedMeaning}
        </>
      )}
      {started && (
        <DiscoveryWordDetails
          key={`${item.meaningId}:${item.meaningRevision ?? ""}`}
          wordSlug={item.wordSlug}
          meaningId={item.meaningId}
          meaningRevision={item.meaningRevision}
          exampleSentence={item.exampleSentence}

          displayedDefinition={getPrimaryDefinition(
            item.shortDefinition,
            item.learnerDefinition,
          )}
        />
      )}
    </div>
  );
}
