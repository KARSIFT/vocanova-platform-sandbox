# VOC-031 — P5 Integrated Core Loop Staging Evidence

## Scope and authority

This document is the `T09` in-repository evidence index for the P5
integrated-core-loop gate (DOC-12 §5 P5). It is a **draft blueprint**
written before adoption and before any `T00`–`T09` implementation: every
row below is `pending`, not `Produced`. It is updated at `T09`
implementation time to record the real evidence the merged `T00`–`T09` PRs
actually produced, and the staged exercises and rollback rehearsal that can
only run once F3 staging exists. This draft text is superseded at that
point, the same way VOC-030's own draft staging-evidence text was
superseded at its `T06`.

## Planned in-repository evidence (to be produced by `T00`–`T09`)

| Evidence | Requirement | Status at draft time |
| --- | --- | --- |
| `EV-00` | `user_onboarding_profiles` migration invariants | pending — `T00` |
| `EV-01` | Migration compatibility, no A1/P1/P2/P3/P4 regression | pending — `T00` |
| `EV-02`–`EV-05` | Onboarding-seeds-`user_settings` logic (`D07`) | pending — `T00`, blocked on `D07` |
| `EV-06`–`EV-11` | Onboarding API and frontend flow | pending — `T01` |
| `EV-12`–`EV-19` | Settings/account write endpoints (`D06`) | pending — `T02`, blocked on `D06` |
| `EV-20`–`EV-23` | Settings/account frontend wiring | pending — `T03` |
| `EV-24`–`EV-28` | Cross-feature integration and navigation consistency | pending — `T04` |
| `EV-29`–`EV-34` | Reliability and recovery across the loop | pending — `T05` |
| `EV-35`–`EV-38` | Accessibility automation and WCAG 2.2 AA pass (`D03`) | pending — `T06` |
| `EV-39`–`EV-41` | Performance automation and Lighthouse thresholds (`D04`) | pending — `T07` |
| `EV-42`–`EV-44` | Final UX consistency, no route renamed (`D08`) | pending — `T08` |
| `EV-45` | Installed deterministic and security suite | pending — `T09` |
| `EV-46` | Staging: full onboard-through-settings loop | **Blocked by `VOC-031-DEP-02`** — F3 staging does not exist |
| `EV-47` | New-table rollback rehearsal | **Blocked by `VOC-031-DEP-02`** — F3 staging does not exist |
| `EV-48` | Exact-SHA independent verification (per PR) | Produced per-PR by Claude Code, not the implementer, as each PR lands |

## Staging exercise plan (blocked by F3 + `D06`–`D08`)

Once `VOC-031-DEP-02` is resolved and a non-production F3 environment is
available, the following exercises must be executed and their results
appended to this document or linked from PR evidence.

### `EV-46` — Full onboard-through-settings loop, at all three supported layouts

1. With a non-production identity that has not completed onboarding,
   complete the DOC-03 §3 question flow and verify redirect into `/home`.
2. Discover and save a word; verify it is reachable from the saved-words
   presentation on Home/Progress.
3. Submit reviews until the daily mission target is met; verify the mission
   completes and the streak advances (P4 behavior, unchanged by this
   package).
4. Submit a sentence for AI feedback from the review-completion flow;
   verify feedback renders.
5. Visit `/progress` and verify Confidence Points, streak, and completion
   history render and agree with what Home showed during the session.
6. Visit `/settings`, change the daily review target and a notification
   toggle, save, and verify the change persists on reload.
7. Visit `/settings/account`, change the display name, save, and verify the
   change persists and renders correctly elsewhere it is shown (if
   anywhere in this package's scope).
8. Repeat steps 1–7 at 360px, 430px, and a desktop width ≥1024px
   (`VOC-031-D05`).
9. Repeat a representative subset with a simulated mid-flow network failure
   (per `T05`) and verify recovery without duplication or data loss.

### `EV-47` — New-table rollback rehearsal

1. Record current `user_onboarding_profiles` and `user_settings` state for
   a test learner.
2. Apply the VOC-031 build and migration in staging.
3. Complete onboarding and update settings/account for that learner, then
   perform a rollback to the previously known-good revision.
4. Verify that:
   - `user_settings`/`users.display_name` state set before the rollback is
     preserved.
   - The P1–P4 write paths remain functional after rollback.
   - No orphaned reference to the rolled-back `user_onboarding_profiles`
     table causes an error in any P1–P4 code path (none should reference
     it).

## `D06`–`D08` interpretation notes

Per whatever the founder actually resolves `D06` (settings/account
write-field boundary) and `D07` (onboarding-seeding rule) into, this
section must be updated at `T09` time to record the resolution actually
implemented, the same way VOC-030's staging evidence recorded its `D01`–`D05`
resolutions. Per the proposed `D08` default, the staging exercise above
does not include any route-rename verification, since none is planned by
this package.

## Cross-capability consistency check (F3-blocked)

1. Load Home and the Review screen for the same non-production identity in
   the same session.
2. Verify the due-review count shown on each is identical
   (`VOC-031-AC-05`).

**In-repository evidence for this consistency property** is planned as
`VOC-031-EV-26` (`T04`) — a shared-data-source assertion analogous to
`VOC-030-EV-28`'s Home/Progress streak consistency.

## Rollback triggers

Per `VOC-031` implementation-plan §Deployment and rollback / release-plan
§Rollback, initiate rollback on:

- False onboarding/settings/account state reaching a learner.
- Suspected cross-user exposure of onboarding/settings/account data.
- A confirmed CSRF or duplicate-write defect on either new write surface.
- A regression in any underlying P1–P4 write path or screen this package's
  reliability/UX work touches.
- An accessibility or performance threshold regression reaching production.
- Migration or schema failure.

## Rollback procedure

1. Preserve `user_onboarding_profiles`/`user_settings`/`users.display_name`
   state.
2. Restore the pre-P5 P1–P4 behavior cleanly.
3. Revert the deployment to the last-known-good revision.
4. Validate with non-production identities.
5. Record the last-known-good revision and the rollback reason.

## Limitations / open dependencies

- **`VOC-031-DEP-02`**: F3 staging does not exist, so `EV-46`/`EV-47` cannot
  be run live. Procedures are documented; live execution is recorded as
  blocked.
- **`VOC-031-DEP-01`**: no prior milestone (A1/P1/P2/P3/P4) has itself
  closed its own DOC-12 live-staging gate yet, for the same F3 reason; this
  package's own gate additionally cannot be declared complete before those
  upstream gates are.
- **`D06`–`D08`**: open founder decisions, not yet resolved at draft time.
- **Accessibility/performance automation itself does not yet exist** at
  draft time (`VOC-031-D00`) — `T06`/`T07` install it; until then, no
  automated accessibility or performance evidence can be produced at all,
  which is a stronger limitation than a prior milestone's "automation not
  yet implemented" note (VOC-030's equivalent limitation), since here even
  the tooling is new.

## P5 gate readiness (draft-time status: not started)

Per the DOC-12 §5 P5 gate wording: **"the full loop works coherently in
staging across supported layouts with no critical product/security/data/
accessibility/reliability defect."**

At draft time, none of this package's tasks have been implemented — this
section records the plan, not a result. Once `T00`–`T09` land:

- **Cross-feature integration** — evidenced by `EV-24..EV-28`.
- **Reliability/recovery** — evidenced by `EV-29..EV-34`.
- **Accessibility** — evidenced by `EV-35..EV-38`.
- **Performance** — evidenced by `EV-39..EV-41`.
- **Final UX consistency** — evidenced by `EV-42..EV-44`.
- **Works coherently in staging across supported layouts** — evidenced by
  `EV-46`, blocked until F3 exists.

The P5 gate **cannot be declared complete by this work alone** because:

1. Live staging evidence (`EV-46`, `EV-47`) is blocked by the missing F3
   environment (`VOC-031-DEP-02`), and the P5 gate's own wording is
   centrally about staging behavior, not one evidence item among several.
2. The upstream P1→P2→P3→P4 chain has not itself closed its own live-staging
   gate yet (`VOC-031-DEP-01`).
3. Exact-SHA independent verification of every PR (`T00`..`T09`) is Claude
   Code's responsibility (`EV-48`); this implementer does not self-approve.
4. Production deployment is separately governed (A-003 §11/12); R3/R4
   founder authority, RL1/RL2 technical activation, and autonomous
   production release are all still disabled.

## Follow-up work

- Execute `EV-46`/`EV-47` once F3 staging exists.
- R1: fold P5's evidence into staging-readiness validation for the release
  candidate.
- A future package: resolve `VOC-031-D08`'s recorded routing drift, either
  by amending DOC-08 or by a dedicated, independently reviewed rename.
