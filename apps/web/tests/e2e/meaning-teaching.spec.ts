import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

const workMeaning = "b571a540-89bb-533b-be81-9cedf443a64a";
const studyMeaning = "f582a137-c77d-58df-942e-884fe336cd64";

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
    { name: "e2e_teaching", value: "true", url: baseURL },
  ]);
});

test("meaning comparison links only real meanings and keeps authored notes aligned", async ({
  page,
}) => {
  await page.goto("/vocabulary/deadline");
  const comparison = page.getByRole("navigation", {
    name: "Compare meanings of deadline",
    exact: true,
  });
  await expect(comparison.getByRole("link")).toHaveCount(2);
  await expect(comparison.getByRole("link").nth(0)).toHaveAttribute(
    "href",
    `#meaning-${workMeaning}`,
  );
  await expect(comparison.getByRole("link").nth(1)).toHaveAttribute(
    "href",
    `#meaning-${studyMeaning}`,
  );
  const work = page.locator(`#meaning-${workMeaning}`);
  const study = page.locator(`#meaning-${studyMeaning}`);
  await work.getByText("Usage tips", { exact: true }).click();
  await study.getByText("Usage tips", { exact: true }).click();
  await expect(
    work.getByText("meet a deadline", { exact: true }),
  ).toBeVisible();
  await expect(
    work.getByText("You meet a deadline when on time and miss it when late.", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    work.getByText("submission deadline", { exact: true }),
  ).toHaveCount(0);
  await expect(
    study.getByText("submission deadline", { exact: true }),
  ).toBeVisible();
  await expect(
    study.getByText(
      "Use by before a deadline time, not until when stating the required submission time.",
      { exact: true },
    ),
  ).toBeVisible();
  await expect(study.getByText("meet a deadline", { exact: true })).toHaveCount(
    0,
  );
  await comparison.getByRole("link").nth(1).focus();
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(new RegExp(`#meaning-${studyMeaning}$`));
  await expect(study).toBeInViewport();
  await page.goto("/vocabulary/invite");
  await expect(
    page.getByRole("navigation", { name: /Compare meanings/ }),
  ).toHaveCount(0);
});

for (const theme of ["light", "dark"] as const) {
  test(`authored combinations can be heard in discovery and saved detail in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    if (!baseURL) throw new Error("Missing app URL");
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await page.addInitScript(() => {
      type Utterance = {
        text: string;
        onstart?: () => void;
        onend?: () => void;
      };
      const capture: string[] = [];
      (window as unknown as { __teachingSpeech: string[] }).__teachingSpeech =
        capture;
      const synth = {
        speaking: false,
        pending: false,
        getVoices: () => [{ lang: "en-GB", localService: true, default: true }],
        speak: (utterance: Utterance) => {
          capture.push(utterance.text);
          utterance.onstart?.();
          utterance.onend?.();
        },
        cancel: () => {},
      };
      class DeviceUtterance {
        text: string;
        constructor(text: string) {
          this.text = text;
        }
      }
      Object.defineProperty(window, "speechSynthesis", {
        configurable: true,
        value: synth,
      });
      Object.defineProperty(window, "SpeechSynthesisUtterance", {
        configurable: true,
        value: DeviceUtterance,
      });
    });
    await page.goto("/discover/daily-conversation/sounds-good");
    const main = page.getByRole("main");
    await main.getByRole("button", { name: /^Save sounds good:/ }).click();
    await expect(
      main.getByRole("button", {
        name: "Remove sounds good from saved words",
        exact: true,
      }),
    ).toBeVisible();
    for (const route of ["/vocabulary/sounds-good", "/words"]) {
      await page.goto(route);
      if (route === "/words")
        await main
          .getByRole("link", {
            name: "Open sounds good details and sentence practice",
            exact: true,
          })
          .click();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await main.getByText("Usage tips", { exact: true }).click();
      const phrase = main.getByRole("group", {
        name: "Pronunciation of combination: That sounds good; sounds good to me",
        exact: true,
      });
      await expect(phrase).toBeVisible();
      const slow = phrase.getByRole("button", {
        name: "Slow pronunciation of combination: That sounds good; sounds good to me",
        exact: true,
      });
      await slow.focus();
      await page.keyboard.press("Space");
      await expect(slow).toHaveAttribute("aria-pressed", "true");
      await phrase
        .getByRole("button", {
          name: "Listen to combination: That sounds good; sounds good to me",
          exact: true,
        })
        .focus();
      await page.keyboard.press("Enter");
      expect(
        await page.evaluate(
          () =>
            (window as unknown as { __teachingSpeech: string[] })
              .__teachingSpeech,
        ),
      ).toEqual(["That sounds good; sounds good to me"]);
      await expect(
        main.getByText(
          "Use sounds good for a singular idea: That sounds good. With plural plans, use sound: Those plans sound good.",
          { exact: true },
        ),
      ).toBeVisible();
      // Scan from a stable viewport after keyboard navigation has scrolled the page.
      await page.evaluate(() => window.scrollTo(0, 0));
      const scan = await scanForAxeViolations(page);
      expect(
        scan.criticalOrSerious,
        formatViolations(scan.criticalOrSerious).join("\n"),
      ).toEqual([]);
      expect(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= window.innerWidth,
        ),
      ).toBe(true);
      await expect(
        main.getByRole("navigation", { name: /Compare meanings/ }),
      ).toHaveCount(0);
    }
  });
}

test("unavailable phrase audio keeps the authored guidance readable", async ({
  page,
}) => {
  await page.addInitScript(() => {
    Object.defineProperty(window, "speechSynthesis", {
      configurable: true,
      value: undefined,
    });
    Object.defineProperty(window, "SpeechSynthesisUtterance", {
      configurable: true,
      value: undefined,
    });
  });
  await page.goto("/vocabulary/deadline");
  const meaning = page.locator(`#meaning-${workMeaning}`);
  await meaning.getByText("Usage tips", { exact: true }).click();
  await meaning
    .getByRole("button", {
      name: "Listen to combination: meet a deadline",
      exact: true,
    })
    .click();
  await expect(meaning.getByRole("status")).toContainText("not available");
  await expect(
    meaning.getByText("meet a deadline", { exact: true }),
  ).toBeVisible();
  await expect(
    meaning.getByText("The deadline is Friday afternoon.", { exact: true }),
  ).toBeVisible();
  await expect(
    meaning.getByRole("button", { name: /^Save deadline:/ }),
  ).toBeVisible();
});
