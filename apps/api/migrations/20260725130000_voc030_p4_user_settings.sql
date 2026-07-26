-- atlas:txmode transaction
-- VOC-030-T00: user_settings table (DOC-05 §6). Schema-complete per D01; this
-- package reads/writes only timezone and daily_review_target. No public
-- Settings API/UI is built.

CREATE TABLE user_settings (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  timezone text NOT NULL DEFAULT 'UTC',
  daily_review_target integer NOT NULL DEFAULT 20,
  review_interval_preset text NOT NULL DEFAULT 'vocanova_default',
  notifications_enabled boolean NOT NULL DEFAULT true,
  marketing_emails_enabled boolean NOT NULL DEFAULT false,
  app_language text NOT NULL DEFAULT 'en',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT user_settings_daily_review_target_in_range CHECK (daily_review_target >= 5 AND daily_review_target <= 100),
  CONSTRAINT user_settings_review_interval_preset_valid CHECK (review_interval_preset IN ('vocanova_default', 'wordup_like', 'custom')),
  CONSTRAINT user_settings_app_language_valid CHECK (app_language ~ '^[A-Za-z]{2,8}$'),
  CONSTRAINT user_settings_timezone_nonempty CHECK (char_length(timezone) > 0)
);

CREATE UNIQUE INDEX user_settings_user_id_key
  ON user_settings (user_id);
