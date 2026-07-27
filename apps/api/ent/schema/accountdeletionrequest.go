package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AccountDeletionRequest is the per-user "the learner requested
// account deletion" record. VOC-031-T04 / DOC-05 §16 / DOC-06 §14 /
// VOC-031-D07. One row per user; the sweep function reads the
// (status, purge_after) pair to decide which rows are eligible
// for anonymization.
//
// The lifecycle is three-valued and strictly ordered:
//   - 'deactivated': the user has been deactivated (status='deleted',
//     deleted_at set on users; every session and unconsumed
//     auth-token revoked). purge_after holds the legal-review
//     deadline. The sweep picks up rows where status='deactivated'
//     AND NOW() >= purge_after.
//   - 'anonymizing': a transient in-flight state the sweep sets
//     to claim the row before it begins mutating per-table data,
//     so two concurrent sweep invocations never double-process.
//   - 'completed': the per-table anonymization is done. completed_at
//     is the timestamp the sweep wrote. The sweep never re-touches
//     a completed row.
type AccountDeletionRequest struct{ ent.Schema }

func (AccountDeletionRequest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "account_deletion_requests",
			Checks: map[string]string{
				"status_valid":              "status IN ('deactivated', 'anonymizing', 'completed')",
				"purge_after_after_request": "purge_after > requested_at",
				"purge_after_within_year":   "purge_after <= requested_at + interval '365 days'",
				"completed_when_status":     "status = 'completed' OR completed_at IS NULL",
				"completed_at_iff_status":   "status <> 'completed' OR completed_at IS NOT NULL",
				"idempotency_key_nonempty":  "char_length(idempotency_key) > 0",
			},
		},
	}
}

func (AccountDeletionRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (AccountDeletionRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Unique().Immutable(),
		field.String("status").Default("deactivated").Immutable(),
		field.Time("requested_at").Immutable(),
		field.Time("purge_after").Immutable(),
		field.Time("completed_at").Optional().Nillable(),
		field.String("idempotency_key").NotEmpty().Immutable(),
	}
}

func (AccountDeletionRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("account_deletion_requests").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (AccountDeletionRequest) Indexes() []ent.Index {
	return []ent.Index{
		// The sweep's hot path is "find every 'deactivated' row
		// whose purge_after has passed". The partial index is
		// what makes that a single index scan instead of a
		// full table scan, and is also why the partial predicate
		// is part of the index's storage.
		index.Fields("status", "purge_after").
			Annotations(entsql.IndexWhere("status = 'deactivated'")),
		index.Fields("user_id"),
	}
}
