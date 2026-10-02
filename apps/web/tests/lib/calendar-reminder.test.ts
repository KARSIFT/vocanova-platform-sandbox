import assert from "node:assert/strict";
import { describe, it } from "node:test";
import {
  createCalendarReminder,
  escapeCalendarText,
  foldCalendarLine,
  localCalendarDate,
  validateReminderDateTime,
  type CalendarReminderInput,
} from "../../src/lib/calendar-reminder";

const now = new Date("2026-10-02T10:00:00Z");
const input: CalendarReminderInput = {
  startDate: "2030-03-10",
  time: "19:30",
  uid: "636d7465-24e8-43cc-a913-e1d4a42d495f",
  appOrigin: "https://vocanova.example",
  updatedAt: now,
};
function properties(calendar: string): string[] {
  return calendar
    .replace(/\r\n[ \t]/g, "")
    .split("\r\n")
    .filter(Boolean);
}

describe("optional calendar reminder", () => {
  it("exports one daily floating-local event with an at-start display alert and no account data", () => {
    const calendar = createCalendarReminder(input, now);
    const lines = properties(calendar);
    assert.equal(lines[0], "BEGIN:VCALENDAR");
    assert.equal(lines.at(-1), "END:VCALENDAR");
    assert.equal(lines.filter((line) => line === "BEGIN:VEVENT").length, 1);
    for (const property of [
      "VERSION:2.0",
      "DTSTART:20300310T193000",
      "DTSTAMP:20261002T100000Z",
      "DURATION:PT10M",
      "RRULE:FREQ=DAILY",
      "ACTION:DISPLAY",
      "TRIGGER:PT0S",
      "URL:https://vocanova.example/plan",
      "TRANSP:TRANSPARENT",
    ])
      assert.ok(lines.includes(property), property);
    assert.doesNotMatch(
      calendar,
      /TZID|VTIMEZONE|ATTENDEE|ORGANIZER|mailto:|ACTION:EMAIL|BEGIN:VTODO/,
    );
    assert.equal(calendar.endsWith("\r\n"), true);
    assert.equal(calendar.replaceAll("\r\n", "").includes("\n"), false);
    for (const line of calendar.split("\r\n"))
      assert.ok(Buffer.byteLength(line, "utf8") <= 75);
  });

  it("retains the draft identity and deterministic bytes while supporting a revised schedule", () => {
    const original = createCalendarReminder(input, now);
    assert.equal(createCalendarReminder(input, now), original);
    const revised = properties(
      createCalendarReminder(
        {
          ...input,
          time: "08:15",
          sequence: 1,
          updatedAt: new Date("2026-10-02T10:05:00Z"),
        },
        now,
      ),
    );
    assert.equal(
      revised.find((line) => line.startsWith("UID:")),
      properties(original).find((line) => line.startsWith("UID:")),
    );
    assert.ok(revised.includes("SEQUENCE:1"));
    assert.ok(revised.includes("DTSTART:20300310T081500"));
    assert.ok(revised.includes("DTSTAMP:20261002T100500Z"));
  });

  it("keeps the chosen clock time through UTC, fractional-offset and daylight-saving device zones", () => {
    const previous = process.env.TZ;
    try {
      for (const zone of ["UTC", "Asia/Tehran", "America/New_York"]) {
        process.env.TZ = zone;
        const calendar = properties(createCalendarReminder(input, now));
        assert.ok(calendar.includes("DTSTART:20300310T193000"), zone);
        assert.ok(calendar.includes("RRULE:FREQ=DAILY"), zone);
        const localNow = new Date(2030, 2, 10, 19, 30, 1);
        assert.equal(localCalendarDate(localNow), "2030-03-10");
        assert.match(
          validateReminderDateTime("2030-03-10", "19:30", localNow) ?? "",
          /later than now/,
        );
        assert.equal(
          validateReminderDateTime("2030-03-10", "19:31", localNow),
          null,
        );
      }
    } finally {
      if (previous === undefined) delete process.env.TZ;
      else process.env.TZ = previous;
    }
  });

  it("accepts leap dates but rejects malformed or impossible dates and times", () => {
    assert.equal(validateReminderDateTime("2028-02-29", "00:00", now), null);
    for (const startDate of [
      "2027-02-29",
      "2100-02-29",
      "2030-04-31",
      "2030-13-01",
      "2030-01-00",
      "2030-1-01",
      "2030-01-01\nEND:VEVENT",
    ]) {
      assert.match(
        validateReminderDateTime(startDate, "19:30", now) ?? "",
        /valid start date/,
      );
    }
    for (const time of [
      "24:00",
      "12:60",
      "9:30",
      "10:00Z",
      "10:00\nATTENDEE:x",
    ]) {
      assert.match(
        validateReminderDateTime("2030-01-01", time, now) ?? "",
        /valid reminder time/,
      );
    }
    assert.match(
      validateReminderDateTime("2020-01-01", "19:30", now) ?? "",
      /later than now/,
    );
    assert.match(
      validateReminderDateTime("2030-01-01", "19:30", new Date("invalid")) ??
        "",
      /device date/,
    );
  });

  it("rejects injected identities and non-origin or unsafe links", () => {
    assert.throws(
      () =>
        createCalendarReminder(
          { ...input, uid: "x\r\nATTENDEE:other@example.test" },
          now,
        ),
      /identity/,
    );
    for (const appOrigin of [
      "javascript:alert(1)",
      "https://example.test/private",
      "https://user:password@example.test",
      "http://example.test",
      "https://example.test/?secret=x",
    ]) {
      assert.throws(() => createCalendarReminder({ ...input, appOrigin }, now));
    }
    assert.ok(
      properties(
        createCalendarReminder(
          { ...input, appOrigin: "http://localhost:3000" },
          now,
        ),
      ).includes("URL:http://localhost:3000/plan"),
    );
    assert.throws(
      () => createCalendarReminder({ ...input, sequence: -1 }, now),
      /revision/,
    );
    assert.throws(
      () =>
        createCalendarReminder(
          { ...input, updatedAt: new Date("invalid") },
          now,
        ),
      /timestamp/,
    );
  });

  it("escapes text and folds long Unicode content without splitting or creating properties", () => {
    const text =
      "Read, repeat; relax\\return\r\nATTENDEE:unexpected\n" + "é🙂".repeat(40);
    const escaped = escapeCalendarText(text);
    assert.equal(escaped.includes("\n"), false);
    assert.ok(
      escaped.includes(
        "Read\\, repeat\\; relax\\\\return\\nATTENDEE:unexpected\\n",
      ),
    );
    const folded = foldCalendarLine("DESCRIPTION:" + escaped);
    for (const line of folded.split("\r\n")) {
      assert.ok(Buffer.byteLength(line, "utf8") <= 75);
      assert.equal(line.includes("\uFFFD"), false);
    }
    assert.equal(folded.replace(/\r\n /g, ""), "DESCRIPTION:" + escaped);
    assert.equal(properties(folded).length, 1);
  });
});
