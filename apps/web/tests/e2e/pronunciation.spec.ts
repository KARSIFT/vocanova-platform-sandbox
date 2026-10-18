import { openOptionalWriting } from "./optional-writing";
import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

interface SpeechCapture {
  calls: { text: string; lang: string; rate: number }[];
  cancellations: number;
  selectedVoiceURI?: string;
  finish: () => void;
  fail: () => void;
}

declare global {
  interface Window {
    __pronunciationTest: SpeechCapture;
  }
}

async function installSpeech(
  page: Page,
  mode: "ready" | "unsupported" | "no-english" | "enhanced" = "ready",
) {
  await page.route("**/api/v1/pronunciation", async (route) => {
    const body = route.request().postDataJSON();
    if (mode === "unsupported" || mode === "no-english") {
      await route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({ title: "Unavailable" }),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        audioBase64: Buffer.from(JSON.stringify({ text: body.text })).toString(
          "base64",
        ),
        mimeType: "audio/mpeg",
        voice: "coral",
        cacheHit: false,
        meaningId: body.meaningId,
      }),
    });
  });
  await page.addInitScript((speechMode) => {
    type Utterance = {
      text: string;
      lang: string;
      rate: number;
      voice?: { voiceURI: string };
      onstart?: () => void;
      onend?: () => void;
      onerror?: () => void;
    };
    let active: Utterance | null = null;
    const capture: SpeechCapture = {
      calls: [],
      cancellations: 0,
      finish: () => {
        const current = active;
        active = null;
        synthesis.speaking = false;
        current?.onend?.();
      },
      fail: () => {
        const current = active;
        active = null;
        synthesis.speaking = false;
        current?.onerror?.();
      },
    };
    const synthesis = {
      speaking: false,
      pending: false,
      getVoices: () =>
        speechMode === "no-english"
          ? [{ lang: "fr-FR", localService: true, default: true }]
          : [
              { lang: "fr-FR", localService: true, default: true },
              {
                lang: "en-GB",
                name: "English UK",
                voiceURI: "english-uk",
                localService: true,
                default: true,
              },
              ...(speechMode === "enhanced"
                ? [
                    {
                      lang: "en-US",
                      name: "Google US English",
                      voiceURI: "google-us",
                      localService: false,
                      default: false,
                    },
                  ]
                : []),
            ],
      speak: (utterance: Utterance) => {
        capture.selectedVoiceURI = utterance.voice?.voiceURI;
        capture.calls.push({
          text: utterance.text,
          lang: utterance.lang,
          rate: utterance.rate,
        });
        active = utterance;
        synthesis.speaking = true;
        utterance.onstart?.();
      },
      cancel: () => {
        capture.cancellations += 1;
        const previous = active;
        active = null;
        synthesis.speaking = false;
        previous?.onend?.();
      },
    };
    class DeviceUtterance {
      text: string;
      constructor(text: string) {
        this.text = text;
      }
    }
    // Synthetic transport/media fixture, not a real MP3 or a voice-quality test.
    const audioBlobs = new Map<string, Blob>();
    const createObjectURL = URL.createObjectURL.bind(URL);
    const revokeObjectURL = URL.revokeObjectURL.bind(URL);
    URL.createObjectURL = (blob) => {
      const url = createObjectURL(blob);
      if (blob instanceof Blob) audioBlobs.set(url, blob);
      return url;
    };
    URL.revokeObjectURL = (url) => {
      audioBlobs.delete(url);
      revokeObjectURL(url);
    };
    class AuthoredAudio {
      playbackRate = 1;
      preservesPitch = false;
      onended: (() => void) | null = null;
      onerror: (() => void) | null = null;
      private utterance: Utterance | null = null;
      private paused = false;
      constructor(private url: string) {}
      async play() {
        this.paused = false;
        const blob = audioBlobs.get(this.url);
        if (!blob) throw new Error("Synthetic audio was disposed");
        const body = JSON.parse(await blob.text()) as {
          text: string;
        };
        if (this.paused) return;
        const utterance: Utterance = {
          text: body.text,
          lang: "en-GB",
          rate: this.playbackRate,
          onend: () => this.onended?.(),
          onerror: () => this.onerror?.(),
        };
        this.utterance = utterance;
        active = utterance;
        capture.calls.push({
          text: utterance.text,
          lang: utterance.lang,
          rate: utterance.rate,
        });
      }
      pause() {
        this.paused = true;
        if (active && active === this.utterance) {
          capture.cancellations++;
          active = null;
        }
      }
    }
    Object.defineProperty(window, "Audio", {
      configurable: true,
      value: AuthoredAudio,
    });
    window.__pronunciationTest = capture;
    Object.defineProperty(window, "speechSynthesis", {
      configurable: true,
      value: speechMode === "unsupported" ? undefined : synthesis,
    });
    Object.defineProperty(window, "SpeechSynthesisUtterance", {
      configurable: true,
      value: speechMode === "unsupported" ? undefined : DeviceUtterance,
    });
  }, mode);
}

for (const theme of ["light", "dark"] as const) {
  test(`word pronunciation plays canonical content on search, discovery and saved detail in ${theme} mode`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    if (!baseURL) throw new Error("A test app URL is required");
    await installSpeech(page);
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await page.goto("/vocabulary/pour");
    const main = page.getByRole("main");
    await expect(
      main.getByRole("link", { name: "Back to word search" }),
    ).toHaveAttribute("href", "/vocabulary");
    const saveRequest = page.waitForRequest(
      (request) =>
        request.method() === "POST" &&
        new URL(request.url()).pathname === "/api/v1/user-words",
    );
    await main.getByRole("button", { name: /^Save pour:/ }).click();
    expect((await saveRequest).postDataJSON()).toEqual({
      meaningId: "mean-pour",
      source: "search",
    });
    await expect(
      main.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();
    await openOptionalWriting(page);
    await expect(
      main.getByRole("textbox", { name: "Write a sentence using pour" }),
    ).toBeVisible();

    for (const [surface, route] of [
      ["search", "/vocabulary/pour"],
      ["discovery", "/discover/ordering-at-a-cafe/pour"],
      ["saved", "/words/uw-mean-pour"],
    ]) {
      await page.goto(route!);
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await expect(
        main.getByRole("heading", { name: "pour", level: 1 }),
      ).toBeVisible();
      expect(
        await page.evaluate(() => window.__pronunciationTest.calls),
      ).toEqual([]);
      const word = main.getByRole("group", {
        name: "Pronunciation of pour",
        exact: true,
      });
      await expect(
        word.getByText("English pronunciation", { exact: true }),
      ).toBeVisible();
      const slow = word.getByRole("button", {
        name: "Slow pronunciation of pour",
      });
      if (surface === "search") {
        await expect(slow).toHaveAttribute("aria-pressed", "false");
        await slow.press("Space");
      }
      await expect(slow).toHaveAttribute("aria-pressed", "true");
      const listen = word.getByRole("button", {
        name: "Listen to pour",
        exact: true,
      });
      await listen.press("Enter");
      await expect(
        word.getByRole("button", { name: "Stop listening to pour" }),
      ).toBeVisible();
      await expect(slow).toBeDisabled();
      // Stop is available while loading too. Wait for actual native playback
      // after the asynchronous natural-audio fallback, rather than that label.
      await expect
        .poll(() => page.evaluate(() => window.__pronunciationTest.calls))
        .toEqual([{ text: "pour", lang: "en-GB", rate: 0.7 }]);

      const example = main.getByRole("button", {
        name: "Listen to example: Could you pour me a cup of coffee?",
        exact: true,
      });
      await example.click();
      await expect(listen).toBeVisible();
      expect(
        await page.evaluate(() => window.__pronunciationTest.cancellations),
      ).toBe(1);
      await expect
        .poll(() =>
          page.evaluate(() => window.__pronunciationTest.calls.at(-1)),
        )
        .toEqual({
          text: "Could you pour me a cup of coffee?",
          lang: "en-GB",
          rate: 0.7,
        });
      await page.evaluate(() => window.__pronunciationTest.finish());
      await expect(example).toBeVisible();
      await listen.click();
      await expect
        .poll(() =>
          page.evaluate(() => window.__pronunciationTest.calls.length),
        )
        .toBe(3);
      const stop = word.getByRole("button", { name: "Stop listening to pour" });
      const box = await stop.boundingBox();
      expect(box!.height).toBeGreaterThanOrEqual(44);
      expect(box!.width).toBeGreaterThanOrEqual(44);
      await stop.click();
      await expect(listen).toBeVisible();
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      await main.screenshot({
        path: testInfo.outputPath(`pronunciation-${surface}-${theme}.png`),
      });
      if (surface === "discovery") {
        await listen.click();
        await expect
          .poll(() =>
            page.evaluate(() => window.__pronunciationTest.calls.length),
          )
          .toBe(4);
        const cancellations = await page.evaluate(
          () => window.__pronunciationTest.cancellations,
        );
        await main
          .getByRole("link", { name: "Back to Journey", exact: true })
          .click();
        await expect(page).toHaveURL(/\/discover\/ordering-at-a-cafe$/);
        expect(
          await page.evaluate(() => window.__pronunciationTest.cancellations),
        ).toBe(cancellations + 1);
      }
    }
  });
}

test("learner pace survives reload and meaning audio stays contextual despite a saved device voice", async ({
  page,
}) => {
  await installSpeech(page, "enhanced");
  await page.goto("/vocabulary/pour");
  const word = page.getByRole("group", {
    name: "Pronunciation of pour",
    exact: true,
  });
  const listen = word.getByRole("button", {
    name: "Listen to pour",
    exact: true,
  });
  await listen.click();
  await expect(
    word.getByRole("button", { name: "Stop listening to pour", exact: true }),
  ).toBeVisible();
  await expect
    .poll(() =>
      page.evaluate(() => window.__pronunciationTest.selectedVoiceURI),
    )
    .toBeUndefined();
  await expect
    .poll(() =>
      page.evaluate(() => window.__pronunciationTest.calls.at(-1)?.rate),
    )
    .toBe(0.85);
  await page.evaluate(() => window.__pronunciationTest.finish());
  await word.getByText("Voice & speed", { exact: true }).click();
  await word
    .getByLabel("English voice", { exact: true })
    .selectOption("english-uk");
  await word.getByLabel("Playback speed", { exact: true }).selectOption("0.7");
  await listen.click();
  await expect(
    word.getByRole("button", { name: "Stop listening to pour", exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(() => window.__pronunciationTest.selectedVoiceURI),
  ).toBeUndefined();
  await expect
    .poll(() =>
      page.evaluate(() => window.__pronunciationTest.calls.at(-1)?.rate),
    )
    .toBe(0.7);
  await page.evaluate(() => window.__pronunciationTest.finish());
  await page.reload();
  await word.getByText("Voice & speed", { exact: true }).click();
  await expect(word.getByLabel("English voice", { exact: true })).toHaveValue(
    "english-uk",
  );
  await expect(word.getByLabel("Playback speed", { exact: true })).toHaveValue(
    "0.7",
  );
  await listen.click();
  await expect(
    word.getByRole("button", { name: "Stop listening to pour", exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(() => window.__pronunciationTest.selectedVoiceURI),
  ).toBeUndefined();
  await expect
    .poll(() =>
      page.evaluate(() => window.__pronunciationTest.calls.at(-1)?.rate),
    )
    .toBe(0.7);
});

test("an unavailable saved device voice does not block contextual audio and can still be cleared", async ({
  page,
  context,
  baseURL,
}) => {
  await installSpeech(page, "no-english");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "reset-voice-csrf", url: baseURL! },
  ]);
  const requests: { text: string; csrf: string | undefined }[] = [];
  await page.route("**/api/v1/pronunciation", async (route) => {
    requests.push({
      text: route.request().postDataJSON().text,
      csrf: route.request().headers()["x-csrf-token"],
    });
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        audioBase64: "aWQz",
        mimeType: "audio/mpeg",
        voice: "luna",
        meaningId: route.request().postDataJSON().meaningId,
        cacheHit: false,
      }),
    });
  });
  await page.addInitScript(() => {
    if (localStorage.getItem("vocanova.speech.v1") === null)
      localStorage.setItem(
        "vocanova.speech.v1",
        JSON.stringify({ rate: 0.7, voiceURI: "previous-device-voice" }),
      );
    const state = { rates: [] as number[] };
    Object.defineProperty(window, "__resetVoiceAudioTest", { value: state });
    class AudioFixture {
      playbackRate = 1;
      preservesPitch = false;
      onended: (() => void) | null = null;
      onerror: (() => void) | null = null;
      preload = "auto";
      async play() {
        state.rates.push(this.playbackRate);
      }
      pause() {}
    }
    Object.defineProperty(window, "Audio", {
      configurable: true,
      value: AudioFixture,
    });
  });
  await page.goto("/vocabulary/pour");
  const word = page.getByRole("group", {
    name: "Pronunciation of pour",
    exact: true,
  });
  const voice = word.getByLabel("English voice", { exact: true });
  const speed = word.getByLabel("Playback speed", { exact: true });
  await word.getByText("Voice & speed", { exact: true }).click();
  await expect(voice).toHaveValue("previous-device-voice");
  await expect(voice.locator("option:checked")).toHaveText(
    "Previous voice unavailable",
  );
  await word
    .getByRole("button", { name: "Listen to pour", exact: true })
    .click();
  await expect(
    word.getByText("AI pronunciation", { exact: true }),
  ).toBeVisible();
  expect(requests).toEqual([{ text: "pour", csrf: "reset-voice-csrf" }]);
  expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
    [],
  );
  await page.reload();
  await word.getByText("Voice & speed", { exact: true }).click();
  // Missing device voices must not silently overwrite an explicit preference.
  await expect(voice).toHaveValue("previous-device-voice");
  await expect(speed).toHaveValue("0.7");
  await expect(voice).toBeEnabled();
  await voice.selectOption("");
  expect(
    await page.evaluate(() =>
      JSON.parse(localStorage.getItem("vocanova.speech.v1")!),
    ),
  ).toEqual({ rate: 0.7, voiceURI: "" });
  await word
    .getByRole("button", { name: "Listen to pour", exact: true })
    .click();
  await expect(
    word.getByText("AI pronunciation", { exact: true }),
  ).toBeVisible();
  expect(requests).toEqual([
    { text: "pour", csrf: "reset-voice-csrf" },
    { text: "pour", csrf: "reset-voice-csrf" },
  ]);
  expect(
    await page.evaluate(
      () =>
        (window as unknown as { __resetVoiceAudioTest: { rates: number[] } })
          .__resetVoiceAudioTest.rates,
    ),
  ).toEqual([0.7]);
  await expect(speed).toBeDisabled();
  await expect(voice).toBeDisabled();
  await word
    .getByRole("button", { name: "Stop listening to pour", exact: true })
    .click();
  await expect(speed).toBeEnabled();
  await page.reload();
  await word.getByText("Voice & speed", { exact: true }).click();
  await expect(voice).toHaveValue("");
  await expect(speed).toHaveValue("0.7");
  expect(requests).toHaveLength(2);
});

test("natural pronunciation requests only authored text with CSRF and replays at learner pace", async ({
  page,
  context,
  baseURL,
}) => {
  await installSpeech(page);
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "natural-audio-csrf", url: baseURL! },
  ]);
  const requests: {
    text: string;
    meaningId: string;
    csrf: string | undefined;
  }[] = [];
  await page.route("**/api/v1/pronunciation", async (route) => {
    requests.push({
      text: route.request().postDataJSON().text,
      meaningId: route.request().postDataJSON().meaningId,
      csrf: route.request().headers()["x-csrf-token"],
    });
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        audioBase64: "aWQz",
        mimeType: "audio/mpeg",
        voice: "luna",
        meaningId: route.request().postDataJSON().meaningId,
        cacheHit: requests.length > 1,
      }),
    });
  });
  await page.addInitScript(() => {
    const state = {
      plays: [] as { rate: number; preservesPitch: boolean }[],
      paused: 0,
      active: null as null | { onended: (() => void) | null },
    };
    Object.defineProperty(window, "__naturalAudioTest", { value: state });
    class AudioFixture {
      playbackRate = 1;
      preservesPitch = false;
      onended: (() => void) | null = null;
      onerror: (() => void) | null = null;
      preload = "auto";
      async play() {
        state.active = this;
        state.plays.push({
          rate: this.playbackRate,
          preservesPitch: this.preservesPitch,
        });
      }
      pause() {
        state.paused++;
      }
    }
    Object.defineProperty(window, "Audio", {
      configurable: true,
      value: AudioFixture,
    });
  });
  await page.goto("/vocabulary/pour");
  const word = page.getByRole("group", {
    name: "Pronunciation of pour",
    exact: true,
  });
  const listen = word.getByRole("button", {
    name: "Listen to pour",
    exact: true,
  });
  await listen.click();
  await expect(
    word.getByRole("button", { name: "Stop listening to pour", exact: true }),
  ).toBeVisible();
  await expect(
    word.getByText("AI pronunciation", { exact: true }),
  ).toBeVisible();
  expect(requests).toEqual([
    { text: "pour", meaningId: "mean-pour", csrf: "natural-audio-csrf" },
  ]);
  const capture = () =>
    page.evaluate(
      () =>
        (
          window as unknown as {
            __naturalAudioTest: {
              plays: { rate: number; preservesPitch: boolean }[];
            };
          }
        ).__naturalAudioTest.plays,
    );
  expect(await capture()).toEqual([{ rate: 0.85, preservesPitch: true }]);
  expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
    [],
  );
  await page.evaluate(() =>
    (
      window as unknown as {
        __naturalAudioTest: { active: { onended: (() => void) | null } };
      }
    ).__naturalAudioTest.active.onended?.(),
  );
  await expect(listen).toBeVisible();
  await word
    .getByRole("button", { name: "Slow pronunciation of pour", exact: true })
    .click();
  await listen.click();
  await expect(
    word.getByRole("button", { name: "Stop listening to pour", exact: true }),
  ).toBeVisible();
  await expect.poll(capture).toEqual([
    { rate: 0.85, preservesPitch: true },
    { rate: 0.7, preservesPitch: true },
  ]);
  expect(requests).toHaveLength(2);
  await word
    .getByRole("button", { name: "Stop listening to pour", exact: true })
    .click();
});

test("blocked natural audio retries the prepared clip inside the next click", async ({
  page,
  context,
  baseURL,
}) => {
  await installSpeech(page);
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "gesture-audio-csrf", url: baseURL! },
  ]);
  let requests = 0;
  await page.route("**/api/v1/pronunciation", async (route) => {
    requests++;
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        audioBase64: "aWQz",
        mimeType: "audio/mpeg",
        voice: "luna",
        meaningId: route.request().postDataJSON().meaningId,
        cacheHit: false,
      }),
    });
  });
  await page.addInitScript(() => {
    const state = { gesture: false, attempts: [] as boolean[], revoked: 0 };
    Object.defineProperty(window, "__gestureAudioTest", { value: state });
    // Keep permission through the click's complete dispatch, including React's
    // root bubble handler. A capture-listener microtask can run before that
    // handler in a trusted browser click, incorrectly denying synchronous play.
    window.addEventListener(
      "click",
      () => {
        state.gesture = true;
      },
      true,
    );
    window.addEventListener("click", () => {
      state.gesture = false;
    });
    const revoke = URL.revokeObjectURL.bind(URL);
    URL.revokeObjectURL = (url) => {
      state.revoked++;
      revoke(url);
    };
    class GestureAudio {
      playbackRate = 1;
      preservesPitch = false;
      preload = "auto";
      onended: (() => void) | null = null;
      onerror: (() => void) | null = null;
      play() {
        state.attempts.push(state.gesture);
        return state.gesture
          ? Promise.resolve()
          : Promise.reject(
              new DOMException("Gesture required", "NotAllowedError"),
            );
      }
      pause() {}
    }
    Object.defineProperty(window, "Audio", {
      configurable: true,
      value: GestureAudio,
    });
  });
  await page.goto("/vocabulary/pour");
  const word = page.getByRole("group", {
    name: "Pronunciation of pour",
    exact: true,
  });
  const listen = word.getByRole("button", {
    name: "Listen to pour",
    exact: true,
  });
  await listen.click();
  await expect(word.getByRole("status")).toHaveText(
    "Tap Listen to play pronunciation.",
  );
  const capture = () =>
    page.evaluate(
      () =>
        (
          window as unknown as {
            __gestureAudioTest: { attempts: boolean[]; revoked: number };
          }
        ).__gestureAudioTest,
    );
  expect((await capture()).attempts).toEqual([false]);
  expect((await capture()).revoked).toBe(0);
  expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
    [],
  );
  await listen.click();
  await expect(
    word.getByText("AI pronunciation", { exact: true }),
  ).toBeVisible();
  expect((await capture()).attempts).toEqual([false, true]);
  expect(requests).toBe(1);
  await word
    .getByRole("button", { name: "Stop listening to pour", exact: true })
    .click();
  expect((await capture()).revoked).toBe(1);
});

for (const mode of ["unsupported", "no-english"] as const) {
  test(`pronunciation explains ${mode} availability without hiding word content`, async ({
    page,
  }) => {
    await installSpeech(page, mode);
    await page.goto("/vocabulary/pour");
    const main = page.getByRole("main");
    await main
      .getByRole("button", { name: "Listen to pour", exact: true })
      .click();
    const status = main
      .getByRole("group", { name: "Pronunciation of pour", exact: true })
      .getByRole("status");
    await expect(status).toHaveText(
      "Pronunciation for this meaning is unavailable. Try Listen again.",
    );
    await expect(
      main.getByText("Could you pour me a cup of coffee?", { exact: true }),
    ).toBeVisible();
    expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
      [],
    );
  });
}

test("meaning-specific playback errors allow another explicit attempt without guessed device audio", async ({
  page,
}) => {
  await installSpeech(page);
  await page.goto("/vocabulary/pour");
  const word = page.getByRole("group", {
    name: "Pronunciation of pour",
    exact: true,
  });
  const listen = word.getByRole("button", {
    name: "Listen to pour",
    exact: true,
  });
  await listen.click();
  await expect
    .poll(() => page.evaluate(() => window.__pronunciationTest.calls.length))
    .toBe(1);
  await page.evaluate(() => window.__pronunciationTest.fail());
  await expect(word.getByRole("status")).toHaveText(
    "Pronunciation for this meaning is unavailable. Try Listen again.",
  );
  await listen.click();
  await expect(
    word.getByRole("button", { name: "Stop listening to pour" }),
  ).toBeVisible();
  await expect
    .poll(() => page.evaluate(() => window.__pronunciationTest.calls))
    .toHaveLength(2);
});

test("a mismatched meaning response never plays and an explicit retry stays available", async ({
  page,
}) => {
  await installSpeech(page);
  let matching = false;
  await page.route("**/api/v1/pronunciation", async (route) => {
    const body = route.request().postDataJSON();
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        audioBase64: Buffer.from(JSON.stringify({ text: body.text })).toString(
          "base64",
        ),
        mimeType: "audio/mpeg",
        voice: "coral",
        cacheHit: false,
        meaningId: matching ? body.meaningId : "an-unrelated-meaning",
      }),
    });
  });
  await page.goto("/vocabulary/pour");
  const word = page.getByRole("group", {
    name: "Pronunciation of pour",
    exact: true,
  });
  const listen = word.getByRole("button", {
    name: "Listen to pour",
    exact: true,
  });
  await listen.click();
  await expect(word.getByRole("status")).toHaveText(
    "Pronunciation for this meaning is unavailable. Try Listen again.",
  );
  expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
    [],
  );
  matching = true;
  await listen.click();
  await expect
    .poll(() => page.evaluate(() => window.__pronunciationTest.calls.length))
    .toBe(1);
  await expect(
    word.getByText("AI pronunciation", { exact: true }),
  ).toBeVisible();
});

test("same-spelling meanings send separate identities and switching stops the previous clip", async ({
  page,
  context,
  baseURL,
}) => {
  await installSpeech(page);
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "meaning-audio-csrf", url: baseURL! },
    { name: "e2e_teaching", value: "true", url: baseURL! },
  ]);
  const requests: { text: string; meaningId: string }[] = [];
  await page.route("**/api/v1/pronunciation", async (route) => {
    const body = route.request().postDataJSON();
    requests.push(body);
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        audioBase64: Buffer.from(JSON.stringify({ text: body.text })).toString(
          "base64",
        ),
        mimeType: "audio/mpeg",
        voice: "coral",
        cacheHit: false,
        meaningId: body.meaningId,
      }),
    });
  });
  await page.goto("/vocabulary/deadline");
  const work = page.locator("#meaning-b571a540-89bb-533b-be81-9cedf443a64a");
  const study = page.locator("#meaning-f582a137-c77d-58df-942e-884fe336cd64");
  await work
    .getByRole("button", { name: "Listen to deadline", exact: true })
    .click();
  await expect
    .poll(() => page.evaluate(() => window.__pronunciationTest.calls.length))
    .toBe(1);
  await study
    .getByRole("button", { name: "Listen to deadline", exact: true })
    .click();
  await expect
    .poll(() => page.evaluate(() => window.__pronunciationTest.calls.length))
    .toBe(2);
  expect(requests).toEqual([
    { text: "deadline", meaningId: "b571a540-89bb-533b-be81-9cedf443a64a" },
    { text: "deadline", meaningId: "f582a137-c77d-58df-942e-884fe336cd64" },
  ]);
  expect(
    await page.evaluate(() => window.__pronunciationTest.cancellations),
  ).toBe(1);
  await expect(
    work.getByRole("button", { name: "Listen to deadline", exact: true }),
  ).toBeVisible();
  await expect(
    study.getByRole("button", {
      name: "Stop listening to deadline",
      exact: true,
    }),
  ).toBeVisible();
});

test("canonical word detail preserves authentication and missing-word handling", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await page.goto("/vocabulary/a-word-not-in-the-catalogue");
  await expect(page.getByText("This page could not be found.")).toBeVisible();
  await context.addCookies([
    { name: "e2e_unauthenticated", value: "1", url: baseURL },
  ]);
  await page.goto("/vocabulary/pour");
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fvocabulary%2Fpour/);
});

test("discovery automatically speaks each word, supports mute and explicit replay, and cancels on exit", async ({
  page,
  context,
  baseURL,
}) => {
  await installSpeech(page);
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "discovery-audio", url: baseURL! },
  ]);
  await page.goto("/discovery");
  expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
    [],
  );
  await page
    .getByRole("button", { name: "Start discovering", exact: true })
    .click();
  await expect
    .poll(() =>
      page.evaluate(() =>
        window.__pronunciationTest.calls.map((call) => call.text),
      ),
    )
    .toEqual(["the"]);
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeFocused();
  await expect
    .poll(() =>
      page.evaluate(() =>
        window.__pronunciationTest.calls.map((call) => call.text),
      ),
    )
    .toEqual(["the", "be"]);
  const beforeMute = await page.evaluate(
    () => window.__pronunciationTest.cancellations,
  );
  const soundOn = page.getByRole("button", {
    name: "Automatic sound on",
    exact: true,
  });
  const soundOff = page.getByRole("button", {
    name: "Automatic sound off",
    exact: true,
  });
  await expect(soundOn).toHaveAttribute("aria-pressed", "true");
  // Start from the new-word heading and use native keyboard navigation. Audio
  // preferences must not send focus back to the title or replay this word.
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("button", { name: "Listen to be", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(soundOn).toBeFocused();
  const wordUrl = page.url();
  await page.keyboard.press("Space");
  await expect(soundOff).toHaveAttribute("aria-pressed", "false");
  await expect(soundOff).toBeFocused();
  await expect(page).toHaveURL(wordUrl);
  expect(
    await page.evaluate(() => window.__pronunciationTest.cancellations),
  ).toBe(beforeMute + 1);
  await page.keyboard.press("Space");
  await expect(soundOn).toHaveAttribute("aria-pressed", "true");
  await expect(soundOn).toBeFocused();
  await page.keyboard.press("Space");
  await expect(soundOff).toHaveAttribute("aria-pressed", "false");
  await expect(soundOff).toBeFocused();
  expect(
    await page.evaluate(() =>
      window.__pronunciationTest.calls.map((call) => call.text),
    ),
  ).toEqual(["the", "be"]);
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 3", exact: true }),
  ).toBeFocused();
  expect(
    await page.evaluate(() =>
      window.__pronunciationTest.calls.map((call) => call.text),
    ),
  ).toEqual(["the", "be"]);
  await page
    .getByRole("button", { name: "Listen to Map word 3", exact: true })
    .click();
  await expect
    .poll(() =>
      page.evaluate(() => window.__pronunciationTest.calls.at(-1)?.text),
    )
    .toBe("Map word 3");
  const beforeExit = await page.evaluate(
    () => window.__pronunciationTest.cancellations,
  );
  await page
    .getByRole("link", { name: "Back to word map", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Knowledge map", exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(() => window.__pronunciationTest.cancellations),
  ).toBe(beforeExit + 1);
});

test("daily lesson audio stops when muted and on route exit without changing grading", async ({
  page,
  context,
  baseURL,
}) => {
  await installSpeech(page);
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "daily-audio", url: baseURL! },
    { name: "e2e_review_fixture_count", value: "2", url: baseURL! },
  ]);
  await page.goto("/reviews");
  expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
    [],
  );
  await page.getByRole("button", { name: "Start review", exact: true }).click();
  await expect
    .poll(() =>
      page.evaluate(() =>
        window.__pronunciationTest.calls.map((call) => call.text),
      ),
    )
    .toEqual(["Review word 1"]);
  const beforeMute = await page.evaluate(
    () => window.__pronunciationTest.cancellations,
  );
  await expect(
    page.getByRole("button", { name: "Automatic audio on", exact: true }),
  ).toHaveAttribute("aria-pressed", "true");
  await page
    .getByRole("button", { name: "Automatic audio on", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Automatic audio off", exact: true }),
  ).toHaveAttribute("aria-pressed", "false");
  expect(
    await page.evaluate(() => window.__pronunciationTest.cancellations),
  ).toBe(beforeMute + 1);
  await page
    .getByRole("radio", { name: "definition for review word 1", exact: true })
    .check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Review word 2", exact: true }),
  ).toBeFocused();
  expect(
    await page.evaluate(() => window.__pronunciationTest.calls),
  ).toHaveLength(1);
  await page
    .getByRole("button", { name: "Listen to Review word 2", exact: true })
    .click();
  await expect
    .poll(() =>
      page.evaluate(() => window.__pronunciationTest.calls.at(-1)?.text),
    )
    .toBe("Review word 2");
  const beforeExit = await page.evaluate(
    () => window.__pronunciationTest.cancellations,
  );
  await page.getByRole("link", { name: "Back to Home", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Today", exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(() => window.__pronunciationTest.cancellations),
  ).toBe(beforeExit + 1);
});

test("discovery brings an unconfirmed choice and its exact retry into view", async ({
  page,
  context,
  baseURL,
}) => {
  await installSpeech(page, "no-english");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "discovery-visible-retry", url: baseURL! },
  ]);
  await page.goto("/discovery");
  await page
    .getByRole("button", { name: "Start discovering", exact: true })
    .click();
  const requests: { key: string | undefined; body: unknown }[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      new URL(request.url()).pathname === "/api/v1/knowledge-map/decisions"
    )
      requests.push({
        key: request.headers()["idempotency-key"],
        body: request.postDataJSON(),
      });
  });
  await page.route(
    "**/api/v1/knowledge-map/decisions",
    (route) => route.abort("failed"),
    { times: 1 },
  );
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  const alert = page.getByRole("main").getByRole("alert");
  const retry = alert.getByRole("button", {
    name: "Retry choice",
    exact: true,
  });
  await expect(retry).toBeInViewport({ ratio: 1 });
  await expect
    .poll(() =>
      alert.evaluate(
        (element) =>
          element === document.activeElement ||
          element.contains(document.activeElement),
      ),
    )
    .toBe(true);
  await expect(
    page.getByRole("button", { name: "Learn", exact: true }),
  ).toBeDisabled();
  await retry.click();
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeFocused();
  expect(requests).toHaveLength(2);
  expect(requests[0]!.key).toBeTruthy();
  expect(requests[1]).toEqual(requests[0]);
});
