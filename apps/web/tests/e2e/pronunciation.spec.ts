import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

interface SpeechCapture {
  calls: { text: string; lang: string; rate: number }[];
  cancellations: number;
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
  mode: "ready" | "unsupported" | "no-english" = "ready",
) {
  await page.addInitScript((speechMode) => {
    type Utterance = {
      text: string;
      lang: string;
      rate: number;
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
              { lang: "en-GB", localService: true, default: false },
            ],
      speak: (utterance: Utterance) => {
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
        word.getByText("Device pronunciation", { exact: true }),
      ).toBeVisible();
      const slow = word.getByRole("button", {
        name: "Slow pronunciation of pour",
      });
      await slow.press("Space");
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
      expect(
        await page.evaluate(() => window.__pronunciationTest.calls),
      ).toEqual([{ text: "pour", lang: "en-GB", rate: 0.75 }]);

      const example = main.getByRole("button", {
        name: "Listen to example: Could you pour me a cup of coffee?",
        exact: true,
      });
      await example.click();
      await expect(listen).toBeVisible();
      expect(
        await page.evaluate(() => window.__pronunciationTest.cancellations),
      ).toBe(1);
      expect(
        await page.evaluate(() => window.__pronunciationTest.calls.at(-1)),
      ).toEqual({
        text: "Could you pour me a cup of coffee?",
        lang: "en-GB",
        rate: 1,
      });
      await page.evaluate(() => window.__pronunciationTest.finish());
      await expect(example).toBeVisible();
      await listen.click();
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
      mode === "unsupported"
        ? "Pronunciation is not available in this browser."
        : "An English voice is not ready on this device. Please try again.",
    );
    await expect(
      main.getByText("Could you pour me a cup of coffee?", { exact: true }),
    ).toBeVisible();
    expect(await page.evaluate(() => window.__pronunciationTest.calls)).toEqual(
      [],
    );
  });
}

test("device playback errors allow another explicit attempt", async ({
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
  await page.evaluate(() => window.__pronunciationTest.fail());
  await expect(word.getByRole("status")).toHaveText(
    "Pronunciation could not play. Please try again.",
  );
  await listen.click();
  await expect(
    word.getByRole("button", { name: "Stop listening to pour" }),
  ).toBeVisible();
  expect(
    await page.evaluate(() => window.__pronunciationTest.calls),
  ).toHaveLength(2);
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
