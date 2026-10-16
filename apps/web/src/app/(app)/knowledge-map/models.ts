import type { KnowledgeMap, KnowledgeMapItem } from "@vocanova/api-client";

export type MapWord = KnowledgeMapItem;
export type KnowledgeMapData = KnowledgeMap;

export function nextDiscoveryTarget(
  data: KnowledgeMapData,
): { band: number; rank?: number } | null {
  const word = data.currentBand.items.find(
    (item) => !item.selfReportedKnown && !item.saved,
  );
  if (word) return { band: data.currentBand.number, rank: word.rank };
  const band = data.bands.find((item) => item.unseenCount > 0);
  return band ? { band: band.number } : null;
}

export function wordState(word: MapWord): "known" | "learning" | "unseen" {
  return word.selfReportedKnown ? "known" : word.saved ? "learning" : "unseen";
}
export function wordStateLabel(word: MapWord): string {
  return word.selfReportedKnown
    ? "Already know"
    : word.saved
      ? "Learning"
      : "New";
}
export function validBand(value: string | string[] | undefined): number {
  const number = typeof value === "string" ? Number(value) : 1;
  return Number.isInteger(number) && number >= 1 && number <= 100 ? number : 1;
}
