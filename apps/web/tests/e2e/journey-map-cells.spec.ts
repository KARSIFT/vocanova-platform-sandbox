import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

test("the Journey map shows all four states at their ranked positions", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_knowledge_map", value: "mixed-progress", url: baseURL! },
  ]);
  await page.goto("/discover?band=1");
  const map = page.getByRole("region", { name: "Your word map" });
  await expect(map.getByRole("list", { name: "Map progress" })).toHaveText(
    /1 already know\s*1 learning\s*2 learned\s*9,996 new/,
  );
  const first =
    page.viewportSize()!.width >= 640
      ? map.getByRole("link", {
          name: "Words 1 to 100, 1 already know, 1 learning, 1 learned, 97 new",
        })
      : map.locator('[data-current="true"]');
  await expect(first).toBeVisible();
  await expect(first.locator('path[data-state="known"]')).toHaveAttribute(
    "d",
    "M1 1h8v8h-8z",
  );
  await expect(first.locator('path[data-state="learning"]')).toHaveAttribute(
    "d",
    "M21 1h8v8h-8z",
  );
  await expect(first.locator('path[data-state="learned"]')).toHaveAttribute(
    "d",
    "M11 11h8v8h-8z",
  );
  const second = first.locator("+ *");
  await expect(second.locator('path[data-state="learned"]')).toHaveAttribute(
    "d",
    "M41 1h8v8h-8z",
  );
  const range = page.getByRole("region", { name: "Words 1–100", exact: true });
  await range.getByText("Words in this range", { exact: true }).click();
  await expect(
    range.getByRole("link", { name: "12 Map word 12 Learned", exact: true }),
  ).toBeVisible();
});

test("Journey shows aggregate saved progress when word states are unavailable", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "e2e_knowledge_map", value: "legacy-progress", url: baseURL! },
  ]);
  await page.goto("/discover?band=1");
  const legend = page.getByRole("list", { name: "Map progress" });
  await expect(legend).toHaveText(/1 already know\s*3 saved\s*9,996 new/);
  const range = page.getByRole("region", { name: "Words 1–100", exact: true });
  await expect(range.getByText("1 known · 2 saved · 97 new")).toBeVisible();
});
