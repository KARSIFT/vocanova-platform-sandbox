"use client";

import Image from "next/image";
import { useState } from "react";
import { getReadyMeaningPicture } from "@/lib/meaning-pictures";

/** Teaching-only visual context. Do not mount this inside graded questions. */
export function MeaningPicture({
  meaningId,
  priority = false,
}: {
  meaningId: string;
  priority?: boolean;
}) {
  const [failedSrc, setFailedSrc] = useState<string>();
  const picture = getReadyMeaningPicture(meaningId, failedSrc);
  if (!picture) return null;

  return (
    <figure
      data-testid="meaning-picture"
      className="mt-5 overflow-hidden rounded-2xl border border-neutral-200 bg-neutral-50"
    >
      <Image
        src={picture.src}
        alt={picture.alt}
        width={960}
        height={640}
        priority={priority}
        sizes="(max-width: 640px) 100vw, 432px"
        className="mx-auto aspect-[3/2] w-full max-w-[27rem] object-contain"
        onError={() => setFailedSrc(picture.src)}
      />
    </figure>
  );
}
