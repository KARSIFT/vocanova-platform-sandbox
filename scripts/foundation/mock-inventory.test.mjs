import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import {
  isRegisteredAPIPath,
  isRegisteredBusinessModule,
  isRegisteredSchemaFile,
  isRegisteredMigrationFile,
  isRegisteredFeatureMigrationTables,
  migrationTableNames,
  validateMockInventory,
} from "./mock-inventory.mjs";

test("current delivery registers only the reviewed learning feature boundaries", () => {
  for (const route of [
    "/api/v1/lessons",
    "/api/v1/lessons/{lessonKey}/sessions",
    "/api/v1/lesson-sessions/{sessionId}",
    "/api/v1/lesson-sessions/{sessionId}/actions",
    "/api/v1/practice-sessions",
    "/api/v1/practice-sessions/{sessionId}",
    "/api/v1/practice-sessions/{sessionId}/actions",
    "/api/v1/meaning-knowledge/{meaningId}",
    "/api/v1/knowledge-summary",
    "/api/v1/learning-preferences",
    "/api/v1/lesson-recommendation",
    "/api/v1/achievements",
  ])
    assert.equal(isRegisteredAPIPath(route), true, route);
  for (const module of ["lessons", "practice", "wordknowledge", "achievements"])
    assert.equal(isRegisteredBusinessModule(module), true, module);
  for (const schema of [
    "lessonaction.go",
    "lessonsession.go",
    "practiceaction.go",
    "practicemistakeresolution.go",
    "practicesession.go",
    "userlearningpreferences.go",
    "userwordknowledge.go",
    "wordknowledgeaction.go",
  ])
    assert.equal(isRegisteredSchemaFile(schema), true, schema);
  assert.deepEqual(validateMockInventory(), []);
});

test("registered feature families do not admit invented routes, modules, or schemas", () => {
  for (const route of [
    "/api/v1/invented-feature",
    "/api/v1/lessons/{lessonKey}/publish",
    "/api/v1/lesson-sessions/{sessionId}/rewards",
    "/api/v1/practice-sessions/{sessionId}/mastery",
    "/api/v1/meaning-knowledge/{meaningId}/share",
    "/api/v1/learning-preferences/admin",
    "/api/v1/lesson-recommendation/override",
    "/api/v1/achievements/claim",
    "/api/v1/lesson-sessions/",
    "/api/v1/meaning-knowledge/",
  ])
    assert.equal(isRegisteredAPIPath(route), false, route);
  for (const module of ["invented", "lessonrewards", "practiceanalytics"])
    assert.equal(isRegisteredBusinessModule(module), false, module);
  for (const schema of [
    "invented.go",
    "lessonreward.go",
    "achievementclaim.go",
  ])
    assert.equal(isRegisteredSchemaFile(schema), false, schema);
});

// VOC-031-T03: the protected-boundary allow list now includes the
// T03 email-change routes (`/api/v1/settings/email-change-links`),
// the `accounts` business module, the `email_change_links` Ent
// schema, and the `email_change_links` migration, in addition to
// the previously adopted A1/P1/P2/P4-T00/P5-T01/T02 boundaries.
test("VOC-031-T03 mock inventory accepts the email-change backend boundary", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-030-T06: the T06 cross-cutting safety test files are present
// and the P5-forbidden invariant holds across the T06 deliverable.
// This test is the same code path as the one above; it is
// separately named to make the T06 acceptance criterion visible in
// the test runner output (the production script's success message
// already prints "VOC-030-T06 mock inventory validation passed").
test("VOC-030-T06 mock inventory accepts the T06 cross-cutting test files", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-031-T06: the P5 cross-cutting reliability test file
// (apps/api/app/api/core_loop_reliability_test.go) and the
// client-side session-expiry helper (apps/web/src/lib/session.ts)
// are both present and the P5-forbidden invariant holds across
// them. This test re-runs the same code path as the two above; it
// is separately named to make the T06 acceptance criterion
// visible in the test runner output.
test("VOC-031-T06 mock inventory accepts the P5 cross-cutting reliability deliverables", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-031-T07a: the accessibility-automation scaffolding
// (Playwright config, tests/e2e/ tree, axe-helper, mock API
// server, CI workflow) is present and the P5-forbidden
// invariant continues to hold. This test re-runs the same
// code path as the ones above; it is separately named to
// make the T07a acceptance criterion visible in the test
// runner output.
test("VOC-031-T07a mock inventory accepts the accessibility-automation scaffolding", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-031-T09: the performance-automation harness
// (apps/web/tests/lighthouse/ tree + CI workflow) is present
// and the P5-forbidden invariant continues to hold across
// it. This test re-runs the same code path as the ones
// above; it is separately named to make the T09 acceptance
// criterion visible in the test runner output.
//
// The T09 acceptance criterion requires that the DOC-08
// thresholds (Performance 85+ / Accessibility 95+ / Best
// Practices 90+) be the single source of truth for the
// runner's assertion logic. The thresholds live in
// `tests/lighthouse/assertions.mjs` (the file the runner
// imports) and are mirrored in `tests/lighthouse/budget.json`
// so the budget file is consumable by a future LHCI
// configuration. To keep the two files in lockstep, this
// test loads `budget.json` and asserts it contains the
// exact values the T09 acceptance criterion names, so a
// drift surfaces here.
test("VOC-031-T09 mock inventory accepts the performance-automation harness", () => {
  assert.deepEqual(validateMockInventory(), []);
});

test("VOC-031-T09 budget.json pins the DOC-08 thresholds verbatim", () => {
  const budgetPath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../../apps/web/tests/lighthouse/budget.json",
  );
  const budget = JSON.parse(readFileSync(budgetPath, "utf8"));
  const scores = budget?.budgets?.[0]?.scores ?? {};
  assert.deepEqual(
    scores,
    { performance: 0.85, accessibility: 0.95, "best-practices": 0.9 },
    "T09 budget.json scores must mirror DOC-08 (Performance 85+ / Accessibility 95+ / Best Practices 90+) exactly",
  );
});

// VOC-031-T11: the P5 mock-inventory completeness check is
// the final piece of the package's "zero legacy mocks, no
// P5-invented route/table/behavior beyond this package's
// own documented scope" guarantee. The static guard rails
// T11 installs are:
//
//  - expectedRouteDirectories now includes the T05 Settings
//    and Settings/account (app) routes;
//  - expectedTopLevelRouteDirectories now includes the
//    T01 /onboarding route (which lives at the top of
//    apps/web/src/app/, not inside (app));
//  - the middleware matcher check now requires /settings
//    and /settings/:path* (the T05 additions);
//  - the D06 SupportedAppLanguages value list is pinned to
//    exactly []string{"en"} at the source level;
//  - the P5 staging-evidence document records the T06/T09/
//    T10 audit findings T11 cites as in-repository evidence.
//
// The first four sub-tests below run the same `validateMockInventory()`
// code path the prior tests run; they are separately
// named so the T11 acceptance criterion is visible in the
// test runner output (the production script's success
// message already prints "VOC-031-T11 mock inventory
// validation passed"). The last sub-test asserts the
// D06 appLanguage invariant directly: the source-level
// declaration is the only canonical source for the
// "no multi-language picker before i18n infrastructure
// exists" guarantee, and a future contributor who
// silently widens the set (e.g. to admit a second
// placeholder locale) would surface here.
test("VOC-031-T11 mock inventory accepts the P5 completeness check", () => {
  assert.deepEqual(validateMockInventory(), []);
});

test("VOC-031-T11 SupportedAppLanguages is restricted to en-only at the source", () => {
  const settingsPath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../../apps/api/business/users/settings.go",
  );
  const content = readFileSync(settingsPath, "utf8");
  assert.match(
    content,
    /SupportedAppLanguages\s*=\s*\[\]string\{"en"\}/,
    'VOC-031-D06: apps/api/business/users/settings.go must declare SupportedAppLanguages = []string{"en"} exactly; admitting a second locale without i18n infrastructure would silently claim a capability the product does not have',
  );
});

test("private lists and original stories admit only implemented route shapes", () => {
  const id = "ac93d068-7c3e-55c9-9d85-3fd3518592dc";
  for (const route of [
    "/api/v1/word-lists",
    "/api/v1/word-lists/{listId}",
    `/api/v1/word-lists/${id}`,
    "/api/v1/word-lists/{listId}/members/{meaningId}",
    `/api/v1/word-lists/${id}/members/${id}`,
    "/api/v1/stories",
    "/api/v1/stories/{storyKey}",
    "/api/v1/stories/a-quiet-lunch",
    "/api/v1/story-sessions",
    "/api/v1/story-sessions/{sessionId}",
    "/api/v1/story-sessions/{sessionId}/actions",
    `/api/v1/story-sessions/${id}`,
    `/api/v1/story-sessions/${id}/actions`,
  ])
    assert.equal(isRegisteredAPIPath(route), true, route);
  for (const route of [
    "/api/v1/word-lists/",
    "/api/v1/word-lists/invalid",
    `/api/v1/word-lists/${id}/share`,
    `/api/v1/word-lists/${id}/members`,
    `/api/v1/word-lists/${id}/members/invalid`,
    `/api/v1/word-lists/${id}/members/${id}/restore`,
    `/api/v1/word-lists/${id}/members/${id}?admin=true`,
    "/api/v1/stories/",
    "/api/v1/stories/bad_key",
    "/api/v1/stories/a-quiet-lunch/publish",
    "/api/v1/story-sessions/",
    "/api/v1/story-sessions/invalid",
    `/api/v1/story-sessions/${id}/rewards`,
    `/api/v1/story-sessions/${id}/actions/admin`,
  ])
    assert.equal(isRegisteredAPIPath(route), false, route);
  for (const module of ["wordlists", "stories"])
    assert.equal(isRegisteredBusinessModule(module), true, module);
  for (const module of ["wordlistsharing", "storyanalytics"])
    assert.equal(isRegisteredBusinessModule(module), false, module);
  for (const schema of ["wordlistshare.go", "storyreward.go"])
    assert.equal(isRegisteredSchemaFile(schema), false, schema);
});

test("list/story real route sources and exact migration table inventories remain bounded", () => {
  const root = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../..",
  );
  for (const file of ["word_lists.go", "stories.go"]) {
    const source = readFileSync(
      path.join(root, "apps/api/app/api", file),
      "utf8",
    );
    const routes = [...source.matchAll(/Path:\s*"([^"\s]+)"/g)].map(
      (match) => match[1],
    );
    assert.ok(routes.length > 0);
    for (const route of routes)
      assert.equal(isRegisteredAPIPath(route), true, `${file}: ${route}`);
  }
  for (const migration of [
    "20261004100000_word_lists.sql",
    "20261004140000_original_stories.sql",
  ]) {
    assert.equal(isRegisteredMigrationFile(migration), true, migration);
    const source = readFileSync(
      path.join(root, "apps/api/migrations", migration),
      "utf8",
    );
    const tables = migrationTableNames(source);
    for (const extra of [
      "CREATE TABLE unexpected_table(id uuid);",
      'CREATE TABLE "unexpected table"(id uuid);',
      "CREATE TABLE private.unexpected_table(id uuid);",
      "CREATE TABLE IF NOT EXISTS unexpected_table(id uuid);",
    ]) {
      assert.equal(
        isRegisteredFeatureMigrationTables(
          migration,
          migrationTableNames(`${source}\n${extra}`),
        ),
        false,
        extra,
      );
    }
    assert.equal(
      isRegisteredFeatureMigrationTables(migration, tables),
      true,
      migration,
    );
    assert.equal(
      isRegisteredFeatureMigrationTables(migration, [
        ...tables,
        "unexpected_table",
      ]),
      false,
    );
    assert.equal(
      isRegisteredFeatureMigrationTables(migration, tables.slice(1)),
      false,
    );
    assert.equal(
      isRegisteredFeatureMigrationTables(migration, [
        ...tables.slice(1),
        tables[1],
      ]),
      false,
    );
  }
  for (const migration of [
    "20261004100001_word_lists.sql",
    "20261004150000_story_rewards.sql",
    "invented.sql",
  ]) {
    assert.equal(isRegisteredMigrationFile(migration), false, migration);
    assert.equal(
      isRegisteredFeatureMigrationTables(migration, [
        "story_sessions",
        "story_actions",
      ]),
      false,
      migration,
    );
  }
  assert.deepEqual(validateMockInventory(), []);
});

test("dictionary registers only its read-only lookup and no learner storage", () => {
  assert.equal(isRegisteredAPIPath("/api/v1/dictionary"), true);
  assert.equal(isRegisteredBusinessModule("dictionary"), true);
  for (const route of [
    "/api/v1/dictionary/",
    "/api/v1/dictionary/book",
    "/api/v1/dictionary/save",
    "/api/v1/dictionary/reviews",
    "/api/v1/dictionary/admin",
    "/api/v1/dictionary-extra",
  ])
    assert.equal(isRegisteredAPIPath(route), false, route);
  for (const module of ["dictionarylearning", "dictionaryadmin"])
    assert.equal(isRegisteredBusinessModule(module), false, module);
  for (const schema of [
    "dictionary.go",
    "dictionaryword.go",
    "userdictionaryword.go",
  ])
    assert.equal(isRegisteredSchemaFile(schema), false, schema);
  assert.equal(
    isRegisteredMigrationFile("20261007010000_dictionary.sql"),
    false,
  );
  const root = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../..",
  );
  const contract = JSON.parse(
    readFileSync(
      path.join(root, "apps/api/openapi/vocanova.openapi.json"),
      "utf8",
    ),
  );
  assert.deepEqual(Object.keys(contract.paths["/api/v1/dictionary"]), ["get"]);
  assert.deepEqual(validateMockInventory(), []);
});
