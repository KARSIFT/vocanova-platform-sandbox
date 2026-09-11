import assert from "node:assert/strict";
import test from "node:test";

import { isPrimaryNavItemActive } from "../../src/app/(app)/_components/bottom-nav-state";

test("keeps Journey active throughout nested discovery routes", () => {
  assert.equal(isPrimaryNavItemActive("/discover", "/discover"), true);
  assert.equal(
    isPrimaryNavItemActive("/discover/ordering-at-a-cafe", "/discover"),
    true,
  );
  assert.equal(
    isPrimaryNavItemActive("/discover/ordering-at-a-cafe/pour", "/discover"),
    true,
  );
  assert.equal(isPrimaryNavItemActive("/words", "/discover"), true);
  assert.equal(
    isPrimaryNavItemActive(
      "/words/00000000-0000-0000-0000-000000000001",
      "/discover",
    ),
    true,
  );
});

test("does not activate Journey for a similarly named route", () => {
  assert.equal(isPrimaryNavItemActive("/discovery", "/discover"), false);
});

test("keeps Home exact and Progress active for nested progress routes", () => {
  assert.equal(isPrimaryNavItemActive("/home", "/home"), true);
  assert.equal(isPrimaryNavItemActive("/home/details", "/home"), false);
  assert.equal(isPrimaryNavItemActive("/progress", "/progress"), true);
  assert.equal(isPrimaryNavItemActive("/progress/sentences", "/progress"), true);
  assert.equal(isPrimaryNavItemActive("/progression", "/progress"), false);
});
