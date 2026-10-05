# Visual learning delivery

This phase follows the owner's 5 October 2026 request: a clearer, easier-to-use product, less explanatory copy, and a useful picture for every supported word meaning. WordUp and Duolingo recordings in the private parent workspace remain the main role models. This document describes the delivery branch until release evidence is recorded.

## Design decisions

Keep the existing blue identity, system sans-serif, neutral reading surfaces and three primary destinations. Home shows today's action before extra tools; Journey puts lessons before secondary navigation; Progress puts the learner's summary before detailed records. Optional explanations open on demand. Definitions, saving and pronunciation remain easy to find.

WordUp's illustrated teaching around 02:42 and 08:10 informed the relationship between word, meaning, example and picture. Duolingo's exercise screen around 03:10 informed a predictable action area. These observations come from inspected frames and the [reference review](../product/reference-video-review.md), not a continuous audio or animation assessment. No competitor artwork or account imagery is included.

Applied [Anthropic frontend-design](https://github.com/anthropics/skills/tree/main/skills/frontend-design) for intentional hierarchy and restrained copy, and [Vercel web-design-guidelines](https://github.com/vercel-labs/agent-skills/tree/main/skills/web-design-guidelines) for interaction/accessibility review. The repository's existing visual identity and sentence case take precedence.

## Pictures

The canonical seed has 89 words and 92 active meanings. The [picture manifest](../../apps/web/src/lib/meaning-pictures.json) maps exact meaning UUIDs to original scenes, alternative descriptions and local WebP assets. Repeated words such as reservation, follow-up and deadline have separate pictures. No keyword matching or third-party image service is used at runtime.

Artwork is generated individually with the built-in image tool, reviewed for its specific meaning and optimized to 960 × 640 WebP for delivery. Original generation files and machine-specific receipts remain private. The common prompt asks for an approachable adult editorial illustration, natural proportions, readable everyday action, a neutral scene with blue/lavender accents, landscape 3:2, and no logos or readable text. Each manifest's word, definition and scene provide the per-image prompt. Pictures illustrate a possible use of the meaning; they do not replace definitions.

Pictures appear on word detail and lesson teaching steps. Graded recall, typing, listening and context questions do not receive a picture or its answer-bearing alt text. Missing, pending, unknown or failed media leaves the readable teaching content and actions intact. Asset readiness requires visual review; file existence alone is insufficient.

## Acceptance record

- **Artwork:** all 92 active canonical meanings have a reviewed, distinct local asset and accurate alternative description. All 92 decode as 960 × 640 WebP; SHA-256 values are distinct. Total source assets: 5,899,628 bytes; largest: 114,618 bytes. Canonical coverage tests require every meaning to be ready. Unknown and failed media still omit the image safely.
- **Source checks:** workspace validation, formatting, lint/Go vet, web/shared types, API build, 258 foundation tests, 52 API-client tests and 138 web-helper tests pass. The final 138-test run includes all-ready picture coverage. The affected production web build passes.
- **Browser checks:** the preceding affected matrix passed 117 cases with three intentional viewport skips. A subsequent 30-case run passed after the final milestone disclosure and real teaching-picture transition assertions. The final combined matrix passed 129 cases with three intentional viewport skips after all artwork was enabled. An additional 51 personal-list and word-knowledge cases passed, for 180 affected browser passes overall. Checks include 360px, 430px and desktop, light/dark, keyboard disclosures, axe scans, image failure, readable definitions/saving and absence of images during grading.
- **Visual review:** inspected mobile Home/Journey/Progress and word teaching plus desktop teaching. Definitions and saving precede the uncropped picture; desktop pictures are capped at 288 CSS pixels high. Secondary explanations and milestones remain reachable through native disclosures. Final 360px Progress is approximately 2,355 CSS pixels tall versus 4,489 before the milestone change under the same fixture.
- **Independent review:** reviewed interface diffs, exact-meaning lookup, teaching-only image placement, failure fallback, milestone state preservation and Docker packaging; no actionable findings remained.
- **Runtime assets:** all 92 local picture URLs returned HTTP 200 and matched source SHA-256 values exactly. Docker now copies public assets and its hosted smoke compares the served file with source bytes. Local Docker/container and database-backed Go tests could not run successfully because this WSL environment has a Docker client but no engine; hosted gates must establish those results.

Hosted CI and consolidated PR review remain pending. Synthetic browser tests do not establish real authentication, persistence, physical speech playback, AI quality or learner effectiveness. Production remains on its previously accepted release until a separate release is verified.
