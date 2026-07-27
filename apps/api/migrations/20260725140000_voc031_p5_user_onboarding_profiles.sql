-- atlas:txmode transaction
-- VOC-031-T00: user_onboarding_profiles (DOC-05 §6) + the
-- onboarding_status='completed' backfill for pre-existing accounts
-- (VOC-031-D03, resolved at adoption 2026-07-27 founder-gate
-- delegation: grandfather pre-existing accounts).
--
-- The backfill keeps every pre-existing account — including
-- non-production staging/test identities created during VOC-025
-- through VOC-030 evidence work — past the /onboarding gate T01 will
-- install. Only accounts created after this migration runs default
-- to onboarding_status='not_started' and see the gate.
--
-- The user_id unique constraint enforces the documented "one row
-- per user" rule at the DB layer. No existing A1–P4 table, column,
-- or constraint is altered; the UPDATE is purely additive on
-- users.onboarding_status and never loosens its check constraint
-- or removes existing rows.

CREATE TABLE user_onboarding_profiles (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
  english_level text NOT NULL DEFAULT 'unknown'
    CHECK (english_level IN ('a1', 'a2', 'b1', 'b2', 'unknown')),
  native_language text NOT NULL CHECK (char_length(native_language) > 0),
  learning_goal text NOT NULL
    CHECK (learning_goal IN ('general', 'work', 'travel', 'study', 'conversation', 'exam')),
  main_use_case text NOT NULL
    CHECK (main_use_case IN ('daily_life', 'work', 'travel', 'study', 'social')),
  daily_review_target integer NOT NULL
    CHECK (daily_review_target >= 5 AND daily_review_target <= 100),
  completed_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX user_onboarding_profiles_user_id_idx
  ON user_onboarding_profiles (user_id);

-- VOC-031-D03: grandfather pre-existing accounts past the
-- /onboarding gate. The created_at < NOW() guard captures every
-- user row that exists at migration time without depending on a
-- hand-supplied timestamp; the WHERE onboarding_status =
-- 'not_started' guard makes the UPDATE idempotent — a re-run
-- after the first successful application is a no-op.
UPDATE users
SET onboarding_status = 'completed', updated_at = NOW()
WHERE onboarding_status = 'not_started' AND created_at < NOW();
