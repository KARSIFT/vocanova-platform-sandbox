export interface PronunciationState {
  status: "idle" | "starting" | "playing" | "unavailable" | "error";
  message?: string;
}

export interface SpeechEnvironment {
  synthesis: Pick<
    SpeechSynthesis,
    "getVoices" | "speak" | "cancel" | "speaking" | "pending"
  >;
  createUtterance: (text: string) => SpeechSynthesisUtterance;
}

/** Prefer an installed English voice, without assuming a particular accent. */
export function selectEnglishVoice(
  voices: readonly SpeechSynthesisVoice[],
): SpeechSynthesisVoice | undefined {
  const english = voices.filter((voice) => /^en(?:[-_]|$)/i.test(voice.lang));
  return (
    english.find((voice) => voice.localService && voice.default) ??
    english.find((voice) => voice.localService) ??
    english.find((voice) => voice.default) ??
    english[0]
  );
}

export function createDeviceSpeechController(
  getEnvironment: () => SpeechEnvironment | null,
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
        const voice = selectEnglishVoice(environment.synthesis.getVoices());
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
        const utterance = environment.createUtterance(text);
        utterance.voice = voice;
        utterance.lang = voice.lang;
        utterance.rate = slow ? 0.75 : 1;
        if (!cancelActive()) {
          notify({
            status: "error",
            message: "Pronunciation could not start. Please try again.",
          });
          return;
        }
        const request = { owner, environment, utterance, notify };
        active = request;
        utterance.onstart = () => {
          if (active === request) notify({ status: "playing" });
        };
        utterance.onend = () => {
          if (active !== request) return;
          active = null;
          notify({ status: "idle" });
        };
        utterance.onerror = () => {
          if (active !== request) return;
          active = null;
          notify({
            status: "error",
            message: "Pronunciation could not play. Please try again.",
          });
        };
        notify({ status: "starting" });
        try {
          environment.synthesis.speak(utterance);
        } catch {
          if (active === request) active = null;
          notify({
            status: "error",
            message: "Pronunciation could not play. Please try again.",
          });
        }
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

export const deviceSpeech = createDeviceSpeechController(() => {
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
