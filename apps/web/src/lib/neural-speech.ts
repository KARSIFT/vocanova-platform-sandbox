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

export interface NeuralSpeechEnvironment {
  request(
    text: string,
    signal: AbortSignal,
  ): Promise<{ audioBase64: string; mimeType: string }>;
  createMedia(response: { audioBase64: string; mimeType: string }): {
    media: PronunciationMedia;
    dispose(): void;
  };
  otherSpeechActive?(): boolean;
}

/** Coordinates one owned request/clip/device utterance across every audio control. */
export function createPronunciationController(
  device: DevicePlayer,
  getEnvironment: () => NeuralSpeechEnvironment | null,
  getPreferences: () => SpeechPreferences,
  timeoutMs = 8000,
) {
  type Request = {
    owner: symbol;
    text: string;
    notify: (state: PronunciationState) => void;
    phase: "request" | "audio" | "device";
    blocked: boolean;
    fallback?: () => void;
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
    } else request.fallback?.();
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
        active.clip &&
        environment &&
        !preferences.voiceURI
      ) {
        const request = active;
        const clip = request.clip!;
        request.notify = notify;
        request.blocked = false;
        notify({ status: "starting" });
        request.timer = setTimeout(() => request.fallback?.(), timeoutMs);
        try {
          clip.media.playbackRate = slow ? 0.7 : preferences.rate;
          // This invocation is intentionally synchronous in play(), before any
          // asynchronous work, so a browser can honor the fresh user gesture.
          const playing = clip.media.play();
          void playing
            .then(() => {
              clearTimeout(request.timer);
              if (active === request && request.phase === "audio")
                notify({ status: "playing", source: "neural" });
            })
            .catch((error) => mediaFailure(request, error));
        } catch (error) {
          mediaFailure(request, error);
        }
        return;
      }
      if (!cancel(true)) return;
      const request: Request = {
        owner,
        text,
        notify,
        phase: "request",
        blocked: false,
        abort: new AbortController(),
      };
      active = request;
      function fallback() {
        if (active !== request || request.phase === "device") return;
        clearTimeout(request.timer);
        request.phase = "device";
        request.abort.abort();
        try {
          clearMedia(request);
        } catch {
          /* Release URLs even on media errors. */
        }
        device.play(owner, text, slow, (state) => {
          if (active !== request || request.phase !== "device") return;
          notify({ ...state, source: "device" });
          if (["idle", "unavailable", "error"].includes(state.status))
            active = null;
        });
      }
      request.fallback = fallback;
      // A selected device voice is an explicit override. Long stories retain
      // device playback; the server accepts bounded public teaching text only.
      if (
        !environment ||
        preferences.voiceURI ||
        Array.from(text).length > 500
      ) {
        fallback();
        return;
      }
      const neural = environment;
      notify({ status: "starting" });
      request.timer = setTimeout(fallback, timeoutMs);
      void (async () => {
        try {
          const response = await neural.request(text, request.abort.signal);
          if (active !== request || request.phase !== "request") return;
          const clip = neural.createMedia(response);
          request.clip = clip;
          request.phase = "audio";
          clip.media.playbackRate = slow ? 0.7 : preferences.rate;
          if ("preservesPitch" in clip.media) clip.media.preservesPitch = true;
          clip.media.onended = () => {
            if (
              active !== request ||
              request.phase !== "audio" ||
              request.blocked
            )
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
            if (!request.blocked) fallback();
          };
          await clip.media.play();
          clearTimeout(request.timer);
          if (active === request && request.phase === "audio")
            notify({ status: "playing", source: "neural" });
        } catch (error) {
          if (active !== request || request.phase === "device") return;
          mediaFailure(request, error);
        }
      })();
    },
    stop(owner: symbol) {
      if (active?.owner === owner) cancel(true);
    },
    release(owner: symbol) {
      if (active?.owner === owner) cancel(false);
    },
  };
}

/** Convert a validated bounded MP3 response into a revocable browser-only URL. */
export function createPronunciationMedia(
  response: { audioBase64: string; mimeType: string },
  browser: Pick<
    typeof globalThis,
    "atob" | "Audio" | "Blob" | "URL"
  > = globalThis,
) {
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
