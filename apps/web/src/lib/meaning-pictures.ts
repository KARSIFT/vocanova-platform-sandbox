import readyPictures from "./meaning-pictures.runtime.json" with { type: "json" };

export interface ReadyMeaningPicture {
  alt: string;
  src: string;
}

/** Browser-only lookup data: editorial scenes and definitions stay out of this graph. */
const pictures: Readonly<Record<string, { alt: string; version: string }>> =
  readyPictures;

/** Unknown, pending or failed media leaves the readable teaching content intact. */
export function getReadyMeaningPicture(
  meaningId: string,
  failedSrc?: string,
): ReadyMeaningPicture | undefined {
  if (!Object.hasOwn(pictures, meaningId)) return undefined;
  const picture = pictures[meaningId];
  if (!picture) return undefined;
  const src = `/images/meanings/${meaningId}.webp?v=${picture.version}`;
  return picture.alt && src !== failedSrc
    ? { alt: picture.alt, src }
    : undefined;
}
