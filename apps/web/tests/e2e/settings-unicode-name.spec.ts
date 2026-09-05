import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

for (const letter of ["𐐀", "a"]) {
  test(`saves 80 ${letter} code points and preserves the name after over-limit edits`, async ({
    page,
    context,
    baseURL,
  }) => {
    if (!baseURL) throw new Error("A test app URL is required");
    await context.addCookies([
      {
        name: "vocanova_session",
        value: `unicode-${randomUUID()}`,
        url: baseURL,
      },
      { name: "vocanova_csrf", value: `csrf-${randomUUID()}`, url: baseURL },
    ]);
    await page.goto("/settings");
    const input = page.getByRole("textbox", { name: "Display name" });
    const name = letter.repeat(80);
    await input.fill(name);
    await expect(input).toHaveValue(name);
    await input.fill(`${name}${letter}`);
    await expect(input).toHaveValue(name);
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();
    await page.reload();
    await expect(input).toHaveValue(name);

    await input.fill("");
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();
    await page.reload();
    await expect(input).toHaveValue("");
  });
}
