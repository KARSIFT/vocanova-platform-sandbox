import type {
  PracticeMode,
  PracticeSessionSummary,
} from "@vocanova/api-client";

export const practiceModes: Record<
  PracticeMode,
  {
    title: string;
    description: string;
    startLabel: string;
  }
> = {
  journey: {
    title: "Today’s practice",
    description: "Listen, complete sentences and use your words.",
    startLabel: "Start practice lesson",
  },
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
    description:
      "Try words you missed in earlier lessons, reviews or practice.",
    startLabel: "Practise past mistakes",
  },
};

export function readPracticeMode(value: unknown): PracticeMode | undefined {
  return typeof value === "string" && Object.hasOwn(practiceModes, value)
    ? (value as PracticeMode)
    : undefined;
}

export function practiceModeHref(
  mode: PracticeMode,
  source?: Pick<PracticeSessionSummary, "listId" | "lessonKey">,
): string {
  const query = new URLSearchParams({ mode });
  if (mode === "typed_recall" || mode === "listening_choice") {
    if (source?.listId) query.set("list", source.listId);
    else if (source?.lessonKey) query.set("lesson", source.lessonKey);
  }
  return `/practice?${query}#${mode === "journey" ? "practice-journey" : `practice-mode-${mode}`}`;
}
