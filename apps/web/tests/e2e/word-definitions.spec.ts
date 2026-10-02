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
