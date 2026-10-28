import type { CSSProperties } from "react";
import Link from "@/ui/link";
import {
  journeyRangeHref,
  nextDiscoveryTarget,
  wordState,
  wordStateLabel,
  type KnowledgeMapData,
} from "./models";
import { AtlasKeys } from "./atlas-keys";
import styles from "./workspace.module.css";

type Band = KnowledgeMapData["bands"][number];

// Known words fill a square from the bottom, then learning words above them.
function fill(band: Band): CSSProperties {
  return {
    "--known": `${(band.knownCount / band.totalWords) * 100}%`,
    "--learning": `${(band.learningCount / band.totalWords) * 100}%`,
  } as CSSProperties;
}

function rangeLabel(band: Pick<Band, "startRank" | "endRank">) {
  return `${band.startRank.toLocaleString()}–${band.endRank.toLocaleString()}`;
}

/** The selected 100 words: where you are, what is next and the one action. */
export function RangeCard({ data }: { data: KnowledgeMapData }) {
  const band = data.currentBand.number;
  const summary = data.bands.find((item) => item.number === band);
  const next = nextDiscoveryTarget(data);
  const nextInRange =
    next?.band === band
      ? data.currentBand.items.find((item) => item.rank === next.rank)
      : undefined;
  const items = data.currentBand.items;
  return (
    <section
      id="range-words"
      aria-labelledby="range-title"
      className={styles.rangeCard}
    >
      <div className={styles.rangeHead}>
        <h2 id="range-title" className={styles.rangeTitle}>
          Words {summary ? rangeLabel(summary) : ""}
        </h2>
        <nav aria-label="Adjacent word ranges" className={styles.rangePaging}>
          {band > 1 && (
            <Link className={styles.pageLink} href={journeyRangeHref(band - 1)}>
              Previous 100
            </Link>
          )}
          {band < 100 && (
            <Link className={styles.pageLink} href={journeyRangeHref(band + 1)}>
              Next 100
            </Link>
          )}
        </nav>
      </div>
      {summary ? (
        <p className={styles.rangeCounts}>
          In this range: {summary.knownCount} already known,{" "}
          {summary.learningCount} learning, {summary.unseenCount} new.
        </p>
      ) : null}
      {/* One mark per word in rank order; the counts above carry the meaning. */}
      <div className={styles.strip} aria-hidden="true">
        {items.map((item) => (
          <span
            key={item.meaningId}
            data-state={wordState(item)}
            data-next={item.rank === nextInRange?.rank ? "true" : undefined}
          />
        ))}
      </div>
      {next ? (
        <div className={styles.nextUp}>
          <p className={styles.nextUpLabel}>
            {nextInRange ? (
              <>
                Next up{" "}
                <span className={`headword ${styles.nextUpWord}`}>
                  {nextInRange.wordText}
                </span>
              </>
            ) : (
              "Every word here is assessed. Continue in the next open range."
            )}
          </p>
          <Link
            className="vn-btn vn-btn-primary"
            href={`/discovery?band=${next.band}${next.rank === undefined ? "" : `&word=${next.rank}`}`}
          >
            Discover words
          </Link>
        </div>
      ) : (
        <p className={styles.status}>No new words left to discover.</p>
      )}
      <details className={styles.wordList}>
        <summary>Words in this range</summary>
        <div className={styles.words}>
          {items.map((word) => (
            <Link
              key={word.meaningId}
              prefetch={false}
              prefetchOnIntent={false}
              href={`/discovery?band=${band}&word=${word.rank}`}
              className={styles.word}
              data-state={wordState(word)}
            >
              <span className={styles.rank}>{word.rank.toLocaleString()}</span>
              <strong>{word.wordText}</strong>
              <span className={styles.state}>{wordStateLabel(word)}</span>
            </Link>
          ))}
        </div>
      </details>
    </section>
  );
}

/** All 10,000 words as 100 ranges: rows of 1,000, filled by what you know. */
export function WordAtlas({ data }: { data: KnowledgeMapData }) {
  const band = data.currentBand.number;
  const known = data.bands.reduce((sum, item) => sum + item.knownCount, 0);
  const learning = data.bands.reduce(
    (sum, item) => sum + item.learningCount,
    0,
  );
  const rows = Array.from({ length: 10 }, (_, row) =>
    data.bands.slice(row * 10, row * 10 + 10),
  );
  return (
    <section aria-labelledby="atlas-title" className={styles.atlasSection}>
      <h2 id="atlas-title" className="section-title">
        Your word map
      </h2>
      <p className={styles.atlasLede}>
        Each square holds 100 words, most common first. Squares fill as you mark
        words you know or are learning.
      </p>
      <ul className={styles.legend} aria-label="Map progress">
        <li>
          <span className={styles.knownDot} aria-hidden="true" />
          {known.toLocaleString()} already know
        </li>
        <li>
          <span className={styles.learningDot} aria-hidden="true" />
          {learning.toLocaleString()} learning
        </li>
        <li>
          <span className={styles.newDot} aria-hidden="true" />
          {(data.totalWords - known - learning).toLocaleString()} new
        </li>
      </ul>
      {/* Wide screens: every square is a link of at least 44px. */}
      <AtlasKeys className={styles.atlasWide}>
        <nav aria-label="Word ranges" className={styles.atlas}>
          {rows.map((row, index) => (
            <div key={index} className={styles.atlasRow}>
              <span className={styles.atlasRowLabel} aria-hidden="true">
                {(index * 1000 + 1).toLocaleString()}
              </span>
              {row.map((item) => (
                <Link
                  key={item.number}
                  prefetch={false}
                  prefetchOnIntent={false}
                  href={journeyRangeHref(item.number)}
                  className={styles.atlasCell}
                  aria-current={item.number === band ? "true" : undefined}
                  aria-label={`Words ${item.startRank} to ${item.endRank}, ${item.knownCount} already know, ${item.learningCount} learning, ${item.unseenCount} new`}
                  style={fill(item)}
                />
              ))}
            </div>
          ))}
        </nav>
      </AtlasKeys>
      {/* Phones: the same map as a picture, and a full-size range chooser
          instead of 100 squares too small to tap. */}
      <div className={styles.atlasCompact}>
        <div className={styles.atlas} aria-hidden="true">
          {rows.map((row, index) => (
            <div key={index} className={styles.atlasRow}>
              <span className={styles.atlasRowLabel}>
                {(index * 1000 + 1).toLocaleString()}
              </span>
              {row.map((item) => (
                <span
                  key={item.number}
                  className={styles.atlasCell}
                  data-current={item.number === band ? "true" : undefined}
                  style={fill(item)}
                />
              ))}
            </div>
          ))}
        </div>
        <form action="/discover#range-words" className={styles.rangeForm}>
          <label htmlFor="range-choice">Go to a range</label>
          <div>
            <select id="range-choice" name="band" defaultValue={band}>
              {data.bands.map((item) => (
                <option key={item.number} value={item.number}>
                  {`Words ${item.startRank.toLocaleString()}–${item.endRank.toLocaleString()}: ${item.knownCount} known, ${item.learningCount} learning`}
                </option>
              ))}
            </select>
            <button type="submit" className="vn-btn vn-btn-secondary">
              Show
            </button>
          </div>
        </form>
      </div>
      <details className={styles.ordering}>
        <summary>Why this order?</summary>
        <p>{data.ordering.title}</p>
        <p>{data.ordering.description}</p>
        {data.ordering.sourceUrl && (
          <a href={data.ordering.sourceUrl} target="_blank" rel="noreferrer">
            Word order source (opens a new tab)
          </a>
        )}
      </details>
    </section>
  );
}
