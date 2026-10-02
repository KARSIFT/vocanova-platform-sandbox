import type { LearningGoal, MainUseCase } from "@vocanova/api-client";

export const learningGoals: Record<LearningGoal, string> = {
  general: "Everyday English",
  work: "English for work",
  travel: "English for travel",
  study: "English for study",
  conversation: "More confident conversations",
  exam: "Preparing for an exam",
};

export const learningFocuses: Record<MainUseCase, string> = {
  daily_life: "Daily life",
  work: "Work",
  travel: "Travel",
  study: "Study",
  social: "Social situations",
};
