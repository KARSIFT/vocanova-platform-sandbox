export function distinctFeedbackHeadline(
  headline: string | undefined,
  statusLabel: string | null,
): string | null {
  if (!headline?.trim()) return null;
  const normalizedHeadline = headline.trim().replace(/\s+/g, " ").toLowerCase();
  const normalizedStatus = statusLabel
    ?.trim()
    .replace(/\s+/g, " ")
    .toLowerCase();
  return normalizedHeadline && normalizedHeadline !== normalizedStatus
    ? headline
    : null;
}
