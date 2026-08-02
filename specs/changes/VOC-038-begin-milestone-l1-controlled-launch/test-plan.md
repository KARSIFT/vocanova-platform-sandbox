# VOC-038 — Test Plan

Not authorized by this package. Once adopted:

- `T02`'s smoke-test suite becomes the primary automated coverage for this milestone and should
  be added to the `scripts/foundation/*.test.mjs` auto-discovery convention where the check is
  static (e.g. kill-switch state assertions), and as a separate scripted runner where it needs
  live HTTP calls against a real deployed environment (mirroring how `VOC-037-EV-03`'s
  verification was performed manually — this task exists specifically to stop that from staying
  manual).
- `T01`/`T04` require the same live-HTTP-surface verification discipline used throughout R2: no
  claim of "works" without an actual request/response against production.
- `T06`'s rollback rehearsal must use a genuinely different artifact (not a label swap), per the
  same standard `VOC-037-EV-03` set.
