import { randomUUID } from "node:crypto";
import { expect, test, type BrowserContext, type Page } from "@playwright/test";
import type { WordListDetail } from "@vocanova/api-client";
import {
  assertKeyboardReachable,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
async function writeHeaders(context: BrowserContext) {
  const csrf = (await context.cookies()).find(
    (cookie) => cookie.name === "vocanova_csrf",
  )?.value;
  if (!csrf) throw new Error("Missing fixture CSRF cookie");
  return { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() };
}
async function seedList(
  context: BrowserContext,
  name = "Conversation plans",
  slugs: string[] = ["invite"],
) {
  const id = randomUUID();
  const response = await context.request.put(
    `${apiURL}/api/v1/word-lists/${id}`,
    {
      headers: await writeHeaders(context),
      data: { name, expectedRevision: 0 },
    },
  );
  expect(response.status()).toBe(200);
  let list = (await response.json()) as WordListDetail;
  for (const slug of slugs) {
    const word = await (
      await context.request.get(`${apiURL}/api/v1/canonical-words/${slug}`)
    ).json();
    const added = await context.request.put(
      `${apiURL}/api/v1/word-lists/${id}/members/${word.word.meanings[0].id}`,
      {
        headers: await writeHeaders(context),
        data: { expectedRevision: list.revision },
      },
    );
    expect(added.status()).toBe(200);
    list = await added.json();
  }
  return list;
}
async function openMembership(page: Page, list: WordListDetail) {
  await page.goto("/vocabulary/invite");
  await page
    .getByRole("button", { name: "Choose personal lists", exact: true })
    .click();
  const select = page.getByLabel("Choose a list for this meaning", {
    exact: true,
  });
  await select.selectOption(list.id);
  await expect(
    page.getByRole("button", { name: /meaning (to|from) list/ }),
  ).toBeVisible();
  return select;
}

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

test("create, search, rename and delete a personal list with keyboard controls", async ({
  page,
}) => {
  await page.goto("/lists");
  const main = page.getByRole("main");
  await expect(
    main.getByRole("heading", { name: "Give your words a place" }),
  ).toBeVisible();
  const name = main.getByLabel("List name", { exact: true });
  await name.fill("Travel conversations");
  await name.press("Tab");
  await expect(
    main.getByRole("button", { name: "Create list", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(/\/lists\/[^/]+$/);
  await expect(
    main.getByRole("heading", { name: "Travel conversations", exact: true }),
  ).toBeVisible();
  await expect(
    main.getByRole("heading", { name: "Your list is empty" }),
  ).toBeVisible();
  await main
    .getByLabel("List name", { exact: true })
    .fill("Plans with friends");
  await main.getByRole("button", { name: "Rename list", exact: true }).click();
  await expect(
    main.getByRole("heading", { name: "Plans with friends", exact: true }),
  ).toBeVisible();
  const listPath = new URL(page.url()).pathname;
  await page.reload();
  await expect(
    main.getByRole("heading", { name: "Plans with friends", exact: true }),
  ).toBeVisible();
  await main
    .getByRole("link", { name: "Back to personal lists", exact: true })
    .click();
  await main.getByLabel("Search your lists").fill("travel");
  await expect(
    main.getByRole("heading", { name: "No lists match your search" }),
  ).toBeVisible();
  await main.getByLabel("Search your lists").fill("friends");
  await main
    .getByRole("link", { name: "Plans with friends", exact: true })
    .click();
  await expect(page).toHaveURL(new RegExp(`${listPath}$`));
  await main.getByRole("button", { name: "Delete list", exact: true }).click();
  await main
    .getByRole("button", { name: "Confirm delete list", exact: true })
    .click();
  await expect(page).toHaveURL(/\/lists$/);
  await expect(
    main.getByRole("link", { name: "Plans with friends", exact: true }),
  ).toHaveCount(0);
});

test("an exact lost membership retry shows newer removal and never restores it", async ({
  page,
  context,
}) => {
  const list = await seedList(context, "Conversation plans", []);
  const select = await openMembership(page, list);
  const requests: { body: unknown; key: string | undefined }[] = [];
  await page.route(
    `**/api/v1/word-lists/${list.id}/members/*`,
    async (route) => {
      if (route.request().method() !== "PUT") return route.continue();
      requests.push({
        body: route.request().postDataJSON(),
        key: route.request().headers()["idempotency-key"],
      });
      if (requests.length === 1) {
        const accepted = await route.fetch();
        expect(accepted.status()).toBe(200);
        const current = (await accepted.json()) as WordListDetail;
        const removed = await context.request.delete(
          `${apiURL}/api/v1/word-lists/${list.id}/members/${current.members[0]!.meaningId}?expectedRevision=${current.revision}`,
          { headers: await writeHeaders(context) },
        );
        expect(removed.status()).toBe(200);
        return route.abort("failed");
      }
      return route.continue();
    },
  );
  await page
    .getByRole("button", { name: "Add meaning to list", exact: true })
    .click();
  await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
  await expect(select).toBeDisabled();
  // Failed reload must not unlock an uncertain change or discard its identity.
  await page.route(`**/api/v1/word-lists/${list.id}`, (route) =>
    route.abort("failed"),
  );
  await page
    .getByRole("button", { name: "Load current status", exact: true })
    .click();
  await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
  await expect(select).toBeDisabled();
  await page
    .getByRole("button", { name: "Retry same change", exact: true })
    .click();
  await expect(
    page
      .getByRole("status")
      .filter({ hasText: "This meaning is currently outside the list." }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Add meaning to list", exact: true }),
  ).toBeEnabled();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toEqual(requests[0]);
  const summary = await (
    await context.request.get(`${apiURL}/api/v1/knowledge-summary`)
  ).json();
  expect(summary).toMatchObject({ saved: 0, selfReportedKnown: 0 });
});

test("stale rename requires current status and preserves the newer name", async ({
  page,
  context,
}) => {
  const list = await seedList(context);
  await page.goto(`/lists/${list.id}`);
  const newer = await context.request.put(
    `${apiURL}/api/v1/word-lists/${list.id}`,
    {
      headers: await writeHeaders(context),
      data: { name: "Newer list name", expectedRevision: list.revision },
    },
  );
  expect(newer.status()).toBe(200);
  await page
    .getByLabel("List name", { exact: true })
    .fill("Older browser edit");
  await page.getByRole("button", { name: "Rename list", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "changed elsewhere",
  );
  await expect(page.getByLabel("List name", { exact: true })).toBeDisabled();
  await expect(
    page.getByRole("button", { name: "Retry same change", exact: true }),
  ).toHaveCount(0);
  await page
    .getByRole("button", { name: "Load current status", exact: true })
    .click();
  await expect(page.getByLabel("List name", { exact: true })).toHaveValue(
    "Newer list name",
  );
  await expect(
    page.getByRole("heading", { name: "Newer list name", exact: true }),
  ).toBeVisible();
});

for (const [mode, label] of [
  ["typed_recall", "Start typed recall"],
  ["listening_choice", "Start listening practice"],
] as const) {
  test(`list selection is retained through a lost ${mode} start response`, async ({
    page,
    context,
  }) => {
    const list = await seedList(context);
    await page.goto(`/practice?list=${list.id}`);
    const selection = page.getByLabel("Practice vocabulary", { exact: true });
    await expect(selection).toHaveValue(`list:${list.id}`);
    const requests: { body: unknown; key: string | undefined }[] = [];
    let sessionId = "";
    await page.route("**/api/v1/practice-sessions", async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      requests.push({
        body: route.request().postDataJSON(),
        key: route.request().headers()["idempotency-key"],
      });
      if (requests.length === 1) {
        const accepted = await route.fetch();
        const session = await accepted.json();
        expect(accepted.status()).toBe(200);
        expect(session).toMatchObject({
          listId: list.id,
          listName: list.name,
          listRevision: list.revision,
          totalSteps: 1,
        });
        sessionId = session.id;
        return route.abort("failed");
      }
      return route.continue();
    });
    await page.getByRole("button", { name: label, exact: true }).click();
    await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
    await expect(selection).toBeDisabled();
    await page
      .getByRole("button", { name: "Retry start", exact: true })
      .click();
    await expect(page).toHaveURL(new RegExp(`/practice/session/${sessionId}$`));
    expect(requests).toHaveLength(2);
    expect(requests[1]).toEqual(requests[0]);
    expect(requests[0]!.body).toEqual({
      mode,
      listId: list.id,
      listRevision: list.revision,
    });
  });
}

test("empty and deleted list practice never silently start a full-course mix", async ({
  page,
  context,
}) => {
  const empty = await seedList(context, "Empty list", []);
  await page.goto(`/practice?list=${empty.id}`);
  await expect(
    page.getByRole("button", { name: "Start typed recall", exact: true }),
  ).toBeDisabled();
  await expect(
    page.getByRole("status").filter({ hasText: "no meanings available" }),
  ).toBeVisible();
  await page.goto(`/practice?list=${randomUUID()}`);
  await expect(
    page.getByRole("button", { name: "Start typed recall", exact: true }),
  ).toBeDisabled();
  await expect(
    page.getByRole("status").filter({ hasText: "could not be loaded" }),
  ).toBeVisible();
  const sessions = await (
    await context.request.get(`${apiURL}/api/v1/practice-sessions`)
  ).json();
  expect(sessions.items).toHaveLength(0);
});

test("changed list practice reloads current membership before another start", async ({
  page,
  context,
}) => {
  const list = await seedList(context);
  await page.goto(`/practice?list=${list.id}`);
  const removed = await context.request.delete(
    `${apiURL}/api/v1/word-lists/${list.id}/members/${list.members[0]!.meaningId}?expectedRevision=${list.revision}`,
    { headers: await writeHeaders(context) },
  );
  expect(removed.status()).toBe(200);
  await page
    .getByRole("button", { name: "Start typed recall", exact: true })
    .click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "available practice has changed",
  );
  await page
    .getByRole("button", { name: "Load saved sessions", exact: true })
    .click();
  await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Start typed recall", exact: true }),
  ).toBeDisabled();
  await expect(
    page.getByRole("status").filter({ hasText: "no meanings available" }),
  ).toBeVisible();
});

test("list load failure and deletion elsewhere have clear recovery", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "e2e_lists", value: "unavailable", url: baseURL },
  ]);
  await page.goto("/lists");
  await expect(
    page.getByRole("heading", { name: "Your lists are unavailable" }),
  ).toBeVisible();
  await context.clearCookies({ name: "e2e_lists" });
  await page
    .getByRole("button", { name: "Try loading lists again", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Give your words a place" }),
  ).toBeVisible();
  const list = await seedList(context);
  await page.goto(`/lists/${list.id}`);
  expect(
    (
      await context.request.delete(
        `${apiURL}/api/v1/word-lists/${list.id}?expectedRevision=${list.revision}`,
        { headers: await writeHeaders(context) },
      )
    ).status(),
  ).toBe(204);
  await page
    .getByRole("button", { name: "Remove invite from this list", exact: true })
    .click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "may have been deleted",
  );
  await page
    .getByRole("button", { name: "Load current status", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "This list is no longer available" }),
  ).toBeVisible();
});

for (const colorScheme of ["light", "dark"] as const) {
  test(`personal lists remain readable and accessible in ${colorScheme}`, async ({
    page,
    context,
  }) => {
    const list = await seedList(
      context,
      "Useful meanings for conversations with friends",
      ["invite", "confirm", "reschedule"],
    );
    await page.emulateMedia({ colorScheme });
    await page.goto(`/lists/${list.id}`);
    const scan = await scanForAxeViolations(page);
    expect(
      scan.criticalOrSerious,
      formatViolations(scan.criticalOrSerious).join("\n"),
    ).toEqual([]);
    await assertKeyboardReachable(page, { minFocusable: 5, minTabStops: 5 });
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth,
      ),
    ).toBe(true);
    await page
      .getByLabel("Search words and meanings in this list")
      .fill("change");
    await expect(
      page
        .getByRole("list", { name: "List meanings", exact: true })
        .getByRole("listitem"),
    ).toHaveCount(1);
  });
}

test("signed-out list detail redirects with the list as returnTo", async ({
  page,
  context,
  baseURL,
}) => {
  await context.clearCookies();
  // The general mock defaults to a synthetic session; force its real 401 boundary.
  await context.addCookies([
    { name: "e2e_unauthenticated", value: "1", url: baseURL! },
  ]);
  const path = `/lists/${randomUUID()}`;
  await page.goto(path);
  await expect(page).toHaveURL(/\/login\?/);
  expect(new URL(page.url()).searchParams.get("returnTo")).toBe(path);
});

test("lost list deletion confirmation retries the exact deletion", async ({
  page,
  context,
}) => {
  const list = await seedList(context);
  await page.goto(`/lists/${list.id}`);
  const requests: { url: string; key: string | undefined }[] = [];
  await page.route(`**/api/v1/word-lists/${list.id}?*`, async (route) => {
    if (route.request().method() !== "DELETE") return route.continue();
    requests.push({
      url: route.request().url(),
      key: route.request().headers()["idempotency-key"],
    });
    if (requests.length === 1) {
      expect((await route.fetch()).status()).toBe(204);
      return route.abort("failed");
    }
    return route.continue();
  });
  await page.getByRole("button", { name: "Delete list", exact: true }).click();
  await page
    .getByRole("button", { name: "Confirm delete list", exact: true })
    .click();
  await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
  await expect(page.getByLabel("List name", { exact: true })).toBeDisabled();
  await page
    .getByRole("button", { name: "Retry same change", exact: true })
    .click();
  await expect(page).toHaveURL(/\/lists$/);
  expect(requests).toHaveLength(2);
  expect(requests[1]).toEqual(requests[0]);
});

test("expired membership session stops changes and returns to sign-in", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("Missing app URL");
  const list = await seedList(context, "Plan words", []);
  await openMembership(page, list);
  await context.addCookies([
    { name: "e2e_lists_auth", value: "expired", url: baseURL },
  ]);
  await page
    .getByRole("button", { name: "Add meaning to list", exact: true })
    .click();
  await expect(page).toHaveURL(/\/login\?/);
  expect(new URL(page.url()).searchParams.get("returnTo")).toBe(
    "/vocabulary/invite",
  );
  expect(new URL(page.url()).searchParams.get("reason")).toBe(
    "session-expired",
  );
});
