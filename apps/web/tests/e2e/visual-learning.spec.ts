import { readFileSync } from "node:fs";
import { createHash, randomUUID } from "node:crypto";
import { expect, test, type Locator } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";
async function waitForPictureVisibility(frame: Locator, visible: boolean) {
  await frame.locator(":scope > div").evaluate((element, expected) =>
    new Promise<void>((resolve) => {
      const observer = new IntersectionObserver(([entry]) => {
        if (entry?.isIntersecting === expected) {
          observer.disconnect();
          resolve();
        }
      });
      observer.observe(element);
    }), visible);
}

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
  await expect(page.getByTestId("meaning-picture").getByRole("status")).toHaveText("Picture unavailable.");
  await expect(page.getByTestId("meaning-picture").getByRole("button", { name: "Retry picture", exact: true })).toBeVisible();
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
  let releasePrevious = () => {};
  const previousResponse = new Promise<void>((resolve) => { releasePrevious = resolve; });
  let firstRequest = true;
  await page.route("https://images.vocanova.site/meanings/**", async (route) => {
    if (firstRequest) {
      firstRequest = false;
      await previousResponse;
    }
    const id = new URL(route.request().url()).pathname.split("/")[2];
    return route.fulfill({ contentType: "image/webp", body: readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url)) });
  });
  await page.clock.install();
  try {
    await page.goto("/learn/conversation-basics");
    await page.getByRole("button", { name: "Start lesson", exact: true }).click();
    const image = page.getByTestId("meaning-picture").getByRole("img");
    await expect(image).toHaveAttribute("src", /^https:\/\/images\.vocanova\.site\/meanings\//);
    const previousSrc = await image.getAttribute("src");
    const previousImage = await image.elementHandle();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(image).not.toHaveAttribute("src", previousSrc!);
    await expect(image).toHaveAttribute("src", /^https:\/\/images\.vocanova\.site\/meanings\//);
    await expect.poll(() => image.evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
    releasePrevious();
    await previousImage?.evaluate((element) => element.dispatchEvent(new Event("error")));
    await page.clock.fastForward(20_000);
    await expect(page.getByTestId("meaning-picture").getByRole("status")).toHaveCount(0);
    await expect(image).toHaveAttribute("src", /^https:\/\/images\.vocanova\.site\/meanings\//);
    for (let index = 0; index < 2; index++) await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(page.getByRole("group", { name: "Answer choices" })).toBeVisible();
    await expect(page.getByTestId("meaning-picture")).toHaveCount(0);
    const previousLookups = lookups;
    await page.evaluate(() => new Promise<void>((resolve) => requestAnimationFrame(() => resolve())));
    expect(lookups).toBe(previousLookups);
  } finally {
    releasePrevious();
  }
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


for (const theme of ["light", "dark"] as const) {
  test(`pending picture keeps its frame and recovers after a late load in ${theme}`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_meaning_image", value: "ready", url: baseURL! },
    ]);
    let release = () => {};
    const delayed = new Promise<void>((resolve) => { release = resolve; });
    await page.route("https://images.vocanova.site/meanings/**", async (route) => {
      await delayed;
      const id = new URL(route.request().url()).pathname.split("/")[2];
      await route.fulfill({ contentType: "image/webp", body: readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url)) });
    });
    await page.clock.install();
    try {
      await page.goto("/vocabulary/cover-letter", { waitUntil: "domcontentloaded" });
      const frame = page.getByTestId("meaning-picture");
      await frame.scrollIntoViewIfNeeded();
      await expect(frame.getByRole("status")).toHaveText("Loading picture…");
      const before = (await frame.boundingBox())!;
      // Offscreen lazy pictures must not acquire a retry just because time passes.
      await frame.evaluate((element) => { element.style.marginTop = "200vh"; });
      await page.evaluate(() => scrollTo(0, 0));
      await expect(frame).not.toBeInViewport();
      await waitForPictureVisibility(frame, false);
      await page.clock.runFor(100);
      await page.clock.fastForward(9_000);
      await expect(frame.getByRole("button", { name: "Retry picture", exact: true })).toHaveCount(0);
      await frame.scrollIntoViewIfNeeded();
      await waitForPictureVisibility(frame, true);
      await page.clock.runFor(100);
      await page.clock.fastForward(9_000);
      const retry = frame.getByRole("button", { name: "Retry picture", exact: true });
      await expect(retry).toBeVisible();
      await retry.focus();
      release();
      await expect.poll(() => frame.getByRole("img").evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
      await expect(frame.getByRole("status")).toHaveCount(0);
      await expect(frame).toBeFocused();
      await expect(frame.getByRole("button", { name: "Retry picture", exact: true })).toHaveCount(0);
      const after = (await frame.boundingBox())!;
      expect(Math.abs(after.height - before.height)).toBeLessThan(1);
      const imageBox = (await frame.getByRole("img").boundingBox())!;
      expect(Math.abs(imageBox.width / imageBox.height - 1.5)).toBeLessThan(0.03);
    } finally {
      release();
    }
  });
}


for (const theme of ["light", "dark"] as const) {
  test(`pending picture offers an explicit retry on the same trusted path in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_meaning_image", value: "ready", url: baseURL! },
    ]);
    let release = () => {};
    const delayed = new Promise<void>((resolve) => { release = resolve; });
    const urls: string[] = [];
    await page.route("https://images.vocanova.site/meanings/**", async (route) => {
      urls.push(route.request().url());
      if (urls.length === 1) await delayed;
      const id = new URL(route.request().url()).pathname.split("/")[2];
      await route.fulfill({ contentType: "image/webp", body: readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url)) });
    });
    await page.clock.install();
    try {
      await page.goto("/vocabulary/cover-letter", { waitUntil: "domcontentloaded" });
      const frame = page.getByTestId("meaning-picture");
      await frame.scrollIntoViewIfNeeded();
      await expect(frame.getByRole("status")).toHaveText("Loading picture…");
      await waitForPictureVisibility(frame, true);
      await page.clock.runFor(100);
      await page.clock.fastForward(9_000);
      const retry = frame.getByRole("button", { name: "Retry picture", exact: true });
      await expect(retry).toBeVisible();
      expect(urls).toHaveLength(1);
      const box = (await retry.boundingBox())!;
      expect(box.width).toBeGreaterThanOrEqual(44);
      expect(box.height).toBeGreaterThanOrEqual(44);
      await page.screenshot({ path: testInfo.outputPath(`pending-picture-${theme}.png`) });
      expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
      expect(await page.evaluate(() => document.visibilityState)).toBe("visible");
      const originalImage = await frame.getByRole("img").elementHandle();
      const responsePromise = page.waitForResponse((response) => new URL(response.url()).searchParams.has("vocanova_retry"));
      if (testInfo.project.name.startsWith("mobile")) {
        await retry.tap();
      } else {
        await retry.focus();
        await page.keyboard.press("Enter");
        await expect(frame).toBeFocused();
      }
      await page.clock.runFor(100);
      await testInfo.attach("retry-request-state", { contentType: "application/json", body: JSON.stringify({
        oldImageStillConnected: await originalImage?.evaluate((element) => element.isConnected),
        currentImage: await frame.getByRole("img").evaluate((element: HTMLImageElement) => ({ src: element.getAttribute("src"), complete: element.complete, naturalWidth: element.naturalWidth, visibility: document.visibilityState })),
        requestedURLs: urls,
      }) });
      await expect.poll(() => urls.length).toBe(2);
      const originalURL = new URL(urls[0]!);
      const retryURL = new URL(urls[1]!);
      expect(originalURL.search).toBe("");
      expect(retryURL.origin).toBe(originalURL.origin);
      expect(retryURL.pathname).toBe(originalURL.pathname);
      expect([...retryURL.searchParams.keys()]).toEqual(["vocanova_retry"]);
      expect(retryURL.searchParams.get("vocanova_retry")).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
      const response = await responsePromise;
      expect(response.status()).toBe(200);
      const id = originalURL.pathname.split("/")[2];
      expect(createHash("sha256").update(await response.body()).digest("hex")).toBe(createHash("sha256").update(readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url))).digest("hex"));
      await expect.poll(() => frame.getByRole("img").evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
      await expect(frame.getByRole("status")).toHaveCount(0);
      // A late result from the replaced attempt cannot restore its loading UI.
      release();
      await page.clock.fastForward(20_000);
      await expect(frame.getByRole("status")).toHaveCount(0);
      await expect(frame.getByRole("button", { name: "Retry picture", exact: true })).toHaveCount(0);
      expect(urls).toHaveLength(2);
    } finally {
      release();
    }
  });
}


test("a cached teaching picture remounts without a stale loading state", async ({ page }) => {
  await page.goto("/");
  const preview = page.getByRole("region", { name: "Learning example" });
  const picture = preview.getByTestId("meaning-picture");
  const image = picture.getByRole("img");
  await expect.poll(() => image.evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
  const originalSrc = await image.getAttribute("src");
  await preview.getByRole("button", { name: "See example feedback", exact: true }).click();
  await expect(picture).toHaveCount(0);
  await preview.getByRole("button", { name: "Back to the word", exact: true }).click();
  await expect(image).toHaveAttribute("src", originalSrc!);
  await expect.poll(() => image.evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
  await page.clock.install();
  await page.clock.fastForward(20_000);
  await expect(picture.getByRole("status")).toHaveCount(0);
  await expect(picture.getByRole("button", { name: "Retry picture", exact: true })).toHaveCount(0);
  await expect(preview.getByRole("button", { name: "See example feedback", exact: true })).toBeFocused();
});


for (const outcome of ["local fallback", "terminal unavailability"] as const) {
  for (const theme of ["light", "dark"] as const) {
    test(`focused picture retry retains its frame through ${outcome} in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
      await context.addCookies([
        { name: "vocanova_theme", value: theme, url: baseURL! },
        { name: "e2e_meaning_image", value: "ready", url: baseURL! },
      ]);
      let release = () => {};
      const pending = new Promise<void>((resolve) => { release = resolve; });
      const requests: string[] = [];
      await page.route("https://images.vocanova.site/meanings/**", async (route) => {
        requests.push(route.request().url());
        if (new URL(route.request().url()).searchParams.has("vocanova_retry")) {
          if (requests.length === 2) {
            await route.abort("failed");
            return;
          }
        } else {
          await pending;
        }
        const id = new URL(route.request().url()).pathname.split("/")[2];
        await route.fulfill({ contentType: "image/webp", body: readFileSync(new URL(`../../public/images/meanings/${id}.webp`, import.meta.url)) });
      });
      if (outcome === "terminal unavailability") {
        await page.route("**/_next/image?**", (route) => route.abort("failed"));
      }
      await page.clock.install();
      try {
        await page.goto("/vocabulary/cover-letter", { waitUntil: "domcontentloaded" });
        const frame = page.getByTestId("meaning-picture");
        await frame.scrollIntoViewIfNeeded();
        await expect(frame.getByRole("status")).toHaveText("Loading picture…");
        const originalFrame = await frame.elementHandle();
        await waitForPictureVisibility(frame, true);
        await page.clock.runFor(100);
        await page.clock.fastForward(9_000);
        const retry = frame.getByRole("button", { name: "Retry picture", exact: true });
        await retry.focus();
        await page.keyboard.press("Enter");
        if (outcome === "local fallback") {
          await expect(frame.getByRole("img")).toHaveAttribute("src", /_next\/image/);
          await expect.poll(() => frame.getByRole("img").evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
        } else {
          await expect(frame.getByRole("status")).toHaveText("Picture unavailable.");
          await expect(retry).toBeVisible();
          await expect(frame.getByRole("img")).toHaveCount(0);
        }
        await expect(frame).toBeFocused();
        expect(await originalFrame?.evaluate((element) => element.isConnected)).toBe(true);
        await expect(page.getByRole("button", { name: /^Save cover letter:/ })).toBeVisible();
        await page.screenshot({ path: testInfo.outputPath(`picture-error-${theme}.png`) });
        if (outcome === "terminal unavailability") {
          await retry.focus();
          await page.keyboard.press("Enter");
          await expect(frame.getByRole("img")).toHaveAttribute("src", /^https:\/\/images\.vocanova\.site\/meanings\//);
          await expect.poll(() => frame.getByRole("img").evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
          await expect(frame.getByRole("status")).toHaveCount(0);
          expect(new URL(requests[2]!).searchParams.get("vocanova_retry")).not.toBe(new URL(requests[1]!).searchParams.get("vocanova_retry"));
        }
        release();
        await page.clock.fastForward(20_000);
        await expect(frame).toBeFocused();
        expect(requests).toHaveLength(outcome === "terminal unavailability" ? 3 : 2);
      } finally {
        release();
      }
    });
  }
}
