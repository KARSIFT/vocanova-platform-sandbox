# Vocanova component system

The production React components are the implementation source for the approved Warm guided coach direction. Read [learning workspace](learning-workspace.md) for product design intent. Use this page to choose existing components before introducing another visual pattern.

## Foundations

`packages/design-tokens/src/` owns palette, spacing, typography, radii and motion values; generated `tokens.generated.css` exposes them to the web. Do not edit generated CSS directly. `globals.css` owns the existing primary/secondary action vocabulary and theme mapping.

## Shared components

| Import                                                  | Purpose and contract                                                                                                                                                                                                          |
| ------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `@/ui/controls` → `Button`                              | Native button, `variant="primary"` (default), `secondary`, `danger`, or `text`. Defaults to `type="button"`; use explicit `type="submit"` for form submission. Native disabled, ref, event, aria and data props pass through. |
| `@/ui/controls` → `ActionLink`                          | A real navigation link with matching action variants; retains the existing intent-prefetch/navigation behavior. Use Button for an action and ActionLink for navigation.                                                       |
| `@/ui/controls` → `TextInput`, `TextArea`, `Select`     | Native labelled form elements with shared field presentation. Caller owns labels, IDs, descriptions, validation and state. Checkbox/radio controls keep their native semantics and specialized presentation.                  |
| `@/ui/surface` → `PageContainer`, `Surface`             | Page width and panel spacing. Surface retains default, primary and secondary tones. Reading surfaces, full-screen players and graded options may have intentionally different layouts.                                        |
| `@/ui/status-message` → `StatusMessage`                 | Info, success and error presentation. Preserve one live-region owner; static examples use `live={false}`. Keep factual server-confirmed messages and recovery actions in the owning feature.                                  |
| `@/ui/learning-guide` → `LearningIntro`, `LearningNote` | Learning hierarchy and short guidance. Notes are not automatic confirmation announcements.                                                                                                                                    |
| `@/ui/session` → `SessionActions`                       | Focused activity action area; callers reserve content space and retain keyboard visibility.                                                                                                                                   |
| `@/ui/meaning-picture` → `MeaningPicture`               | Meaning/revision-specific teaching imagery. Do not mount inside graded questions.                                                                                                                                             |
| `@/ui/pronunciation` → `ListenButton`, `SpeechSettings` | Existing pronunciation and preferences. Playback is not an assessment.                                                                                                                                                        |

Feature components remain responsible for meaning identity, scheduling, saving, feedback history and draft recovery. Reuse WordDetailContent, MeaningTeaching and SentenceFeedback rather than reconstructing their state behavior. Preserve server authority and idempotent writes.

## Browser gallery

`/component-gallery` is a development-only, synthetic interactive reference. It is not linked from learner navigation. Without the exact server-side `VOCANOVA_COMPONENT_GALLERY=true`, it returns 404, including deployed builds. The flag is read at request time and is not a NEXT_PUBLIC value; this change does not enable it in staging or production.

From the repository root, activate pinned Node/pnpm and run the local launcher with the explicit flag:

```bash
VOCANOVA_COMPONENT_GALLERY=true bash .work/vocanova-local.sh build
VOCANOVA_COMPONENT_GALLERY=true bash .work/vocanova-local.sh preview
```

Open `http://127.0.0.1:3100/component-gallery`. The `.work` launcher is local ignored tooling; if it is absent in a fresh checkout, build/start the web with PORT=3100, PORT_WEB=3100, MOCK_API_PORT=8180 and both API URLs set to `http://127.0.0.1:8180`, following the local port policy. Never occupy Vergecloud's 3000/8080. Public API URL changes require rebuilding.

The gallery demonstrates native field validation, textarea refs/focus, explicit submission, disabled/pending actions, navigation links, feedback, panels, typography and learning guidance. It makes no learning mutations or provider requests. Review both themes, 360/430/desktop, keyboard navigation and 200% text. Visual examples do not establish AI quality, database persistence or learner effectiveness.

Run the gallery browser spec with the flag enabled for interactive examples; run it against an unflagged server to verify 404. The ordinary CI run verifies the disabled route and explicitly skips only the enabled-gallery scenarios.

## Intentional exceptions

Keep navigation icons, sound toggles, range selectors with dedicated interaction behavior, graded answer choices, chips and disclosure summaries specialized. Do not turn links into buttons or add a wrapper that loses native semantics. Prefer the shared components for ordinary controls. `className` remains an escape hatch for layout and documented contextual variants, not a reason to duplicate the entire control style.

## Review changes

Review screenshots and actual interactions alongside automated checks. Look for one primary next action, readable vocabulary, preserved drafts, announced errors, reachable recovery, visible focus, no overflow, and truthful completion states. A passing accessibility scan supplements visual judgment and learner observations.
