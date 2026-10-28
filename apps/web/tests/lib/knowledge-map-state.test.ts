import assert from "node:assert/strict";
import { test } from "node:test";

import type { KnowledgeMap, KnowledgeMapDecision } from "@vocanova/api-client";

import {
  applyWordState,
  confirmedWordState,
  confirmsChoice,
  nextDiscoveryTarget,
} from "../../src/app/(app)/knowledge-map/models";

const revision = (value: number) => String(value).padStart(64, "0");

function range(): KnowledgeMap {
  return {
    attribution: {} as KnowledgeMap["attribution"],
    catalogVersion: "test",
    totalWords: 200,
    bandSize: 100,
    ordering: { title: "", description: "", sourceUrl: "" },
    bands: [
      { number: 1, startRank: 1, endRank: 100, totalWords: 100, knownCount: 1, learningCount: 0, unseenCount: 99 },
      { number: 2, startRank: 101, endRank: 200, totalWords: 100, knownCount: 0, learningCount: 0, unseenCount: 100 },
    ],
    currentBand: {
      number: 1,
      items: Array.from({ length: 100 }, (_, index) => ({
        rank: index + 1,
        meaningId: `m-${index + 1}`,
        wordId: `w-${index + 1}`,
        wordText: `word ${index + 1}`,
        wordSlug: `word-${index + 1}`,
        partOfSpeech: "noun",
        shortDefinition: "A test meaning.",
        selfReportedKnown: index === 0,
        saved: false,
        due: false,
        revision: revision(index === 0 ? 1 : 0),
      })),
    },
  };
}

const learnSecond: KnowledgeMapDecision = {
  meaningId: "m-2",
  decision: "learn",
  expectedRevision: revision(0),
};

test("the decided meaning's state comes from the response, not the request", () => {
  // Replaying an old Learn receipt after the meaning was marked known elsewhere.
  const state = confirmedWordState(
    {
      meaningId: "m-2",
      decision: "learn",
      revision: revision(3),
      selfReportedKnown: true,
      saved: false,
      due: false,
    },
    learnSecond,
  );
  assert.deepEqual(state, {
    selfReportedKnown: true,
    saved: false,
    due: false,
    revision: revision(3),
  });
  assert.equal(confirmsChoice(state!, "learn"), false);
  assert.equal(confirmsChoice(state!, "known"), true);
});

test("incomplete or mismatched responses are not trusted", () => {
  const complete = {
    meaningId: "m-2",
    decision: "learn",
    revision: revision(1),
    selfReportedKnown: false,
    saved: true,
    due: true,
  };
  assert.notEqual(confirmedWordState(complete, learnSecond), null);
  for (const broken of [
    null,
    "ok",
    { ...complete, meaningId: "m-3" },
    { ...complete, revision: "not-a-revision" },
    { meaningId: "m-2", decision: "learn", revision: revision(1) },
    { ...complete, saved: "true" },
  ])
    assert.equal(confirmedWordState(broken, learnSecond), null);
});

test("a removal reported on replay leaves the meaning new and the counts consistent", () => {
  const data = range();
  const removed = applyWordState(data, "m-2", {
    selfReportedKnown: false,
    saved: false,
    due: false,
    revision: revision(4),
  });
  assert.deepEqual(removed.bands[0], data.bands[0]);
  assert.equal(nextDiscoveryTarget(removed)?.rank, 2);
});

test("applying confirmed states moves exactly one word between disjoint counts", () => {
  let data = range();
  data = applyWordState(data, "m-2", {
    selfReportedKnown: false,
    saved: true,
    due: true,
    revision: revision(1),
  });
  assert.deepEqual(
    [data.bands[0]!.knownCount, data.bands[0]!.learningCount, data.bands[0]!.unseenCount],
    [1, 1, 98],
  );
  // The same word later reported as known (another tab) moves, not duplicates.
  data = applyWordState(data, "m-2", {
    selfReportedKnown: true,
    saved: false,
    revision: revision(2),
  });
  assert.deepEqual(
    [data.bands[0]!.knownCount, data.bands[0]!.learningCount, data.bands[0]!.unseenCount],
    [2, 0, 98],
  );
  const word = data.currentBand.items[1]!;
  assert.equal(word.revision, revision(2));
  assert.equal(word.due, true, "an absent due flag keeps the previous value");
  assert.equal(nextDiscoveryTarget(data)?.rank, 3);
  // Other ranges are untouched.
  assert.deepEqual(data.bands[1], range().bands[1]);
});
