# Calendar reminders

Vocanova supports an optional daily practice reminder through a downloadable calendar file. Open **Settings → Calendar reminder**, choose a future start date and time, then download and import the `.ics` file into a calendar you use. The learning plan links to the same section.

The file contains one daily repeating, ten-minute event and a display alert at its start. It links to the learning plan on the Vocanova site that generated the file. It contains no learner name, email, private notes or account identifier. Downloading does not change account preferences, request browser notification permission, subscribe to a service or send a message.

## Time and alerts

The event uses a **floating local time**: its start has neither a UTC suffix nor a fixed timezone. The importing calendar decides how that clock time relates to its timezone, daylight-saving changes and travel. The browser's current timezone is shown as a reference, not as a server-controlled schedule.

After importing, check the first event's date, time and alert in the calendar. Calendar applications can handle imported alerts differently; a generated file alone does not prove that an alert will appear. Review the event around daylight-saving changes and when changing calendar timezone.

Edit or delete the repeating event in the calendar to change or stop the reminder. Changing the form in Vocanova does not update an already imported event. Import only once: repeated imports may produce duplicate events even when the file retains its identifier.

## Consent and stored preferences

The old reminder-preference switch is no longer shown. Its existing `notificationsEnabled` account value remains stored for compatibility and is left untouched by the calendar feature and unrelated settings changes. A legacy true value is **not consent to send reminders**.

Email and push reminders are not delivered by this feature. Any later server-delivered reminder requires a separate explicit opt-in, scheduling and delivery safeguards, and tested opt-out behavior.

## Export behavior and verification

Each form draft uses a random identifier unrelated to the learner. Repeated downloads of the unchanged draft keep the same identifier, timestamp and revision; changing the chosen schedule increments its revision. Reloading the page starts a new draft. Calendar applications decide whether they merge or duplicate imports.

The exporter validates a real future device-local date/time, uses CRLF line endings, escapes calendar text and folds UTF-8 content lines at 75 octets. It uses a fixed learning-plan path on the app's own origin and creates no attendees or email alarms.

Automated coverage checks export bytes, validation, identity and revision stability, local-time behavior across device zones, explicit browser download, unchanged settings, keyboard access, small-screen layout and accessible light/dark presentation. Passing those checks does not establish alarm delivery in a particular calendar application; that requires an actual import and observed alert.

## References

- [iCalendar specification, RFC 5545](https://www.rfc-editor.org/rfc/rfc5545): floating date-time, recurrence, alarms and content-line formatting.
- [Import events into Google Calendar](https://support.google.com/calendar/answer/37118?hl=en).
- [Import calendars in Calendar on Mac](https://support.apple.com/guide/calendar/import-or-export-calendars-icl1023/mac).
