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
  // Scope to the active page, excluding inactive streamed copies.
  const chooser = page.getByRole("main").getByText("Change meaning: invite", { exact: true });
  await expect(chooser).toHaveCount(1);
  await expect(chooser).toBeVisible();
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
test("Save and write preserves keyboard focus through a safe retry", async ({
  page,
}) => {
  await page.goto("/writing?situation=daily-conversation");
  const main = page.getByRole("main");
  await main.getByRole("link", { name: "invite", exact: true }).click();
  const selected = main.getByRole("region", {
    name: "Writing with your selected meaning",
  });
  const saves: Array<{ body: unknown; key: string | undefined }> = [];
  let releaseFailure!: () => void;
  const delayedFailure = new Promise<void>((resolve) => {
    releaseFailure = resolve;
  });
  await page.route("**/api/v1/user-words", async (route) => {
    const request = route.request();
    if (request.method() !== "POST") {
      await route.continue();
      return;
    }
    saves.push({
      body: request.postDataJSON(),
      key: request.headers()["idempotency-key"],
    });
    if (saves.length === 1) {
      await delayedFailure;
      await route.abort("failed");
      return;
    }
    await route.continue();
  });

  try {
    await selected
      .getByRole("button", { name: "Save and write", exact: true })
      .press("Enter");
    await expect.poll(() => saves.length).toBe(1);
    await expect
      .soft(selected.getByRole("button", { name: /sav(e|ing)/i }))
      .toBeFocused();
    await expect(selected.getByRole("status")).toContainText(/sav|confirm/i);
  } finally {
    releaseFailure();
  }

  await expect(selected.getByRole("alert")).toBeVisible();
  const retry = selected.getByRole("button", { name: /retry.*sav|try.*sav/i });
  await expect.soft(retry).toBeFocused();
  await retry.press("Enter");
  await expect(
    selected.getByRole("textbox", { name: "Write a sentence using invite" }),
  ).toBeFocused();
  expect(saves).toHaveLength(2);
  expect(saves[0]?.key).toBeTruthy();
  expect(saves[1]).toEqual(saves[0]);
});

test("a save conflict keeps keyboard focus while the learner checks saved status", async ({
  page,
}) => {
  await page.goto("/writing?situation=daily-conversation");
  const main = page.getByRole("main");
  await main.getByRole("link", { name: "invite", exact: true }).click();
  const selected = main.getByRole("region", {
    name: "Writing with your selected meaning",
  });
  const save = selected.getByRole("button", { name: "Save and write" });
  const check = selected.getByRole("button", { name: "Check saved status" });
  await save.focus();

  let saves = 0;
  let reads = 0;
  let releaseRead!: () => void;
  const delayedRead = new Promise<void>((resolve) => {
    releaseRead = resolve;
  });
  await page.route("**/api/v1/user-words", async (route) => {
    if (route.request().method() !== "POST") return route.continue();
    saves += 1;
    return route.fulfill({
      status: 409,
      json: { detail: "The save request conflicts with current state." },
      headers: {
        "access-control-allow-origin": new URL(page.url()).origin,
        "access-control-allow-credentials": "true",
      },
    });
  });
  await page.route("**/api/v1/canonical-words/invite", async (route) => {
    reads += 1;
    if (reads === 1) return route.abort("failed");
    await delayedRead;
    return route.continue();
  });

  try {
    await page.keyboard.press("Enter");
    await expect(selected.getByRole("alert")).toBeVisible();
    await expect(check).toBeFocused();
    expect(reads).toBe(1);

    await page.keyboard.press("Enter");
    await expect.poll(() => reads).toBe(2);
    await expect(check).toHaveAttribute("aria-busy", "true");
    await expect(check).toBeFocused();
  } finally {
    releaseRead();
  }

  await expect(save).toBeFocused();
  await expect(selected.getByRole("alert")).toHaveCount(0);
  expect(saves).toBe(1);
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
