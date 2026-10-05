import manifest from "./meaning-pictures.json" with { type: "json" };

export interface MeaningPictureContent {
  meaningId: string;
  word: string;
  definition: string;
  /** Editorial brief, never learner-facing copy. */
  scene: string;
  alt: string;
  src: string;
  /** Mark ready only after the local asset has been generated and reviewed. */
  status: "pending" | "ready";
}

/** Original scene briefs, aligned to the canonical seed by meaning UUID. */
export const MEANING_PICTURES: Readonly<Record<string, MeaningPictureContent>> =
  manifest as Record<string, MeaningPictureContent>;

export function getMeaningPicture(
  meaningId: string,
): MeaningPictureContent | undefined {
  return Object.hasOwn(MEANING_PICTURES, meaningId)
    ? MEANING_PICTURES[meaningId]
    : undefined;
}

/** Pending or failed media must never replace the readable teaching content. */
export function getReadyMeaningPicture(
  meaningId: string,
  failedSrc?: string,
): MeaningPictureContent | undefined {
  const picture = getMeaningPicture(meaningId);
  return picture?.status === "ready" && picture.src !== failedSrc
    ? picture
    : undefined;
}
