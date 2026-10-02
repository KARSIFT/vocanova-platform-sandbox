"use client";

import { useEffect, useRef, useState } from "react";
import {
  createCalendarReminder,
  localCalendarDate,
  validateReminderDateTime,
} from "@/lib/calendar-reminder";
import { Surface } from "@/ui/surface";

const inputStyle =
  "mt-2 min-h-12 w-full min-w-0 max-w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-base text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700";

export function CalendarReminder() {
  const [startDate, setStartDate] = useState("");
  const [time, setTime] = useState("19:00");
  const [today, setToday] = useState("");
  const [timezone, setTimezone] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [downloaded, setDownloaded] = useState(false);
  const draft = useRef<{
    uid: string;
    signature: string;
    sequence: number;
    updatedAt: Date;
  } | null>(null);
  const downloads = useRef(new Map<string, ReturnType<typeof setTimeout>>());

  useEffect(() => {
    const now = new Date();
    setToday(localCalendarDate(now));
    const first = new Date(now);
    if (now.getHours() >= 19) first.setDate(first.getDate() + 1);
    setStartDate(localCalendarDate(first));
    try {
      setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone || null);
    } catch {
      setTimezone(null);
    }
    const activeDownloads = downloads.current;
    return () => {
      for (const [url, timer] of activeDownloads) {
        clearTimeout(timer);
        URL.revokeObjectURL(url);
      }
      activeDownloads.clear();
    };
  }, []);

  function changed() {
    setError("");
    setDownloaded(false);
  }

  function download(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setDownloaded(false);
    const now = new Date();
    const validation = validateReminderDateTime(startDate, time, now);
    if (validation) {
      setError(validation);
      return;
    }
    try {
      const signature = startDate + "T" + time;
      if (!draft.current) {
        draft.current = {
          uid: crypto.randomUUID(),
          signature,
          sequence: 0,
          updatedAt: now,
        };
      } else if (draft.current.signature !== signature) {
        draft.current = {
          ...draft.current,
          signature,
          sequence: draft.current.sequence + 1,
          updatedAt: now,
        };
      }
      const content = createCalendarReminder(
        {
          startDate,
          time,
          ...draft.current,
          appOrigin: window.location.origin,
        },
        now,
      );
      const url = URL.createObjectURL(
        new Blob([content], { type: "text/calendar;charset=utf-8" }),
      );
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = "vocanova-daily-practice.ics";
      document.body.append(anchor);
      anchor.click();
      anchor.remove();
      downloads.current.set(
        url,
        setTimeout(() => {
          URL.revokeObjectURL(url);
          downloads.current.delete(url);
        }, 1000),
      );
      setError("");
      setDownloaded(true);
    } catch {
      setError("We could not prepare the calendar file. Please try again.");
    }
  }

  return (
    <Surface
      id="practice-reminder"
      aria-labelledby="calendar-reminder-heading"
      className="mt-6 scroll-mt-6"
    >
      <h2
        id="calendar-reminder-heading"
        className="text-xl font-bold text-neutral-900"
      >
        Calendar reminder
      </h2>
      <p className="mt-3 text-neutral-700">
        Make a little time for English each day. Download a repeating event,
        then import it into the calendar you use.
      </p>
      <p className="mt-2 text-sm text-neutral-600">
        This is optional. Vocanova does not send email or push reminders, and
        downloading does not change your account preferences.
      </p>
      <form
        aria-label="Reminder calendar file"
        onSubmit={download}
        className="mt-4"
      >
        <div className="grid min-w-0 gap-4 sm:grid-cols-2">
          <label
            className="min-w-0 font-semibold text-neutral-900"
            htmlFor="reminder-start-date"
          >
            Start date
            <input
              id="reminder-start-date"
              type="date"
              required
              min={today || undefined}
              value={startDate}
              onChange={(event) => {
                setStartDate(event.target.value);
                changed();
              }}
              className={inputStyle}
            />
          </label>
          <label
            className="min-w-0 font-semibold text-neutral-900"
            htmlFor="reminder-time"
          >
            Reminder time
            <input
              id="reminder-time"
              type="time"
              required
              value={time}
              onChange={(event) => {
                setTime(event.target.value);
                changed();
              }}
              className={inputStyle}
            />
          </label>
        </div>
        <p className="mt-3 text-sm text-neutral-600">
          {timezone ? `Your browser currently uses ${timezone}. ` : ""}The event
          repeats at the chosen clock time in your calendar’s local time.
          Confirm its time zone and alert after importing, including around
          daylight-saving changes.
        </p>
        <button
          type="submit"
          disabled={!startDate}
          className="mt-4 inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-700 px-5 py-3 font-semibold text-white hover:bg-primary-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 disabled:opacity-60"
        >
          Download calendar reminder
        </button>
        {error && (
          <p role="alert" className="mt-3 text-sm text-red-700">
            {error}
          </p>
        )}
        {downloaded && (
          <p role="status" className="mt-3 text-sm text-neutral-700">
            Calendar file prepared. Open the download and import it into your
            calendar to finish setting up your reminder.
          </p>
        )}
      </form>
      <details className="mt-4 rounded-xl border border-neutral-200 bg-neutral-50 px-4">
        <summary className="min-h-12 cursor-pointer py-3 font-semibold text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700">
          After downloading
        </summary>
        <ol className="mb-4 list-decimal space-y-2 pl-5 text-sm text-neutral-700">
          <li>
            Import the .ics file into your calendar. The event repeats daily and
            includes an alert at its start; your calendar may ask you to enable
            or adjust that alert.
          </li>
          <li>
            Check the first event’s time and your calendar’s time zone. Your
            calendar controls alerts, daylight-saving behavior and what happens
            when you travel.
          </li>
          <li>
            To change or stop reminders, edit or delete the repeating event in
            your calendar. Import only once: importing again can create
            duplicates, and changes here do not update an event you already
            imported.
          </li>
        </ol>
      </details>
    </Surface>
  );
}
