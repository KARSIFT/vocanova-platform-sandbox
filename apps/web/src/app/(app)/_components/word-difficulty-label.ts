import { formatLevelBand } from "../discover/_components/level-band";

export function wordDifficultyLabel(level?: string | null): string | null {
  const availableLevel = level?.trim();
  if (!availableLevel || availableLevel.toLowerCase() === "unknown")
    return null;
  return formatLevelBand(availableLevel);
}
