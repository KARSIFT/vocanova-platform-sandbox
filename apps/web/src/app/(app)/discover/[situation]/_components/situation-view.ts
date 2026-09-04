/**
 * Pure view-selection logic for a single Discover situation screen (VOC-1179).
 *
 * Kept separate from `page.tsx` so the "which state renders" decision is
 * unit-testable without a DOM/React renderer.
 */
export type SituationDetailView = "empty" | "list";

/**
 * Decide whether a situation's detail screen should show the "no words here
 * yet" empty state or the list of meanings for that situation.
 */
export function getSituationDetailView(
  meaningCount: number,
): SituationDetailView {
  return meaningCount > 0 ? "list" : "empty";
}
