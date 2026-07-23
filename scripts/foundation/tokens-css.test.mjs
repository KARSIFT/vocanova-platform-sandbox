import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";

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

import { generateTokensCss } from "../../apps/web/scripts/generate-tokens-css.mjs";

const tokensCssPath = path.resolve("apps/web/src/app/tokens.generated.css");

const EASING_KEY_NAMES = {
  linear: "linear",
  easeIn: "in",
  easeOut: "out",
  easeInOut: "in-out",
};

function expectedProperties() {
  return [
    ...Object.entries(spacing).map(([key, value]) => [
      `--spacing-${key}`,
      value,
    ]),
    ...Object.entries(neutral).map(([key, value]) => [
      `--color-neutral-${key}`,
      value,
    ]),
    ...Object.entries(brand.primary).map(([key, value]) => [
      `--color-primary-${key}`,
      value,
    ]),
    ...Object.entries(brand.secondary).map(([key, value]) => [
      `--color-secondary-${key}`,
      value,
    ]),
    ...Object.entries(fontSize).map(([key, value]) => [`--text-${key}`, value]),
    ...Object.entries(radius).map(([key, value]) => [`--radius-${key}`, value]),
    ...Object.entries(elevation).map(([key, value]) => [
      `--shadow-${key}`,
      value,
    ]),
    ...Object.entries(easing).map(([key, value]) => [
      `--ease-${EASING_KEY_NAMES[key]}`,
      value,
    ]),
    ...Object.entries(duration).map(([key, value]) => [
      `--duration-${key}`,
      value,
    ]),
  ];
}

// Ordering dependency: this test imports @vocanova/design-tokens from dist,
// which is emitted by `tsc -b` in `pnpm run typecheck` / `pnpm run build`
// before `pnpm run test` executes in CI.
test("tokens.generated.css includes every token property/value pair", () => {
  const css = readFileSync(tokensCssPath, "utf8");
  const entries = expectedProperties();

  assert.equal(entries.length, 64);

  for (const [property, value] of entries) {
    assert.match(
      css,
      new RegExp(
        `\\s${property}:\\s${value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")};`,
      ),
    );
  }
});

test("tokens.generated.css is byte-equal to generator output", () => {
  const css = readFileSync(tokensCssPath, "utf8");
  const generated = generateTokensCss();

  assert.equal(css, generated);
});
