import type { CompletionDay } from "@vocanova/api-client";

export function getCompletionDayView(day: CompletionDay) {
  if (day.status === "protected") {
    return {
      label: "Streak protected",
      className: "border-secondary-700 bg-secondary-50 text-secondary-900",
    };
  }
  if (day.status === "completed") {
    return {
      label: "Completed",
      className: "border-primary-700 bg-primary-50 text-primary-900",
    };
  }
  if (day.status === undefined && day.completed) {
    // Older APIs merge completion and streak protection into one boolean.
    return {
      label: "Completed or protected",
      className: "border-primary-700 bg-primary-50 text-primary-900",
    };
  }
  return {
    label: "Not complete",
    className: "border-neutral-400 bg-neutral-100 text-neutral-700",
  };
}
