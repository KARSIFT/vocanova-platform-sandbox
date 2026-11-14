# Shared component consistency — 14 November 2026

The owner requested consolidation of existing controls, panels and feedback plus a browser component gallery, with feedback after staging and production rollout. The delivery retains the approved Warm guided coach direction.

## Delivered behavior

- Ordinary actions share Button and ActionLink; form fields share TextInput, TextArea and Select. Labels, refs, input attributes, native validation, pending locks and explicit submission types stay with their features. Button defaults to a non-submitting action.
- StatusMessage shares feedback presentation while retaining existing paragraph/div tags, announcement roles, IDs, refs and mounted-region ownership. Dark success text uses the existing readable blue role.
- Dictionary result panels use Surface. Specialized reading surfaces, graded answers, review ratings, audio toggles, navigation, meaning cards and list rows retain their appropriate presentation.
- Authentication uses the shared controls; obsolete parallel auth control CSS is removed. The connection recovery action remains a full document navigation.
- Profile's avatar/description group wraps at narrow widths with enlarged text. Native mobile navigation remains distinct from action buttons.
- The synthetic `/component-gallery` demonstrates actions, fields, feedback, theme changes and learning guidance. It returns 404 unless the exact server-side development flag is enabled. No deployment enables that flag, and no learner navigation entry is added.

See the [component reference](../design/component-system.md) and [implementation plan](../superpowers/plans/2026-11-14-component-consistency.md). This is frontend consolidation; learning mutations and provider integrations retain their existing owners.

## Verification

Fresh production web build, full web lint, application/browser TypeScript, formatting and all 302 helper/middleware tests passed. The direct-Node server-page concurrency test loader now understands CSS-module imports; actual browser tests exercise compiled CSS.

The final shared-control gallery and Profile/navigation enlarged-text and Journey document-retry matrix passed **54 cases**, zero failures or retries, across Chrome, Firefox and WebKit at 360px, 430px and desktop in both themes. It includes native form submission, disabled controls, ref focus, validation errors, action-link navigation, 200% text, no horizontal overflow and critical/serious axe checks on the gallery. Nine explicit skips select the opposite gallery configuration. Against an unflagged server, all three 404 checks passed; six enabled-gallery scenarios were explicitly skipped.

The independent visual pass inspected 24 final Profile/Settings/Lists states across three widths and both themes, including Profile at 200% text, plus four keyboard-focus captures. Root inspected the 360px light contact sheet. Earlier screenshots exposed navigation-label splitting caused by generic action padding and the narrow Profile explanation; those failures are preserved and covered by new browser regressions. The focused baseline failed both 360px theme cases before repair.

Broad Chrome component-integration acceptance passed **290 cases with 16 existing layout-scoped skips**, zero failures or retries. It covers authentication recovery, password/mobile lifecycle, Profile, settings and pending saves, lists, Home, review, writing, practice, guided lessons, stories and the saved-meaning bridge. After the final native recovery-anchor repairs, Chrome navigation/recovery acceptance passed **135 cases with six explicit configuration/layout skips**, zero failures or retries; this includes the 18 Chrome cases in the cross-engine matrix above. Journey document retry failed both focused baseline cases before repair and passed afterward. The existing Practice fallback checks passed before and after; its native-anchor restoration is supported by source review without a separate behavioral RED claim. Independent whole-diff review and the focused repair re-review found no remaining blocker. Hosted CI/rollout are separate gates.

Local foundation checks passed **333/337** with a real macOS temporary directory. The four existing platform failures concern GNU sed, Bash coproc and the Chromium-install fixture; Linux CI must pass its full checks. Default macOS temporary-directory symlinks caused additional initial harness failures. No test checks were removed. The pinned Chromium download returned HTTP 403 for this location, so local Chrome uses the installed browser through the repository-pinned Playwright runner; cached Firefox/WebKit ran normally.

## Acceptance limits

Synthetic browser fixtures do not establish real authentication delivery, AI-feedback quality, database durability, actual screen-reader speech or learner effectiveness. Component consistency does not complete the product or replace the prepared learner study. Gallery screenshots are development examples and must not be presented as earned learner progress.

## Release status

Implementation remains on the consolidated delivery branch until independent review and applicable CI pass. Staging and production deployment are separate observed outcomes; no rollout is claimed by local verification alone.
