-- atlas:txmode file
-- Issue #1429: DOC-05 §§2, 9, and 16 define review_attempts as immutable
-- learner history. Account deletion is the sole deletion path and must use
-- the same transaction-local purge gate as the other learning histories.

CREATE FUNCTION vocanova_reject_review_attempt_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'DELETE' AND current_setting('vocanova.ledger_purge', true) = 'on' THEN
    RETURN OLD;
  END IF;

  RAISE EXCEPTION '% rows are immutable', TG_TABLE_NAME
    USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER review_attempts_immutable
  BEFORE UPDATE OR DELETE ON review_attempts
  FOR EACH ROW EXECUTE FUNCTION vocanova_reject_review_attempt_mutation();
