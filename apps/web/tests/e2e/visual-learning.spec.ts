import { readFileSync } from "node:fs";
import { createHash, randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";
test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_teaching", value: "true", url: baseURL! },
  ]);
});
for (const theme of ["light", "dark"] as const) {
  test(`canonical picture and optional usage help stay usable in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/vocabulary/cover-letter");
    const picture = page.getByTestId("meaning-picture").getByRole("img");
    await expect(picture).toBeVisible();
    await expect.poll(() => picture.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
    await expect(picture).toHaveAttribute("alt", /.+/);
    const tips = page.locator("summary").filter({ hasText: "Usage tips" });
    if (await tips.count()) {
      await expect(tips.locator("..")).not.toHaveAttribute("open");
      await tips.focus();
      await page.keyboard.press("Enter");
      await expect(tips.locator("..")).toHaveAttribute("open", "");
      await expect(tips).toBeFocused();
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.evaluate(() => scrollTo(0, 0));
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    await page.screenshot({ path: testInfo.outputPath(`word-picture-${theme}.png`), fullPage: true });
  });
}
test("broken picture leaves the meaning and save action usable", async ({ page }) => {
  await page.route("**/_next/image?**", (route) => route.abort());
  await page.goto("/vocabulary/cover-letter");
  await expect(page.getByRole("heading", { name: "cover letter", exact: true })).toBeVisible();
  await expect(page.getByTestId("meaning-picture")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /^Save cover letter:/ })).toBeVisible();
  await expect(page.getByRole("heading", { name: "In a sentence" })).toBeVisible();
});
test("a teaching picture disappears before graded lesson questions", async ({ page }) => {
  await page.goto("/learn/conversation-basics");
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const teachingImage = page.getByTestId("meaning-picture").getByRole("img");
  await expect(page.getByRole("heading", { name: "invite", exact: true })).toBeVisible();
  await expect(teachingImage).toBeVisible();
  await expect.poll(() => teachingImage.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
  expect((await teachingImage.boundingBox())!.height).toBeLessThanOrEqual(288);
  for (let step = 0; step < 3; step++) {
    await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
  }
  await expect(page.getByRole("group", { name: "Answer choices" })).toBeVisible();
  await expect(page.getByTestId("meaning-picture")).toHaveCount(0);
});

for (const theme of ["light", "dark"] as const) {
  test(`approved CDN picture loads directly without a catalog waterfall in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_meaning_image", value: "ready", url: baseURL! },
    ]);
    const requests: string[] = [];
    page.on("request", (request) => requests.push(request.url()));
    await page.route("https://images.vocanova.site/meanings/**", (route) => {
      const id = new URL(route.request().url()).pathname.split("/")[2];
      return route.fulfill({ contentType: "image/webp", body: readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url)) });
    });
    await page.goto("/vocabulary/cover-letter");
    const picture = page.getByTestId("meaning-picture").getByRole("img");
    await expect(picture).toHaveAttribute("src", /^https:\/\/images\.vocanova\.site\/meanings\//);
    await expect(picture).toHaveAttribute("loading", "lazy");
    await expect.poll(() => picture.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
    expect(requests.filter((url) => url.includes("/api/v1/meaning-images"))).toEqual([]);
    expect(requests.filter((url) => url.includes("/_next/image") && url.includes("images.vocanova"))).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    const box = (await picture.boundingBox())!;
    expect(Math.abs(box.width / box.height - 1.5)).toBeLessThan(0.02);
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    await picture.evaluate((element) => element.scrollIntoView({ block: "center" }));
    await page.screenshot({ path: testInfo.outputPath(`cdn-picture-${theme}.png`) });
  });

  test(`failed CDN picture preserves existing local artwork in ${theme}`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_meaning_image", value: "ready", url: baseURL! },
    ]);
    await page.route("https://images.vocanova.site/meanings/**", (route) => route.abort());
    await page.goto("/vocabulary/cover-letter");
    const picture = page.getByTestId("meaning-picture").getByRole("img");
    await expect(picture).toHaveAttribute("src", /_next\/image/);
    await expect.poll(() => picture.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
    await expect(page.getByRole("button", { name: /^Save cover letter:/ })).toBeVisible();
  });
}


test("revision-bound lesson lookup shows CDN artwork only during teaching", async ({ page }) => {
  const approved = new Map<string, { meaningId: string; meaningRevision: string; sha256: string; imageUrl: string; alt: string; width: number; height: number }>();
  let lookups = 0;
  await page.route((url) => url.pathname.startsWith("/api/v1/lesson-sessions/") || /^\/api\/v1\/lessons\/[^/]+\/sessions$/.test(url.pathname), async (route) => {
    const response = await route.fetch();
    const data = await response.json();
    const step = data.currentStep;
    if (step?.kind === "teach") {
      const word = step.word;
      const revision = createHash("sha256").update(JSON.stringify([word.meaningId, word.wordText, word.partOfSpeech, word.definition])).digest("hex");
      step.pronunciationRevision = revision;
      const sha256 = "a".repeat(64);
      approved.set(word.meaningId, {
        meaningId: word.meaningId, meaningRevision: revision, sha256,
        imageUrl: `https://images.vocanova.site/meanings/${word.meaningId}/${revision}/${sha256}.webp`,
        alt: "Friends greet a visitor at an open doorway.", width: 960, height: 640,
      });
    }
    await route.fulfill({ response, json: data });
  });
  await page.route("**/api/v1/meaning-images?**", async (route) => {
    lookups++;
    const ids = new URL(route.request().url()).searchParams.get("meaningIds")!.split(",");
    expect(ids.length).toBeLessThanOrEqual(100);
    await route.fulfill({ json: { items: ids.flatMap((id) => approved.has(id) ? [approved.get(id)] : []) } });
  });
  await page.route("https://images.vocanova.site/meanings/**", (route) => {
    const id = new URL(route.request().url()).pathname.split("/")[2];
    return route.fulfill({ contentType: "image/webp", body: readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url)) });
  });
  await page.goto("/learn/conversation-basics");
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const image = page.getByTestId("meaning-picture").getByRole("img");
  await expect(image).toHaveAttribute("src", /^https:\/\/images\.vocanova\.site\/meanings\//);
  await expect.poll(() => image.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
  for (let index = 0; index < 3; index++) await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("group", { name: "Answer choices" })).toBeVisible();
  await expect(page.getByTestId("meaning-picture")).toHaveCount(0);
  const previousLookups = lookups;
  await page.evaluate(() => new Promise<void>((resolve) => requestAnimationFrame(() => resolve())));
  expect(lookups).toBe(previousLookups);
});


test("changed canonical meaning hides obsolete local artwork without an empty box", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "e2e_meaning_image", value: "stale", url: baseURL! }]);
  await page.goto("/vocabulary/cover-letter");
  await expect(page.getByRole("heading", { name: "cover letter", exact: true })).toBeVisible();
  await expect(page.getByTestId("meaning-picture")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /^Save cover letter:/ })).toBeVisible();
  await expect(page.getByRole("heading", { name: "In a sentence" })).toBeVisible();
});


test("Home streams the saved-word picture without delaying Review", async ({ page, context, baseURL }) => {
  await page.goto("/vocabulary/invite");
  await page.getByRole("button", { name: /^Save invite:/ }).click();
  await expect(page.getByRole("button", { name: "Remove invite from saved words", exact: true })).toBeVisible();
  await context.addCookies([
    { name: "e2e_review_fixture_count", value: "2", url: baseURL! },
    { name: "e2e_meaning_image", value: "ready", url: baseURL! },
    { name: "e2e_meaning_images_delay_ms", value: "2000", url: baseURL! },
  ]);
  await page.route("https://images.vocanova.site/meanings/**", (route) => {
    const id = new URL(route.request().url()).pathname.split("/")[2];
    return route.fulfill({ contentType: "image/webp", body: readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url)) });
  });
  await page.goto("/home", { waitUntil: "commit" });
  const review=page.getByRole("region", {name:"Review words",exact:true});
  await expect(review.getByRole("button",{name:"Start review",exact:true})).toBeEnabled();
  const picture=page.getByRole("region",{name:"Your vocabulary",exact:true}).getByTestId("meaning-picture").getByRole("img");
  await expect(picture).toHaveAttribute("src", /_next\/image/);
  await expect(picture).toHaveAttribute("src", /^https:\/\/images\.vocanova\.site\/meanings\//);
  await expect.poll(()=>picture.evaluate((element)=>(element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
  await expect(review.getByRole("button",{name:"Start review",exact:true})).toBeEnabled();
});

for (const mode of ["stale", "unavailable"] as const) {
  test(`Home retains its reviewed local fallback when image metadata is ${mode}`, async ({ page,context,baseURL })=>{
    await page.goto("/vocabulary/invite");
    await page.getByRole("button",{name:/^Save invite:/}).click();
    await expect(page.getByRole("button",{name:"Remove invite from saved words",exact:true})).toBeVisible();
    await context.addCookies([
      {name:"e2e_review_fixture_count",value:"2",url:baseURL!},
      {name:"e2e_meaning_image",value:mode==="stale"?"stale":"ready",url:baseURL!},
      ...(mode==="unavailable"?[{name:"e2e_meaning_images_status",value:"503",url:baseURL!}]:[]),
    ]);
    await page.goto("/home");
    const picture=page.getByRole("region",{name:"Your vocabulary",exact:true}).getByTestId("meaning-picture").getByRole("img");
    await expect(picture).toHaveAttribute("src",/_next\/image/);
    await expect.poll(()=>picture.evaluate((element)=>(element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
    await expect(page.getByRole("region",{name:"Review words",exact:true}).getByRole("button",{name:"Start review",exact:true})).toBeEnabled();
    expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  });
}
