import { randomUUID } from "node:crypto";
import { test, expect, type Page } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";
test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_daily_conversation", value: "true", url: baseURL! },
  ]);
});

// At the end of the writing step the meaning chooser can rest under the sticky
// header with only a sliver below it, and how much shows depends on font
// metrics. Focus must bring its whole target clear of the header.
async function expectChooserRevealedByFocus(page: Page) {
  const chooser = page.getByText("Change meaning: invite", { exact: true });
  await chooser.focus();
  const [header, box] = await Promise.all([
    page.getByRole("banner").boundingBox(),
    chooser.boundingBox(),
  ]);
  expect(box!.height).toBeGreaterThanOrEqual(44);
  expect(box!.y).toBeGreaterThanOrEqual(header!.y + header!.height);
  expect(
    await chooser.evaluate((summary) => {
      const { x, y, width, height } = summary.getBoundingClientRect();
      return summary.contains(
        document.elementFromPoint(x + width / 2, y + height / 2),
      );
    }),
  ).toBe(true);
}

for (const theme of ["light", "dark"])
  test(`topic writing selects exact meaning and retains draft on failure in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.goto("/writing");
    await page
      .getByRole("navigation", { name: "Writing situations" })
      .getByRole("link", { name: "Daily Conversation", exact: true })
      .click();
    await expect(
      page.getByText("A plan with a friend has changed.", { exact: false }),
    ).toBeVisible();
    await page.getByRole("link", { name: "invite", exact: true }).click();
    const saves: string[] = [];
    page.on("request", (request) => {
      if (
        request.method() === "POST" &&
        new URL(request.url()).pathname === "/api/v1/user-words"
      )
        saves.push(request.url());
    });
    await expect(
      page.getByRole("button", { name: "Save and write", exact: true }),
    ).toBeVisible();
    const selectedURL = page.url();
    expect(saves).toEqual([]);
    await page
      .getByRole("button", { name: "Save and write", exact: true })
      .click();
    await expect(
      page.getByRole("textbox", { name: "Write a sentence using invite" }),
    ).toBeVisible();
    expect(saves).toHaveLength(1);
    expect(page.url()).toBe(selectedURL);
    await expect(
      page
        .getByRole("main")
        .locator("details:visible")
        .filter({
          has: page.locator("summary").filter({ hasText: "Change topic:" }),
        }),
    ).not.toHaveAttribute("open");
    await expect(
      page
        .getByRole("main")
        .locator("details:visible")
        .filter({
          has: page.locator("summary").filter({ hasText: "Change meaning:" }),
        }),
    ).not.toHaveAttribute("open");
    const draft = page.getByRole("textbox", {
      name: "Write a sentence using invite",
    });
    await draft.fill("I would like to invite you to dinner on Friday.");
    let lost = false;
    const intents: unknown[] = [];
    await page.route("**/api/v1/learner-sentences", async (route) => {
      intents.push([
        route.request().postDataJSON(),
        route.request().headers()["idempotency-key"],
      ]);
      if (!lost) {
        lost = true;
        await route.abort("failed");
        return;
      }
      await route.continue();
    });
    await page
      .getByRole("button", { name: "Check my sentence", exact: true })
      .click();
    await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
    await expect(draft).toHaveValue(
      "I would like to invite you to dinner on Friday.",
    );
    await page.reload();
    await expect(draft).toHaveValue(
      "I would like to invite you to dinner on Friday.",
    );
    await expectChooserRevealedByFocus(page);
    // Audit the restored page from a stable position, as dictionary-search
    // does; the chooser's actual target is checked above.
    await page.evaluate(() => window.scrollTo(0, 0));
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page
      .getByRole("button", { name: "Check my sentence", exact: true })
      .focus();
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("button", { name: "Try another sentence", exact: true }),
    ).toBeVisible();
    expect(intents[1]).toEqual(intents[0]);
    await page
      .getByRole("link", { name: "Your writing history", exact: true })
      .click();
    await expect(
      page
        .getByText("I would like to invite you to dinner on Friday.", {
          exact: true,
        })
        .first(),
    ).toBeVisible();
  });
test("signed-out writing does not appear as empty vocabulary", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_unauthenticated", value: "1", url: baseURL! },
  ]);
  await page.goto("/writing?situation=daily-conversation");
  await expect(page).toHaveURL(/\/login\?/);
});
