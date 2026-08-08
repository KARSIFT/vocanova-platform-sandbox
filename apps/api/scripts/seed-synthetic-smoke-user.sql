-- VOC-050-T00: idempotent seed for the synthetic smoke-test account.
--
-- Invoked by apps/api/scripts/seed-synthetic-smoke-user.sh on every
-- staging and production deploy, immediately after migrations apply.
-- Rerunning it against a database that already holds the account is a
-- no-op apart from refreshing its mutable fields, so a deploy never
-- duplicates or errors on the second and subsequent runs
-- (VOC-050-AC-00 / VOC-050-TEST-00).
--
-- The caller supplies two psql variables, which psql quotes and
-- escapes as SQL literals via the :'name' form, so no shell value is
-- ever interpolated into this statement text unescaped:
--
--   synthetic_email         reserved, non-deliverable identity
--   synthetic_display_name  human-obvious label for the account
--
-- The values are handed to the DO block below through session
-- settings, because psql deliberately does not expand :'name' inside
-- a dollar-quoted body.
SELECT
  set_config('vocanova.synthetic_smoke_test_email', :'synthetic_email', false),
  set_config('vocanova.synthetic_smoke_test_display_name', :'synthetic_display_name', false);

DO $seed$
DECLARE
  -- users_active_email_key is a partial unique index on lower(email)
  -- for rows with deleted_at IS NULL, so every lookup and comparison
  -- here matches on the lowercased address and ignores soft-deleted
  -- rows.
  target_email text := lower(trim(current_setting('vocanova.synthetic_smoke_test_email')));
  target_display_name text := trim(current_setting('vocanova.synthetic_smoke_test_display_name'));
  retired_account_count integer;
BEGIN
  IF target_email = '' THEN
    RAISE EXCEPTION 'VOC-050 seed: synthetic_email must not be empty';
  END IF;

  -- Fail closed rather than adopting an account somebody else owns.
  -- Without this guard the refresh below would silently convert a real
  -- user's account into the privileged synthetic identity if that user
  -- ever came to hold the reserved address (VOC-050-AC-01).
  IF EXISTS (
    SELECT 1
    FROM users
    WHERE lower(email) = target_email
      AND deleted_at IS NULL
      AND NOT is_synthetic_test_account
  ) THEN
    RAISE EXCEPTION
      'VOC-050 seed refused: % is already registered to a non-synthetic account', target_email;
  END IF;

  -- Retire any previously-seeded synthetic account under a different
  -- address (the reserved identity is configurable, so it can rotate).
  -- Soft-deleting rather than merely clearing the marker keeps every
  -- synthetic row identifiable and frees users_single_synthetic_test_account_idx
  -- for the account seeded below.
  UPDATE users
  SET status = 'deleted',
      deleted_at = now(),
      is_synthetic_test_account = false,
      updated_at = now()
  WHERE is_synthetic_test_account
    AND deleted_at IS NULL
    AND lower(email) IS DISTINCT FROM target_email;
  GET DIAGNOSTICS retired_account_count = ROW_COUNT;
  IF retired_account_count > 0 THEN
    RAISE NOTICE 'VOC-050 seed: retired % synthetic account(s) under a previous address', retired_account_count;
  END IF;

  UPDATE users
  SET display_name = target_display_name,
      status = 'active',
      onboarding_status = 'completed',
      email_verified_at = COALESCE(email_verified_at, now()),
      is_synthetic_test_account = true,
      updated_at = now()
  WHERE lower(email) = target_email
    AND deleted_at IS NULL;

  IF NOT FOUND THEN
    -- A random id (rather than a fixed one) keeps the insert safe when
    -- a soft-deleted synthetic row from an earlier rotation is still
    -- present and would otherwise collide on the primary key.
    INSERT INTO users (
      id,
      email,
      display_name,
      status,
      onboarding_status,
      email_verified_at,
      is_synthetic_test_account,
      created_at,
      updated_at
    )
    VALUES (
      gen_random_uuid(),
      target_email,
      target_display_name,
      'active',
      'completed',
      now(),
      true,
      now(),
      now()
    );
    RAISE NOTICE 'VOC-050 seed: created synthetic smoke-test account %', target_email;
  ELSE
    RAISE NOTICE 'VOC-050 seed: refreshed existing synthetic smoke-test account %', target_email;
  END IF;
END
$seed$;
