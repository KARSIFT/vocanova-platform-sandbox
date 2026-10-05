import readyPictureAlts from "./meaning-pictures.runtime.json" with { type: "json" };

export interface ReadyMeaningPicture {
  alt: string;
  src: string;
}

/** Browser-only lookup data: editorial scenes and definitions stay out of this graph. */
const pictureAlts: Readonly<Record<string, string>> = readyPictureAlts;

/** Unknown, pending or failed media leaves the readable teaching content intact. */
export function getReadyMeaningPicture(
  meaningId: string,
  failedSrc?: string,
): ReadyMeaningPicture | undefined {
  if (!Object.hasOwn(pictureAlts, meaningId)) return undefined;
  const alt = pictureAlts[meaningId];
  const src = `/images/meanings/${meaningId}.webp`;
  return alt && src !== failedSrc ? { alt, src } : undefined;
}
