import type { KnowledgeMapData } from "./models";
type Band = KnowledgeMapData["bands"][number];

export function learnedCount(band: Band) {
  if (band.learningCount === 0) return 0;
  // Older responses counted all saved words together.
  if (!band.wordStates) return undefined;
  return band.wordStates.match(/m/g)?.length ?? 0;
}

export function rangeProgressLabel(
  band: Band,
  separator: string,
  knownLabel: string,
) {
  const learned = learnedCount(band);
  const counts = [`${band.knownCount} ${knownLabel}`];
  if (learned === undefined) {
    counts.push(`${band.learningCount} saved`);
  } else {
    counts.push(`${band.learningCount - learned} learning`);
    if (learned > 0) counts.push(`${learned} learned`);
  }
  counts.push(`${band.unseenCount} new`);
  return counts.join(separator);
}

export function rangePreviewPaths(band: Band) {
  if (!band.wordStates && band.unseenCount < band.totalWords) return null;
  const paths = { k: "", l: "", m: "" };
  for (const [index, code] of Array.from(band.wordStates ?? "").entries()) {
    if (code !== "k" && code !== "l" && code !== "m") continue;
    const x = (index % 10) * 10 + 1;
    const y = Math.floor(index / 10) * 10 + 1;
    paths[code] += `M${x} ${y}h8v8h-8z`;
  }
  return paths;
}
