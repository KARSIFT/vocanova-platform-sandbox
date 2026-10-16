import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import {
  nextDiscoveryTarget,
  validBand,
  type KnowledgeMapData,
} from "../knowledge-map/models";
import { Discovery } from "./discovery";
import styles from "../knowledge-map/workspace.module.css";

export default async function DiscoveryPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  let band = validBand(query.band);
  const requestedRank = typeof query.word === "string" ? Number(query.word) : 0;
  const returnQuery = new URLSearchParams();
  if (typeof query.band === "string") returnQuery.set("band", query.band);
  if (typeof query.word === "string") returnQuery.set("word", query.word);
  const returnTo = `/discovery${returnQuery.size ? `?${returnQuery}` : ""}`;
  let data: KnowledgeMapData | null = null;
  try {
    const client = await createServerApiClient();
    const current = (await client.getKnowledgeMap(band)).data;
    const explicitWord = current.currentBand.items.find(
      (word) => word.rank === requestedRank,
    );
    // A deliberate word link can revisit an assessment. Guided discovery
    // instead continues to unseen words, using the existing range metadata.
    const target = explicitWord ? null : nextDiscoveryTarget(current);
    data =
      target && target.band !== band
        ? (await client.getKnowledgeMap(target.band)).data
        : current;
    band = data.currentBand.number;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(error, returnTo);
  }
  const initialRank =
    data?.currentBand.items.find((word) => word.rank === requestedRank)?.rank ??
    data?.currentBand.items.find(
      (word) => !word.selfReportedKnown && !word.saved,
    )?.rank;
  const changedElsewhere = Boolean(
    data &&
    initialRank === undefined &&
    data.bands.some((item) => item.unseenCount > 0),
  );
  if (!data || !data.currentBand.items.length || changedElsewhere)
    return (
      <PageContainer>
        <h1 className={styles.title}>Discover words</h1>
        <p role="alert">
          {changedElsewhere
            ? "New words changed elsewhere. Try again to load your next word."
            : "Words could not load."}
        </p>
        <a href={returnTo} className={styles.secondary}>
          Try again
        </a>
        <Link href="/knowledge-map" className={styles.back}>
          Knowledge map
        </Link>
      </PageContainer>
    );
  if (initialRank === undefined)
    return (
      <PageContainer>
        <h1 className={styles.title}>You’ve explored the whole map</h1>
        <p className={styles.definition}>
          Choose a word on the map to revisit it.
        </p>
        <Link
          href={`/knowledge-map?band=${band}#range-words`}
          className={styles.primary}
        >
          Knowledge map
        </Link>
      </PageContainer>
    );
  return (
    <Discovery
      key={`${data.catalogVersion}:${band}:${initialRank}`}
      initialData={data}
      initialRank={initialRank}
    />
  );
}
