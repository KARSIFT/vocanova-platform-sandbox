import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

const editorFor = (page: Page) =>
  page.getByRole("region", { name: "Your knowledge and note", exact: true });

async function openEditor(page: Page) {
  await page.locator("summary").filter({ hasText: "Personal tools" }).click();
  const editor = editorFor(page);
  await editor.getByRole("button", { name: "Edit knowledge and note" }).click();
  await expect(editor.getByRole("textbox", { name: "My note" })).toBeEnabled();
  return editor;
}

for (const theme of ["light", "dark"] as const) {
  test(`personal knowledge and notes persist separately from saved practice in ${theme} mode`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    if (!baseURL) throw new Error("Missing app URL");
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await page.goto("/vocabulary/pour");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    let editor = await openEditor(page);
    const note = "I pour tea for my friends.";
    await editor
      .getByRole("checkbox", { name: "I already know this meaning" })
      .check();
    await editor.getByRole("textbox", { name: "My note" }).fill(note);
    await editor
      .getByRole("button", { name: "Save my changes", exact: true })
      .click();
    await expect(editor.getByRole("status")).toHaveText(
      "Your knowledge and note are saved.",
    );
    await page.reload();
    editor = await openEditor(page);
    await expect(editor.getByRole("checkbox")).toBeChecked();
    await expect(editor.getByRole("textbox", { name: "My note" })).toHaveValue(
      note,
    );
    await expect(
      editor.getByText("This is your own assessment.", { exact: false }),
    ).toBeVisible();
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth,
      ),
    ).toBe(true);
    await page.screenshot({
      path: testInfo.outputPath(`word-knowledge-${theme}.png`),
      fullPage: true,
    });

    await page.goto("/progress");
    let overview = page.getByRole("region", { name: "Your vocabulary map" });
    await overview.locator("summary").click();
    await expect(
      overview.getByText("0 saved · 0 ready for review", {
        exact: true,
      }),
    ).toBeVisible();
    await overview
      .getByRole("link", {
        name: "1 meaning marked already known",
        exact: true,
      })
      .click();
    await expect(page).toHaveURL(/knowledge=known/);
    let tile = page
      .getByRole("region", { name: "Word knowledge map" })
      .getByRole("link")
      .filter({ has: page.getByText("pour", { exact: true }) });
    await expect(tile).toHaveCount(1);
    await expect(tile).toHaveAttribute("href", "/vocabulary/pour");
    await expect(
      tile.getByText("Already known", { exact: true }),
    ).toBeVisible();
    await expect(tile.getByText("Due for review", { exact: true })).toHaveCount(
      0,
    );
    await tile.click();
    await page.getByRole("button", { name: /^Save pour:/ }).click();
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();

    await page.goto("/progress");
    overview = page.getByRole("region", { name: "Your vocabulary map" });
    await overview.locator("summary").click();
    await expect(
      overview.getByText("1 saved · 1 ready for review", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      overview.getByRole("link", { name: "Start review", exact: true }),
    ).toHaveAttribute("href", "/review");
    await overview
      .getByRole("link", {
        name: "1 meaning marked already known",
        exact: true,
      })
      .click();
    tile = page
      .getByRole("region", { name: "Word knowledge map" })
      .getByRole("link")
      .filter({ has: page.getByText("pour", { exact: true }) });
    await expect(
      tile.getByText("Already known", { exact: true }),
    ).toBeVisible();
    await expect(
      tile.getByText("Also saved for practice", { exact: true }),
    ).toBeVisible();
    await expect(
      tile.getByText("Due for review", { exact: true }),
    ).toBeVisible();
    await tile.click();
    editor = await openEditor(page);
    await expect(editor.getByRole("textbox", { name: "My note" })).toHaveValue(
      note,
    );
    await editor.getByRole("checkbox").uncheck();
    await editor.getByRole("textbox", { name: "My note" }).fill("");
    await editor
      .getByRole("button", { name: "Save my changes", exact: true })
      .click();
    await expect(editor.getByRole("status")).toHaveText(
      "Your knowledge and note are saved.",
    );
    await page.reload();
    editor = await openEditor(page);
    await expect(editor.getByRole("checkbox")).not.toBeChecked();
    await expect(editor.getByRole("textbox", { name: "My note" })).toHaveValue(
      "",
    );
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();
    await page.goto("/progress");
    overview = page.getByRole("region", { name: "Your vocabulary map" });
    await overview.locator("summary").click();
    await expect(
      overview.getByRole("link", {
        name: "0 meanings marked already known",
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      overview.getByText("1 saved · 1 ready for review", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      overview.getByRole("link", { name: "Start review", exact: true }),
    ).toHaveAttribute("href", "/review");
  });
}

test("a lost knowledge response retries the exact update before allowing a newer edit", async ({
  page,
}) => {
  await page.goto("/vocabulary/pour");
  const editor = await openEditor(page);
  const original = {
    selfReportedKnown: true,
    note: "I pour tea every morning.",
  };
  const requests: { key: string | undefined; body: string | null }[] = [];
  await page.route("**/api/v1/meaning-knowledge/*", async (route) => {
    if (route.request().method() !== "PUT") return route.continue();
    requests.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postData(),
    });
    if (requests.length === 1) {
      // Apply the update to the fixture, then lose only its response.
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      return route.abort("failed");
    }
    return route.continue();
  });
  await editor.getByRole("checkbox").check();
  await editor.getByRole("textbox", { name: "My note" }).fill(original.note);
  await editor
    .getByRole("button", { name: "Save my changes", exact: true })
    .click();
  await expect(editor.getByRole("alert")).toContainText("could not confirm");
  await expect(editor.getByRole("checkbox")).toBeDisabled();
  await expect(editor.getByRole("textbox", { name: "My note" })).toBeDisabled();
  await expect(editor.getByRole("textbox", { name: "My note" })).toHaveValue(
    original.note,
  );
  expect(requests).toHaveLength(1);
  await editor
    .getByRole("button", { name: "Retry saving changes", exact: true })
    .click();
  await expect(editor.getByRole("status")).toHaveText(
    "Your knowledge and note are saved.",
  );
  await expect(editor.getByRole("textbox", { name: "My note" })).toBeEnabled();
  expect(requests).toHaveLength(2);
  expect(requests[0]!.key).toBeTruthy();
  expect(JSON.parse(requests[0]!.body!)).toEqual(original);
  expect(requests[1]).toEqual(requests[0]);

  const newerNote = "I pour water for my guests.";
  await editor.getByRole("textbox", { name: "My note" }).fill(newerNote);
  await expect(editor.getByRole("status")).toHaveCount(0);
  expect(requests).toHaveLength(2);
  await editor
    .getByRole("button", { name: "Save my changes", exact: true })
    .click();
  await expect(editor.getByRole("status")).toHaveText(
    "Your knowledge and note are saved.",
  );
  expect(requests).toHaveLength(3);
  expect(requests[2]!.key).toBeTruthy();
  expect(requests[2]!.key).not.toBe(requests[0]!.key);
  expect(JSON.parse(requests[2]!.body!)).toEqual({
    ...original,
    note: newerNote,
  });
  await page.reload();
  await openEditor(page);
  await expect(editor.getByRole("checkbox")).toBeChecked();
  await expect(editor.getByRole("textbox", { name: "My note" })).toHaveValue(
    newerNote,
  );
});

test("failed initial knowledge loading cannot replace an existing private note", async ({
  page,
}) => {
  await page.goto("/vocabulary/pour");
  const editor = await openEditor(page);
  const existingNote = "Pour slowly so the cup does not overflow.";
  await editor.getByRole("checkbox").check();
  await editor.getByRole("textbox", { name: "My note" }).fill(existingNote);
  await editor
    .getByRole("button", { name: "Save my changes", exact: true })
    .click();
  await expect(editor.getByRole("status")).toHaveText(
    "Your knowledge and note are saved.",
  );
  await page.reload();
  let reads = 0;
  let writes = 0;
  await page.route("**/api/v1/meaning-knowledge/*", async (route) => {
    if (route.request().method() === "PUT") writes += 1;
    if (route.request().method() === "GET" && ++reads === 1) {
      return route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          code: "unavailable",
          message: "Temporarily unavailable",
        }),
      });
    }
    return route.continue();
  });
  await page.locator("summary").filter({ hasText: "Personal tools" }).click();
  await editor.getByRole("button", { name: "Edit knowledge and note" }).click();
  await expect(editor.getByRole("alert")).toBeVisible();
  await expect(editor.getByRole("textbox", { name: "My note" })).toHaveCount(0);
  await expect(editor.getByRole("checkbox")).toHaveCount(0);
  await expect(
    editor.getByRole("button", { name: "Save my changes", exact: true }),
  ).toHaveCount(0);
  expect(writes).toBe(0);
  await editor
    .getByRole("button", { name: "Try loading again", exact: true })
    .click();
  await expect(editor.getByRole("textbox", { name: "My note" })).toHaveValue(
    existingNote,
  );
  await expect(editor.getByRole("checkbox")).toBeChecked();
  await expect(editor.getByRole("alert")).toHaveCount(0);
  expect(reads).toBe(2);
  expect(writes).toBe(0);
});
