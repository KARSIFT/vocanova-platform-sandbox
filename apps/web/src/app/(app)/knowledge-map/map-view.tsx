import type { CSSProperties } from "react";
import Link from "@/ui/link";
import {
  earlierDiscoveryBand,
  journeyRangeHref,
  nextDiscoveryTarget,
  wordState,
  wordStateLabel,
  type KnowledgeMapData,
} from "./models";
import { AtlasKeys } from "./atlas-keys";
import { JourneyRangeAnchor } from "./journey-range-anchor";
import { JourneyRangeSelect } from "./journey-range-select";
import styles from "./workspace.module.css";

type Band = KnowledgeMapData["bands"][number];

function legacyFill(band: Band): CSSProperties | undefined {
  if (band.wordStates || band.unseenCount === band.totalWords) return;
  return {
    "--known": `${(band.knownCount / band.totalWords) * 100}%`,
    "--learning": `${(band.learningCount / band.totalWords) * 100}%`,
  } as CSSProperties;
}

function learnedCount(band: Band) {
  return band.wordStates?.match(/m/g)?.length ?? 0;
}

function progressLabel(band: Band, learned = learnedCount(band)) {
  return `${band.knownCount} already know, ${band.learningCount - learned} learning, ${learned ? `${learned} learned, ` : ""}${band.unseenCount} new`;
}

/** One path per tracked status keeps 10,000 marks out of the DOM. */
function AtlasGrid({ band }: { band: Band }) {
  if (!band.wordStates && band.unseenCount < band.totalWords) return null;
  const paths = { k: "", l: "", m: "" };
  for (const [index, code] of Array.from(band.wordStates ?? "").entries()) {
    if (code !== "k" && code !== "l" && code !== "m") continue;
    const x = (index % 10) * 10 + 1;
    const y = Math.floor(index / 10) * 10 + 1;
    paths[code] += `M${x} ${y}h8v8h-8z`;
  }
  return (
    <svg
      viewBox="0 0 100 100"
      className={styles.atlasGrid}
      aria-hidden="true"
      focusable="false"
    >
      <rect width="100" height="100" fill="url(#word-atlas-unseen)" />
      {paths.k ? <path data-state="known" d={paths.k} /> : null}
      {paths.l ? <path data-state="learning" d={paths.l} /> : null}
      {paths.m ? <path data-state="learned" d={paths.m} /> : null}
    </svg>
  );
}

function rangeLabel(band: Pick<Band, "startRank" | "endRank">) {
  return `${band.startRank.toLocaleString()}–${band.endRank.toLocaleString()}`;
}

/** The selected 100 words: where you are, what is next and the one action. */
export function RangeCard({ data }: { data: KnowledgeMapData }) {
  const band = data.currentBand.number;
  const summary = data.bands.find((item) => item.number === band);
  const learned = summary ? learnedCount(summary) : 0;
  const next = nextDiscoveryTarget(data);
  const targetBand = next?.band ?? earlierDiscoveryBand(data)?.number;
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
      <JourneyRangeAnchor key={band} />
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
          {summary.knownCount} known · {summary.learningCount - learned}{" "}
          learning
          {learned > 0 ? ` · ${learned} learned` : ""} · {summary.unseenCount}{" "}
          new
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
      {targetBand !== undefined ? (
        <div className={styles.nextUp}>
          <p className={styles.nextUpLabel}>
            {nextInRange ? (
              <>
                Next up{" "}
                <span className={`headword ${styles.nextUpWord}`}>
                  {nextInRange.wordText}
                </span>
              </>
            ) : next ? (
              "Every word here is assessed. Continue in the next open range."
            ) : (
              "New words are available in an earlier range."
            )}
          </p>
          <Link
            className="vn-btn vn-btn-primary"
            href={`/discovery?band=${targetBand}${next?.rank === undefined ? "" : `&word=${next.rank}`}`}
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
              href={`/discovery?band=${band}&word=${word.rank}&revisit=1`}
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

/** All 10,000 words as 100 ranges with their states in rank order. */
export function WordAtlas({ data }: { data: KnowledgeMapData }) {
  const band = data.currentBand.number;
  const known = data.bands.reduce((sum, item) => sum + item.knownCount, 0);
  const saved = data.bands.reduce((sum, item) => sum + item.learningCount, 0);
  const learned = data.bands.reduce((sum, item) => sum + learnedCount(item), 0);
  const rows = Array.from({ length: 10 }, (_, row) =>
    data.bands.slice(row * 10, row * 10 + 10),
  );
  return (
    <section aria-labelledby="atlas-title" className={styles.atlasSection}>
      <h2 id="atlas-title" className="section-title">
        Your word map
      </h2>
      <p className={styles.atlasLede}>Each square shows 100 words in order.</p>
      <ul className={styles.legend} aria-label="Map progress">
        <li>
          <span className={styles.knownDot} aria-hidden="true" />
          {known.toLocaleString()} already know
        </li>
        <li>
          <span className={styles.learningDot} aria-hidden="true" />
          {(saved - learned).toLocaleString()} learning
        </li>
        <li>
          <span className={styles.learnedDot} aria-hidden="true" />
          {learned.toLocaleString()} learned
        </li>
        <li>
          <span className={styles.newDot} aria-hidden="true" />
          {(data.totalWords - known - saved).toLocaleString()} new
        </li>
      </ul>
      <svg
        className={styles.atlasDefinitions}
        width="0"
        height="0"
        aria-hidden="true"
        focusable="false"
      >
        <defs>
          <pattern
            id="word-atlas-unseen"
            width="10"
            height="10"
            patternUnits="userSpaceOnUse"
          >
            <rect
              x="1"
              y="1"
              width="8"
              height="8"
              className={styles.atlasUnseen}
            />
          </pattern>
        </defs>
      </svg>
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
                  href={`/discovery?band=${item.number}&active=1`}
                  className={styles.atlasCell}
                  aria-current={item.number === band ? "true" : undefined}
                  aria-label={`Words ${item.startRank} to ${item.endRank}, ${progressLabel(item)}`}
                  style={legacyFill(item)}
                >
                  <AtlasGrid band={item} />
                </Link>
              ))}
            </div>
          ))}
        </nav>
      </AtlasKeys>
      {/* Phones: the same map as a picture, and a full-size range chooser
          instead of 100 squares too small to tap. */}
      <div className={styles.atlasCompact}>
        <form action="/discovery" className={styles.rangeForm}>
          <input type="hidden" name="active" value="1" />
          <label htmlFor="range-choice">Go to a range</label>
          <span id="range-choice-navigation" className="sr-only">
            Selecting a range opens Discovery.
          </span>
          <div>
            <JourneyRangeSelect band={band}>
              {data.bands.map((item) => (
                <option key={item.number} value={item.number}>
                  {`${item.startRank.toLocaleString()}–${item.endRank.toLocaleString()}`}
                </option>
              ))}
            </JourneyRangeSelect>
          </div>
        </form>
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
                  style={legacyFill(item)}
                >
                  <AtlasGrid band={item} />
                </span>
              ))}
            </div>
          ))}
        </div>
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
