-- Negative control: delete one known nonempty projection only on the disposable target.
-- The runner must never call this against its source. No ledger bypass is required.
\set ON_ERROR_STOP on
BEGIN;
DO $fault$
DECLARE
  removed integer;
BEGIN
  IF current_database() <> 'vocanova_rehearsal' THEN
    RAISE EXCEPTION 'rehearsal_fault_wrong_database';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.users WHERE id='d0c0a001-0000-4000-8000-000000000001' AND email='restore-rehearsal@vocanova.invalid' AND is_synthetic_test_account) THEN
    RAISE EXCEPTION 'rehearsal_fault_synthetic_identity_missing';
  END IF;
  DELETE FROM public.daily_activity_summaries WHERE id='d0c0a001-0000-4000-8000-000000000008' AND user_id='d0c0a001-0000-4000-8000-000000000001' AND local_date='2026-10-02';
  GET DIAGNOSTICS removed = ROW_COUNT;
  IF removed <> 1 THEN
    RAISE EXCEPTION 'rehearsal_fault_expected_one_activity_record';
  END IF;
END;
$fault$;
COMMIT;
