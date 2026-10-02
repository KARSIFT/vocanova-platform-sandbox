-- Explicit staging-only fixture preparation, never the ordinary deploy seed.
-- The caller must serialize preparation through the entire browser journey.
-- This transaction prevents partial retirement; it is not a journey-wide lock.
BEGIN;
SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '30s';
SELECT set_config('vocanova.journey_environment', :'journey_environment', true) AS journey_environment_setting,
       set_config('vocanova.api_environment', :'api_environment', true) AS api_environment_setting,
       set_config('vocanova.journey_email', :'synthetic_email', true) AS journey_email_setting
\gset

DO $prepare$
DECLARE
  target_email text := current_setting('vocanova.journey_email');
  prior_user users%ROWTYPE;
  prepared_at timestamptz := now();
BEGIN
  IF current_setting('vocanova.journey_environment') <> 'staging'
     OR current_setting('vocanova.api_environment') NOT IN ('', 'staging') THEN
    RAISE EXCEPTION 'synthetic journey preparation requires staging';
  END IF;
  -- Keep email limits and pattern in sync with prepare-synthetic-staging-journey.sh.
  IF length(target_email) > 254 OR target_email !~
     '^[a-z0-9][a-z0-9._%+-]{0,63}@[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*\.invalid$' THEN
    RAISE EXCEPTION 'synthetic journey requires a canonical lowercase .invalid email';
  END IF;

  PERFORM pg_advisory_xact_lock(hashtextextended('vocanova:prepare-staging-synthetic-journey', 0));
  IF EXISTS (
    SELECT 1 FROM users
    WHERE is_synthetic_test_account AND deleted_at IS NULL
      AND lower(email) IS DISTINCT FROM target_email
  ) THEN
    RAISE EXCEPTION 'synthetic journey refused a different active synthetic identity';
  END IF;

  SELECT * INTO prior_user FROM users
  WHERE lower(email) = target_email AND deleted_at IS NULL
  FOR UPDATE;
  IF FOUND THEN
    IF NOT prior_user.is_synthetic_test_account OR prior_user.status <> 'active' THEN
      RAISE EXCEPTION 'synthetic journey refused an unmarked or inactive reserved identity';
    END IF;

    UPDATE sessions SET revoked_at = prepared_at
    WHERE user_id = prior_user.id AND revoked_at IS NULL AND expires_at > prepared_at;
    UPDATE users
    SET status = 'deleted', deleted_at = prepared_at, updated_at = prepared_at
    WHERE id = prior_user.id;
  END IF;

  INSERT INTO users (
    id, email, display_name, status, onboarding_status, email_verified_at,
    is_synthetic_test_account, created_at, updated_at
  ) VALUES (
    gen_random_uuid(), target_email, 'Staging synthetic journey', 'active',
    'completed', prepared_at, true, prepared_at, prepared_at
  );
END
$prepare$;
COMMIT;
SELECT 'prepared' AS synthetic_staging_journey;
