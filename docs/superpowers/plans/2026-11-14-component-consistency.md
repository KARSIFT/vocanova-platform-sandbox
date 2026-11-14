# Component consistency implementation plan

**Goal:** Consolidate Vocanova controls, panels and feedback into reusable components, migrate existing screens, and provide a browser gallery for owner and agent review.

**Authority:** Owner requested all proposed consolidation and gallery work, with feedback after staging and production delivery. Existing Warm guided coach direction applies. Work in the established checkout on one new delivery branch; no additional design approval round is needed.

**Architecture:** Keep React server-compatible native-element wrappers, existing tokens and primary/secondary action classes. Preserve element semantics, native validation, refs, event handlers, IDs, server-confirmed progress and idempotent writes. Domain-specific interactions remain in feature components.

**Constraints:** No new production dependencies. No navigation or learning-model changes. Local ports 3100/8180; preserve Vergecloud 3000/8080 and prototype 3102. CI and production port defaults remain unchanged. No personal data or production secrets in tests/gallery. Root coordinates builds and Git.

## Tasks

- [x] Shared controls: Button, ActionLink, TextInput, TextArea and Select in ui/controls.tsx with scoped field styles. Migrate authentication, onboarding, settings and lists; retain form behavior and labels.
- [x] Shared panels/feedback: preserve Surface public API, add shared StatusMessage with explicit semantic role and tone. Migrate learning/domain buttons and reusable panel/status patterns where semantics match. Keep specialized selectors, icon controls and graded choices explicit.
- [x] Development gallery: /component-gallery only when VOCANOVA_COMPONENT_GALLERY=true on the server; otherwise 404. Interactive synthetic examples show controls, loading/disabled, validation, panels, feedback, typography and real learning guidance/media controls. No backend writes; no navigation entry. Document component APIs and intentional exceptions.
- [x] Verify native submission, disabled controls, labels, focus, long text, light/dark, 200% text, 360/430/desktop; run focused existing journeys and web checks. Independent whole-diff review and fix findings.
- [ ] Record exact evidence and limitations, commit integrated result, create one draft PR with hold, inspect CI/reviews, use merge queue then strict staging acceptance and manual production deployment after readiness.

## Review focus

Native button types and pending locks; forwarded refs/aria/data attributes; error announcements without duplicate live regions; component defaults vs authored styles; gallery unavailable without explicit server configuration. Test control semantics in browser and relevant real feature journeys rather than static implementation-mirroring tests.
