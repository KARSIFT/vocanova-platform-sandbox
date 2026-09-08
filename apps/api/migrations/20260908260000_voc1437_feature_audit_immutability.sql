-- atlas:txmode file
-- Issue #1437: feature audit rows are immutable operational history. Account
-- deletion may only clear learner linkage and metadata through the shared,
-- transaction-local purge gate; it may not rewrite or delete the event.

CREATE FUNCTION vocanova_guard_feature_audit_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'UPDATE'
     AND current_setting('vocanova.ledger_purge', true) = 'on'
     AND NEW.id = OLD.id
     AND NEW.user_id IS NULL
     AND NEW.action = OLD.action
     AND NEW.entity_type = OLD.entity_type
     AND NEW.entity_id IS NULL
     AND NEW.request_id IS NOT DISTINCT FROM OLD.request_id
     AND NEW.actor_type = OLD.actor_type
     AND NEW.actor_id IS NULL
     AND NEW.metadata = '{}'::jsonb
     AND NEW.created_at = OLD.created_at
  THEN
    RETURN NEW;
  END IF;

  RAISE EXCEPTION 'feature_audit_logs rows are immutable'
    USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER feature_audit_logs_immutable
  BEFORE UPDATE OR DELETE ON feature_audit_logs
  FOR EACH ROW EXECUTE FUNCTION vocanova_guard_feature_audit_mutation();
