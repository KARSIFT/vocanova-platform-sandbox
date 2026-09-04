/**
 * Pure view-selection logic for the Discover (Journey) list screen (VOC-1179).
 *
 * Kept separate from `page.tsx` so the "which state renders" decision is
 * unit-testable without a DOM/React renderer.
 */
export type DiscoverListView = "empty" | "list";

/**
 * Decide whether the Discover list screen should show the "no situations
 * available yet" empty state or the grid of journey situations.
 */
export function getDiscoverListView(situationCount: number): DiscoverListView {
  return situationCount > 0 ? "list" : "empty";
}
