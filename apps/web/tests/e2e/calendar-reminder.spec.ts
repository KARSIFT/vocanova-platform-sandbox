import { randomUUID } from "node:crypto";
import { readFile } from "node:fs/promises";
import { expect, test, type Download, type Page } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

function properties(content: string) {
  return content
    .replace(/\r\n[ \t]/g, "")
    .split("\r\n")
    .filter(Boolean);
}
async function readCalendar(download: Download) {
  expect(download.suggestedFilename()).toBe("vocanova-daily-practice.ics");
  expect(await download.failure()).toBeNull();
  const path = await download.path();
  if (!path) throw new Error("The calendar download was not saved");
  return readFile(path, "utf8");
}
async function downloadCalendar(page: Page) {
  const download = page.waitForEvent("download");
  await page
    .getByRole("button", { name: "Download calendar reminder", exact: true })
    .click();
  return readCalendar(await download);
}

test.beforeEach(async ({ page, context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
  await page.clock.setFixedTime(new Date("2030-01-15T12:00:00Z"));
  await page.addInitScript(() => {
    const state = window as typeof window & {
      reminderPermissionRequests: number;
    };
    state.reminderPermissionRequests = 0;
    if ("Notification" in window) {
      Object.defineProperty(Notification, "requestPermission", {
        configurable: true,
        value: async () => {
          state.reminderPermissionRequests += 1;
          return "denied";
        },
      });
    }
  });
});

for (const theme of ["light", "dark"] as const) {
  test(
    "downloads an explicit local-calendar reminder without changing account preferences (" +
      theme +
      ")",
    async ({ page, context, baseURL }, testInfo) => {
      await context.addCookies([
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      const mutations: string[] = [];
      page.on("request", (request) => {
        if (
          new URL(request.url()).pathname.startsWith("/api/v1/") &&
          ["POST", "PUT", "PATCH", "DELETE"].includes(request.method())
        ) {
          mutations.push(
            request.method() + " " + new URL(request.url()).pathname,
          );
        }
      });
      const mockAPI = "http://127.0.0.1:" + (process.env.MOCK_API_PORT ?? 8080);
      const before = await (
        await context.request.get(mockAPI + "/api/v1/settings")
      ).json();
      await page.goto("/settings#practice-reminder");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await expect(page).toHaveTitle(/Settings/);
      const section = page.getByRole("region", {
        name: "Calendar reminder",
        exact: true,
      });
      await expect(section).toBeVisible();
      await expect(
        page.getByRole("switch", { name: /^Daily review reminder/ }),
      ).toHaveCount(0);
      await expect(
        section.getByText(/Vocanova does not send email or push reminders/),
      ).toBeVisible();
      await section
        .getByLabel("Start date", { exact: true })
        .fill("2030-02-01");
      await section.getByLabel("Reminder time", { exact: true }).fill("18:45");
      const button = section.getByRole("button", {
        name: "Download calendar reminder",
        exact: true,
      });
      // Native time inputs can expose several keyboard segments before the button.
      for (let tabs = 0; tabs < 5; tabs += 1) {
        await page.keyboard.press("Tab");
        if (
          await button.evaluate((element) => element === document.activeElement)
        )
          break;
      }
      await expect(button).toBeFocused();
      const box = await button.boundingBox();
      expect(box!.width).toBeGreaterThanOrEqual(44);
      expect(box!.height).toBeGreaterThanOrEqual(44);

      const firstEvent = page.waitForEvent("download");
      await page.keyboard.press("Enter");
      const first = await readCalendar(await firstEvent);
      const lines = properties(first);
      expect(lines).toEqual(
        expect.arrayContaining([
          "BEGIN:VCALENDAR",
          "BEGIN:VEVENT",
          "DTSTART:20300201T184500",
          "RRULE:FREQ=DAILY",
          "ACTION:DISPLAY",
          "TRIGGER:PT0S",
          "URL:" + new URL("/plan", baseURL).href,
        ]),
      );
      expect(first).not.toMatch(/TZID|ATTENDEE|ORGANIZER|mailto:/);
      await expect(section.getByRole("status")).toHaveText(
        "Calendar file prepared. Open the download and import it into your calendar to finish setting up your reminder.",
      );
      expect(await downloadCalendar(page)).toBe(first);

      await section.getByLabel("Reminder time", { exact: true }).fill("19:00");
      await expect(section.getByRole("status")).toHaveCount(0);
      const revised = properties(await downloadCalendar(page));
      expect(revised.find((line) => line.startsWith("UID:"))).toBe(
        lines.find((line) => line.startsWith("UID:")),
      );
      expect(revised).toContain("SEQUENCE:1");
      expect(revised).toContain("DTSTART:20300201T190000");
      await section.getByText("After downloading", { exact: true }).click();
      await expect(
        section.getByText(
          /Import only once: importing again can create duplicates/,
        ),
      ).toBeVisible();
      await expect(
        section.getByText(
          /Your calendar controls alerts, daylight-saving behavior/,
        ),
      ).toBeVisible();

      expect(mutations).toEqual([]);
      expect(
        await page.evaluate(
          () =>
            (window as typeof window & { reminderPermissionRequests: number })
              .reminderPermissionRequests,
        ),
      ).toBe(0);
      const after = await (
        await context.request.get(mockAPI + "/api/v1/settings")
      ).json();
      expect(after).toEqual(before);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      const scan = await scanForAxeViolations(page);
      expect(
        scan.criticalOrSerious,
        formatViolations(scan.criticalOrSerious).join("\n"),
      ).toEqual([]);
      await section.screenshot({
        path: testInfo.outputPath("calendar-reminder-" + theme + ".png"),
      });
    },
  );
}

test("rejects an elapsed time and a failed file preparation without claiming a reminder is active", async ({
  page,
}) => {
  await page.goto("/settings#practice-reminder");
  const section = page.getByRole("region", {
    name: "Calendar reminder",
    exact: true,
  });
  const date = await page.evaluate(() => {
    const now = new Date();
    return [
      now.getFullYear(),
      String(now.getMonth() + 1).padStart(2, "0"),
      String(now.getDate()).padStart(2, "0"),
    ].join("-");
  });
  await section.getByLabel("Start date", { exact: true }).fill(date);
  await section.getByLabel("Reminder time", { exact: true }).fill("00:00");
  const downloads: Download[] = [];
  page.on("download", (download) => downloads.push(download));
  await section
    .getByRole("button", { name: "Download calendar reminder", exact: true })
    .click();
  await expect(section.getByRole("alert")).toHaveText(
    "Choose a start date and time later than now on this device.",
  );
  expect(downloads).toEqual([]);
  await expect(section.getByRole("status")).toHaveCount(0);

  await section.getByLabel("Start date", { exact: true }).fill("2030-02-01");
  await page.evaluate(() => {
    URL.createObjectURL = () => {
      throw new Error("Synthetic download preparation failure");
    };
  });
  await section
    .getByRole("button", { name: "Download calendar reminder", exact: true })
    .click();
  await expect(section.getByRole("alert")).toHaveText(
    "We could not prepare the calendar file. Please try again.",
  );
  expect(downloads).toEqual([]);
  await expect(section.getByRole("status")).toHaveCount(0);
});
