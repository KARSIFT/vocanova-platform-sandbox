import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`discovery resumes exact 4330 from Home and explicit links win in ${theme}`, async ({ page, baseURL }) => {
    // Lighthouse injects this header into all requests, including save preflights.
    await page.context().setExtraHTTPHeaders({ "X-Lighthouse-T09-Screen": "discovery", "X-Lighthouse-T09-Layout": "mobile-360" });
    await page.context().addCookies([
      { name: "vocanova_session", value: `position-${randomUUID()}`, url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_knowledge_map", value: "complete", url: baseURL! },
    ]);
    await page.goto("/discovery?band=44&word=4330&active=1&revisit=1");
    const title = page.getByRole("heading", { name: "Map word 4330", exact: true });
    await expect(title).toBeVisible();
    const card = page.getByRole("article");
    const previous = page.getByRole("button", { name: "Previous word", exact: true });
    await expect(previous).toBeVisible();
    expect((await previous.boundingBox())!.y).toBeGreaterThan((await card.boundingBox())!.y + (await card.boundingBox())!.height);
    await expect(page.getByRole("status", { name: "Discovery position" })).toHaveText("Your place is saved.");
    await page.goto("/home");
    await page.goto("/discovery");
    await expect(title).toBeVisible();
    await page.goto("/discovery?band=12&word=1150&active=1&revisit=1");
    await expect(page.getByRole("heading", { name: "Map word 1150", exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Word range", exact: true }).click();
    const picker = page.getByRole("dialog", { name: "Choose a range" });
    const current = picker.getByRole("button", { name: "1,101–1,200", exact: true });
    await expect(current).toHaveAttribute("aria-current", "true");
    await expect(current.locator("svg")).toHaveCount(1);
    await expect(current).toContainText(/already know.*(?:learning|saved).*new/);
    await page.keyboard.press("Escape");
    await expect(page.getByRole("button", { name: "Word range", exact: true })).toBeFocused();
  });
}

test("discovery positions are isolated between accounts", async ({browser,baseURL})=>{
 const a=await browser.newContext(), b=await browser.newContext();
 await a.addCookies([{name:"vocanova_session",value:`position-a-${randomUUID()}`,url:baseURL!}]);
 await b.addCookies([{name:"vocanova_session",value:`position-b-${randomUUID()}`,url:baseURL!}]);
 const one=await a.newPage(),two=await b.newPage();
 await one.goto("/discovery?band=44&word=4330&active=1&revisit=1");
 await expect(one.getByRole("status",{name:"Discovery position"})).toHaveText("Your place is saved.");
 await two.goto("/discovery");
 await expect(two.getByRole("heading",{name:"Map word 4330",exact:true})).toHaveCount(0);
 await a.close();await b.close();
});


test("range previews preserve legacy aggregate counts without per-word claims",async({page,baseURL})=>{
 await page.context().addCookies([
  {name:"vocanova_session",value:`position-legacy-${randomUUID()}`,url:baseURL!},
  {name:"e2e_knowledge_map",value:"legacy-progress",url:baseURL!},
 ]);
 await page.goto("/discovery?band=1&word=2&active=1&revisit=1");
 await page.getByRole("button",{name:"Word range",exact:true}).click();
 const row=page.getByRole("dialog",{name:"Choose a range"}).getByRole("button",{name:"1–100",exact:true});
 await expect(row).toContainText("saved");
 await expect(row).not.toContainText("learned");
 await expect(row.locator("svg")).toHaveCount(1);
 await expect(row.locator("svg pattern")).toHaveCount(0);
});

test("optional position failures leave navigation usable and allow explicit retry",async({page,baseURL})=>{
 await page.context().addCookies([{name:"vocanova_session",value:`position-outage-${randomUUID()}`,url:baseURL!}]);
 await page.route("**/api/v1/discovery-position",async(route)=>{
  if(route.request().method()==="PUT")await route.abort("failed");else await route.continue();
 });
 await page.goto("/discovery?band=44&word=4330&active=1&revisit=1");
 await expect(page.getByRole("status",{name:"Discovery position"})).toContainText("could not be saved");
 await expect(page.getByRole("button",{name:"Previous word",exact:true})).toBeEnabled();
 await expect(page.getByRole("button",{name:"Learn",exact:true})).toBeEnabled();
 await page.unroute("**/api/v1/discovery-position");
 await page.getByRole("button",{name:"Save this place again",exact:true}).click();
 await expect(page.getByRole("status",{name:"Discovery position"})).toHaveText("Your place is saved.");
});


for (const optionalReadFails of [false, true]) {
 test(`stale discovery tab cannot save or retry into a switched account (optional read fails: ${optionalReadFails})`, async ({page, context, baseURL}) => {
  const cookies = (session: string, alternate: boolean) => [
   {name: "vocanova_session", value: session, url: baseURL!},
   {name: "e2e_user_id", value: alternate ? "alternate" : "default", url: baseURL!},
   {name: "e2e_position_read_failure", value: optionalReadFails ? "1" : "0", url: baseURL!},
  ];
  await context.addCookies(cookies(`switch-a-${randomUUID()}`, false));
  let failedReads = 0;
  page.on("response", response => {
    if (new URL(response.url()).pathname === "/api/v1/discovery-position" && response.request().method() === "GET" && response.status() === 503) failedReads++;
  });
  await page.goto("/discovery?band=44&word=4330&active=1&revisit=1");
  await expect(page.getByRole("status", {name: "Discovery position"})).toContainText(optionalReadFails ? "could not be saved" : "Your place is saved.");
  if (optionalReadFails) expect(failedReads).toBeGreaterThan(0);
  const other = await context.newPage();
  await context.addCookies(cookies(`switch-b-${randomUUID()}`, true));
  await context.addCookies([{name:"e2e_position_read_failure",value:"0",url:baseURL!}]);
  await other.goto("/discovery?band=12&word=1150&active=1&revisit=1");
  await expect(other.getByRole("status", {name: "Discovery position"})).toHaveText("Your place is saved.");
  // Read through the observed API origin; this is the same browser cookie jar.
  const apiURL = new URL("/api/v1/discovery-position", process.env.NEXT_PUBLIC_API_BASE_URL ?? `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`).href;
  const before = await other.evaluate(async url => (await (await fetch(url, {credentials:"include"})).json()).position, apiURL);
  const rejected = page.waitForResponse(response => new URL(response.url()).pathname === "/api/v1/discovery-position" && response.request().method() === "PUT");
  await page.getByRole("button", {name:"Previous word",exact:true}).click();
  if (optionalReadFails) await page.getByRole("button", {name:"Save this place again",exact:true}).click();
  const response = await rejected;
  const after = await other.evaluate(async url => (await (await fetch(url, {credentials:"include"})).json()).position, apiURL);
  expect(after).toEqual(before);
  expect(response.status()).toBe(403);
  expect(response.request().postDataJSON().expectedAccountId).toBe("00000000-0000-0000-0000-000000000001");
  await expect(page.getByRole("status",{name:"Discovery position"})).toContainText("Your account changed");
  await expect(page.getByRole("button",{name:"Save this place again",exact:true})).toHaveCount(0);
  await expect(page.getByRole("link",{name:"Reload discovery",exact:true})).toBeVisible();
  await page.getByRole("button", {name:"Previous word",exact:true}).click();
  await expect(page.getByRole("status",{name:"Discovery position"})).toContainText("Your account changed");
  await other.close();
 });
}
