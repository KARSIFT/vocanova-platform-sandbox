import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";

// Catch a return to an inverted blue reading panel or a white next-action
// button: learning copy belongs on a calm surface and the action carries blue.
async function expectCalmNextStep(panel: Locator, action: Locator) {
  const surface = await panel.evaluate((element) => {
    const style = getComputedStyle(element);
    return { background: style.backgroundColor, color: style.color };
  });
  const button = await action.evaluate((element) => {
    const style = getComputedStyle(element);
    return { background: style.backgroundColor, color: style.color };
  });
  const rgb = (color: string) =>
    color
      .match(/[\d.]+/g)!
      .slice(0, 3)
      .map(Number) as [number, number, number];
  const luminance = (color: string) =>
    rgb(color).reduce((sum, value, index) => {
      const channel = value / 255;
      const linear =
        channel <= 0.04045
          ? channel / 12.92
          : ((channel + 0.055) / 1.055) ** 2.4;
      return sum + linear * [0.2126, 0.7152, 0.0722][index]!;
    }, 0);
  const contrast = (foreground: string, background: string) => {
    const foregroundLight = luminance(foreground);
    const backgroundLight = luminance(background);
    return (
      (Math.max(foregroundLight, backgroundLight) + 0.05) /
      (Math.min(foregroundLight, backgroundLight) + 0.05)
    );
  };
  const surfaceChannels = rgb(surface.background);
  expect(
    Math.max(...surfaceChannels) - Math.min(...surfaceChannels),
  ).toBeLessThan(35);
  const [red, , blue] = rgb(button.background);
  expect(blue - red).toBeGreaterThan(60);
  expect(contrast(surface.color, surface.background)).toBeGreaterThanOrEqual(
    4.5,
  );
  expect(contrast(button.color, button.background)).toBeGreaterThanOrEqual(4.5);
}

async function expectFits(page: Page, action: Locator) {
  await expect(action).toBeInViewport();
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
  expect((await action.boundingBox())!.height).toBeGreaterThanOrEqual(44);
}

// Include the streamed secondary practice action before checking every visible
// Home action. This catches competing fills even without a primary class name.
async function expectOneFilledAction(page: Page, nextAction: Locator) {
  await expect(
    page
      .getByRole("region", { name: "Your learning path", exact: true })
      .getByRole("button", { name: "Start practice lesson", exact: true }),
  ).toBeVisible();
  const filledActions = await nextAction.evaluate((expected) => {
    const actions = Array.from(
      document.querySelectorAll("main a, main button"),
    );
    const filled = actions.filter((action) => {
      const style = getComputedStyle(action);
      if (style.visibility === "hidden" || action.getClientRects().length === 0)
        return false;
      const channels =
        style.backgroundColor.match(/[\d.]+/g)?.map(Number) ?? [];
      const [red = 0, , blue = 0, alpha = 1] = channels;
      return alpha > 0.5 && blue - red > 60;
    });
    return {
      labels: filled.map((action) => action.textContent?.trim()),
      isNextAction: filled.length === 1 && filled[0] === expected,
    };
  });
  expect(
    filledActions.labels,
    "Only the next activity has a blue fill",
  ).toHaveLength(1);
  expect(filledActions.isNextAction).toBe(true);
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`Warm Home guidance (${theme})`, () => {
    test.beforeEach(async ({ context, baseURL }) => {
      await context.addCookies([
        {
          name: "vocanova_session",
          value: `warm-home-${randomUUID()}`,
          url: baseURL!,
        },
        { name: "vocanova_csrf", value: "warm-home", url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
    });

    test("a first word has concise guidance and one blue discovery action", async ({
      page,
    }) => {
      await page.goto("/home");
      const panel = page.getByRole("region", {
        name: "Review words",
        exact: true,
      });
      const action = panel.getByRole("link", {
        name: "Discover words",
        exact: true,
      });
      await expect(
        page.getByRole("heading", { name: "Today", level: 1, exact: true }),
      ).toBeVisible();
      await expect(
        panel.getByRole("heading", { name: "Discover new words", exact: true }),
      ).toBeVisible();
      await expect(panel.getByText(/Choose a useful word\./)).toBeVisible();
      await expectOneFilledAction(page, action);
      await expectCalmNextStep(panel, action);
      await expectFits(page, action);
      await expect(action).toHaveAttribute("href", "/discovery");
      await action.focus();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(/\/discovery(?:\?|$)/);
    });

    test("ready words start the real lesson and the active review remains primary", async ({
      page,
      context,
      baseURL,
    }) => {
      await context.addCookies([
        { name: "e2e_review_fixture_count", value: "3", url: baseURL! },
      ]);
      await page.goto("/home");
      const panel = page.getByRole("region", {
        name: "Review words",
        exact: true,
      });
      const start = panel.getByRole("button", {
        name: "Start review",
        exact: true,
      });
      await expect(
        panel.getByText("3 words are ready to remember.", { exact: true }),
      ).toBeVisible();
      await expectCalmNextStep(panel, start);
      await expectOneFilledAction(page, start);
      await expectFits(page, start);
      const started = page.waitForResponse(
        (response) =>
          response.url().endsWith("/api/v1/daily-lessons") &&
          response.request().method() === "POST",
      );
      await start.click();
      expect((await started).ok()).toBe(true);
      await expect(page).toHaveURL(/\/review\/session\?lesson=/);
      await page.goto("/home");
      const resume = panel.getByRole("link", {
        name: "Continue review",
        exact: true,
      });
      await expect(resume).toHaveAttribute("href", "/review");
      await expect(
        panel.getByText("Review in progress", { exact: true }),
      ).toBeVisible();
      await expect(
        panel.getByRole("link", { name: "Discover words", exact: true }),
      ).toHaveCount(0);
      await expectCalmNextStep(panel, resume);
      await expectOneFilledAction(page, resume);
      await expectFits(page, resume);
      await resume.click();
      await expect(page).toHaveURL(/\/review(?:\/session\?lesson=.*)?$/);
    });

    test("no due words explains returning without inventing completion or a schedule", async ({
      page,
    }) => {
      await page.goto("/home");
      const panel = page.getByRole("region", {
        name: "Review words",
        exact: true,
      });
      await expect(
        panel.getByText(/Come back to see which words are ready\./),
      ).toBeVisible();
      await expect(
        panel.getByText("Reviewed today: 0", { exact: true }),
      ).toBeVisible();
      await expect(
        panel.getByText(/mastered|review complete|tomorrow/i),
      ).toHaveCount(0);
      const goal = page.getByRole("region", {
        name: "Today’s goal",
        exact: true,
      });
      await expect(
        goal.getByText("In progress", { exact: true }),
      ).toBeVisible();
      await expect(
        goal.locator(".goal-row").filter({ hasText: "Reviews" }).locator("dd"),
      ).toHaveText("0 of 20");
      await expect(
        panel.getByRole("link", { name: "Discover words", exact: true }),
      ).toHaveAttribute("href", "/discovery");
      await expectOneFilledAction(
        page,
        panel.getByRole("link", { name: "Discover words", exact: true }),
      );
    });
  });
}
