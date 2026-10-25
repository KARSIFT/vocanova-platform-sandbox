import type { MeaningImage } from "@vocanova/api-client";
import readyPictures from "./meaning-pictures.runtime.json" with { type: "json" };

export interface ReadyMeaningPicture {
  alt: string;
  src: string;
  external?: boolean;
  width?: number;
  height?: number;
}

/** Browser-only lookup data: editorial scenes and definitions stay out of this graph. */
const pictures: Readonly<
  Record<string, { alt: string; version: string; meaningRevision: string }>
> = readyPictures;

/** Unknown, pending or failed media leaves the readable teaching content intact. */
export function getReadyMeaningPicture(
  meaningId: string,
  failedSrc?: string,
  expectedRevision?: string,
): ReadyMeaningPicture | undefined {
  if (!Object.hasOwn(pictures, meaningId)) return undefined;
  const picture = pictures[meaningId];
  // Legacy views without a canonical revision retain the reviewed local fallback.
  // When a trusted current/frozen revision exists, mismatched old artwork is hidden.
  if (
    !picture ||
    (expectedRevision && picture.meaningRevision !== expectedRevision)
  )
    return undefined;
  const src = `/images/meanings/${meaningId}.webp?v=${picture.version}`;
  return picture.alt && src !== failedSrc
    ? { alt: picture.alt, src }
    : undefined;
}

/** Only the API may approve artwork. Never accept arbitrary provider or catalog URLs. */
export function getExternalMeaningPicture(
  meaningId: string,
  image?: MeaningImage,
  expectedRevision?: string,
  failedSources: readonly string[] = [],
): ReadyMeaningPicture | undefined {
  if (
    !image ||
    image.meaningId !== meaningId ||
    !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(
      meaningId,
    ) ||
    !/^[a-f0-9]{64}$/.test(image.meaningRevision) ||
    !/^[a-f0-9]{64}$/.test(image.sha256) ||
    (expectedRevision && image.meaningRevision !== expectedRevision) ||
    typeof image.alt !== "string" ||
    !image.alt.trim() ||
    image.alt.length > 500 ||
    !Number.isInteger(image.width) ||
    image.width < 1 ||
    image.width > 960 ||
    !Number.isInteger(image.height) ||
    image.height < 1 ||
    image.height > 640
  )
    return undefined;
  const approvedUrl = `https://images.vocanova.site/meanings/${meaningId}/${image.meaningRevision}/${image.sha256}.webp`;
  if (image.imageUrl !== approvedUrl || failedSources.includes(approvedUrl))
    return undefined;
  return {
    src: approvedUrl,
    alt: image.alt,
    external: true,
    width: image.width,
    height: image.height,
  };
}
