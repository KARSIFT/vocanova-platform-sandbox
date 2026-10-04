import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";
const slugs = [
  "airport",
  "restaurant",
  "hotel-check-in",
  "job-interview",
  "daily-conversation",
  "work-meeting",
  "university-class",
  "shopping",
  "home-and-renting",
  "public-transport",
  "deliveries",
  "everyday-payments",
  "everyday-services",
  "health-appointments",
  "phone-calls",
  "email-and-online-tasks",
  "everyday-work",
];
test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_unit_guides", value: "true", url: baseURL! },
  ]);
});
test("all seventeen situation guides have original phrases, canonical help and writing links", async ({
  page,
}) => {
  test.setTimeout(120_000);
  for (const slug of slugs) {
    await page.goto(`/discover/${slug}`);
    const guide = page.getByRole("region", {
      name: "A quick guide",
      exact: true,
    });
    await expect(guide).toBeVisible();
    await guide.getByText("Useful phrases", { exact: true }).click();
    await expect(
      guide.getByRole("button", { name: /^Play audio: guide phrase/ }),
    ).toHaveCount(3);
    await guide.getByText("Words and examples", { exact: true }).click();
    await expect(guide.getByRole("link").first()).toHaveAttribute(
      "href",
      new RegExp(`/discover/${slug}/[^#]+#meaning-[0-9a-f-]{36}$`),
    );
    await expect(
      guide.getByRole("link", {
        name: "Write about this situation",
        exact: true,
      }),
    ).toHaveAttribute("href", `/writing?situation=${slug}`);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
  }
});
test("guide connects to matching guided lesson and original story", async ({
  page,
}) => {
  await page.goto("/discover/daily-conversation");
  const guide = page.getByRole("region", {
    name: "A quick guide",
    exact: true,
  });
  await expect(
    guide.getByRole("link", { name: /guided lesson$/ }),
  ).toHaveAttribute("href", /^\/learn\//);
  const story = guide.getByRole("link", {
    name: "Story: Plans for Saturday",
    exact: true,
  });
  await expect(story).toHaveAttribute("href", "/stories/plans-for-saturday");
  await story.click();
  await page
    .getByRole("link", { name: "Open the situation guide", exact: true })
    .click();
  await expect(page).toHaveURL(/\/discover\/daily-conversation#unit-guide$/);
  await guide.getByText("Words and examples", { exact: true }).click();
  const meaningLink = guide.getByRole("link").first();
  const href = await meaningLink.getAttribute("href");
  expect(href).toMatch(
    /^\/discover\/daily-conversation\/[^#]+#meaning-[0-9a-f-]{36}$/,
  );
  const anchor = href!.split("#")[1]!;
  await meaningLink.click();
  await expect(page).toHaveURL(new RegExp(`#${anchor}$`));
  await expect(
    page.getByRole("main").locator(`[id="${anchor}"]`),
  ).toBeVisible();
});
test("unavailable secondary catalogs preserve canonical words and guide", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_lessons", value: "unavailable", url: baseURL! },
    { name: "e2e_stories", value: "unavailable", url: baseURL! },
  ]);
  await page.goto("/discover/restaurant");
  await expect(
    page.getByRole("heading", { name: "Restaurant", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("region", { name: "A quick guide", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("list", { name: "Words in this situation", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText(
      "Guided lessons could not load. You can still explore this guide and its words.",
      { exact: true },
    ),
  ).toBeVisible();
  await expect(
    page.getByText(
      "Story links could not load. You can try the story library later.",
      { exact: true },
    ),
  ).toBeVisible();
});

test("guide phrases and examples support both themes and keyboard reading", async ({
  page,
  context,
  baseURL,
}) => {
  for (const theme of ["light", "dark"] as const) {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.goto("/discover/daily-conversation");
    const guide = page.getByRole("region", {
      name: "A quick guide",
      exact: true,
    });
    const phrases = guide
      .locator("summary")
      .filter({ hasText: "Useful phrases" });
    await phrases.focus();
    await page.keyboard.press("Enter");
    await expect(
      guide.getByRole("button", { name: /^Play audio: guide phrase/ }),
    ).toHaveCount(3);
    const words = guide
      .locator("summary")
      .filter({ hasText: "Words and examples" });
    await words.focus();
    await page.keyboard.press("Enter");
    // Scan from a stable viewport after keyboard scrolling near the sticky header.
    await page.evaluate(() => window.scrollTo(0, 0));
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
  }
});
