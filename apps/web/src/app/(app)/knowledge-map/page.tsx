import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { VocabularySources } from "@/ui/vocabulary-sources";
import {
  validBand,
  nextDiscoveryTarget,
  wordState,
  wordStateLabel,
  type KnowledgeMapData,
} from "./models";
import styles from "./workspace.module.css";

export default async function KnowledgeMapPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  const band = validBand(query.band);
  let data: KnowledgeMapData | null = null;
  try {
    const client = await createServerApiClient();
    data = (await client.getKnowledgeMap(band)).data;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, `/knowledge-map?band=${band}`);
  }
  if (!data)
    return (
      <PageContainer>
        <h1 className={styles.title}>Knowledge map</h1>
        <p role="alert">Your map could not load.</p>
        <a className={styles.secondary} href={`/knowledge-map?band=${band}`}>
          Try again
        </a>
      </PageContainer>
    );
  const current = data.bands.find(
    (item) => item.number === data.currentBand.number,
  );
  const known = data.bands.reduce((sum, item) => sum + item.knownCount, 0);
  const learning = data.bands.reduce(
    (sum, item) => sum + item.learningCount,
    0,
  );
  const next = nextDiscoveryTarget(data);
  return (
    <PageContainer className={styles.wide}>
      <header className={`${styles.header} ${styles.mapHeader}`}>
        <div>
          <h1 className={styles.title}>Knowledge map</h1>
          {!next && (
            <p className={styles.status}>No new words left to discover.</p>
          )}
        </div>
        {next && (
          <Link
            className={styles.primary}
            href={`/discovery?band=${next.band}${next.rank === undefined ? "" : `&word=${next.rank}`}`}
          >
            Discover
          </Link>
        )}
      </header>
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
      <details className={styles.rangePicker}>
        <summary>Choose a range · 1–10,000</summary>
        <nav aria-label="Word ranges" className={styles.bands}>
          {data.bands.map((item) => (
            <Link
              key={item.number}
              prefetch={false}
              href={`/knowledge-map?band=${item.number}#range-words`}
              className={styles.band}
              aria-current={item.number === band ? "true" : undefined}
              aria-label={`Words ${item.startRank} to ${item.endRank}, ${item.knownCount} already know, ${item.learningCount} learning, ${item.unseenCount} new`}
            >
              <span>
                {item.startRank.toLocaleString()}–
                {item.endRank.toLocaleString()}
              </span>
              <span className={styles.bandMeter} aria-hidden="true">
                <span
                  className={styles.knownFill}
                  style={{
                    width: `${(item.knownCount / item.totalWords) * 100}%`,
                  }}
                />
                <span
                  className={styles.learningFill}
                  style={{
                    width: `${(item.learningCount / item.totalWords) * 100}%`,
                  }}
                />
              </span>
            </Link>
          ))}
        </nav>
      </details>
      <section
        id="range-words"
        className={styles.range}
        aria-labelledby="range-title"
      >
        <div className={styles.header}>
          <h2 id="range-title" className={styles.subtitle}>
            Words {current?.startRank.toLocaleString()}–
            {current?.endRank.toLocaleString()}
          </h2>
          <nav aria-label="Adjacent word ranges" className={styles.rangePaging}>
            {band > 1 && (
              <Link
                className={styles.back}
                href={`/knowledge-map?band=${band - 1}#range-words`}
              >
                Previous 100
              </Link>
            )}
            {band < 100 && (
              <Link
                className={styles.back}
                href={`/knowledge-map?band=${band + 1}#range-words`}
              >
                Next 100
              </Link>
            )}
          </nav>
        </div>
        <div className={styles.words}>
          {data.currentBand.items.map((word) => (
            <Link
              key={word.meaningId}
              prefetch={false}
              href={`/discovery?band=${band}&word=${word.rank}`}
              className={`${styles.word} ${styles[wordState(word)]}`}
            >
              <span className={styles.rank}>{word.rank.toLocaleString()}</span>
              <strong>{word.wordText}</strong>
              <span className={styles.state}>{wordStateLabel(word)}</span>
            </Link>
          ))}
        </div>
      </section>
      <VocabularySources attribution={data.attribution} />
    </PageContainer>
  );
}
