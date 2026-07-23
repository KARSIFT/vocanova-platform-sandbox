import { readFileSync } from "node:fs";
import { globSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

// Deterministic, cheap checks for two specific mistakes independent AI
// review caught by hand on VOC-019-T00 (2026-07-24) - catching them here
// means CI fails fast, before an expensive independent-review cycle ever
// spends tokens re-discovering the same two things. Not a general-purpose
// Tailwind linter: narrowly targets patterns with no legitimate exception,
// not the broader "every utility must resolve to a wired token" judgment
// call review.yml's reviewer still makes (e.g. the touch-target spacing
// question from that same PR, which had a real defensible exception).

const webSrc = fileURLToPath(
  new URL("../../apps/web/src/app/(app)/", import.meta.url),
);

const files = globSync("**/*.tsx", { cwd: webSrc }).map((f) =>
  path.join(webSrc, f),
);

const errors = [];

// Tailwind v4 auto-generates utilities from @theme custom properties for
// namespaces it recognizes (colors, spacing) but NOT for --duration-*/
// --ease-* - those only work as arbitrary-value utilities
// (duration-[var(--duration-fast)]). A bare `duration-fast` class matches
// no Tailwind candidate and silently emits no rule - the exact bug found
// live on VOC-019-T00's first attempt.
const bareDurationPattern =
  /(?<!--)\bduration-(instant|fast|base|slow|slower)\b(?!\])/;

// apps/web/src/app/(app)/layout.tsx already renders the one <main> for
// every route in this group - a route page.tsx must not render its own,
// or the page ends up with two nested <main> landmarks (invalid per the
// HTML spec, breaks assistive-tech landmark navigation). Found live on
// the same PR's first attempt.
const nestedMainPattern = /<main[\s>]/;

for (const file of files) {
  const rel = path.relative(webSrc, file);
  const content = readFileSync(file, "utf8");

  if (bareDurationPattern.test(content)) {
    errors.push(
      `${rel}: bare "duration-<name>" utility found - Tailwind v4 does not ` +
        `generate utilities from --duration-* custom properties. Use the ` +
        `arbitrary-value form instead, e.g. duration-[var(--duration-fast)].`,
    );
  }

  // layout.tsx itself is exempt - it's the one file allowed to render the
  // group's single <main>.
  if (path.basename(file) !== "layout.tsx" && nestedMainPattern.test(content)) {
    errors.push(
      `${rel}: renders its own <main> - apps/web/src/app/(app)/layout.tsx ` +
        `already provides one for every route in this group. Use <div> or ` +
        `<section> instead.`,
    );
  }
}

if (errors.length > 0) {
  process.stderr.write(
    `Tailwind token / landmark check failed:\n${errors.map((e) => `  - ${e}`).join("\n")}\n`,
  );
  process.exitCode = 1;
}
