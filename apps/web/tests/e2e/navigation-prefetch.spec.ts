// Every protected route prefetch runs the server identity check. These checks
// exercise the shipped policy without suppressing viewport observers: primary
// destinations and prominent links into routes with a loading state prefetch
// on view (so that state shows at once), other links on intent, the search
// form never; and a slow navigation is acknowledged at once.

import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

const VIEWPORT = new Set([
  "/home",
  "/discover",
  "/progress",
  "/practice",
  "/stories",
  "/words",
]);
const viewportPrefetch = (path: string) =>
  VIEWPORT.has(path) ||
  path.startsWith("/words/") ||
  path.startsWith("/discover/");
const INTENT_ONLY = [
  "/vocabulary",
  "/settings",
  "/lists",
  "/plan",
  "/courses",
  "/discovery",
];

async function signIn(page: Page, cookies: Record<string, string> = {}) {
  const baseURL = test.info().project.use.baseURL!;
  await page.context().addCookies([
    { name: "vocanova_session", value: `nav-${randomUUID()}`, url: baseURL },
    { name: "vocanova_csrf", value: "navigation-prefetch", url: baseURL },
    ...Object.entries(cookies).map(([name, value]) => ({
      name,
      value,
      url: baseURL,
    })),
  ]);
}

function recordPrefetches(page: Page) {
  const paths: string[] = [];
  const writes: string[] = [];
  page.on("request", (request) => {
    if (request.headers()["next-router-prefetch"])
      paths.push(new URL(request.url()).pathname);
    if (request.method() !== "GET" && request.method() !== "HEAD")
      writes.push(request.url());
  });
  return { paths, writes };
}

async function settle(page: Page) {
  // Prefetches start right after hydration; give them time to be observed.
  await page.waitForLoadState("networkidle");
}

for (const route of ["/home", "/discover", "/progress"] as const) {
  test(`${route} prefetches only its planned destinations while idle`, async ({
    page,
  }) => {
    await signIn(page);
    const { paths, writes } = recordPrefetches(page);
    await page.goto(route);
    await settle(page);
    await page.mouse.wheel(0, 2000);
    await settle(page);
    // Two requests (route tree and segment) per destination, at most seven.
    expect(paths.length).toBeLessThanOrEqual(14);
    for (const path of paths) expect(viewportPrefetch(path), path).toBe(true);
    for (const path of INTENT_ONLY) expect(paths).not.toContain(path);
    expect(paths).not.toContain(route);
    expect(writes).toEqual([]);
  });
}

test("other links prefetch once on intent and never write", async ({
  page,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "home-desktop-1280",
    "Pointer hover is a desktop interaction; focus intent is covered below.",
  );
  await signIn(page);
  const { paths, writes } = recordPrefetches(page);
  await page.goto("/home");
  await settle(page);
  expect(paths).not.toContain("/plan");
  const plan = page
    .getByRole("navigation", { name: "Explore your English" })
    .getByRole("link", { name: "Your plan", exact: true });
  await plan.hover();
  await expect
    .poll(() => paths.filter((path) => path === "/plan").length)
    .toBeGreaterThan(0);
  await page.mouse.move(0, 0);
  await plan.hover();
  await settle(page);
  // The router keeps the first result; hovering again does not refetch.
  expect(paths.filter((path) => path === "/plan").length).toBeLessThanOrEqual(2);
  expect(writes).toEqual([]);
});

test("keyboard focus is intent; the search form does not prefetch", async ({
  page,
}) => {
  await signIn(page);
  const { paths } = recordPrefetches(page);
  await page.goto("/home");
  await settle(page);
  const search = page.getByRole("searchbox", { name: "Search any English word" });
  await search.focus();
  await settle(page);
  expect(paths).not.toContain("/vocabulary");
  await page
    .getByRole("navigation", { name: "Explore your English" })
    .getByRole("link", { name: "Personal lists", exact: true })
    .focus();
  await expect.poll(() => paths.includes("/lists")).toBe(true);
});

test("a slow navigation is acknowledged immediately and clears when it arrives", async ({
  page,
}) => {
  await signIn(page);
  await page.goto("/home");
  await settle(page);
  // Delay every later API response, including the identity check.
  await page.context().addCookies([
    { name: "e2e_response_delay_ms", value: "1500", url: page.url() },
  ]);
  const progress = page.locator(".navigation-progress");
  await expect(progress).not.toHaveAttribute("data-active", "true");
  await page
    .getByRole("navigation", { name: "Explore your English" })
    .getByRole("link", { name: "Situations", exact: true })
    .click();
  // Well before the 1.5 s identity check could finish.
  await expect(progress).toHaveAttribute("data-active", "true", {
    timeout: 500,
  });
  await expect(
    page.getByRole("heading", { name: "Situation courses", level: 1 }),
  ).toBeVisible({ timeout: 15_000 });
  await expect(progress).not.toHaveAttribute("data-active", "true");
});

test("a prominent link shows its destination's loading state at once", async ({
  page,
}) => {
  await signIn(page);
  await page.goto("/home");
  const practice = page
    .getByRole("navigation", { name: "Explore your English" })
    .getByRole("link", { name: "Practice", exact: true });
  // Viewport prefetch happens when the link is on screen (below the fold on
  // small phones), as it is when a learner reaches it.
  await practice.scrollIntoViewIfNeeded();
  await settle(page);
  // From now on every server read is slow, including the identity check; only
  // the loading state prefetched on view can appear immediately.
  await page.context().addCookies([
    { name: "e2e_response_delay_ms", value: "1500", url: page.url() },
  ]);
  await practice.click();
  const status = page.getByRole("status").filter({ hasText: "Loading practice" });
  await expect(status).toBeVisible({ timeout: 700 });
  await expect(
    page.getByRole("heading", { name: "Today’s practice", level: 1 }),
  ).toBeVisible({ timeout: 15_000 });
});

test("choosing the current page supersedes a pending navigation", async ({
  page,
}) => {
  await signIn(page);
  await page.goto("/home");
  await settle(page);
  await page.context().addCookies([
    { name: "e2e_response_delay_ms", value: "1500", url: page.url() },
  ]);
  const progress = page.locator(".navigation-progress");
  await page
    .getByRole("navigation", { name: "Explore your English" })
    .getByRole("link", { name: "Situations", exact: true })
    .click();
  await expect(progress).toHaveAttribute("data-active", "true", {
    timeout: 500,
  });
  await page.getByRole("banner").getByRole("link", { name: "VocaNova" }).click();
  await expect(progress).not.toHaveAttribute("data-active", "true", {
    timeout: 500,
  });
  // The router keeps the newer request: the learner stays on Today.
  await page.waitForTimeout(4000);
  await expect(page).toHaveURL(/\/home$/);
  await expect(progress).not.toHaveAttribute("data-active", "true");
});

test("Back during a pending navigation clears the indicator", async ({
  page,
}) => {
  await signIn(page);
  await page.goto("/progress");
  await page
    .getByRole("navigation", { name: "Primary" })
    .filter({ visible: true })
    .getByRole("link", { name: "Home" })
    .click();
  await expect(page).toHaveURL(/\/home$/);
  await settle(page);
  await page.context().addCookies([
    { name: "e2e_response_delay_ms", value: "1500", url: page.url() },
  ]);
  const progress = page.locator(".navigation-progress");
  await page
    .getByRole("navigation", { name: "Explore your English" })
    .getByRole("link", { name: "Situations", exact: true })
    .click();
  await expect(progress).toHaveAttribute("data-active", "true", {
    timeout: 500,
  });
  await page.goBack({ waitUntil: "commit" });
  await expect(progress).not.toHaveAttribute("data-active", "true", {
    timeout: 1000,
  });
  await expect(page).toHaveURL(/\/progress$/, { timeout: 15_000 });
});

test("a search submission is acknowledged immediately", async ({ page }) => {
  await signIn(page);
  await page.goto("/home");
  await settle(page);
  await page.context().addCookies([
    { name: "e2e_response_delay_ms", value: "1500", url: page.url() },
  ]);
  const search = page.getByRole("searchbox", { name: "Search any English word" });
  await search.fill("invite");
  await search.press("Enter");
  await expect(page.locator(".navigation-progress")).toHaveAttribute(
    "data-active",
    "true",
    { timeout: 500 },
  );
  await expect(page).toHaveURL(/\/vocabulary\?q=invite$/, { timeout: 15_000 });
  await expect(
    page.getByRole("link", { name: "invite", exact: true }),
  ).toBeVisible({ timeout: 15_000 });
});
