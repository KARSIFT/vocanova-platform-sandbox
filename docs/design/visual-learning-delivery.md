# Visual learning delivery

This phase follows the owner's 5 October 2026 request: a clearer, easier-to-use product, less explanatory copy, and a useful picture for every supported word meaning. WordUp and Duolingo recordings in the private parent workspace remain the main role models. The consolidated delivery is [PR #1486](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1486); production deployment is separate.

## Design decisions

Keep the existing blue identity, system sans-serif, neutral reading surfaces and three primary destinations. Home shows today's action before extra tools; Journey puts lessons before secondary navigation; Progress puts the learner's summary before detailed records. Optional explanations open on demand. Definitions, saving and pronunciation remain easy to find.

WordUp's illustrated teaching around 02:42 and 08:10 informed the relationship between word, meaning, example and picture. Duolingo's exercise screen around 03:10 informed a predictable action area. These observations come from inspected frames and the [reference review](../product/reference-video-review.md), not a continuous audio or animation assessment. No competitor artwork or account imagery is included.

Applied [Anthropic frontend-design](https://github.com/anthropics/skills/tree/main/skills/frontend-design) for intentional hierarchy and restrained copy, and [Vercel web-design-guidelines](https://github.com/vercel-labs/agent-skills/tree/main/skills/web-design-guidelines) for interaction/accessibility review. The repository's existing visual identity and sentence case take precedence.

## Pictures

The canonical seed has 89 words and 92 active meanings. The [picture manifest](../../apps/web/src/lib/meaning-pictures.json) maps exact meaning UUIDs to original scenes, alternative descriptions and local WebP assets. Repeated words such as reservation, follow-up and deadline have separate pictures. No keyword matching or third-party image service is used at runtime.

The owner clarified that the initial collection looked too similar and requested both variety and movement, while preserving its picture shape. All 92 current pictures were revised in the selected mix of realistic photos, illustrations and simple visual diagrams. People, ages, clothes, objects, colors, settings, lighting and viewpoints vary across the collection. Actions and gestures explain concrete meanings; arrows and short sequences explain timing, direction and transactions. Quiet object photographs remain useful for nouns. The [artwork guidance](learning-workspace.md#vocabulary-artwork) applies to the current collection and future generation.

Artwork is generated individually with the built-in image tool, inspected against each exact definition and optimized to 960 × 640 WebP. Prompts prohibit readable text, logos and captions; documents and screens use non-readable marks. The final collection was also reviewed together to detect repetition. Original generation files, rejected attempts, the first collection and machine-specific receipts remain private. Pictures illustrate a possible use of the meaning; they do not replace definitions.

A generated runtime catalogue contains only ready meaning IDs and alt text; editorial definitions and scene briefs stay out of browser bundles. Run `pnpm --filter @vocanova/web generate:meaning-pictures` after an authoring change. Synchronization and exact canonical coverage are checked by tests.

Pictures appear on word detail and lesson teaching steps. Graded recall, typing, listening and context questions do not receive a picture or its answer-bearing alt text. Missing, pending, unknown or failed media leaves the readable teaching content and actions intact. Asset readiness requires visual review; file existence alone is insufficient.

## Acceptance record

- **Revised artwork:** all 92 active canonical meanings have a reviewed, distinct asset and matching alternative description. Exactly 92 WebPs decode at 960 × 640 with distinct SHA-256 values. Total source assets: 5,152,740 bytes; largest: 141,536 bytes. Public copies match every reviewed private asset byte for byte. Canonical coverage tests require every meaning to be ready; unknown and failed media omit the image safely.
- **Source checks:** final integrated workspace validation, runtime catalogue sync, formatting, lint/Go vet, package/web/browser types and package/web/API builds pass. Tests pass: 258 foundation, 52 API-client and 140 web-helper cases. The compact runtime follow-up also passed 12 focused picture browser cases. Authored scene strings occur in none of the final build's 51 browser JavaScript chunks.
- **Browser checks:** the original affected matrix passed 180 cases overall, with three intentional viewport skips. The initial hosted full suite exposed older expectations for copy and examples that now open through disclosures. Those tests were corrected to exercise keyboard opening and retain content and save-state assertions; focused reruns passed 39 and 66 cases across 360px, 430px and desktop. The final integrated full suite is running.
- **Visual review:** inspected mobile Home/Journey/Progress and word teaching plus desktop teaching. Definitions and saving precede the uncropped picture; desktop pictures are capped at 288 CSS pixels high. Secondary explanations and milestones remain reachable through native disclosures. Under the same fixture, 360px Progress is approximately 2,355 CSS pixels tall versus 4,489 before the milestone change.
- **Independent review:** all 92 revised pictures and descriptions were inspected directly, including corrected variants and the complete collection. No actionable meaning or quality findings remain. Interface diffs, exact-meaning lookup, teaching-only placement, failure fallback, compact runtime catalogue, selector corrections and Docker packaging were reviewed independently.
- **Runtime assets:** all 92 revised picture URLs returned HTTP 200 from the local production server and matched source bytes exactly. Docker copies public assets, and the initial hosted smoke passed its source-byte comparison. Fresh hosted gates for the revised head remain pending. Local container/database-backed checks are unavailable because WSL has a Docker client without an engine; hosted checks establish those results.

The final integrated browser suite, hosted CI and consolidated PR review remain pending. Synthetic browser tests do not establish real authentication, persistence, physical speech playback, AI quality or learner effectiveness. Production remains on its previously accepted release until a separate release is verified.
