// This is display-only state. The API, not the browser, decides whether a
// saved word is due (DOC-05 §9).
export function formatWordReviewState(
  reviewState: string | undefined,
  due: boolean | undefined,
): string | null {
  if (due) {
    return "Due today";
  }

  switch (reviewState) {
    case "new":
      return "New";
    case "learning":
    case "reviewing":
      return "Learning";
    case "mastered":
      return "Mastered";
    default:
      return null;
  }
}
