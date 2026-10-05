import { unitGuides } from "../discover/[situation]/_components/unit-guide-content";

// Reviewed original guide goals describe the available practice, not a
// competence claim. Keep this presentation separate from session snapshots.
export function getSituationOutcome(slug: string): string | undefined {
  return unitGuides[slug]?.goal;
}
