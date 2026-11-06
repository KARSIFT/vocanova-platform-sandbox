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

export function formatSavedWordStatus(
  status: string,
  reviewState?: string,
): string | null {
  switch (reviewState ?? status) {
    case "new":
      return "New";
    case "learning":
      return "Learning";
    case "reviewing":
      return "Reviewing";
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
