-- atlas:txmode file
-- VOC-1402: DOC-05 §12 makes both learning ledgers append-only. Account
-- deletion (§16) is the sole intentional deletion path, enabled only for its
-- containing transaction through vocanova.ledger_purge.

CREATE FUNCTION vocanova_reject_learning_ledger_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'DELETE' AND current_setting('vocanova.ledger_purge', true) = 'on' THEN
    RETURN OLD;
  END IF;

  RAISE EXCEPTION '% rows are append-only', TG_TABLE_NAME
    USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER confidence_point_ledger_append_only
  BEFORE UPDATE OR DELETE ON confidence_point_ledger
  FOR EACH ROW EXECUTE FUNCTION vocanova_reject_learning_ledger_mutation();

CREATE TRIGGER grace_day_ledger_append_only
  BEFORE UPDATE OR DELETE ON grace_day_ledger
  FOR EACH ROW EXECUTE FUNCTION vocanova_reject_learning_ledger_mutation();
