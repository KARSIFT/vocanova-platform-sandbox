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
    description:
      "Read a meaning and recall the word or phrase you learned, without answer choices.",
    startLabel: "Start typed recall",
  },
  listening_choice: {
    title: "Listen for meaning",
    description:
      "Hear a word with device pronunciation, then choose its meaning. No microphone is needed.",
    startLabel: "Start listening practice",
  },
  mistakes: {
    title: "Revisit mistakes",
    description:
      "Give words you missed another try, with focused questions from supported lessons, reviews and practice.",
    startLabel: "Practise past mistakes",
  },
};
