# Discovery usability implementation plan

**Goal:** Make discovery easy to navigate and hear, with concise learner-facing presentation.
**Architecture:** Retain Next/React and existing server-confirmed learning mutations. Navigation uses explicit range/rank URLs and never records knowledge. Canonical word details load on demand. Shared speech retains meaning/revision correctness and bounded cancellation.
**Execution:** Parallel bounded ownership; the owner explicitly requests autonomous execution and has opted out of further questions.

## Accepted direction and implementation choices

The owner's seven annotated examples and request define acceptance: clickable word progress, previous word, range switching, direct atlas entry, aligned controls, a recognizable settings gear, useful word details instead of repeated sources, and less text across related screens. Existing blue/lavender tokens and three primary destinations remain.

Use a compact word picker rather than 100 tiny interactive buttons. It provides ranked word selection with readable state, keyboard access and mobile targets. Map selection enters the actual selected word directly; generic discovery entry can retain the first-use Start gesture. Explicit range/word navigation must honor completed ranges and preserve URL state on reload/Back. Audio autoplay remains subject to browser gesture policy, with an explicit Listen control.

Keep short labels and accessible names; use standard icons for familiar controls. Move detailed teaching/help and catalog credits to optional disclosures or their dedicated destination. Do not remove attribution from the product or hide errors/privacy guidance. Follow the same simplification principle on adjacent Home/Practice presentation where supported by inspection.

References: repository learning-workspace and reference-video-review; Duolingo's published core-tabs redesign and home-screen rationale. No copied artwork/characters or new icon library.

## Task 1 — Discovery and range navigation (manual_audit_map)

Files: discovery/discovery.tsx, discovery/page.tsx, knowledge-map/map-view.tsx, knowledge-map/workspace.module.css; focused E2E navigation specs.

- [x] Pin current failures: decorative strip, no previous word, atlas detour, assessed-range redirect, action misalignment.
- [x] Add accessible word/range pickers and previous-word navigation, locked while a decision is unresolved.
- [x] Enter active discovery directly from atlas selection; preserve rank/range/active state and pending request identity.
- [x] Replace discovery source block with lazy canonical word meanings/examples; keep attribution reachable on Journey.
- [x] Test no navigation POSTs, bounds, completed ranges, keyboard/focus and meaningful error recovery.

## Task 2 — Pronunciation reliability and catalog evidence (backend_audit)

Files: speech/neural helpers, ui/pronunciation.tsx if needed, dedicated audio tests; API package only after root coordination.

- [x] Reproduce request deadline mismatch, provider/media failures and cancellation in deterministic tests.
- [x] Separate request and media-start deadlines; retain sense/revision matching and stale-response cancellation.
- [x] Audit actual catalog audio/image/definition coverage with safe aggregate evidence; distinguish metadata, cached audio and audible playback.
- [x] Identify an affordable durable coverage path; do not silently substitute a wrong heteronym or invent media availability.

### Catalog audio delivery extension

Read-only production evidence established only14 current word clips for10,051
meanings. Complete the requirement with an operator-only, default-read-only
command and a durable job/attempt ledger. Reuse canonical utterance/context,
provider/model/voice/instruction hashes and decoded-audio validation. Generation
requires an explicit job identity and immutable attempt/character/observed-audio
limits. Failed and cancelled attempts consume the budget; resume cannot pay for
the same attempt again. Preserve ready cache entries and public daily caps.

Review kill-switch parsing, secret-safe errors, provider identity on historical
receipts, shared cache leases and catalog-change publication before any paid
execution. Verify using disposable PostgreSQL, then a small live pilot before
choosing the bounded backfill run. Token-based billing estimates are not hard
provider spending limits. Report exact current ready/missing counts afterward.

## Task 3 — Shared presentation and integrated acceptance (root)

Files: app-header/nav-icons, Home/adjacent copy, product/design documentation; targeted existing tests.

- [x] Replace malformed gear drawing with balanced standard inline SVG; preserve Settings label, target and focus.
- [x] Simplify repeated copy on adjacent high-use screens, using familiar icons with accessible names and keeping essential learning/error information.
- [x] Review integrated result independently; build once under coordinated local3100/8180 runtime.
- [ ] Finish manual responsive acceptance and hosted performance checks (local370pass/5skip; provider/media remain separate).
- [ ] One consolidated PR with hold, independent review and full relevant CI; normal queue and separate authorized production promotion after staging acceptance.

## Review focus

Unresolved mutation retries must never apply to a different word; navigation must not claim learning; fully assessed ranges remain reachable; untrusted media/details must remain bound to meaning/revision; keyboard/touch targets and long text must stay usable. Missing media/provider failures need truthful recovery rather than silent success.
