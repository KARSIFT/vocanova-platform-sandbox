import assert from "node:assert/strict";
import test from "node:test";

import {
  createPronunciationController,
  createPronunciationMedia,
  type NeuralSpeechEnvironment,
  type PronunciationMedia,
} from "../../src/lib/neural-speech";

import {
  createDeviceSpeechController,
  readSpeechPreferences,
  selectEnglishVoice,
  type PronunciationState,
  type SpeechEnvironment,
} from "../../src/lib/speech";

function neuralFixture() {
  const requests: {
    text: string;
    signal: AbortSignal;
    resolve: (value: { audioBase64: string; mimeType: string }) => void;
    reject: (error: Error) => void;
  }[] = [];
  const clips: {
    media: PronunciationMedia;
    disposed: number;
    paused: number;
  }[] = [];
  const deviceCalls: { owner: symbol; text: string; slow: boolean }[] = [];
  const releases: symbol[] = [];
  const environment: NeuralSpeechEnvironment = {
    request(text, signal) {
      return new Promise((resolve, reject) =>
        requests.push({ text, signal, resolve, reject }),
      );
    },
    createMedia() {
      const clip = {
        media: {
          playbackRate: 1,
          preservesPitch: false,
          onended: null,
          onerror: null,
          play: async () => {},
          pause: () => {
            clip.paused++;
          },
        } as PronunciationMedia,
        disposed: 0,
        paused: 0,
      };
      clips.push(clip);
      return {
        media: clip.media,
        dispose: () => {
          clip.disposed++;
        },
      };
    },
  };
  let preferences = { rate: 0.85, voiceURI: "" };
  const controller = createPronunciationController(
    {
      prepare() {},
      play(owner, text, slow, notify) {
        deviceCalls.push({ owner, text, slow });
        notify({ status: "playing" });
      },
      stop() {},
      release(owner) {
        releases.push(owner);
      },
    },
    () => environment,
    () => preferences,
  );
  return {
    controller,
    requests,
    clips,
    deviceCalls,
    releases,
    environment,
    setPreferences(next: typeof preferences) {
      preferences = next;
    },
  };
}

const audioResponse = { audioBase64: "aWQz", mimeType: "audio/mpeg" };
async function settleSpeech() {
  await new Promise<void>((resolve) => setImmediate(resolve));
}

function voice(
  lang: string,
  localService = true,
  isDefault = false,
  name = lang,
) {
  return {
    lang,
    localService,
    default: isDefault,
    name,
    voiceURI: name,
  } as SpeechSynthesisVoice;
}

function fixture() {
  const utterances: SpeechSynthesisUtterance[] = [];
  let cancellations = 0;
  let voices = [voice("en-GB")];
  const synthesis = {
    speaking: false,
    pending: false,
    getVoices: () => voices,
    speak: (utterance: SpeechSynthesisUtterance) => {
      utterances.push(utterance);
      synthesis.speaking = true;
      utterance.onstart?.({} as SpeechSynthesisEvent);
    },
    cancel: () => {
      cancellations += 1;
      synthesis.speaking = false;
      synthesis.pending = false;
    },
  };
  const environment: SpeechEnvironment = {
    synthesis,
    createUtterance: (text) =>
      ({
        text,
        onstart: null,
        onend: null,
        onerror: null,
      }) as unknown as SpeechSynthesisUtterance,
  };
  return {
    controller: createDeviceSpeechController(() => environment),
    environment,
    synthesis,
    utterances,
    get cancellations() {
      return cancellations;
    },
    setVoices: (next: SpeechSynthesisVoice[]) => {
      voices = next;
    },
  };
}

test("selects enhanced English voices before basic installed defaults and honors the chosen voice", () => {
  const foreign = voice("fr-FR", true, true);
  const remote = voice("en-US", false, true);
  const local = voice("en-GB");
  const localDefault = voice("en-AU", true, true);
  assert.equal(
    selectEnglishVoice([foreign, remote, local, localDefault]),
    remote,
  );
  assert.equal(selectEnglishVoice([foreign, remote, local]), remote);
  assert.equal(selectEnglishVoice([foreign, remote]), remote);
  assert.equal(selectEnglishVoice([foreign]), undefined);
  assert.equal(selectEnglishVoice([]), undefined);
  const google = voice("en-US", false, false, "Google US English");
  const natural = voice("en-GB", true, false, "English Natural");
  assert.equal(selectEnglishVoice([localDefault, google, natural]), natural);
  assert.equal(selectEnglishVoice([localDefault, google]), google);
  assert.equal(
    selectEnglishVoice([localDefault, google], localDefault.voiceURI),
    localDefault,
  );
  assert.equal(selectEnglishVoice([foreign, google], foreign.voiceURI), google);
  assert.equal(
    selectEnglishVoice([localDefault, google], "removed-voice"),
    google,
  );
});

test("speech preferences accept only supported speeds and a bounded voice identifier", () => {
  const read = (value: unknown) =>
    readSpeechPreferences({ getItem: () => JSON.stringify(value) });
  assert.deepEqual(read(null), { rate: 0.85, voiceURI: "" });
  assert.deepEqual(
    read({ rate: 0.7, voiceURI: "Google US English", text: "never retained" }),
    { rate: 0.7, voiceURI: "Google US English" },
  );
  assert.deepEqual(read({ rate: 5, voiceURI: "x".repeat(513) }), {
    rate: 0.85,
    voiceURI: "",
  });
  assert.deepEqual(readSpeechPreferences({ getItem: () => "invalid JSON" }), {
    rate: 0.85,
    voiceURI: "",
  });
  assert.deepEqual(
    readSpeechPreferences({
      getItem: () => {
        throw new Error("private storage error");
      },
    }),
    { rate: 0.85, voiceURI: "" },
  );
});

test("preparation does not play; explicit playback uses the authored text and speed", () => {
  const f = fixture();
  const owner = Symbol();
  const states: PronunciationState[] = [];
  f.controller.prepare();
  assert.equal(f.utterances.length, 0);
  f.controller.play(
    owner,
    "Could you pour me a cup of coffee?",
    true,
    (state) => states.push(state),
  );
  assert.equal(f.utterances[0]!.text, "Could you pour me a cup of coffee?");
  assert.equal(f.utterances[0]!.lang, "en-GB");
  assert.equal(f.utterances[0]!.rate, 0.7);
  assert.deepEqual(
    states.map((state) => state.status),
    ["starting", "playing"],
  );
  f.controller.stop(owner);
  assert.equal(states.at(-1)!.status, "idle");
});

test("switching controls cancels owned speech; stale events and other unmounts cannot stop the new voice", () => {
  const f = fixture();
  const first = Symbol();
  const second = Symbol();
  const firstStates: PronunciationState[] = [];
  const secondStates: PronunciationState[] = [];
  f.controller.play(first, "pour", false, (state) => firstStates.push(state));
  const previous = f.utterances[0]!;
  f.controller.play(second, "I pour the tea.", false, (state) =>
    secondStates.push(state),
  );
  assert.equal(f.cancellations, 1);
  assert.equal(firstStates.at(-1)!.status, "idle");
  assert.equal(f.utterances[1]!.rate, 0.85);
  f.controller.release(first);
  previous.onend?.({} as SpeechSynthesisEvent);
  previous.onerror?.({} as SpeechSynthesisErrorEvent);
  assert.equal(f.cancellations, 1);
  assert.equal(secondStates.at(-1)!.status, "playing");
  f.controller.release(second);
  assert.equal(f.cancellations, 2);
  assert.equal(
    secondStates.at(-1)!.status,
    "playing",
    "unmounted controls are not notified",
  );
});

test("does not interrupt an unowned speech queue", () => {
  const f = fixture();
  const owner = Symbol();
  const states: PronunciationState[] = [];
  f.synthesis.pending = true;
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  f.controller.stop(owner);
  f.controller.release(owner);
  assert.equal(f.utterances.length, 0);
  assert.equal(f.cancellations, 0);
  assert.equal(states.at(-1)!.status, "error");
});

test("missing support and delayed English voices give a recoverable explanation", () => {
  const states: PronunciationState[] = [];
  const notify = (state: PronunciationState) => states.push(state);
  createDeviceSpeechController(() => null).play(
    Symbol(),
    "pour",
    false,
    notify,
  );
  assert.equal(states.at(-1)!.status, "unavailable");
  const f = fixture();
  const owner = Symbol();
  f.setVoices([voice("fr-FR")]);
  f.controller.play(owner, "pour", false, notify);
  assert.equal(f.utterances.length, 0);
  assert.match(states.at(-1)!.message!, /English voice/);
  f.setVoices([voice("en-US")]);
  f.controller.play(owner, "pour", false, notify);
  assert.equal(states.at(-1)!.status, "playing");
  f.controller.stop(owner);
});

test("speech failures stay recoverable without exposing the native error", () => {
  const f = fixture();
  const owner = Symbol();
  const states: PronunciationState[] = [];
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  f.synthesis.speaking = false;
  f.utterances[0]!.onerror?.({ error: "network" } as SpeechSynthesisErrorEvent);
  assert.equal(states.at(-1)!.status, "error");
  assert.match(states.at(-1)!.message!, /Please try again/);
  f.environment.synthesis.speak = () => {
    throw new Error("private native error");
  };
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  assert.equal(states.at(-1)!.status, "error");
  assert.doesNotMatch(states.at(-1)!.message!, /private/);
  f.controller.release(owner);
});

test("remote voice failures fall back once to local English, preserving speed and ownership", () => {
  const f = fixture();
  const local = voice("en-GB", true, true, "Basic English");
  const remote = voice("en-US", false, false, "Google US English");
  f.setVoices([voice("fr-FR"), local, remote]);
  const states: PronunciationState[] = [];
  const owner = Symbol();
  f.controller.play(owner, "An authored example.", false, (state) =>
    states.push(state),
  );
  assert.equal(f.utterances[0]!.voice, remote);
  f.utterances[0]!.onerror?.({ error: "network" } as SpeechSynthesisErrorEvent);
  assert.equal(f.utterances.length, 2);
  assert.equal(f.utterances[1]!.voice, local);
  assert.equal(f.utterances[1]!.rate, 0.85);
  f.utterances[0]!.onend?.({} as SpeechSynthesisEvent);
  assert.equal(states.at(-1)!.status, "playing");
  f.utterances[1]!.onerror?.({ error: "network" } as SpeechSynthesisErrorEvent);
  assert.equal(f.utterances.length, 2);
  assert.equal(states.at(-1)!.status, "error");
  f.controller.release(owner);
});

test("denied autoplay never retries or bypasses browser permission", () => {
  const f = fixture();
  f.setVoices([
    voice("en-US", false, false, "Google US English"),
    voice("en-GB"),
  ]);
  const states: PronunciationState[] = [];
  f.controller.play(Symbol(), "pour", false, (state) => states.push(state));
  f.utterances[0]!.onerror?.({
    error: "not-allowed",
  } as SpeechSynthesisErrorEvent);
  assert.equal(f.utterances.length, 1);
  assert.equal(f.cancellations, 0);
  assert.equal(states.at(-1)!.status, "error");
});

test("every replay reads the latest speed and voice preferences", () => {
  const f = fixture();
  const local = voice("en-GB", true, true, "Basic English");
  const remote = voice("en-US", false, false, "Google US English");
  f.setVoices([local, remote]);
  let preferences = { rate: 0.7, voiceURI: local.voiceURI };
  const controller = createDeviceSpeechController(
    () => f.environment,
    () => preferences,
  );
  const owner = Symbol();
  controller.play(owner, "pour", false, () => {});
  assert.equal(f.utterances[0]!.rate, 0.7);
  assert.equal(f.utterances[0]!.voice, local);
  controller.stop(owner);
  preferences = { rate: 1, voiceURI: remote.voiceURI };
  controller.play(owner, "pour", false, () => {});
  assert.equal(f.utterances[1]!.rate, 1);
  assert.equal(f.utterances[1]!.voice, remote);
  controller.stop(owner);
});

test("natural audio uses the saved learner pace, preserves pitch and releases its clip on completion", async () => {
  const f = neuralFixture();
  const states: PronunciationState[] = [];
  f.controller.play(Symbol(), "pour", false, (state) => states.push(state));
  assert.deepEqual(states, [{ status: "starting" }]);
  assert.equal(f.requests[0]!.text, "pour");
  f.requests[0]!.resolve(audioResponse);
  await settleSpeech();
  assert.equal(f.clips[0]!.media.playbackRate, 0.85);
  assert.equal(f.clips[0]!.media.preservesPitch, true);
  assert.deepEqual(states.at(-1), { status: "playing", source: "neural" });
  f.clips[0]!.media.onended?.();
  assert.equal(f.clips[0]!.disposed, 1);
  assert.deepEqual(states.at(-1), { status: "idle", source: "neural" });
  assert.equal(f.deviceCalls.length, 0);
});

test("stopping a pending natural voice aborts it and ignores late results; unrelated owners cannot cancel it", async () => {
  const f = neuralFixture();
  const owner = Symbol();
  const states: PronunciationState[] = [];
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  f.controller.release(Symbol());
  assert.equal(f.requests[0]!.signal.aborted, false);
  f.controller.stop(owner);
  assert.equal(f.requests[0]!.signal.aborted, true);
  f.requests[0]!.resolve(audioResponse);
  await settleSpeech();
  assert.equal(f.clips.length, 0);
  assert.equal(states.at(-1)!.status, "idle");
  f.controller.play(owner, "coffee", true, (state) => states.push(state));
  f.requests[1]!.resolve(audioResponse);
  await settleSpeech();
  assert.equal(f.clips[0]!.media.playbackRate, 0.7);
  f.controller.release(owner);
  assert.equal(f.clips[0]!.disposed, 1);
});

test("natural audio failure and decode/CSP failure use an honest device fallback", async () => {
  for (const failure of ["provider", "decode"] as const) {
    const f = neuralFixture();
    const states: PronunciationState[] = [];
    if (failure === "decode")
      f.environment.createMedia = () => {
        throw new Error("private CSP/decode detail");
      };
    const owner = Symbol();
    f.controller.play(owner, "pour", false, (state) => states.push(state));
    if (failure === "provider")
      f.requests[0]!.reject(new Error("private quota or provider detail"));
    else f.requests[0]!.resolve(audioResponse);
    await settleSpeech();
    assert.equal(f.deviceCalls.length, 1);
    assert.deepEqual(states.at(-1), { status: "playing", source: "device" });
    assert.doesNotMatch(JSON.stringify(states), /private/);
    f.controller.release(owner);
  }
});

test("natural media playback errors release the blob and fall back, but denied autoplay never bypasses permission", async () => {
  for (const errorName of ["NotSupportedError", "NotAllowedError"]) {
    const f = neuralFixture();
    const create = f.environment.createMedia;
    f.environment.createMedia = (response) => {
      const clip = create(response);
      clip.media.play = async () => {
        throw new DOMException("private media detail", errorName);
      };
      return clip;
    };
    const owner = Symbol();
    const states: PronunciationState[] = [];
    f.controller.play(owner, "pour", false, (state) => states.push(state));
    f.requests[0]!.resolve(audioResponse);
    await settleSpeech();
    assert.equal(f.clips[0]!.disposed, errorName === "NotAllowedError" ? 0 : 1);
    assert.equal(f.deviceCalls.length, errorName === "NotAllowedError" ? 0 : 1);
    assert.equal(
      states.at(-1)!.status,
      errorName === "NotAllowedError" ? "error" : "playing",
    );
    assert.doesNotMatch(JSON.stringify(states), /private/);
    f.controller.release(owner);
    assert.equal(f.clips[0]!.disposed, 1);
  }
});

test("a denied ready clip retries directly in the next gesture without fetching or retaining another clip", async () => {
  const f = neuralFixture();
  const create = f.environment.createMedia;
  let gesture = false;
  const attemptedGestures: boolean[] = [];
  f.environment.createMedia = (response) => {
    const clip = create(response);
    clip.media.play = () => {
      attemptedGestures.push(gesture);
      return gesture
        ? Promise.resolve()
        : Promise.reject(
            new DOMException("gesture required", "NotAllowedError"),
          );
    };
    return clip;
  };
  const owner = Symbol();
  const states: PronunciationState[] = [];
  gesture = true;
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  gesture = false;
  f.requests[0]!.resolve(audioResponse);
  await settleSpeech();
  assert.deepEqual(attemptedGestures, [false]);
  assert.equal(states.at(-1)!.status, "error");
  assert.equal(f.clips[0]!.disposed, 0);
  assert.equal(
    f.deviceCalls.length,
    0,
    "denied autoplay must not trigger device speech",
  );
  f.setPreferences({ rate: 0.7, voiceURI: "" });
  gesture = true;
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  assert.deepEqual(
    attemptedGestures,
    [false, true],
    "play() must run before play() returns to the event loop",
  );
  gesture = false;
  await settleSpeech();
  assert.equal(f.requests.length, 1);
  assert.equal(f.clips.length, 1);
  assert.equal(f.clips[0]!.media.playbackRate, 0.7);
  assert.deepEqual(states.at(-1), { status: "playing", source: "neural" });
  f.controller.release(owner);
  assert.equal(f.clips[0]!.disposed, 1);
});

test("a denied ready clip is disposed on stop, text change, owner change and voice override", async () => {
  for (const action of ["stop", "text", "owner", "voice"] as const) {
    const f = neuralFixture();
    const create = f.environment.createMedia;
    f.environment.createMedia = (response) => {
      const clip = create(response);
      clip.media.play = async () => {
        throw new DOMException("gesture required", "NotAllowedError");
      };
      return clip;
    };
    const owner = Symbol();
    f.controller.play(owner, "pour", false, () => {});
    f.requests[0]!.resolve(audioResponse);
    await settleSpeech();
    if (action === "stop") f.controller.stop(owner);
    else {
      if (action === "voice")
        f.setPreferences({ rate: 0.85, voiceURI: "offline-choice" });
      f.controller.play(
        action === "owner" ? Symbol() : owner,
        action === "text" ? "coffee" : "pour",
        false,
        () => {},
      );
    }
    assert.equal(f.clips[0]!.disposed, 1, action);
    assert.equal(
      f.requests.length,
      action === "text" || action === "owner" ? 2 : 1,
      action,
    );
    // Release the only remaining request, including a differently owned one.
    if (action === "text" || action === "owner")
      f.requests[1]!.reject(new Error("test cleanup"));
    await settleSpeech();
    f.controller.release(owner);
  }
});

test("a timeout uses device audio and a late completed generation cannot replace it", async (t) => {
  t.mock.timers.enable({ apis: ["setTimeout"] });
  const f = neuralFixture();
  const states: PronunciationState[] = [];
  const owner = Symbol();
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  t.mock.timers.tick(8000);
  assert.equal(f.requests[0]!.signal.aborted, true);
  assert.equal(f.deviceCalls.length, 1);
  f.requests[0]!.resolve(audioResponse);
  await settleSpeech();
  assert.equal(f.clips.length, 0);
  assert.deepEqual(states.at(-1), { status: "playing", source: "device" });
  f.controller.release(owner);
});

test("an explicit device voice skips generation and a newer word owns the only active clip", async () => {
  const f = neuralFixture();
  const first = Symbol();
  f.setPreferences({ rate: 1, voiceURI: "chosen-device-voice" });
  f.controller.play(first, "pour", false, () => {});
  assert.equal(f.requests.length, 0);
  assert.equal(f.deviceCalls.length, 1);
  f.setPreferences({ rate: 0.85, voiceURI: "" });
  const second = Symbol();
  f.controller.play(second, "coffee", false, () => {});
  f.requests[0]!.resolve(audioResponse);
  await settleSpeech();
  const oldEnd = f.clips[0]!.media.onended;
  f.controller.play(first, "tea", false, () => {});
  assert.equal(f.clips[0]!.disposed, 1);
  oldEnd?.();
  f.controller.release(second);
  assert.equal(f.requests[1]!.signal.aborted, false);
  f.requests[1]!.resolve(audioResponse);
  await settleSpeech();
  assert.equal(f.clips.length, 2);
  f.controller.release(first);
  assert.equal(f.clips[1]!.disposed, 1);
});

test("media factory rejects unsupported/oversized/invalid clips and revokes a URL if constructing audio is blocked", () => {
  assert.throws(() =>
    createPronunciationMedia({ audioBase64: "aWQz", mimeType: "text/html" }),
  );
  assert.throws(() =>
    createPronunciationMedia({
      audioBase64: "x".repeat(2_000_001),
      mimeType: "audio/mpeg",
    }),
  );
  assert.throws(() =>
    createPronunciationMedia({ audioBase64: "%%%", mimeType: "audio/mpeg" }),
  );
  const revoked: string[] = [];
  const browser = {
    atob,
    Blob,
    Audio: class {
      constructor() {
        throw new DOMException("blocked", "SecurityError");
      }
    },
    URL: {
      createObjectURL: () => "blob:private-test",
      revokeObjectURL: (value: string) => revoked.push(value),
    },
  } as unknown as Pick<typeof globalThis, "atob" | "Audio" | "Blob" | "URL">;
  assert.throws(() => createPronunciationMedia(audioResponse, browser));
  assert.deepEqual(revoked, ["blob:private-test"]);
});

test("a released clip whose play promise settles later cannot change the next word", async () => {
  const f = neuralFixture();
  const create = f.environment.createMedia;
  let finishPlay: (() => void) | undefined;
  f.environment.createMedia = (response) => {
    const clip = create(response);
    clip.media.play = () =>
      new Promise((resolve) => {
        finishPlay = resolve;
      });
    return clip;
  };
  const first = Symbol();
  const states: PronunciationState[] = [];
  f.controller.play(first, "pour", false, (state) => states.push(state));
  f.requests[0]!.resolve(audioResponse);
  await settleSpeech();
  const second = Symbol();
  f.controller.play(second, "coffee", false, () => {});
  const beforeLatePlay = states.length;
  finishPlay?.();
  await settleSpeech();
  assert.equal(states.length, beforeLatePlay);
  assert.equal(f.clips[0]!.paused, 1);
  assert.equal(f.clips[0]!.disposed, 1);
  assert.equal(f.requests[1]!.signal.aborted, false);
  f.controller.release(second);
});

test("a media error after playback begins switches once to device audio and frees the clip", async () => {
  const f = neuralFixture();
  const states: PronunciationState[] = [];
  const owner = Symbol();
  f.controller.play(owner, "pour", false, (state) => states.push(state));
  f.requests[0]!.resolve(audioResponse);
  await settleSpeech();
  const staleEnd = f.clips[0]!.media.onended;
  f.clips[0]!.media.onerror?.();
  staleEnd?.();
  assert.equal(f.deviceCalls.length, 1);
  assert.equal(f.clips[0]!.disposed, 1);
  assert.deepEqual(states.at(-1), { status: "playing", source: "device" });
  f.controller.release(owner);
});

test("natural audio never interrupts speech outside the shared player", () => {
  const f = neuralFixture();
  f.environment.otherSpeechActive = () => true;
  const states: PronunciationState[] = [];
  f.controller.play(Symbol(), "pour", false, (state) => states.push(state));
  assert.equal(f.requests.length, 0);
  assert.equal(f.deviceCalls.length, 0);
  assert.equal(f.releases.length, 0);
  assert.equal(states.at(-1)!.status, "error");
});
