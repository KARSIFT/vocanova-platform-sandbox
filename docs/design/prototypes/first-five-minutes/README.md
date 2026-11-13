# First five minutes — interactive design comparison

Three visual directions for the same Vocanova journey. These are disposable
design previews; the production app does not load them. No real accounts, audio,
API calls, analytics, saved progress or AI evaluation are involved.

From the repository root:

```sh
python3 -m http.server 3100 --bind 127.0.0.1 --directory docs/design/prototypes/first-five-minutes
```

Open one direction:

- [Calm dictionary](http://127.0.0.1:3100/?direction=dictionary)
- [Warm guided coach — recommended for discussion](http://127.0.0.1:3100/?direction=coach)
- [Visual situations](http://127.0.0.1:3100/?direction=situations)

Use **Compare designs and preview states** above the product to switch directions,
themes or scenarios. Collapse it before a learner's first attempt. Scenario values
are `new`, `returning`, `completed`, `unavailable`; `theme=dark` opens dark mode.
Restart clears the preview interaction state. Reload also resets it; drafts stay
only in the open page's memory. No account data is read or changed.

Start with Find your first word, discover invite, Save for review, Review ready
words, Show meaning, choose one of the existing four review ratings, write a
sentence, inspect the explicitly fixed example feedback and return Home. The
sample review queue contains only invite. A recorded preview review cannot be
repeated in that preview day. Production scheduling and confirmation are not
implemented here.

The feedback is **for the separate displayed example**, not the visitor's draft.
The original approved invite illustration is copied from the app's public assets.
Its meaning and context remain intact; this work does not replace catalog content.

See the [design proposal](../../../superpowers/specs/2026-11-13-first-five-minutes-design.md)
and [learner protocol](../../../product/first-five-minutes-study-2026-11-13.md).
The owner will select a direction before production implementation. No learner
sessions or design winner are claimed.
