import type { KnowledgeMapData } from "../knowledge-map/models";

export interface SavedDiscoveryPosition {
  catalogVersion: string;
  band: number;
  rank: number;
  active: boolean;
  revision: number;
}

export function explicitDiscoverySelection(
  query: Record<string, string | string[] | undefined>,
) {
  const band = typeof query.band === "string" ? Number(query.band) : 0;
  const rank = typeof query.word === "string" ? Number(query.word) : 0;
  return (
    (Number.isInteger(band) && band >= 1 && band <= 100) ||
    (Number.isInteger(rank) && rank >= 1 && rank <= 10000)
  );
}

/** Resume exact words, including known/saved words, only against a fresh catalog. */
export function resumableDiscoveryPosition(
  position: SavedDiscoveryPosition | null,
  data: KnowledgeMapData,
) {
  if (
    !position ||
    position.catalogVersion !== data.catalogVersion ||
    !Number.isInteger(position.band) ||
    !Number.isInteger(position.rank) ||
    position.band !== data.currentBand.number ||
    position.band !== Math.floor((position.rank - 1) / data.bandSize) + 1 ||
    !data.currentBand.items.some((word) => word.rank === position.rank)
  )
    return null;
  return position;
}
