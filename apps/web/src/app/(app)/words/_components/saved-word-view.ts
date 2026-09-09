export type SavedWordsView = "empty" | "list" | "exhausted";

export function getSavedWordsView(
  itemCount: number,
  hasCursor: boolean,
): SavedWordsView {
  if (itemCount > 0) {
    return "list";
  }
  return hasCursor ? "exhausted" : "empty";
}

export function formatSavedWordStatus(status: string): string | null {
  switch (status) {
    case "new":
      return "New";
    case "learning":
    case "reviewing":
      return "Learning";
    case "mastered":
      return "Mastered";
    case "ignored":
      return "Ignored";
    case "archived":
      return "Archived";
    default:
      return null;
  }
}
