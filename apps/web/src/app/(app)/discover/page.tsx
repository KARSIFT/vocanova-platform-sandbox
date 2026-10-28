import Link from "next/link";
import { Suspense } from "react";
import { ApiResponseError, type Situation } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { KnowledgeMapView } from "../knowledge-map/map-view";
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
 * read never holds back the knowledge map or the other ways to learn.
 */
async function PlaceCards({ read }: { read: Promise<SituationsRead> }) {
  const { items: situations, error } = await read;
  if (error instanceof ApiResponseError && error.status === 401)
    requireAuthRedirect(error, "/discover");
  const places = [...(situations ?? [])]
    .sort((first, second) => first.displayOrder - second.displayOrder)
    .slice(0, PLACE_LIMIT);
  return (
    <>
      {places.length > 0 ? (
        <ul className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {places.map((situation) => {
            const description =
              getSituationOutcome(situation.slug) ?? situation.shortDescription;
            const descriptionId = `place-${situation.slug}`;
            return (
              <li
                key={situation.slug}
                className="group relative flex min-h-24 min-w-0 items-start gap-3 rounded-2xl border border-neutral-200 bg-white p-4 shadow-[var(--vn-shadow-card)] transition-[border-color,transform] duration-[var(--duration-fast)] hover:-translate-y-px hover:border-primary-300 has-[a:focus-visible]:outline has-[a:focus-visible]:outline-[3px] has-[a:focus-visible]:outline-offset-[3px] has-[a:focus-visible]:outline-primary-500"
              >
                <span
                  aria-hidden="true"
                  className="grid size-11 shrink-0 place-items-center rounded-xl bg-primary-50"
                >
                  <SituationIcon slug={situation.slug} />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
                    <Link
                      href={`/discover/${encodeURIComponent(situation.slug)}`}
                      aria-describedby={description ? descriptionId : undefined}
                      className="font-semibold text-neutral-900 after:absolute after:inset-0 after:rounded-2xl after:content-[''] group-hover:text-primary-800 focus-visible:outline-none"
                    >
                      {situation.title}
                    </Link>
                    {situation.levelBand ? (
                      <span className="rounded-full bg-secondary-50 px-2 py-0.5 text-xs font-semibold text-secondary-800">
                        {formatLevelBand(situation.levelBand)}
                      </span>
                    ) : null}
                  </span>
                  {description ? (
                    <span
                      id={descriptionId}
                      className="mt-1 line-clamp-2 block text-sm text-neutral-600"
                    >
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
      )}
    </>
  );
}

function PlaceCardsPending() {
  return (
    <div aria-busy="true" className="mt-3">
      <p role="status" className="sr-only">
        Loading situations…
      </p>
      <ul
        aria-hidden="true"
        className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3"
      >
        {[0, 1, 2].map((index) => (
          <li
            key={index}
            className="min-h-24 rounded-2xl border border-neutral-200 bg-white/60"
          />
        ))}
      </ul>
    </div>
  );
}

/** Journey's alternatives stay available whatever the map or situations do. */
function JourneyPlaces({ read }: { read: Promise<SituationsRead> }) {
  return (
    <section aria-labelledby="journey-places-heading" className="my-5">
      <div className="flex flex-wrap items-end justify-between gap-x-4 gap-y-1">
        <div>
          <h2
            id="journey-places-heading"
            className="text-xl font-bold tracking-tight text-neutral-900"
          >
            Places to use English
          </h2>
          <p className="mt-0.5 text-sm text-neutral-600">
            Learn the words that belong to a real situation.
          </p>
        </div>
        <Link
          href="/courses"
          className="inline-flex min-h-11 items-center font-semibold text-primary-700 hover:text-primary-800"
        >
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
          <Link key={href} href={href} className="vn-chip-link">
            {label}
          </Link>
        ))}
      </nav>
    </section>
  );
}

export default async function JourneyPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const situationsRead = readSituations();
  return (
    <KnowledgeMapView
      searchParams={searchParams}
      intro={<JourneyPlaces read={situationsRead} />}
    />
  );
}
