/** Keep a fuller explanation only when it adds text to the short definition. */
export function getAdditionalDefinition(
  shortDefinition: string,
  learnerDefinition?: string | null,
): string | null {
  const explanation = learnerDefinition?.trim();
  if (!explanation) return null;

  const normalize = (text: string) =>
    text.trim().replace(/\s+/g, " ").toLowerCase();
  return normalize(explanation) === normalize(shortDefinition)
    ? null
    : explanation;
}

export function formatNoteType(noteType: string): string {
  return noteType
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
