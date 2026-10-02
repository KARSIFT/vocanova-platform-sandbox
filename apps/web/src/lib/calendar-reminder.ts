export interface CalendarReminderInput {
  startDate: string;
  time: string;
  /** An opaque UUID, retained for repeated downloads of this draft. */
  uid: string;
  appOrigin: string;
  updatedAt: Date;
  sequence?: number;
}

export function localCalendarDate(now: Date): string {
  return [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, "0"),
    String(now.getDate()).padStart(2, "0"),
  ].join("-");
}

export function validateReminderDateTime(
  startDate: string,
  time: string,
  now: Date,
): string | null {
  if (!Number.isFinite(now.getTime()))
    return "We could not read your device date. Please try again.";
  if (!/^\d{4}-\d{2}-\d{2}$/.test(startDate))
    return "Choose a valid start date.";
  const [year, month, day] = startDate.split("-").map(Number) as [
    number,
    number,
    number,
  ];
  const date = new Date(Date.UTC(year, month - 1, day));
  if (
    year < 1000 ||
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  )
    return "Choose a valid start date.";
  if (!/^(?:[01]\d|2[0-3]):[0-5]\d$/.test(time))
    return "Choose a valid reminder time.";
  const localNow = `${localCalendarDate(now)}T${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;
  if (`${startDate}T${time}` <= localNow)
    return "Choose a start date and time later than now on this device.";
  return null;
}

export function escapeCalendarText(value: string): string {
  return value
    .replace(/\\/g, "\\\\")
    .replace(/\r\n|\r|\n/g, "\\n")
    .replace(/;/g, "\\;")
    .replace(/,/g, "\\,");
}

/** RFC5545 content lines fold at 75 UTF-8 octets, never inside a code point. */
export function foldCalendarLine(value: string): string {
  const encoder = new TextEncoder();
  const lines: string[] = [];
  let line = "";
  let bytes = 0;
  for (const character of value) {
    const length = encoder.encode(character).length;
    if (bytes + length > 75) {
      lines.push(line);
      line = " ";
      bytes = 1;
    }
    line += character;
    bytes += length;
  }
  lines.push(line);
  return lines.join("\r\n");
}

export function createCalendarReminder(
  input: CalendarReminderInput,
  now: Date = new Date(),
): string {
  const error = validateReminderDateTime(input.startDate, input.time, now);
  if (error) throw new Error(error);
  if (!/^[a-f\d]{8}-(?:[a-f\d]{4}-){3}[a-f\d]{12}$/i.test(input.uid))
    throw new Error("Invalid calendar identity.");
  if (!Number.isFinite(input.updatedAt.getTime()))
    throw new Error("Invalid calendar timestamp.");
  const sequence = input.sequence ?? 0;
  if (!Number.isSafeInteger(sequence) || sequence < 0)
    throw new Error("Invalid calendar revision.");
  const origin = new URL(input.appOrigin);
  const localHost = ["localhost", "127.0.0.1", "[::1]"].includes(
    origin.hostname,
  );
  if (
    origin.username ||
    origin.password ||
    origin.origin !== input.appOrigin ||
    (origin.protocol !== "https:" &&
      !(origin.protocol === "http:" && localHost))
  )
    throw new Error("Invalid app origin.");
  const url = new URL("/plan", origin).href;
  const stamp = input.updatedAt
    .toISOString()
    .replace(/[-:]/g, "")
    .replace(/\.\d{3}Z$/, "Z");
  // Deliberately floating local time: the importing calendar owns its timezone,
  // daylight-saving behavior and alarm policy. Do not convert daily time to UTC.
  const start =
    input.startDate.replaceAll("-", "") +
    "T" +
    input.time.replace(":", "") +
    "00";
  const description =
    "A little English practice: open your Vocanova learning plan.\nThis reminder is managed in your calendar. Check its time zone and alert after import. Edit or delete the repeating event there to change or stop it.";
  return (
    [
      "BEGIN:VCALENDAR",
      "VERSION:2.0",
      "PRODID:-//Vocanova//Practice reminder//EN",
      "CALSCALE:GREGORIAN",
      "BEGIN:VEVENT",
      `UID:${input.uid}@vocanova`,
      `DTSTAMP:${stamp}`,
      `SEQUENCE:${sequence}`,
      `DTSTART:${start}`,
      "DURATION:PT10M",
      "RRULE:FREQ=DAILY",
      "SUMMARY:Practise English with Vocanova",
      `DESCRIPTION:${escapeCalendarText(description)}`,
      `URL:${url}`,
      "TRANSP:TRANSPARENT",
      "BEGIN:VALARM",
      "ACTION:DISPLAY",
      "TRIGGER:PT0S",
      "DESCRIPTION:Time for a little English practice",
      "END:VALARM",
      "END:VEVENT",
      "END:VCALENDAR",
    ]
      .map(foldCalendarLine)
      .join("\r\n") + "\r\n"
  );
}
