/** Only the positively identified app-bundle failure needs a fresh document. */
export function isSameAppChunkLoadError(
  error: unknown,
  origin: string,
): boolean {
  if (!(error instanceof Error) || error.name !== "ChunkLoadError")
    return false;
  // Turbopack's emitted error includes the failed URL before its module context.
  const match = /^Failed to load chunk (\S+)(?: from module(?:[\s\S]*))?$/.exec(
    error.message,
  );
  if (!match) return false;
  try {
    const url = new URL(match[1]!, origin);
    return (
      url.origin === origin &&
      /^\/_next\/static\/chunks\/[^/]+\.js$/.test(url.pathname)
    );
  } catch {
    return false;
  }
}
