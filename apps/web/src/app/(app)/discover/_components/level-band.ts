export function formatLevelBand(levelBand: string): string {
  if (levelBand.toLowerCase() === "mixed") {
    return "Mixed levels";
  }

  return levelBand
    .split("_")
    .map((level) => level.toUpperCase())
    .join("–");
}
