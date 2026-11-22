"use client";

import { RangeMiniGrid } from "../knowledge-map/range-mini-grid";
import { rangeProgressLabel } from "../knowledge-map/range-preview-model";
import { StatusMessage } from "@/ui/status-message";
import { Select, TextInput } from "@/ui/controls";
import { useEffect, useRef, useState, type MouseEvent } from "react";
import {
  WORD_STATES,
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
  onRange,
  onWord,
}: {
  data: KnowledgeMapData;
  rank: number;
  locked: boolean;
  onRange: (band: number) => void;
  onWord: (rank: number) => void;
}) {
  const [picker, setPicker] = useState<Picker | null>(null);
  const [query, setQuery] = useState("");
  const [stateFilter, setStateFilter] = useState("");
  const [grid, setGrid] = useState(false);
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
    setStateFilter("");
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
      (!stateFilter || wordState(word) === stateFilter) &&
      (!filter ||
        (byRank
          ? String(word.rank).startsWith(numericFilter)
          : word.wordText.toLocaleLowerCase().includes(filter))),
  );
  const choosingRange = picker === "range";

  return (
    <>
      <div className={styles.controls}>
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
          <div
            className={styles.panel}
            onFocusCapture={(event) => {
              // Enlarged text can make the panel scroll as well as the results.
              const control = event.target;
              requestAnimationFrame(() => {
                if (
                  control === document.activeElement &&
                  control.matches(":focus-visible")
                )
                  control.scrollIntoView({
                    block: "nearest",
                    inline: "nearest",
                  });
              });
            }}
          >
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
              <TextInput
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
            {!choosingRange && (
              <div className={styles.scanControls}>
                <Select
                  aria-label="Knowledge state"
                  value={stateFilter}
                  onChange={(event) => {
                    setStateFilter(event.target.value);
                    if (options.current) options.current.scrollTop = 0;
                  }}
                >
                  <option value="">All states</option>
                  {Object.entries(WORD_STATES).map(([state, { label }]) => (
                    <option key={state} value={state}>
                      {label}
                    </option>
                  ))}
                </Select>
                <button
                  type="button"
                  aria-pressed={grid}
                  className={styles.gridToggle}
                  onClick={() => setGrid(!grid)}
                >
                  Grid view
                </button>
              </div>
            )}
            {(choosingRange ? ranges.length : words.length) === 0 ? (
              <StatusMessage as="p" className={styles.empty} role="status">
                {choosingRange ? "No ranges match." : "No words match."}
              </StatusMessage>
            ) : (
              <ul
                ref={options}
                className={styles.options}
                data-grid={!choosingRange && grid}
              >
                {choosingRange
                  ? ranges.map((range) => (
                      <li
                        key={range.number}
                        ref={range.number === band ? current : undefined}
                      >
                        <button
                          type="button"
                          className={`${styles.option} ${styles.rangeOption}`}
                          aria-label={rangeLabel(range)}
                          aria-describedby={`discovery-range-counts-${range.number}`}
                          aria-current={
                            range.number === band ? "true" : undefined
                          }
                          disabled={locked}
                          onClick={() => {
                            dialog.current?.close();
                            onRange(range.number);
                          }}
                        >
                          <span className={styles.rangePreview}>
                            <RangeMiniGrid band={range} />
                          </span>
                          <span>
                            <span className={styles.rangeName}>
                              {rangeLabel(range)}
                            </span>
                            <span
                              id={`discovery-range-counts-${range.number}`}
                              className={styles.rangeCounts}
                            >
                              {rangeProgressLabel(range, " · ", "already know")}
                            </span>
                          </span>
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
