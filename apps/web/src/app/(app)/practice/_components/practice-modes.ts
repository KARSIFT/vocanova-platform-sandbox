import type { PracticeMode } from "@vocanova/api-client";

export const practiceModes: Record<
  PracticeMode,
  {
    title: string;
    description: string;
    startLabel: string;
  }
> = {
  typed_recall: {
    title: "Type the word",
    description: "See a meaning, then type the word you remember.",
    startLabel: "Start typed recall",
  },
  listening_choice: {
    title: "Listen for meaning",
    description: "Hear a word, then choose its meaning. No microphone needed.",
    startLabel: "Start listening practice",
  },
  mistakes: {
    title: "Revisit mistakes",
    description: "Try again with words you missed in earlier practice.",
    startLabel: "Practise past mistakes",
  },
};
