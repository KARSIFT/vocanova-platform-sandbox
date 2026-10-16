import Link from "next/link";
import { ApiResponseError } from "@vocanova/api-client";
import { createServerApiClient, requireAuthRedirect } from "@/lib/api-server";
import { PageContainer } from "@/ui/surface";
import { validBand, type KnowledgeMapData } from "../knowledge-map/models";
import { Discovery } from "./discovery";
import styles from "../knowledge-map/workspace.module.css";

export default async function DiscoveryPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const query = await searchParams;
  const band = validBand(query.band);
  const requestedRank = typeof query.word === "string" ? Number(query.word) : 0;
  let data: KnowledgeMapData | null = null;
  try {
    data = (await (await createServerApiClient()).getKnowledgeMap(band)).data;
  } catch (error) {
    if (error instanceof ApiResponseError && error.status === 401)
      requireAuthRedirect(
        error,
        `/discovery?band=${band}&word=${requestedRank}`,
      );
  }
  if (!data || !data.currentBand.items.length)
    return (
      <PageContainer>
        <h1 className={styles.title}>Discover words</h1>
        <p role="alert">Words could not load.</p>
        <a href={`/discovery?band=${band}`} className={styles.secondary}>
          Try again
        </a>
        <Link href="/knowledge-map" className={styles.back}>
          Knowledge map
        </Link>
      </PageContainer>
    );
  const initialRank =
    data.currentBand.items.find((word) => word.rank === requestedRank)?.rank ??
    data.currentBand.items.find(
      (word) => !word.selfReportedKnown && !word.saved,
    )?.rank ??
    data.currentBand.items[0]?.rank ??
    1;
  return (
    <Discovery
      key={`${data.catalogVersion}:${band}:${initialRank}`}
      initialData={data}
      initialRank={initialRank}
    />
  );
}
