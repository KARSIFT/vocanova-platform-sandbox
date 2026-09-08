-- atlas:txmode file
-- DOC-05 §13/§15: immutable, learner-scoped audit records. User linkage is
-- nullable so the account-deletion transaction can retain an aggregate event
-- without retaining a deleted learner's identifier or metadata.

CREATE TABLE feature_audit_logs (
    id uuid PRIMARY KEY,
    user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    action text NOT NULL CHECK (action <> ''),
    entity_type text NOT NULL CHECK (entity_type <> ''),
    entity_id uuid,
    request_id text,
    actor_type text NOT NULL CHECK (actor_type IN ('user', 'system', 'admin', 'ai')),
    actor_id uuid,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX feature_audit_logs_user_created_at_idx
    ON feature_audit_logs (user_id, created_at DESC);

CREATE INDEX feature_audit_logs_entity_idx
    ON feature_audit_logs (entity_type, entity_id);
