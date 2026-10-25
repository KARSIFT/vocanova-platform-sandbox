"use client";

import Image from "next/image";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type RefObject,
} from "react";
import type { MeaningImage } from "@vocanova/api-client";
import {
  getExternalMeaningPicture,
  getReadyMeaningPicture,
  type ReadyMeaningPicture,
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
  );
  const local = getReadyMeaningPicture(
    meaningId,
    undefined,
    meaningRevision ?? image?.meaningRevision,
  );
  if (!external && !local) return null;

  return (
    <PictureFrame
      key={`${meaningId}:${meaningRevision ?? image?.meaningRevision ?? ""}`}
      external={external}
      local={local}
      priority={priority}
    />
  );
}

// This is a presentation threshold, not a request deadline or automatic retry.
const SLOW_PICTURE_MS = 8_000;

function PictureFrame({
  external,
  local,
  priority,
}: {
  external?: ReadyMeaningPicture;
  local?: ReadyMeaningPicture;
  priority: boolean;
}) {
  const [failedSources, setFailedSources] = useState<string[]>([]);
  const [retryNonce, setRetryNonce] = useState<string>();
  const figure = useRef<HTMLElement>(null);
  const retry = useRef<HTMLButtonElement>(null);
  const generation = useRef(0);
  const renderedGeneration = generation.current;
  const isCurrent = useCallback(
    () => generation.current === renderedGeneration,
    [renderedGeneration],
  );
  const retainRetryFocus = useCallback(() => {
    if (retry.current && document.activeElement === retry.current) {
      figure.current?.focus({ preventScroll: true });
    }
  }, []);
  const picture =
    external && !failedSources.includes(external.src)
      ? external
      : local && !failedSources.includes(local.src)
        ? local
        : undefined;
  const requestSource = picture
    ? retryNonce
      ? `${picture.src}${picture.src.includes("?") ? "&" : "?"}vocanova_retry=${retryNonce}`
      : picture.src
    : undefined;
  const retryPicture = () => {
    generation.current += 1;
    retainRetryFocus();
    // Only an explicit retry restores failed sources and requests fresh bytes.
    setFailedSources([]);
    setRetryNonce(crypto.randomUUID());
  };

  return (
    <figure
      ref={figure}
      data-testid="meaning-picture"
      aria-label="Meaning picture"
      tabIndex={-1}
      className="mt-5 overflow-hidden rounded-2xl border border-neutral-200 bg-neutral-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
    >
      <div className="relative mx-auto aspect-[3/2] w-full max-w-[27rem]">
        {picture && requestSource ? (
          <PictureSource
            key={requestSource}
            picture={picture}
            src={requestSource}
            priority={priority}
            isCurrent={isCurrent}
            retryRef={retry}
            retainRetryFocus={retainRetryFocus}
            onRetry={retryPicture}
            onError={() => {
              if (!isCurrent()) return;
              retainRetryFocus();
              generation.current += 1;
              setFailedSources((sources) => [...sources, picture.src]);
            }}
          />
        ) : (
          <PicturePlaceholder
            message="Picture unavailable."
            retryRef={retry}
            onRetry={retryPicture}
          />
        )}
      </div>
    </figure>
  );
}

/** Source changes reset loading state while the named focus destination stays mounted. */
function PictureSource({
  picture,
  src,
  priority,
  isCurrent,
  retryRef,
  retainRetryFocus,
  onRetry,
  onError,
}: {
  picture: ReadyMeaningPicture;
  src: string;
  priority: boolean;
  isCurrent: () => boolean;
  retryRef: RefObject<HTMLButtonElement | null>;
  retainRetryFocus: () => void;
  onRetry: () => void;
  onError: () => void;
}) {
  const [loaded, setLoaded] = useState(false);
  const [slow, setSlow] = useState(false);
  const image = useRef<HTMLImageElement>(null);
  const active = useRef(true);
  const finishLoading = useCallback(() => {
    if (!active.current || !isCurrent()) return;
    retainRetryFocus();
    setLoaded(true);
    setSlow(false);
  }, [isCurrent, retainRetryFocus]);

  useLayoutEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
      retainRetryFocus();
    };
  }, [retainRetryFocus]);

  useEffect(() => {
    const element = image.current;
    const container = element?.parentElement;
    if (loaded || !element || !container) return;
    let cancelled = false;
    let timer: number | undefined;
    const current = () => !cancelled && active.current && isCurrent();
    // A cached image may finish before React attaches its load listener.
    if (element.complete && element.naturalWidth > 0) {
      finishLoading();
      return;
    }
    const observer = new IntersectionObserver(([entry]) => {
      window.clearTimeout(timer);
      // Lazy images below the viewport have not had a chance to load yet.
      if (!entry?.isIntersecting || !current()) return;
      timer = window.setTimeout(() => {
        if (!current()) return;
        if (element.complete && element.naturalWidth > 0) {
          finishLoading();
        } else {
          setSlow(true);
        }
      }, SLOW_PICTURE_MS);
    });
    observer.observe(container);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
      observer.disconnect();
    };
  }, [loaded, isCurrent, finishLoading]);

  return (
    <>
      <Image
        ref={image}
        src={src}
        alt={picture.alt}
        width={picture.width ?? 960}
        height={picture.height ?? 640}
        unoptimized={picture.external}
        priority={priority}
        sizes="(max-width: 640px) 100vw, 432px"
        className={`h-full w-full object-contain ${loaded ? "" : "opacity-0"}`}
        onLoad={finishLoading}
        onError={() => {
          if (active.current && isCurrent()) onError();
        }}
      />
      {!loaded && (
        <PicturePlaceholder
          message={slow ? "Picture is taking longer." : "Loading picture…"}
          retryRef={retryRef}
          onRetry={slow ? onRetry : undefined}
        />
      )}
    </>
  );
}

function PicturePlaceholder({
  message,
  retryRef,
  onRetry,
}: {
  message: string;
  retryRef: RefObject<HTMLButtonElement | null>;
  onRetry?: () => void;
}) {
  return (
    <div className="absolute inset-0 flex flex-col items-center justify-center gap-3 p-3 text-center text-neutral-600">
      <svg
        aria-hidden="true"
        viewBox="0 0 32 24"
        fill="none"
        className="h-6 w-8 text-neutral-400"
      >
        <rect
          x="1"
          y="1"
          width="30"
          height="22"
          rx="4"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <circle
          cx="22"
          cy="8"
          r="2.5"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <path
          d="m3 20 8-9 6 7 4-4 8 6"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinejoin="round"
        />
      </svg>
      <p role="status" className="text-sm">
        {message}
      </p>
      {onRetry && (
        <button
          ref={retryRef}
          type="button"
          className="min-h-[44px] min-w-[44px] rounded-xl border border-neutral-300 bg-white px-4 text-sm font-semibold text-primary-700 hover:bg-neutral-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
          onClick={onRetry}
        >
          Retry picture
        </button>
      )}
    </div>
  );
}
