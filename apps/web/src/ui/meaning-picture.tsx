"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import type { MeaningImage } from "@vocanova/api-client";
import {
  getExternalMeaningPicture,
  getReadyMeaningPicture,
} from "@/lib/meaning-pictures";
import { requestMeaningPicture } from "@/lib/meaning-picture-requests";

/** Teaching-only visual context. Do not mount this inside graded questions. */
export function MeaningPicture({
  meaningId,
  image,
  meaningRevision,
  priority = false,
  resolveMissing = false,
}: {
  meaningId: string;
  image?: MeaningImage;
  meaningRevision?: string;
  priority?: boolean;
  resolveMissing?: boolean;
}) {
  const [failedSources, setFailedSources] = useState<string[]>([]);
  const [resolved, setResolved] = useState<MeaningImage>();
  useEffect(() => {
    // Embedded metadata avoids a second API read on discovery/detail screens.
    // A frozen lesson may only display an external image for its exact revision.
    if (!resolveMissing || image || !meaningRevision) return;
    let active = true;
    void requestMeaningPicture(meaningId).then((result) => {
      if (active) setResolved(result);
    });
    return () => {
      active = false;
    };
  }, [meaningId, meaningRevision, image, resolveMissing]);
  const external = getExternalMeaningPicture(
    meaningId,
    image ?? resolved,
    meaningRevision,
    failedSources,
  );
  const local = getReadyMeaningPicture(
    meaningId,
    undefined,
    meaningRevision ?? image?.meaningRevision,
  );
  const picture =
    external ??
    (local && !failedSources.includes(local.src) ? local : undefined);
  if (!picture) return null;

  return (
    <figure
      data-testid="meaning-picture"
      className="mt-5 overflow-hidden rounded-2xl border border-neutral-200 bg-neutral-50"
    >
      <Image
        src={picture.src}
        alt={picture.alt}
        width={picture.width ?? 960}
        height={picture.height ?? 640}
        unoptimized={picture.external}
        priority={priority}
        sizes="(max-width: 640px) 100vw, 432px"
        className="mx-auto aspect-[3/2] w-full max-w-[27rem] object-contain"
        onError={() => setFailedSources((sources) => [...sources, picture.src])}
      />
    </figure>
  );
}
