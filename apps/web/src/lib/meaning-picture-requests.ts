import type { MeaningImage } from "@vocanova/api-client";

// Only teaching steps whose frozen canonical revision is known use this fallback.
// Discovery and word detail carry approved image metadata in their existing read.
const pending = new Map<string, ((image?: MeaningImage) => void)[]>();
const cache = new Map<
  string,
  { expires: number; result: Promise<MeaningImage | undefined> }
>();
const MAX_CACHE = 512;
let scheduled = false;

export function requestMeaningPicture(
  meaningId: string,
): Promise<MeaningImage | undefined> {
  const previous = cache.get(meaningId);
  if (previous && previous.expires > Date.now()) return previous.result;
  const result = new Promise<MeaningImage | undefined>((resolve) => {
    pending.set(meaningId, [...(pending.get(meaningId) ?? []), resolve]);
  });
  cache.set(meaningId, { expires: Date.now() + 60_000, result });
  if (cache.size > MAX_CACHE) cache.delete(cache.keys().next().value!);
  if (!scheduled) {
    scheduled = true;
    queueMicrotask(() => void flush());
  }
  return result;
}

async function flush() {
  scheduled = false;
  const entries = [...pending.entries()];
  pending.clear();
  try {
    const { createApiClient } = await import("./api");
    const client = createApiClient();
    await Promise.all(
      Array.from(
        { length: Math.ceil(entries.length / 100) },
        async (_, batch) => {
          const group = entries.slice(batch * 100, (batch + 1) * 100);
          try {
            const { data } = await client.getMeaningImages(
              group.map(([id]) => id),
              { signal: AbortSignal.timeout(3000) },
            );
            for (const [id, listeners] of group) {
              const image = data.items.find((item) => item.meaningId === id);
              for (const resolve of listeners) resolve(image);
            }
          } catch {
            for (const [, listeners] of group)
              for (const resolve of listeners) resolve();
          }
        },
      ),
    );
  } catch {
    for (const [, listeners] of entries)
      for (const resolve of listeners) resolve();
  }
}
