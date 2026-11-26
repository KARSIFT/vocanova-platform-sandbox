import type { PronunciationState, SpeechPreferences } from "./speech";

interface DevicePlayer {
  prepare(): void;
  play(
    owner: symbol,
    text: string,
    slow: boolean,
    notify: (state: PronunciationState) => void,
  ): void;
  stop(owner: symbol): void;
  release(owner: symbol): void;
}

export interface PronunciationMedia {
  playbackRate: number;
  preservesPitch?: boolean;
  onended: (() => void) | null;
  onerror: (() => void) | null;
  play(): Promise<void>;
  pause(): void;
}

interface PronunciationAudioSource {
  audioBase64: string;
  audioUrl?: string;
  mimeType: string;
  meaningId?: string;
  meaningRevision?: string;
}

export interface NeuralSpeechEnvironment {
  request(
    text: string,
    signal: AbortSignal,
    meaningId?: string,
    meaningRevision?: string,
    preferCdn?: boolean,
  ): Promise<PronunciationAudioSource>;
  createMedia(response: PronunciationAudioSource): {
    media: PronunciationMedia;
    dispose(): void;
  };
  otherSpeechActive?(): boolean;
}

/** Only use status codes, never provider/server error text, in learner messages. */
function meaningFailureMessage(error: unknown): string {
  const status =
    error instanceof Error && "status" in error ? error.status : undefined;
  if (status === 401 || status === 403)
    return "Sign in again to play pronunciation.";
  if (status === 404)
    return "This meaning may have changed. Refresh the page to try pronunciation.";
  if (status === 429)
    return "Pronunciation is busy or its daily limit has been reached. Try again later.";
  return "Pronunciation for this meaning is unavailable. Try again later.";
}

/** Coordinates one owned request/clip/device utterance across every audio control. */
export function createPronunciationController(
  device: DevicePlayer,
  getEnvironment: () => NeuralSpeechEnvironment | null,
  getPreferences: () => SpeechPreferences,
  timeoutMs = 8000,
  meaningRequestTimeoutMs = 16000,
) {
  type Request = {
    owner: symbol;
    text: string;
    meaningId?: string;
    meaningRevision?: string;
    notify: (state: PronunciationState) => void;
    phase: "request" | "audio" | "device";
    blocked: boolean;
    fallback?: (error?: unknown) => void;
    abort: AbortController;
    timer?: ReturnType<typeof setTimeout>;
    clip?: { media: PronunciationMedia; dispose(): void };
  };
  let active: Request | null = null;

  function clearMedia(request: Request) {
    const clip = request.clip;
    request.clip = undefined;
    if (clip) {
      clip.media.onended = null;
      clip.media.onerror = null;
      try {
        clip.media.pause();
      } finally {
        clip.dispose();
      }
    }
  }

  function cancel(notify: boolean): boolean {
    const previous = active;
    if (!previous) return true;
    active = null;
    clearTimeout(previous.timer);
    previous.abort.abort();
    try {
      clearMedia(previous);
      device.release(previous.owner);
      if (notify)
        previous.notify({
          status: "idle",
          source:
            previous.phase === "audio"
              ? "neural"
              : previous.phase === "device"
                ? "device"
                : undefined,
        });
      return true;
    } catch {
      if (notify)
        previous.notify({
          status: "error",
          message: "Pronunciation could not stop. Please try again.",
        });
      return false;
    }
  }

  function mediaFailure(request: Request, error: unknown) {
    if (active !== request || request.phase === "device") return;
    clearTimeout(request.timer);
    if (error instanceof Error && error.name === "NotAllowedError") {
      // Keep just this owner's prepared, paused clip. An explicit retry can
      // invoke play() directly inside the next gesture, without another fetch
      // consuming that gesture. Switching words, stopping or unmounting still
      // disposes it immediately; there is no history or persistent audio cache.
      request.blocked = true;
      try {
        request.clip?.media.pause();
      } catch {
        /* Denied media is not playing. */
      }
      request.notify({
        status: "error",
        message: "Tap Listen to play pronunciation.",
      });
    } else request.fallback?.(error);
  }

  return {
    prepare() {
      device.prepare();
    },
    play(
      owner: symbol,
      text: string,
      slow: boolean,
      notify: (state: PronunciationState) => void,
      meaningId?: string,
      meaningRevision?: string,
    ) {
      if (!text.trim()) return;
      let environment: NeuralSpeechEnvironment | null = null;
      try {
        environment = getEnvironment();
      } catch {
        /* Device fallback remains available. */
      }
      const preferences = getPreferences();
      if ((!active || active.blocked) && environment?.otherSpeechActive?.()) {
        notify({
          status: "error",
          message:
            "Other speech is playing. Wait for it to finish, then try again.",
        });
        return;
      }
      if (
        active?.blocked &&
        active.owner === owner &&
        active.text === text &&
        active.meaningId === meaningId &&
        active.meaningRevision === meaningRevision &&
        active.clip &&
        environment &&
        (!preferences.voiceURI || meaningId)
      ) {
        const request = active;
        const clip = request.clip!;
        const isCurrentClip = () =>
          active === request &&
          request.phase === "audio" &&
          request.clip === clip;
        request.notify = notify;
        request.blocked = false;
        notify({ status: "starting" });
        request.timer = setTimeout(() => {
          if (isCurrentClip()) request.fallback?.();
        }, timeoutMs);
        try {
          clip.media.playbackRate = slow ? 0.7 : preferences.rate;
          // This invocation is intentionally synchronous in play(), before any
          // asynchronous work, so a browser can honor the fresh user gesture.
          const playing = clip.media.play();
          void playing
            .then(() => {
              if (isCurrentClip()) {
                clearTimeout(request.timer);
                notify({ status: "playing", source: "neural" });
              }
            })
            .catch((error) => {
              if (isCurrentClip()) mediaFailure(request, error);
            });
        } catch (error) {
          mediaFailure(request, error);
        }
        return;
      }
      if (!cancel(true)) return;
      const request: Request = {
        owner,
        text,
        meaningId,
        meaningRevision,
        notify,
        phase: "request",
        blocked: false,
        abort: new AbortController(),
      };
      active = request;
      function fallback(error?: unknown) {
        if (active !== request || request.phase === "device") return;
        clearTimeout(request.timer);
        request.phase = "device";
        request.abort.abort();
        try {
          clearMedia(request);
        } catch {
          /* Release URLs even on media errors. */
        }
        if (meaningId) {
          active = null;
          notify({
            status: "unavailable",
            message: meaningFailureMessage(error),
          });
          return;
        }
        device.play(owner, text, slow, (state) => {
          if (active !== request || request.phase !== "device") return;
          notify({ ...state, source: "device" });
          if (["idle", "unavailable", "error"].includes(state.status))
            active = null;
        });
      }
      // A device override is safe only for general audio. Meaning-specific
      // playback needs the server's confirmed context and must never guess.
      if (
        !environment ||
        (preferences.voiceURI && !meaningId) ||
        Array.from(text).length > 500
      ) {
        fallback();
        return;
      }
      const neural = environment;
      async function start(preferCdn: boolean, revision = meaningRevision) {
        const attempt = request.abort;
        const isCurrent = () =>
          active === request &&
          request.abort === attempt &&
          request.phase !== "device";
        request.fallback = fallback;
        request.notify({ status: "starting" });
        if (!isCurrent()) return;
        // Give API generation and playback startup separate time budgets.
        request.timer = setTimeout(
          () => {
            if (isCurrent()) request.fallback?.();
          },
          meaningId ? meaningRequestTimeoutMs : timeoutMs,
        );
        try {
          const response = await neural.request(
            text,
            attempt.signal,
            meaningId,
            revision,
            preferCdn,
          );
          if (!isCurrent()) return;
          if (meaningId && response.meaningId !== meaningId)
            throw new Error("Pronunciation meaning did not match");
          if (revision && response.meaningRevision !== revision)
            throw new Error("Pronunciation meaning revision did not match");

          if (response.audioUrl !== undefined) {
            if (!preferCdn) throw new Error("Unexpected pronunciation URL");
            request.fallback = () => {
              if (!isCurrent()) return;
              clearTimeout(request.timer);
              request.abort = new AbortController();
              request.phase = "request";
              request.blocked = false;
              attempt.abort();
              try {
                clearMedia(request);
              } catch {
                /* The failed clip is no longer owned by this request. */
              }
              void start(false, response.meaningRevision ?? revision);
            };
          }
          clearTimeout(request.timer);
          request.timer = setTimeout(() => {
            if (isCurrent()) request.fallback?.();
          }, timeoutMs);
          const clip = neural.createMedia(response);
          request.clip = clip;
          request.phase = "audio";
          clip.media.playbackRate = slow ? 0.7 : preferences.rate;
          if ("preservesPitch" in clip.media) clip.media.preservesPitch = true;
          clip.media.onended = () => {
            if (!isCurrent() || request.phase !== "audio" || request.blocked)
              return;
            active = null;
            try {
              clearMedia(request);
            } catch {
              /* Playback already ended. */
            }
            request.notify({ status: "idle", source: "neural" });
          };
          clip.media.onerror = () => {
            if (isCurrent() && !request.blocked) request.fallback?.();
          };
          await clip.media.play();
          if (isCurrent()) {
            clearTimeout(request.timer);
            request.notify({ status: "playing", source: "neural" });
          }
        } catch (error) {
          if (isCurrent()) mediaFailure(request, error);
        }
      }
      void start(Boolean(meaningId));
    },
    stop(owner: symbol) {
      if (active?.owner === owner) cancel(true);
    },
    release(owner: symbol) {
      if (active?.owner === owner) cancel(false);
    },
  };
}

const pronunciationCdnUrl =
  /^https:\/\/audio\.vocanova\.site\/meanings\/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})\/([a-f0-9]{64})\/[a-f0-9]{64}\/[a-f0-9]{64}\.mp3$/;

/** Play a confirmed CDN clip or turn bounded API bytes into a revocable URL. */
export function createPronunciationMedia(
  response: PronunciationAudioSource,
  browser: Pick<
    typeof globalThis,
    "atob" | "Audio" | "Blob" | "URL"
  > = globalThis,
) {
  if (response.audioUrl !== undefined) {
    const url = response.audioUrl;
    const match = typeof url === "string" && pronunciationCdnUrl.exec(url);
    if (
      response.mimeType !== "audio/mpeg" ||
      response.audioBase64 !== "" ||
      !match ||
      match[0] !== url ||
      match[1] !== response.meaningId ||
      match[2] !== response.meaningRevision
    ) {
      throw new Error("Unsupported pronunciation URL");
    }
    const media = new browser.Audio();
    // Anonymous CORS omits the application's domain-wide session cookies.
    media.crossOrigin = "anonymous";
    media.preload = "auto";
    media.src = url;
    return {
      media: media as unknown as PronunciationMedia,
      dispose: () => {
        media.removeAttribute("src");
        media.load();
      },
    };
  }
  if (
    response.mimeType !== "audio/mpeg" ||
    !response.audioBase64 ||
    response.audioBase64.length > 2_000_000
  ) {
    throw new Error("Unsupported pronunciation audio");
  }
  const decoded = browser.atob(response.audioBase64);
  const bytes = Uint8Array.from(decoded, (character) =>
    character.charCodeAt(0),
  );
  const url = browser.URL.createObjectURL(
    new browser.Blob([bytes], { type: "audio/mpeg" }),
  );
  try {
    const media = new browser.Audio(url);
    media.preload = "auto";
    return {
      media: media as unknown as PronunciationMedia,
      dispose: () => browser.URL.revokeObjectURL(url),
    };
  } catch (error) {
    browser.URL.revokeObjectURL(url);
    throw error;
  }
}
