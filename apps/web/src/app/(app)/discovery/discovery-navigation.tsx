"use client";

import { useEffect, useRef, useState, type MouseEvent } from "react";
import {
  wordState,
  wordStateLabel,
  type KnowledgeMapData,
} from "../knowledge-map/models";
import workspace from "../knowledge-map/workspace.module.css";
import styles from "./discovery-navigation.module.css";

type Picker = "range" | "word";

function rangeLabel(range: KnowledgeMapData["bands"][number]) {
  return `${range.startRank.toLocaleString()}–${range.endRank.toLocaleString()}`;
}

export function DiscoveryNavigation({
  data,
  rank,
  locked,
  onPrevious,
  onRange,
  onWord,
}: {
  data: KnowledgeMapData;
  rank: number;
  locked: boolean;
  onPrevious: () => void;
  onRange: (band: number) => void;
  onWord: (rank: number) => void;
}) {
  const [picker, setPicker] = useState<Picker | null>(null);
  const [query, setQuery] = useState("");
  const dialog = useRef<HTMLDialogElement>(null);
  const search = useRef<HTMLInputElement>(null);
  const options = useRef<HTMLUListElement>(null);
  const current = useRef<HTMLLIElement>(null);
  const band = data.currentBand.number;
  const selectedRange = data.bands.find((range) => range.number === band);

  useEffect(() => {
    if (!picker || !dialog.current) return;
    if (!dialog.current.open) dialog.current.showModal();
    search.current?.focus({ preventScroll: true });
    if (options.current && current.current) {
      options.current.scrollTop =
        current.current.offsetTop - options.current.clientHeight / 3;
    }
  }, [picker]);

  function openPicker(event: MouseEvent<HTMLButtonElement>, next: Picker) {
    // Safari does not always focus a button on tap. Make it the return target.
    event.currentTarget.focus({ preventScroll: true });
    setQuery("");
    setPicker(next);
  }

  const filter = query.trim().toLocaleLowerCase();
  const numericFilter = filter.replaceAll(",", "");
  const byRank = /^\d+$/.test(numericFilter);
  const ranges = data.bands.filter(
    (range) =>
      !filter ||
      (byRank
        ? Number(numericFilter) >= range.startRank &&
          Number(numericFilter) <= range.endRank
        : rangeLabel(range).includes(filter.replace(/\s*-\s*/g, "–"))),
  );
  const words = data.currentBand.items.filter(
    (word) =>
      !filter ||
      (byRank
        ? String(word.rank).startsWith(numericFilter)
        : word.wordText.toLocaleLowerCase().includes(filter)),
  );
  const choosingRange = picker === "range";

  return (
    <>
      <div className={styles.controls}>
        <button
          type="button"
          className={workspace.secondary}
          onClick={onPrevious}
          disabled={locked || rank <= 1}
        >
          Previous word
        </button>
        <button
          type="button"
          className={styles.rangeTrigger}
          aria-label="Word range"
          aria-describedby="discovery-current-range"
          aria-haspopup="dialog"
          disabled={locked}
          onClick={(event) => openPicker(event, "range")}
        >
          <span>
            <span className={styles.caption}>Word range</span>
            <span id="discovery-current-range">
              {selectedRange ? rangeLabel(selectedRange) : band}
            </span>
          </span>
          <span aria-hidden="true">⌄</span>
        </button>
      </div>
      <button
        type="button"
        className={styles.wordTrigger}
        aria-label="Choose a word"
        aria-haspopup="dialog"
        disabled={locked}
        onClick={(event) => openPicker(event, "word")}
      >
        <span className={styles.triggerLabel}>
          Choose a word <span aria-hidden="true">⌄</span>
        </span>
        <span
          className={`${workspace.strip} ${styles.progress}`}
          aria-hidden="true"
        >
          {data.currentBand.items.map((word) => (
            <span
              key={word.meaningId}
              data-state={wordState(word)}
              data-next={word.rank === rank ? "true" : undefined}
            />
          ))}
        </span>
      </button>
      <dialog
        ref={dialog}
        className={styles.dialog}
        aria-labelledby="discovery-picker-title"
        onClose={() => setPicker(null)}
        onKeyDown={(event) => {
          // A search input can consume Escape to clear its value first.
          if (event.key === "Escape") {
            event.preventDefault();
            event.currentTarget.close();
          }
        }}
      >
        {picker && (
          <div className={styles.panel}>
            <header className={styles.panelHeader}>
              <h2 id="discovery-picker-title">
                {choosingRange ? "Choose a range" : "Choose a word"}
              </h2>
              <button
                type="button"
                className={styles.close}
                aria-label="Close picker"
                onClick={() => dialog.current?.close()}
              >
                <span aria-hidden="true">×</span>
              </button>
            </header>
            <label className={styles.search}>
              {choosingRange ? "Filter ranges" : "Filter words"}
              <input
                ref={search}
                type="search"
                value={query}
                placeholder={
                  choosingRange
                    ? "Enter a word number"
                    : "Enter a word or its number"
                }
                onChange={(event) => {
                  setQuery(event.target.value);
                  if (options.current) options.current.scrollTop = 0;
                }}
              />
            </label>
            {(choosingRange ? ranges.length : words.length) === 0 ? (
              <p className={styles.empty} role="status">
                {choosingRange ? "No ranges match." : "No words match."}
              </p>
            ) : (
              <ul ref={options} className={styles.options}>
                {choosingRange
                  ? ranges.map((range) => (
                      <li
                        key={range.number}
                        ref={range.number === band ? current : undefined}
                      >
                        <button
                          type="button"
                          className={styles.option}
                          aria-current={
                            range.number === band ? "true" : undefined
                          }
                          disabled={locked}
                          onClick={() => {
                            dialog.current?.close();
                            onRange(range.number);
                          }}
                        >
                          {rangeLabel(range)}
                        </button>
                      </li>
                    ))
                  : words.map((word) => (
                      <li
                        key={word.meaningId}
                        ref={word.rank === rank ? current : undefined}
                      >
                        <button
                          type="button"
                          className={`${styles.option} ${styles.wordOption}`}
                          aria-current={word.rank === rank ? "true" : undefined}
                          disabled={locked}
                          onClick={() => {
                            dialog.current?.close();
                            onWord(word.rank);
                          }}
                        >
                          <span className={styles.rank}>
                            {word.rank.toLocaleString()}
                          </span>
                          <span className={styles.word}>{word.wordText}</span>
                          <span
                            className={styles.state}
                            data-state={wordState(word)}
                          >
                            {wordStateLabel(word)}
                          </span>
                        </button>
                      </li>
                    ))}
              </ul>
            )}
          </div>
        )}
      </dialog>
    </>
  );
}
