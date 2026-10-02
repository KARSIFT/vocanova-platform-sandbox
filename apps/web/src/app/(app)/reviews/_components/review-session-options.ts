import type { DueWord } from "@vocanova/api-client";

interface ReviewOption {
  meaningId: string;
  label: string;
}

export function buildMultipleChoiceOptions(
  dueWords: readonly DueWord[],
  currentIndex: number,
  sessionSeed: string,
): ReviewOption[] {
  const current = dueWords[currentIndex];
  if (!current) {
    return [];
  }
  const distractors = dueWords
    .filter((_, index) => index !== currentIndex)
    .slice(0, 3)
    .map((dueWord) => ({
      meaningId: dueWord.meaningId,
      label: `${dueWord.partOfSpeech} — ${dueWord.shortDefinition}`,
    }));
  const all = [
    {
      meaningId: current.meaningId,
      label: `${current.partOfSpeech} — ${current.shortDefinition}`,
    },
    ...distractors,
  ];

  // Initial options render on both the server and client. Seed the shuffle
  // from the serialized session seed and card identity so hydration agrees
  // while positions vary between sessions. Keep scheduling metadata out so
  // queue refreshes within a session stay stable.
  const identity = JSON.stringify([
    sessionSeed,
    current.userWordId,
    current.meaningId,
  ]);
  let seed = 2166136261;
  for (let index = 0; index < identity.length; index++) {
    seed = Math.imul(seed ^ identity.charCodeAt(index), 16777619);
  }
  for (let index = all.length - 1; index > 0; index--) {
    // Mulberry32: deterministic 32-bit mixing for presentation, not security.
    seed = (seed + 0x6d2b79f5) | 0;
    let mixed = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    mixed ^= mixed + Math.imul(mixed ^ (mixed >>> 7), 61 | mixed);
    const fraction = ((mixed ^ (mixed >>> 14)) >>> 0) / 4294967296;
    const swapIndex = Math.floor(fraction * (index + 1));
    [all[index], all[swapIndex]] = [all[swapIndex]!, all[index]!];
  }
  return all;
}
