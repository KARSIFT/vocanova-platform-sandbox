import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`discovery resumes exact 4330 from Home and explicit links win in ${theme}`, async ({ page, baseURL }) => {
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
