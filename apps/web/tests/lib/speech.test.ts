import assert from "node:assert/strict";
import test from "node:test";

import {
  createDeviceSpeechController,
  selectEnglishVoice,
  type PronunciationState,
  type SpeechEnvironment,
} from "../../src/lib/speech";

function voice(lang: string, localService = true, isDefault = false) {
  return { lang, localService, default: isDefault } as SpeechSynthesisVoice;
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

test("selects a real English voice, preferring an installed default", () => {
  const foreign = voice("fr-FR", true, true);
  const remote = voice("en-US", false, true);
  const local = voice("en-GB");
  const localDefault = voice("en-AU", true, true);
  assert.equal(
    selectEnglishVoice([foreign, remote, local, localDefault]),
    localDefault,
  );
  assert.equal(selectEnglishVoice([foreign, remote, local]), local);
  assert.equal(selectEnglishVoice([foreign, remote]), remote);
  assert.equal(selectEnglishVoice([foreign]), undefined);
  assert.equal(selectEnglishVoice([]), undefined);
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
  assert.equal(f.utterances[0]!.rate, 0.75);
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
  assert.equal(f.utterances[1]!.rate, 1);
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
