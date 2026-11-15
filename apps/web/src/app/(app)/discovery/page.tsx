import { StatusMessage } from "@/ui/status-message";
import Link from "@/ui/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import {
  discoveryWordInRange,
  nextDiscoveryTarget,
  validBand,
  type KnowledgeMapData,
} from "../knowledge-map/models";
import {
  explicitDiscoverySelection,
  resumableDiscoveryPosition,
  type SavedDiscoveryPosition,
} from "./discovery-position";
import { Discovery } from "./discovery";
import { DiscoveryFinished } from "./discovery-finished";
import styles from "../knowledge-map/workspace.module.css";

export default async function DiscoveryPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  let band = validBand(query.band);
  let active = query.active === "1";
  const revisit = query.revisit === "1";
  let requestedRank = typeof query.word === "string" ? Number(query.word) : 0;
  if (
    !query.band &&
    Number.isInteger(requestedRank) &&
    requestedRank >= 1 &&
    requestedRank <= 10000
  )
    band = Math.floor((requestedRank - 1) / 100) + 1;
  const returnQuery = new URLSearchParams();
  if (typeof query.band === "string") returnQuery.set("band", query.band);
  if (typeof query.word === "string") returnQuery.set("word", query.word);
  if (active) returnQuery.set("active", "1");
  if (revisit) returnQuery.set("revisit", "1");
  const returnTo = `/discovery${returnQuery.size ? `?${returnQuery}` : ""}`;
  let data: KnowledgeMapData | null = null;
  let initialRank: number | undefined;
  let explicitlySelected = false;
  let changedElsewhere = false;
  let savedPosition: SavedDiscoveryPosition | null = null;
  let resumed = false;
  let positionRevision: number | null = null;
  try {
    const client = await createServerApiClient();
    try {
      savedPosition = (await client.getDiscoveryPosition()).data.position;
      positionRevision = savedPosition?.revision ?? 0;
    } catch (error) {
      if (error instanceof ApiResponseError && error.status === 401)
        throw error;
      // A navigation preference outage must not prevent discovery.
    }
    const explicit = explicitDiscoverySelection(query);
    const savedBand = savedPosition?.band;
    if (
      !explicit &&
      savedBand &&
      Number.isInteger(savedBand) &&
      savedBand >= 1 &&
      savedBand <= 100
    )
      band = savedBand;
    let current = (await client.getKnowledgeMap(band)).data;
    const resume = !explicit
      ? resumableDiscoveryPosition(savedPosition, current)
      : null;
    if (resume) {
      resumed = true;
      requestedRank = resume.rank;
      active = resume.active;
    } else if (!explicit && band !== 1) {
      band = 1;
      current = (await client.getKnowledgeMap(band)).data;
    }
    const explicitWord =
      revisit || Boolean(resume)
        ? current.currentBand.items.find((word) => word.rank === requestedRank)
        : undefined;
    const target = explicitWord ? null : nextDiscoveryTarget(current);
    data =
      target && target.band !== band
        ? (await client.getKnowledgeMap(target.band)).data
        : current;
    const word =
      explicitWord ??
      discoveryWordInRange(
        data,
        data.currentBand.number === band ? requestedRank : undefined,
      );
    initialRank = word?.rank;
    explicitlySelected = Boolean(explicitWord);
    changedElsewhere = Boolean(target && !word);
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, returnTo);
  }
  if (!data || !data.currentBand.items.length || changedElsewhere)
    return (
      <PageContainer>
        <h1 className={styles.title}>Discover words</h1>
        <StatusMessage as="p" tone="error" role="alert">
          {changedElsewhere
            ? "New words changed elsewhere. Try again to load your next word."
            : "Words could not load."}
        </StatusMessage>
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <a href={returnTo} className={styles.secondary}>
            Try again
          </a>
          <Link href="/discover" className={styles.back}>
            Open your word map
          </Link>
        </div>
      </PageContainer>
    );
  if (initialRank === undefined)
    return (
      <PageContainer>
        <DiscoveryFinished data={data} active={active} />
      </PageContainer>
    );
  return (
    <Discovery
      key={`${data.catalogVersion}:${returnTo}:${data.currentBand.number}:${initialRank}`}
      initialData={data}
      initialRank={initialRank}
      initiallyActive={active}
      initiallyRevisiting={explicitlySelected}
      initialPositionRevision={positionRevision}
      initiallyResumed={resumed}
    />
  );
}
