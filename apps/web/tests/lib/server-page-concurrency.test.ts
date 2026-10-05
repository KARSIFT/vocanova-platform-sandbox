import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { registerHooks } from "node:module";
import { after, afterEach, describe, it } from "node:test";
import { setImmediate } from "node:timers/promises";
import ts from "typescript";
import { ApiResponseError } from "@vocanova/api-client";
import { isValidElement, type ReactElement } from "react";

// Import the actual server pages, rather than a duplicated data loader. Node's
// existing test hooks handle extensionless imports and next/headers; this hook
// only adds the bundler's source alias and JSX transformation for these tests.
const sourceRoot = new URL("../../src/", import.meta.url);
const pageHooks = registerHooks({
  resolve(specifier, context, nextResolve) {
    if (specifier.startsWith("@/")) {
      return nextResolve(new URL(specifier.slice(2), sourceRoot).href, context);
    }
    return nextResolve(specifier, context);
  },
  load(url, context, nextLoad) {
    if (url.startsWith(sourceRoot.href) && url.endsWith(".tsx")) {
      return {
        format: "module",
        source: ts.transpileModule(readFileSync(new URL(url), "utf8"), {
          compilerOptions: {
            target: ts.ScriptTarget.ESNext,
            module: ts.ModuleKind.ESNext,
            jsx: ts.JsxEmit.ReactJSX,
          },
        }).outputText,
        shortCircuit: true,
      };
    }
    return nextLoad(url, context);
  },
});
after(() => pageHooks.deregister());

type ServerPage = (props?: {
  searchParams: Promise<Record<string, string>>;
}) => Promise<unknown>;
async function importPage(path: string): Promise<ServerPage> {
  return (await import(new URL(`app/(app)/${path}/page.tsx`, sourceRoot).href))
    .default;
}
const HomePage = await importPage("home");
const JourneyPage = await importPage("discover");
const ProgressPage = await importPage("progress");
const WritingPage = await importPage("writing");

const originalFetch = globalThis.fetch;
const originalApiBaseURL = process.env.API_BASE_URL;
afterEach(() => {
  globalThis.fetch = originalFetch;
  if (originalApiBaseURL === undefined) delete process.env.API_BASE_URL;
  else process.env.API_BASE_URL = originalApiBaseURL;
});

const api = (path: string) => `/api/v1/${path}`;
const emptyCollection = { items: [], hasMore: false };
const fixture: Record<string, unknown> = {
  [api("me")]: { id: "test-learner", onboardingStatus: "completed" },
  [api("user-words")]: emptyCollection,
  [api("reviews/due")]: { ...emptyCollection, totalCount: 0 },
  [api("daily-mission")]: {
    reviewTarget: 5,
    reviewsCompleted: 0,
    newWordTarget: 1,
    newWordsCompleted: 0,
    sentencePracticeTarget: 1,
    sentencePracticesCompleted: 0,
    status: "in_progress",
    streak: { currentStreakCount: 0, longestStreakCount: 0 },
  },
  [api("lesson-recommendation")]: { status: "content_unavailable" },
  [api("journey-situations")]: emptyCollection,
  [api("journey-situations/daily-conversation")]: {
    situation: { slug: "daily-conversation", title: "Daily Conversation" },
    meanings: [],
  },
  [api("lessons")]: emptyCollection,
  [api("progress")]: {
    confidencePointsBalance: 0,
    streak: { currentStreakCount: 0, longestStreakCount: 0 },
    completionHistory: [],
  },
  [api("knowledge-summary")]: {},
  [api("achievements")]: {},
  [api("learner-sentences")]: emptyCollection,
  [api("canonical-words/invite")]: {
    word: {
      slug: "invite",
      meanings: [
        {
          id: "invite-meaning",
          shortDefinition: "Ask someone to join you.",
          saved: false,
        },
      ],
    },
  },
};

function mockFetch(
  statuses: Record<string, number> = {},
  held = false,
  overrides: Record<string, unknown> = {},
) {
  process.env.API_BASE_URL = "http://page-test.internal";
  const started: string[] = [];
  const release: (() => void)[] = [];
  globalThis.fetch = (async (input, init) => {
    const path = new URL(String(input)).pathname;
    assert.equal(
      init?.method,
      "GET",
      "rendering must not mutate learner state",
    );
    assert.ok(Object.hasOwn(fixture, path), `Unexpected request: ${path}`);
    started.push(path);
    if (held) await new Promise<void>((resolve) => release.push(resolve));
    return Response.json(overrides[path] ?? fixture[path], {
      status: statuses[path] ?? 200,
    });
  }) as typeof fetch;
  return {
    started,
    release() {
      held = false;
      for (const resolve of release) resolve();
    },
  };
}

function elements(node: unknown): ReactElement<Record<string, unknown>>[] {
  if (Array.isArray(node)) return node.flatMap(elements);
  if (!isValidElement<Record<string, unknown>>(node)) return [];
  return [node, ...elements(node.props.children)];
}
function text(node: unknown): string {
  if (Array.isArray(node)) return node.map(text).join(" ");
  if (isValidElement<Record<string, unknown>>(node))
    return text(node.props.children);
  return typeof node === "string" ? node : "";
}

describe("independent server page reads", () => {
  for (const [name, page, expected] of [
    [
      "Home",
      HomePage,
      [
        "user-words",
        "reviews/due",
        "daily-mission",
        "me",
        "lesson-recommendation",
      ],
    ],
    [
      "Journey",
      JourneyPage,
      ["journey-situations", "lessons", "lesson-recommendation"],
    ],
    [
      "Progress",
      ProgressPage,
      [
        "user-words",
        "progress",
        "knowledge-summary",
        "lessons",
        "achievements",
        "learner-sentences",
      ],
    ],
    [
      "Writing topic",
      () =>
        WritingPage({
          searchParams: Promise.resolve({ situation: "daily-conversation" }),
        }),
      ["journey-situations", "me", "journey-situations/daily-conversation"],
    ],
    [
      "Writing without topic",
      () => WritingPage({ searchParams: Promise.resolve({}) }),
      ["journey-situations", "me"],
    ],
  ] as const) {
    it(`${name} starts independent GETs before any response is released`, async () => {
      const requests = mockFetch({}, true);
      const rendering = page();
      // Drain the event-loop's microtasks while every response remains blocked.
      // This asserts causality, not elapsed time or a machine-dependent budget.
      await setImmediate();
      try {
        assert.deepEqual(
          [...requests.started].sort(),
          expected.map(api).sort(),
        );
      } finally {
        requests.release();
        await rendering;
      }
    });
  }
});

describe("page error semantics", () => {
  for (const [name, page, required, optional, returnTo] of [
    ["Home", HomePage, "daily-mission", "lesson-recommendation", "/home"],
    [
      "Journey",
      JourneyPage,
      "journey-situations",
      "lesson-recommendation",
      "/discover",
    ],
    ["Progress", ProgressPage, "progress", "learner-sentences", "/progress"],
  ] as const) {
    it(`${name} propagates mandatory failure`, async () => {
      mockFetch({ [api(required)]: 503 });
      await assert.rejects(
        page(),
        (error: unknown) =>
          error instanceof ApiResponseError && error.status === 503,
      );
    });
    it(`${name} retains optional-service fallback`, async () => {
      mockFetch({ [api(optional)]: 503 });
      const rendered = await page();
      if (name === "Progress") {
        assert.ok(
          text(rendered).includes("Your recent writing could not load."),
        );
      } else {
        assert.ok(
          elements(rendered).some(
            (element) =>
              typeof element.type === "function" &&
              element.type.name === "RecommendedLesson" &&
              element.props.data === null,
          ),
        );
      }
    });
    it(`${name} redirects expired optional authentication`, async () => {
      mockFetch({ [api(optional)]: 401 });
      await assert.rejects(
        page(),
        (error: unknown) =>
          error instanceof Error &&
          "digest" in error &&
          String(error.digest).includes(
            `/login?returnTo=${encodeURIComponent(returnTo)}`,
          ),
      );
    });
    it(`${name} redirects expired mandatory authentication`, async () => {
      mockFetch({ [api(required)]: 401 });
      await assert.rejects(
        page(),
        (error: unknown) =>
          error instanceof Error &&
          "digest" in error &&
          String(error.digest).includes(
            `/login?returnTo=${encodeURIComponent(returnTo)}`,
          ),
      );
    });
  }
  for (const [optional, unavailable] of [
    ["knowledge-summary", "Your vocabulary map is unavailable right now."],
    ["lessons", null],
    ["achievements", "Your milestones could not load."],
  ] as const) {
    it(`Progress redirects ${optional} 401`, async () => {
      mockFetch({ [api(optional)]: 401 });
      await assert.rejects(
        ProgressPage(),
        (error: unknown) =>
          error instanceof Error &&
          "digest" in error &&
          String(error.digest).includes("/login?returnTo=%2Fprogress"),
      );
    });
    it(`Progress retains ${optional} non-auth fallback`, async () => {
      mockFetch({ [api(optional)]: 503 });
      const renderedText = text(await ProgressPage());
      assert.ok(renderedText.includes("Confidence Points"));
      if (unavailable) assert.ok(renderedText.includes(unavailable));
      else assert.ok(!renderedText.includes("Your learning path"));
    });
  }
});

describe("writing topic and meaning validation", () => {
  it("starts canonical lookup only after the server confirms topic membership", async () => {
    const requests = mockFetch({}, true, {
      [api("journey-situations/daily-conversation")]: {
        situation: { slug: "daily-conversation", title: "Daily Conversation" },
        meanings: [
          {
            meaningId: "invite-meaning",
            wordSlug: "invite",
            wordText: "invite",
            shortDefinition: "Ask someone to join you.",
          },
        ],
      },
    });
    const rendering = WritingPage({
      searchParams: Promise.resolve({
        situation: "daily-conversation",
        meaning: "invite-meaning",
      }),
    });
    await setImmediate();
    try {
      assert.deepEqual(
        [...requests.started].sort(),
        ["journey-situations", "me", "journey-situations/daily-conversation"]
          .map(api)
          .sort(),
      );
    } finally {
      requests.release();
      await rendering;
    }
    assert.equal(
      requests.started.filter((path) => path === api("canonical-words/invite"))
        .length,
      1,
    );
  });
  it("rejects a meaning outside the server-provided topic", async () => {
    const requests = mockFetch();
    await assert.rejects(
      WritingPage({
        searchParams: Promise.resolve({
          situation: "daily-conversation",
          meaning: "invite-meaning",
        }),
      }),
      (error: unknown) =>
        error instanceof Error &&
        "digest" in error &&
        String(error.digest).includes("404"),
    );
    assert.ok(
      !requests.started.some((path) => path.includes("canonical-words")),
    );
  });
  it("redirects an expired topic request with the selected meaning intact", async () => {
    mockFetch({ [api("journey-situations/daily-conversation")]: 401 });
    const returnTo =
      "/writing?situation=daily-conversation&meaning=invite-meaning";
    await assert.rejects(
      WritingPage({
        searchParams: Promise.resolve({
          situation: "daily-conversation",
          meaning: "invite-meaning",
        }),
      }),
      (error: unknown) =>
        error instanceof Error &&
        "digest" in error &&
        String(error.digest).includes(
          `/login?returnTo=${encodeURIComponent(returnTo)}`,
        ),
    );
  });
});
