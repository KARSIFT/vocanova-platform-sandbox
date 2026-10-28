import type {
  KnowledgeMap,
  KnowledgeMapDecision,
  KnowledgeMapItem,
} from "@vocanova/api-client";

export type MapWord = KnowledgeMapItem;
export type KnowledgeMapData = KnowledgeMap;

/** One meaning's state exactly as a decision response reported it. */
export interface ConfirmedWordState {
  selfReportedKnown: boolean;
  saved: boolean;
  due?: boolean;
  revision: string;
}

/**
 * Reads the decided meaning's current state from the server's response. The
 * requested choice is never used to infer it: replaying an earlier receipt
 * returns the meaning as it is now. Anything incomplete yields null, so the
 * caller re-reads the map instead of guessing.
 */
export function confirmedWordState(
  result: unknown,
  request: KnowledgeMapDecision,
): ConfirmedWordState | null {
  if (!result || typeof result !== "object") return null;
  const value = result as Record<string, unknown>;
  if (
    value.meaningId !== request.meaningId ||
    typeof value.revision !== "string" ||
    !/^[0-9a-f]{64}$/.test(value.revision) ||
    typeof value.selfReportedKnown !== "boolean" ||
    typeof value.saved !== "boolean"
  )
    return null;
  return {
    selfReportedKnown: value.selfReportedKnown,
    saved: value.saved,
    due: typeof value.due === "boolean" ? value.due : undefined,
    revision: value.revision,
  };
}

/** Whether the reported state is exactly what this choice produces. */
export function confirmsChoice(
  state: ConfirmedWordState,
  decision: KnowledgeMapDecision["decision"],
): boolean {
  return decision === "known"
    ? state.selfReportedKnown && !state.saved
    : state.saved && !state.selfReportedKnown;
}

/** Puts one meaning's reported state into a loaded range, moving its counts. */
export function applyWordState(
  data: KnowledgeMapData,
  meaningId: string,
  state: ConfirmedWordState,
): KnowledgeMapData {
  const before = data.currentBand.items.find(
    (item) => item.meaningId === meaningId,
  );
  if (!before) return data;
  const from = wordState(before);
  const to = state.selfReportedKnown
    ? "known"
    : state.saved
      ? "learning"
      : "unseen";
  const delta = (value: "known" | "learning" | "unseen") =>
    (to === value ? 1 : 0) - (from === value ? 1 : 0);
  return {
    ...data,
    bands: data.bands.map((band) =>
      band.number === data.currentBand.number
        ? {
            ...band,
            knownCount: band.knownCount + delta("known"),
            learningCount: band.learningCount + delta("learning"),
            unseenCount: band.unseenCount + delta("unseen"),
          }
        : band,
    ),
    currentBand: {
      ...data.currentBand,
      items: data.currentBand.items.map((item) =>
        item.meaningId === meaningId
          ? {
              ...item,
              selfReportedKnown: state.selfReportedKnown,
              saved: state.saved,
              due: state.due ?? item.due,
              revision: state.revision,
            }
          : item,
      ),
    },
  };
}

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
/** The word map lives on Journey; every range link, retry and return uses it. */
export function journeyRangeHref(band: number): string {
  return `/discover?band=${band}#range-words`;
}

export function validBand(value: string | string[] | undefined): number {
  const number = typeof value === "string" ? Number(value) : 1;
  return Number.isInteger(number) && number >= 1 && number <= 100 ? number : 1;
}
