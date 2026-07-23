import { writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  brand,
  duration,
  easing,
  elevation,
  fontSize,
  neutral,
  radius,
  spacing,
} from "@vocanova/design-tokens";

const EASING_KEY_NAMES = {
  linear: "linear",
  easeIn: "in",
  easeOut: "out",
  easeInOut: "in-out",
};

function themeProperties(prefix, scale, keyNames) {
  return Object.entries(scale).map(
    ([key, value]) =>
      `  --${prefix}-${keyNames ? keyNames[key] : key}: ${value};`,
  );
}

export function generateTokensCss() {
  const properties = [
    ...themeProperties("spacing", spacing),
    ...themeProperties("color-neutral", neutral),
    ...themeProperties("color-primary", brand.primary),
    ...themeProperties("color-secondary", brand.secondary),
    ...themeProperties("text", fontSize),
    ...themeProperties("radius", radius),
    ...themeProperties("shadow", elevation),
    ...themeProperties("ease", easing, EASING_KEY_NAMES),
    ...themeProperties("duration", duration),
  ];

  return `/*
 * GENERATED FILE — do not edit.
 * Run \`pnpm --filter @vocanova/web generate:tokens\` to regenerate.
 * Source of truth: packages/design-tokens/src/*.
 */

/* prettier-ignore */
@theme static {
${properties.join("\n")}
}
`;
}

const outputPath = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../src/app/tokens.generated.css",
);

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  writeFileSync(outputPath, generateTokensCss());
  process.stdout.write(`Wrote ${path.relative(process.cwd(), outputPath)}\n`);
}
