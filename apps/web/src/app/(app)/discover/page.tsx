import Link from "@/ui/link";
import { Suspense } from "react";
import { ApiResponseError, type Situation } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { VocabularySources } from "@/ui/vocabulary-sources";
import { RangeCard, WordAtlas } from "../knowledge-map/map-view";
import { validBand, type KnowledgeMapData } from "../knowledge-map/models";
import styles from "../knowledge-map/workspace.module.css";
import { SituationIcon } from "../_components/situation-icon";
import { getSituationOutcome } from "../_components/situation-presentation";
import { formatLevelBand } from "./_components/level-band";

const PLACE_LIMIT = 6;

type SituationsRead =
  { items: Situation[]; error?: undefined } | { items: null; error: unknown };

// Settles instead of rejecting, so it can start alongside the map read safely.
async function readSituations(): Promise<SituationsRead> {
  try {
    const client = await createServerApiClient();
    return { items: (await client.listJourneySituations()).data.items };
  } catch (error) {
    return { items: null, error };
  }
}

/**
 * Optional situation cards stream in their own boundary, so a slow or failed
 * read never holds back the word map or the other ways to learn.
 */
async function PlaceCards({ read }: { read: Promise<SituationsRead> }) {
  const { items: situations, error } = await read;
  if (error instanceof ApiResponseError && error.status === 401)
    requireAuthRedirect(error, "/discover");
  const places = [...(situations ?? [])]
    .sort((first, second) => first.displayOrder - second.displayOrder)
    .slice(0, PLACE_LIMIT);
  return places.length > 0 ? (
    <ul className={styles.places}>
      {places.map((situation) => {
        const description =
          getSituationOutcome(situation.slug) ?? situation.shortDescription;
        const descriptionId = `place-${situation.slug}`;
        return (
          <li key={situation.slug} className={styles.place}>
            <span aria-hidden="true" className={styles.placeIcon}>
              <SituationIcon slug={situation.slug} />
            </span>
            <span className="min-w-0 flex-1">
              <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
                <Link
                  href={`/discover/${encodeURIComponent(situation.slug)}`}
                  // Situation pages have an instant loading state.
                  prefetch="auto"
                  aria-describedby={description ? descriptionId : undefined}
                  className={styles.placeLink}
                >
                  {situation.title}
                </Link>
                {situation.levelBand ? (
                  <span className={styles.level}>
                    {formatLevelBand(situation.levelBand)}
                  </span>
                ) : null}
              </span>
              {description ? (
                <span id={descriptionId} className={styles.placeText}>
                  {description}
                </span>
              ) : null}
            </span>
          </li>
        );
      })}
    </ul>
  ) : (
    <p role="status" className="mt-3 text-neutral-700">
      {situations
        ? "New situations are on the way."
        : "Situations could not load right now. The other ways to learn below are still available."}
    </p>
  );
}

function PlaceCardsPending() {
  return (
    <div aria-busy="true">
      <p role="status" className="sr-only">
        Loading situations…
      </p>
      <ul aria-hidden="true" className={styles.places}>
        {[0, 1, 2].map((index) => (
          <li
            key={index}
            className={`${styles.place} ${styles.placePending}`}
          />
        ))}
      </ul>
    </div>
  );
}

/** Journey's alternatives stay available whatever the map or situations do. */
function JourneyPlaces({ read }: { read: Promise<SituationsRead> }) {
  return (
    <section
      aria-labelledby="journey-places-heading"
      className={styles.placesSection}
    >
      <div className="flex flex-wrap items-end justify-between gap-x-4 gap-y-1">
        <div>
          <h2 id="journey-places-heading" className="section-title">
            Places to use English
          </h2>
          <p className="mt-0.5 text-sm text-neutral-600">
            Learn the words that belong to a real situation.
          </p>
        </div>
        <Link href="/courses" className="vn-text-link">
          All situations
        </Link>
      </div>
      <Suspense fallback={<PlaceCardsPending />}>
        <PlaceCards read={read} />
      </Suspense>
      <nav
        aria-label="More ways to learn"
        className="mt-4 flex flex-wrap gap-2"
      >
        {(
          [
            ["/courses", "Situation courses"],
            ["/stories", "Short stories"],
            ["/lists", "Personal lists"],
          ] as const
        ).map(([href, label]) => (
          <Link
            key={href}
            href={href}
            prefetch={href === "/stories" ? "auto" : false}
            className="vn-chip-link"
          >
            {label}
          </Link>
        ))}
      </nav>
    </section>
  );
}

type MapRead =
  { data: KnowledgeMapData; failed?: undefined } | { data: null; failed: true };

/**
 * An explicit `?band=` opens that range. Plain Journey opens where the learner
 * is: the first range that still has new words, using the same range metadata
 * guided discovery uses.
 */
async function readMap(
  requested: string | string[] | undefined,
): Promise<MapRead> {
  const band = validBand(requested);
  try {
    const client = await createServerApiClient();
    const first = (await client.getKnowledgeMap(band)).data;
    if (requested !== undefined) return { data: first };
    const frontier = first.bands.find((item) => item.unseenCount > 0)?.number;
    return {
      data:
        frontier && frontier !== band
          ? (await client.getKnowledgeMap(frontier)).data
          : first,
    };
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(
        error,
        requested === undefined ? "/discover" : `/discover?band=${band}`,
      );
    return { data: null, failed: true };
  }
}

export default async function JourneyPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const situationsRead = readSituations();
  const requested = (await searchParams).band;
  const map = await readMap(requested);
  const retryHref =
    requested === undefined
      ? "/discover"
      : `/discover?band=${validBand(requested)}`;
  return (
    <PageContainer className={styles.journey}>
      <header className={styles.journeyHead}>
        <h1 className="page-title">Journey</h1>
        <p className="page-lede">
          Learn the 10,000 most useful English words, 100 at a time, and the
          words that belong to real places.
        </p>
      </header>
      {/* Phones read range, map, then places; wide screens keep the map
          beside the range and places. */}
      <div className={styles.mapLayout}>
        {map.data ? (
          <>
            <RangeCard data={map.data} />
            <WordAtlas data={map.data} />
          </>
        ) : (
          <section aria-labelledby="map-error" className={styles.rangeCard}>
            <h2 id="map-error" className="section-title">
              Your word map
            </h2>
            <p role="alert" className="mt-2 text-neutral-800">
              Your map could not load.
            </p>
            <p className="mt-1 text-sm text-neutral-600">
              Your saved words are safe. Try again, or learn with a place below.
            </p>
            <a className="vn-btn vn-btn-secondary mt-3" href={retryHref}>
              Try again
            </a>
          </section>
        )}
        <JourneyPlaces read={situationsRead} />
      </div>
      {map.data ? (
        <VocabularySources attribution={map.data.attribution} />
      ) : null}
    </PageContainer>
  );
}
