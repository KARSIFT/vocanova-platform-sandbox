import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";
import { pressTab } from "./keyboard-navigation";

// Catch an off-screen decision, overlapping actions, or a date that remains
// accessible by name while its actual visible text is clipped.
async function expectReadableControl(
  page: Page,
  control: Locator,
  texts: Locator[],
) {
  await control.focus();
  await pressTab(page);
  await pressTab(page, { backwards: true });
  await expect(control).toBeFocused();
  await expect(control).toBeInViewport({ ratio: 1 });
  const bounds = await control.boundingBox();
  const headerBottom = await page
    .getByRole("banner")
    .evaluate((element) => element.getBoundingClientRect().bottom);
  expect(
    bounds!.y,
    "The complete focused control must clear the sticky header",
  ).toBeGreaterThanOrEqual(headerBottom);
  expect(bounds!.width).toBeGreaterThanOrEqual(44);
  expect(bounds!.height).toBeGreaterThanOrEqual(44);
  for (const text of texts) {
    await expect(text).toBeVisible();
    const visible = await text.evaluate((element) => {
      const button = element.closest("button, a")!;
      const buttonRect = button.getBoundingClientRect();
      const ownRect = element.getBoundingClientRect();
      const range = document.createRange();
      range.selectNodeContents(element);
      const rects = Array.from(range.getClientRects()).filter(
        (rect) => rect.width > 0 && rect.height > 0,
      );
      return (
        rects.length > 0 &&
        rects.every((rect) => {
          const hit = document.elementFromPoint(
            rect.left + rect.width / 2,
            rect.top + rect.height / 2,
          );
          return (
            rect.left >= Math.max(0, buttonRect.left, ownRect.left) - 1 &&
            rect.right <=
              Math.min(innerWidth, buttonRect.right, ownRect.right) + 1 &&
            rect.top >= Math.max(0, buttonRect.top, ownRect.top) - 1 &&
            rect.bottom <=
              Math.min(innerHeight, buttonRect.bottom, ownRect.bottom) + 1 &&
            hit !== null &&
            button.contains(hit)
          );
        })
      );
    });
    expect(
      visible,
      `The full text of ${await text.textContent()} must be readable and unobscured`,
    ).toBe(true);
  }
  expect(
    await control.evaluate((element) => {
      const style = getComputedStyle(element);
      // Fractional device scales can rasterize authored 2px outlines to fewer
      // CSS pixels (WebKit at 2.75x reports 5 physical pixels as 1.818182px).
      const physicalWidth = Math.round(
        parseFloat(style.outlineWidth) * devicePixelRatio,
      );
      return (
        element.matches(":focus-visible") &&
        style.outlineStyle !== "none" &&
        physicalWidth >= Math.floor(2 * devicePixelRatio)
      );
    }),
  ).toBe(true);
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
}

for (const theme of ["light", "dark"] as const) {
  for (const textSize of [100, 200]) {
    test(`discovery decisions stay readable at ${textSize}% text in ${theme}`, async ({
      page,
      context,
      baseURL,
    }, info) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      // Avoid audio in local geometry checks; no playback is part of this task.
      await page.route("**/api/v1/pronunciation", (route) =>
        route.fulfill({ status: 503, json: { title: "Unavailable" } }),
      );
      await page.goto("/discovery?band=1&word=2&active=1");
      await expect(
        page.getByRole("heading", { name: "be", exact: true }),
      ).toBeVisible();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page.evaluate((size) => {
        document.documentElement.style.fontSize = `${size}%`;
      }, textSize);
      for (const name of ["Learn", "Already know"]) {
        const action = page.getByRole("button", { name, exact: true });
        await expectReadableControl(page, action, [action]);
      }
      await page.screenshot({
        path: info.outputPath(`discovery-${theme}-${textSize}.png`),
      });
    });

    test(`review actions retain the full schedule at ${textSize}% text in ${theme}`, async ({
      page,
      context,
      baseURL,
    }, info) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
        { name: "e2e_review_fixture_count", value: "1", url: baseURL! },
        {
          name: "e2e_daily_review_at",
          value: "2026-10-18T22:30:00.000Z",
          url: baseURL!,
        },
      ]);
      await page.route("**/api/v1/pronunciation", (route) =>
        route.fulfill({ status: 503, json: { title: "Unavailable" } }),
      );
      await page.goto("/review");
      await page
        .getByRole("button", { name: "Start review", exact: true })
        .click();
      await page
        .getByRole("radio", { name: "A different meaning.", exact: true })
        .check();
      await page
        .getByRole("button", { name: "Check answer", exact: true })
        .click();
      await expect(
        page.getByRole("status").filter({ hasText: /^Not quite/ }),
      ).toBeVisible();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page.evaluate((size) => {
        document.documentElement.style.fontSize = `${size}%`;
      }, textSize);
      // The correctness marker must not squeeze the answer into fragments of
      // ordinary words, even when the card itself stays within the viewport.
      const correctText = page
        .getByRole("group", { name: "Meaning of Review word 1", exact: true })
        .getByText("definition for review word 1", { exact: true });
      expect(
        await correctText.evaluate((element) => {
          const node = element.firstChild!;
          const word = document.createRange();
          word.setStart(node, 0);
          word.setEnd(node, "definition".length);
          return word.getClientRects().length;
        }),
        "The first answer word must remain readable without forced splitting",
      ).toBe(1);
      const date = new Intl.DateTimeFormat("en", {
        month: "short",
        day: "numeric",
        timeZone: await page.evaluate(
          () => Intl.DateTimeFormat().resolvedOptions().timeZone,
        ),
      }).format(new Date("2026-10-18T22:30:00.000Z"));
      const schedule = page.getByRole("button", {
        name: `Show this word again: Automatic · ${date}`,
        exact: true,
      });
      await expectReadableControl(page, schedule, [
        schedule.getByText("Next review", { exact: true }),
        schedule.getByText(date, { exact: true }),
      ]);
      const next = page.getByRole("button", { name: "Continue", exact: true });
      await expectReadableControl(page, next, [next]);
      const [scheduleBounds, continueBounds] = await Promise.all([
        schedule.boundingBox(),
        next.boundingBox(),
      ]);
      expect(
        scheduleBounds!.x + scheduleBounds!.width <= continueBounds!.x + 1 ||
          scheduleBounds!.y + scheduleBounds!.height <= continueBounds!.y + 1,
      ).toBe(true);
      await page.screenshot({
        path: info.outputPath(`review-${theme}-${textSize}.png`),
      });
    });
  }
}

for (const theme of ["light", "dark"] as const) {
  test(`Home preserves a usable reading width at the desktop breakpoint with 200% text in ${theme}`, async ({
    page,
    context,
    baseURL,
  }, info) => {
    test.skip(
      info.project.use.viewport?.width !== 1280,
      "The desktop project covers this 1024px breakpoint.",
    );
    await page.setViewportSize({ width: 1024, height: 900 });
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.goto("/home");
    await page.evaluate(() => {
      document.documentElement.style.fontSize = "200%";
    });
    const panel = page.getByRole("region", {
      name: "Review words",
      exact: true,
    });
    const guidance = panel.getByText(
      "Choose a useful word. Learn its meaning, then try it in your own sentence.",
      { exact: true },
    );
    await expect(guidance).toBeVisible();
    // Leave room for a short phrase per line instead of squeezing the main
    // guidance into the fixed sidebar's leftover width at enlarged text.
    expect(
      await guidance.evaluate(
        (element) =>
          element.getBoundingClientRect().width /
          parseFloat(getComputedStyle(element).fontSize),
      ),
    ).toBeGreaterThanOrEqual(10);
    const next = panel.getByRole("link", {
      name: "Discover words",
      exact: true,
    });
    await expectReadableControl(page, next, [next]);
    await page.screenshot({
      path: info.outputPath(`home-1024-${theme}-200.png`),
    });
  });
}
