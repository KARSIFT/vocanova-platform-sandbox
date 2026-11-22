import { meaningAnchor } from "../_components/meaning-comparison";

export function discoveryDetailHref(
  wordSlug: string,
  meaningId: string,
  rank: number,
  { finished, writing }: { finished?: boolean; writing?: boolean } = {},
) {
  const anchor = writing
    ? `sentence-practice-${meaningId}`
    : meaningAnchor(meaningId);
  return `/vocabulary/${encodeURIComponent(wordSlug)}?fromDiscovery=${rank}${finished ? "&finished=1" : ""}#${encodeURIComponent(anchor)}`;
}

/** Rebuild an internal return destination from one ranked word. */
export function discoveryReturnHref(
  rank: string | string[] | undefined,
  finished: boolean,
) {
  if (typeof rank !== "string" || !/^[1-9]\d{0,4}$/.test(rank)) return null;
  const word = Number(rank);
  if (word > 10_000) return null;
  const band = Math.ceil(word / 100);
  return `/discovery?band=${band}&word=${word}&active=1${finished ? "" : "&revisit=1"}`;
}
