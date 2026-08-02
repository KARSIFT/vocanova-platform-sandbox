# VOC-038 — Specification: Milestone L1, Controlled Launch

## 1. Source objective and gate (DOC-12 §5, quoted verbatim)

> **L1 — Controlled Launch.** Objective: release to a limited audience, monitor, expand only
> with evidence. Rollout order: deploy with risky features disabled where appropriate → smoke
> tests → founder/internal allowlist → validate non-AI core loop → enable AI for the
> allowlisted cohort → monitored limited cohort → gradual expansion only after thresholds pass
> → pause/rollback immediately on a trigger (cross-user exposure, auth failure, unsafe AI
> feedback, injection exposing protected info, migration-caused inconsistency, unreliable
> mission/progress state, unacceptable error-rate/latency, material quality-regression reports,
> incorrect provider privacy config, AI cost overrun, insufficient monitoring). **Gate:**
> governed release running in production, controlled audience completes the core loop,
> monitoring functions, no launch-blocking incident remains, rollback controls are proven,
> founder records hold/expand decision. General public availability is a separate, later
> expansion decision.

## 2. What already exists that L1 can build on (from R2)

R2 (VOC-037) already built the exact mechanisms L1's rollout order depends on:

- **"deploy with risky features disabled"** — the four kill switches
  (`EMAIL_MAGIC_LINK_ENABLED`, `GOOGLE_OAUTH_ENABLED`, `NEW_USER_SIGNUP_ENABLED`,
  `AI_FEATURES_ENABLED`) are real, independently verified against production at the HTTP
  surface (`VOC-037-EV-03`), and `deploy-production.yml` already writes them to their safe
  (mostly `false`) defaults on every deploy.
- **"pause/rollback immediately on a trigger"** — a genuine two-artifact rollback rehearsal
  against real production is documented (`VOC-037-EV-03`).
- **"monitoring functions"** — Sentry (real test event verified) and self-hosted Uptime Kuma
  with a live down/up alert rehearsal to Telegram are both active (`VOC-037-EV-04`).

What does **not** yet exist, and is genuinely new L1 scope:

- A **cohort/allowlist mechanism** — right now `NEW_USER_SIGNUP_ENABLED=false` blocks all new
  signups uniformly; there is no way to admit a named small group while keeping signup closed
  to everyone else. L1 needs this distinction (`T00`).
- A **production smoke-test suite** — no file matching `*smoke*` exists in the repository as of
  this drafting pass. R1/R2 verification so far has been manual `curl`/SSH checks, sufficient
  for one-time verification but not for "deploy with risky features disabled → smoke tests" as
  a repeatable step (`T02`).
- **Expansion thresholds** — DOC-12's trigger list (cross-user exposure, auth failure, unsafe
  AI feedback, etc.) names *what* forces a pause, but no document yet states the *quantitative*
  thresholds (e.g., what error rate, what latency, what AI cost per day) that gate moving from
  "monitored limited cohort" to wider release (`T05`).

## 3. Open questions (genuine founder decisions, not filled in by this package)

**Open question 1 — initial allowlist composition.** Who is actually in the founder/internal
allowlist for the first cohort? Candidates in increasing order of exposure: founder-only,
founder + a small named group of trusted testers, founder + testers + a small waitlist. This
package does not choose one; `T00` is scoped as a decision-record-only task, exactly like
VOC-037's `T00`/`T02` were for the production-hosting and legal-content decisions.

**Open question 2 — expansion thresholds.** What error-rate, latency, and AI-cost-per-day
numbers, sustained over what window, justify moving from the initial cohort to the next
expansion tranche? This is a business/risk-tolerance decision, not a technical one — `T05` is
scoped as a decision-record-only task for the same reason.

## 4. Non-goals

- General public availability (DOC-12 §5 explicitly separates this out as "a separate, later
  expansion decision" — out of scope for L1 and therefore for this package).
- Any change to the AI feature set, review/spaced-repetition logic, or UI beyond what's needed
  to support cohort gating and smoke testing.
