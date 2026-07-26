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

// ConfidencePointLedger is the append-only history of every Confidence Point
// change. The running balance is balance_after of the latest row per user.
// reason and source_type follow DOC-05 §12 (with the D02 reconciliation:
// `word_added` reason / `user_word` source_type added for the Add word reward).
// Owned by gamification.
type ConfidencePointLedger struct{ ent.Schema }

func (ConfidencePointLedger) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "confidence_point_ledger",
			Checks: map[string]string{
				"amount_nonzero": "amount <> 0",
			},
		},
	}
}

func (ConfidencePointLedger) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (ConfidencePointLedger) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Int("amount"),
		field.Int("balance_after"),
		field.Enum("reason").
			Values(
				"word_added",
				"review_correct",
				"daily_mission_completed",
				"sentence_submitted",
				"ai_feedback_received",
				"streak_bonus",
				"admin_adjustment",
			),
		field.Enum("source_type").
			Values(
				"user_word",
				"review_attempt",
				"daily_mission",
				"learner_sentence",
				"ai_feedback_attempt",
				"streak",
				"admin",
			),
		field.UUID("source_id", uuid.UUID{}).Optional().Nillable(),
		field.String("idempotency_key").Optional().Nillable(),
		field.JSON("metadata", map[string]any{}).Optional(),
		field.Time("occurred_at"),
	}
}

func (ConfidencePointLedger) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("confidence_point_ledger").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (ConfidencePointLedger) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "idempotency_key").
			Unique().
			Annotations(entsql.IndexWhere("idempotency_key IS NOT NULL")),
		index.Fields("user_id", "occurred_at"),
	}
}
