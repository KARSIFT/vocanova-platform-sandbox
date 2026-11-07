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
import { Discovery } from "./discovery";
import { DiscoveryFinished } from "./discovery-finished";
import styles from "../knowledge-map/workspace.module.css";

export default async function DiscoveryPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  const band = validBand(query.band);
  const active = query.active === "1";
  const revisit = query.revisit === "1";
  const requestedRank = typeof query.word === "string" ? Number(query.word) : 0;
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
  try {
    const client = await createServerApiClient();
    const current = (await client.getKnowledgeMap(band)).data;
    const explicitWord = revisit
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
        <p role="alert">
          {changedElsewhere
            ? "New words changed elsewhere. Try again to load your next word."
            : "Words could not load."}
        </p>
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
    />
  );
}
