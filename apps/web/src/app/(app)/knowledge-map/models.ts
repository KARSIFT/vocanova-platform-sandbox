import type {
  KnowledgeMap,
  KnowledgeMapDecision,
  KnowledgeMapItem,
} from "@vocanova/api-client";

export type MapWord = KnowledgeMapItem;
export type KnowledgeMapData = KnowledgeMap;

const WORD_STATES = {
  known: { code: "k", label: "Already know" },
  learning: { code: "l", label: "Learning" },
  learned: { code: "m", label: "Learned" },
  unseen: { code: "n", label: "New" },
} as const;

/** One meaning's state exactly as a decision response reported it. */
export interface ConfirmedWordState {
  selfReportedKnown: boolean;
  saved: boolean;
  learned?: boolean;
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
    ...(value.learned === true ? { learned: true } : {}),
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
  const to = wordState(state);
  const counted = (status: keyof typeof WORD_STATES) =>
    status === "learned" ? "learning" : status;
  const delta = (status: ReturnType<typeof counted>) =>
    Number(counted(to) === status) - Number(counted(from) === status);
  return {
    ...data,
    bands: data.bands.map((band) => {
      if (band.number !== data.currentBand.number || from === to) return band;
      const unseenCount = band.unseenCount + delta("unseen");
      const codes =
        band.wordStates ??
        (band.unseenCount === band.totalWords
          ? "n".repeat(band.totalWords)
          : undefined);
      const index = before.rank - band.startRank;
      const wordStates =
        codes && unseenCount < band.totalWords
          ? codes.slice(0, index) +
            WORD_STATES[to].code +
            codes.slice(index + 1)
          : undefined;
      return {
        ...band,
        knownCount: band.knownCount + delta("known"),
        learningCount: band.learningCount + delta("learning"),
        unseenCount,
        wordStates,
      };
    }),
    currentBand: {
      ...data.currentBand,
      items: data.currentBand.items.map((item) =>
        item.meaningId === meaningId
          ? {
              ...item,
              selfReportedKnown: state.selfReportedKnown,
              saved: state.saved,
              learned: state.learned,
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
  const word = discoveryWordInRange(data);
  if (word) return { band: data.currentBand.number, rank: word.rank };
  const band = data.bands.find(
    (item) => item.number > data.currentBand.number && item.unseenCount > 0,
  );
  return band ? { band: band.number } : null;
}

export function earlierDiscoveryBand(data: KnowledgeMapData) {
  return data.bands.findLast(
    (item) => item.number < data.currentBand.number && item.unseenCount > 0,
  );
}

export function isUnseen(word: MapWord) {
  return !word.selfReportedKnown && !word.saved;
}

export function discoveryWordInRange(
  data: KnowledgeMapData,
  requestedRank = 0,
) {
  const items = data.currentBand.items;
  return (
    items.find((word) => word.rank >= requestedRank && isUnseen(word)) ??
    items.find(isUnseen)
  );
}

export function wordState(
  word: Pick<MapWord, "selfReportedKnown" | "saved" | "learned">,
): keyof typeof WORD_STATES {
  if (word.selfReportedKnown) return "known";
  if (word.saved) return word.learned ? "learned" : "learning";
  return "unseen";
}
export function wordStateLabel(word: MapWord): string {
  return WORD_STATES[wordState(word)].label;
}
/** The word map lives on Journey; every range link, retry and return uses it. */
export function journeyRangeHref(band: number): string {
  return `/discover?band=${band}#range-words`;
}

export function validBand(value: string | string[] | undefined): number {
  const number = typeof value === "string" ? Number(value) : 1;
  return Number.isInteger(number) && number >= 1 && number <= 100 ? number : 1;
}
