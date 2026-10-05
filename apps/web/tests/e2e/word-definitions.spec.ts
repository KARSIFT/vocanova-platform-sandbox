import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

const shortDefinition = "to make liquid flow into a container";

for (const fixture of ["duplicate", "distinct", "blank", "absent"]) {
  test(`word pages retain useful definitions without ${fixture} repetition`, async ({
    page,
  }, testInfo) => {
    const url = testInfo.project.use.baseURL!;
    await page.context().addCookies([
      { name: "vocanova_session", value: randomUUID(), url },
      { name: "vocanova_csrf", value: "definitions-fixture", url },
      { name: "e2e_definition_fixture", value: fixture, url },
    ]);
    await page.goto("/discover/ordering-at-a-cafe/pour");
    await page.getByRole("button", { name: /^Save pour:/ }).click();
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();

    for (const route of [
      "/discover/ordering-at-a-cafe/pour",
      "/words/uw-mean-pour",
    ]) {
      await page.goto(route);
      const main = page.getByRole("main");
      const definition = main
        .getByText(shortDefinition, { exact: true })
        .first();
      await expect(definition).toBeVisible();
      const paragraphs = definition.locator("..").locator(":scope > p");
      const exactDefinitions = await paragraphs.evaluateAll((elements) =>
        elements
          .map((element) =>
            element.textContent?.trim().replace(/\s+/g, " ").toLowerCase(),
          )
          .filter((text) => text === "to make liquid flow into a container"),
      );
      expect(exactDefinitions).toHaveLength(1);
      await expect(paragraphs.filter({ hasText: /^\s*$/ })).toHaveCount(0);
      if (fixture === "distinct") {
        await expect(
          main.getByText(
            `${shortDefinition}. You control where the liquid goes by tipping its container.`,
            { exact: true },
          ),
        ).toBeVisible();
      }
      await expect(
        main.getByText("Could you pour me a cup of coffee?", { exact: true }),
      ).toBeVisible();
    }
  });
}

for (const theme of ["light", "dark"] as const) {
  test(`discovery and saved words explain usage in plain English in ${theme} mode`, async ({
    page,
  }, testInfo) => {
    const url = testInfo.project.use.baseURL!;
    await page.context().addCookies([
      { name: "vocanova_session", value: randomUUID(), url },
      { name: "vocanova_csrf", value: "usage-notes-fixture", url },
      { name: "e2e_daily_conversation", value: "true", url },
      { name: "vocanova_theme", value: theme, url },
    ]);
    await page.goto("/discover/daily-conversation/sounds-good");
    const main = page.getByRole("main");
    const tips = main.locator("summary").filter({ hasText: "Usage tips" });
    const notes = tips.locator("..");
    const expectedNotes = [
      {
        heading: "Often used with",
        text: "That sounds good; sounds good to me",
      },
      {
        heading: "When to use it",
        text: "Sounds good! is a natural short response in friendly conversation and messages. In sentence practice, use a complete example such as That sounds good.",
      },
      {
        heading: "Watch out",
        text: "Use sounds good for a singular idea: That sounds good. With plural plans, use sound: Those plans sound good.",
      },
    ];

    for (const surface of ["discovery", "saved"]) {
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await expect(
        main.getByRole("heading", { name: "sounds good", level: 1 }),
      ).toBeVisible();
      await expect(notes.getByRole("heading", { name: "Often used with", exact: true })).toBeHidden();
      await tips.focus();
      await page.keyboard.press("Enter");
      await expect(tips).toBeFocused();
      await expect(notes.getByRole("heading", { level: 4 })).toHaveText(
        expectedNotes.map((note) => note.heading),
      );
      for (const note of expectedNotes) {
        await expect(
          notes
            .getByRole("heading", { name: note.heading, exact: true })
            .locator("..")
            .getByText(note.text, { exact: true }),
        ).toBeVisible();
      }
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      await page.screenshot({
        path: testInfo.outputPath(`usage-notes-${surface}-${theme}.png`),
        fullPage: true,
      });

      if (surface === "discovery") {
        await main
          .getByRole("button", { name: /^Save sounds good:/ })
          .click();
        await expect(
          main.getByRole("button", {
            name: "Remove sounds good from saved words",
          }),
        ).toBeVisible();
        await page.goto("/words");
        await main
          .getByRole("link", {
            name: "Open sounds good details and sentence practice",
            exact: true,
          })
          .click();
        await expect(page).toHaveURL(/\/words\/uw-/);
      }
    }
  });
}
