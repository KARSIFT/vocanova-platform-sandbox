import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`explicit Retry recovers a failed Settings chunk (${theme})`, async ({ page, context, baseURL }) => {
    const chunks = path.resolve(".next/static/chunks");
    const targets = readdirSync(chunks).filter(name => name.endsWith(".js") && readFileSync(path.join(chunks,name),"utf8").includes("Review rhythm preset"));
    expect(targets).toHaveLength(1);
    const chunkPath = `/_next/static/chunks/${targets[0]}`;
    await context.addCookies([{name:"vocanova_session",value:randomUUID(),url:baseURL!},{name:"vocanova_theme",value:theme,url:baseURL!}]);
    await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference",value),theme);
    let block = true, blocked = 0, documents = 0;
    const documentPaths: string[] = [];
    page.on("request",request=>{if(request.isNavigationRequest() && request.frame()===page.mainFrame()) {documents++;documentPaths.push(new URL(request.url()).pathname);}});
    // Block this real route module through the runtime's own network retry.
    await page.route(`**${chunkPath}`,route=>{if(block){blocked++;return route.abort("failed");}return route.continue();});
    await page.goto("/home");
    await expect(page.getByRole("heading",{name:"Today",exact:true})).toBeVisible();
    const settings=page.getByRole("link",{name:"Settings",exact:true});
    await settings.focus();await page.keyboard.press("Enter");
    const error=page.getByRole("heading",{name:"We couldn't load your settings",exact:true});
    await expect(error).toBeVisible({timeout:15000});
    expect(blocked).toBeGreaterThanOrEqual(2);
    // React may log only "Error" in Firefox. The blocked real module,
    // friendly boundary and explicit document recovery are the contract.
    expect(documents).toBe(1); // No automatic reload on failure.
    block=false;
    await page.getByRole("button",{name:"Try again",exact:true}).focus();
    await page.keyboard.press("Enter");
    await expect(page.getByRole("heading",{name:"Settings",exact:true})).toBeVisible();
    expect(documents).toBe(2);
    expect(documentPaths).toEqual(["/home", "/settings"]);
    await expect(page.getByRole("group",{name:"Daily review target",exact:true})).toBeVisible(); // Explicit retry performs exactly one document navigation.
    await expect(error).toHaveCount(0);
    await expect(page.locator("html")).toHaveAttribute("data-theme",theme);
  });

  test(`ordinary Settings API error retries without reloading (${theme})`, async ({page,context,baseURL})=>{
    await context.addCookies([{name:"vocanova_session",value:randomUUID(),url:baseURL!},{name:"vocanova_theme",value:theme,url:baseURL!},{name:"e2e_settings_read_failure",value:"503",url:baseURL!}]);
    let documents=0;
    page.on("request",request=>{if(request.isNavigationRequest() && request.frame()===page.mainFrame())documents++;});
    await page.goto("/settings");
    await expect(page.getByRole("heading",{name:"We couldn't load your settings",exact:true})).toBeVisible();
    await context.clearCookies({name:"e2e_settings_read_failure"});
    await page.getByRole("button",{name:"Try again",exact:true}).click();
    await expect(page.getByRole("heading",{name:"Settings",exact:true})).toBeVisible();
    expect(documents).toBe(1);
  });
}
