import {
  createPronunciationController,
  createPronunciationMedia,
} from "./neural-speech";

export interface PronunciationState {
  status: "idle" | "starting" | "playing" | "unavailable" | "error";
  message?: string;
  source?: "neural" | "device";
}

export function pronunciationCaption(state: PronunciationState) {
  return state.source === "neural"
    ? "Natural pronunciation"
    : state.source === "device"
      ? "Device pronunciation"
      : "English pronunciation";
}

export interface SpeechEnvironment {
  synthesis: Pick<
    SpeechSynthesis,
    "getVoices" | "speak" | "cancel" | "speaking" | "pending"
  >;
  createUtterance: (text: string) => SpeechSynthesisUtterance;
}

export const DEFAULT_SPEECH_PREFERENCES = { rate: 0.85, voiceURI: "" };
export const SPEECH_PREFERENCES_KEY = "vocanova.speech.v1";
const preferencesChanged = "vocanova-speech-preferences";
let temporaryPreferences = { ...DEFAULT_SPEECH_PREFERENCES };
let persistenceUnavailable = false;

export interface SpeechPreferences {
  rate: number;
  voiceURI: string;
}

export function readSpeechPreferences(
  storage: Pick<Storage, "getItem"> | null = typeof window === "undefined" ||
  persistenceUnavailable
    ? null
    : safeSpeechStorage(),
): SpeechPreferences {
  try {
    if (!storage) return { ...temporaryPreferences };
    const stored = JSON.parse(
      storage?.getItem(SPEECH_PREFERENCES_KEY) ?? "null",
    );
    return {
      rate: [0.7, 0.85, 1].includes(stored?.rate) ? stored.rate : 0.85,
      voiceURI:
        typeof stored?.voiceURI === "string" && stored.voiceURI.length <= 512
          ? stored.voiceURI
          : "",
    };
  } catch {
    return { ...DEFAULT_SPEECH_PREFERENCES };
  }
}

function safeSpeechStorage(): Storage | null {
  try {
    return window.localStorage;
  } catch {
    return null;
  }
}

export function saveSpeechPreferences(patch: Partial<SpeechPreferences>) {
  const next = { ...readSpeechPreferences(), ...patch };
  const validated = readSpeechPreferences({
    getItem: () => JSON.stringify(next),
  });
  temporaryPreferences = validated;
  try {
    const storage = safeSpeechStorage();
    if (storage)
      storage.setItem(SPEECH_PREFERENCES_KEY, JSON.stringify(validated));
    persistenceUnavailable = !storage;
  } catch {
    persistenceUnavailable = true;
    // Playback remains available when private browsing disables persistence.
  }
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(preferencesChanged));
  }
  return validated;
}

export function subscribeSpeechPreferences(refresh: () => void) {
  if (typeof window === "undefined") return () => {};
  const onStorage = (event: StorageEvent) => {
    if (!event.key || event.key === SPEECH_PREFERENCES_KEY) refresh();
  };
  window.addEventListener(preferencesChanged, refresh);
  window.addEventListener("storage", onStorage);
  return () => {
    window.removeEventListener(preferencesChanged, refresh);
    window.removeEventListener("storage", onStorage);
  };
}

function voiceQuality(voice: SpeechSynthesisVoice): number {
  // Browser names are hints, not an acoustic quality guarantee. Local service
  // describes delivery only; it must not outrank an available enhanced voice.
  if (/natural|neural|premium|enhanced/i.test(voice.name ?? "")) return 4;
  if (/google/i.test(voice.name ?? "")) return 3;
  if (voice.default) return 2;
  return 1;
}

/** Prefer enhanced English voices when supplied, and honor the learner's choice. */
export function selectEnglishVoice(
  voices: readonly SpeechSynthesisVoice[],
  preferredURI = "",
): SpeechSynthesisVoice | undefined {
  const english = voices.filter((voice) => /^en(?:[-_]|$)/i.test(voice.lang));
  return (
    (preferredURI
      ? english.find((voice) => voice.voiceURI === preferredURI)
      : undefined) ??
    english.slice().sort((a, b) => voiceQuality(b) - voiceQuality(a))[0]
  );
}

export function availableEnglishVoices(): SpeechSynthesisVoice[] {
  try {
    return typeof window === "undefined"
      ? []
      : (window.speechSynthesis?.getVoices() ?? []).filter((voice) =>
          /^en(?:[-_]|$)/i.test(voice.lang),
        );
  } catch {
    return [];
  }
}

export function createDeviceSpeechController(
  getEnvironment: () => SpeechEnvironment | null,
  getPreferences: () => SpeechPreferences = readSpeechPreferences,
) {
  let active: {
    owner: symbol;
    environment: SpeechEnvironment;
    utterance: SpeechSynthesisUtterance;
    notify: (state: PronunciationState) => void;
  } | null = null;

  function cancelActive(notify = true): boolean {
    const previous = active;
    if (!previous) return true;
    active = null;
    // cancel() is global to the speech queue. Only invoke it for speech this
    // controller owns, never when an unrelated control unmounts.
    try {
      previous.environment.synthesis.cancel();
      if (notify) previous.notify({ status: "idle" });
      return true;
    } catch {
      if (notify) {
        previous.notify({
          status: "error",
          message: "Pronunciation could not stop. Please try again.",
        });
      }
      return false;
    }
  }

  return {
    prepare() {
      try {
        // Some browsers populate voices only after the first request.
        getEnvironment()?.synthesis.getVoices();
      } catch {
        // Availability is checked again when the learner presses Listen.
      }
    },
    play(
      owner: symbol,
      text: string,
      slow: boolean,
      notify: (state: PronunciationState) => void,
    ) {
      try {
        const environment = getEnvironment();
        if (!environment) {
          notify({
            status: "unavailable",
            message: "Pronunciation is not available in this browser.",
          });
          return;
        }
        const preferences = getPreferences();
        const voices = environment.synthesis.getVoices();
        const voice = selectEnglishVoice(voices, preferences.voiceURI);
        if (!voice) {
          notify({
            status: "unavailable",
            message:
              "An English voice is not ready on this device. Please try again.",
          });
          return;
        }
        if (
          !active &&
          (environment.synthesis.speaking || environment.synthesis.pending)
        ) {
          notify({
            status: "error",
            message:
              "Other speech is playing. Wait for it to finish, then try again.",
          });
          return;
        }
        if (!text.trim()) return;
        if (!cancelActive()) {
          notify({
            status: "error",
            message: "Pronunciation could not start. Please try again.",
          });
          return;
        }
        function speak(
          selectedVoice: SpeechSynthesisVoice,
          allowFallback: boolean,
        ) {
          const utterance = environment!.createUtterance(text);
          utterance.voice = selectedVoice;
          utterance.lang = selectedVoice.lang;
          utterance.rate = slow ? 0.7 : preferences.rate;
          const request = {
            owner,
            environment: environment!,
            utterance,
            notify,
          };
          active = request;
          utterance.onstart = () => {
            if (active === request) notify({ status: "playing" });
          };
          utterance.onend = () => {
            if (active !== request) return;
            active = null;
            notify({ status: "idle" });
          };
          utterance.onerror = (event) => {
            if (active !== request) return;
            active = null;
            // A remote voice may disappear or fail offline. Retry at most once
            // with an available local English voice; never retry denied autoplay,
            // interrupted/cancelled speech, or unrelated native failures.
            const fallback =
              allowFallback &&
              !selectedVoice.localService &&
              [
                "network",
                "voice-unavailable",
                "synthesis-unavailable",
              ].includes(event?.error)
                ? voices.find(
                    (candidate) =>
                      candidate.localService &&
                      /^en(?:[-_]|$)/i.test(candidate.lang),
                  )
                : undefined;
            if (fallback) {
              try {
                environment!.synthesis.cancel();
                speak(fallback, false);
                return;
              } catch {
                // The ordinary recoverable error below remains visible.
              }
            }
            notify({
              status: "error",
              message: "Pronunciation could not play. Please try again.",
            });
          };
          notify({ status: "starting" });
          try {
            environment!.synthesis.speak(utterance);
          } catch {
            if (active === request) active = null;
            notify({
              status: "error",
              message: "Pronunciation could not play. Please try again.",
            });
          }
        }
        speak(voice, true);
      } catch {
        notify({
          status: "unavailable",
          message:
            "Pronunciation is not available right now. Please try again.",
        });
      }
    },
    stop(owner: symbol) {
      if (active?.owner === owner) cancelActive();
    },
    release(owner: symbol) {
      if (active?.owner === owner) cancelActive(false);
    },
  };
}

const localSpeech = createDeviceSpeechController(() => {
  if (
    typeof window === "undefined" ||
    !window.speechSynthesis ||
    !window.SpeechSynthesisUtterance
  ) {
    return null;
  }
  return {
    synthesis: window.speechSynthesis,
    createUtterance: (text) => new window.SpeechSynthesisUtterance(text),
  };
});

// Keep one shared interface across discovery, review, lessons and examples.
// These controls accept public teaching text, never a learner's draft.
export const deviceSpeech = createPronunciationController(
  localSpeech,
  () => {
    if (
      typeof window === "undefined" ||
      !window.Audio ||
      !window.URL?.createObjectURL ||
      !window.atob
    )
      return null;
    return {
      async request(text: string, signal: AbortSignal) {
        const [{ createApiClient }, { getOrRefreshCSRFToken }] =
          await Promise.all([import("./api"), import("./csrf")]);
        if (signal.aborted) throw new DOMException("Cancelled", "AbortError");
        const token = await getOrRefreshCSRFToken();
        if (signal.aborted) throw new DOMException("Cancelled", "AbortError");
        if (!token) throw new Error("Pronunciation session is unavailable");
        const { data } = await createApiClient().getPronunciation(text, {
          signal,
          headers: { "X-CSRF-Token": token },
        });
        return data;
      },
      createMedia: createPronunciationMedia,
      otherSpeechActive: () =>
        Boolean(
          window.speechSynthesis?.speaking || window.speechSynthesis?.pending,
        ),
    };
  },
  readSpeechPreferences,
);
