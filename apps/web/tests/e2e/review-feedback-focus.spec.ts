import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";

for (const font of ["native", "fallback"] as const) {
test(`checked review feedback keeps targets clear of persistent bars across window heights (${font})`, async ({ page, baseURL }, testInfo) => {
  const width = page.viewportSize()!.width;
  const observations = [];
  for (const height of [480, 540, 600, 640, 720, 736, 800, 900]) {
    await page.setViewportSize({ width, height });
    await page.context().addCookies([
      { name: "vocanova_session", value: `feedback-focus-${randomUUID()}`, url: baseURL! },
      { name: "vocanova_csrf", value: "feedback-focus-csrf", url: baseURL! },
      { name: "e2e_review_fixture_count", value: "1", url: baseURL! },
      { name: "e2e_daily_review_target", value: "20", url: baseURL! },
    ]);
    await page.goto("/reviews");
    // Exercise a shipped fallback font when Avenir/Segoe are unavailable.
    if (font === "fallback") await page.addStyleTag({ content: "body { font-family: Arial, sans-serif; }" });
    await page.getByRole("button", { name: "Start review", exact: true }).click();
    await page.getByRole("radio", { name: "definition for review word 1", exact: true }).check();
    await page.getByRole("button", { name: "Check answer", exact: true }).click();
    const feedback = page.getByRole("status").filter({ hasText: /^Correct/ });
    await expect(feedback).toBeFocused();
    const header = (await page.getByRole("banner").boundingBox())!;
    const rail = (await page.getByRole("region", { name: "Session actions", exact: true }).boundingBox())!;
    const bounds = (await feedback.boundingBox())!;
    const audio = (await page.getByRole("button", { name: "Automatic audio off", exact: true }).boundingBox())!;
    observations.push({ width, height, header, rail, feedback: bounds, audio });
    await testInfo.attach(`feedback-${height}`, { body: JSON.stringify(observations.at(-1)), contentType: "application/json" });
    expect(bounds.y).toBeGreaterThanOrEqual(header.y + header.height + 8);
    expect(bounds.y + bounds.height).toBeLessThanOrEqual(rail.y - 8);
    const exposedAudioHeight = Math.min(audio.y + audio.height, rail.y) - Math.max(audio.y, header.y + header.height);
    if (exposedAudioHeight > 0) expect(exposedAudioHeight).toBeGreaterThanOrEqual(24);
    const scan = await scanForAxeViolations(page);
    expect(scan.criticalOrSerious, JSON.stringify(observations.at(-1))).toEqual([]);
  }
});
}
