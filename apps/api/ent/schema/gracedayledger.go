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

// GraceDayLedger is the append-only history of every grace-day change.
// The current available balance is balance_after of the latest row per user.
// reason and source_type follow DOC-05 §12. Owned by gamification.
type GraceDayLedger struct{ ent.Schema }

func (GraceDayLedger) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "grace_day_ledger",
			Checks: map[string]string{
				"amount_nonzero": "amount <> 0",
			},
		},
	}
}

func (GraceDayLedger) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (GraceDayLedger) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Int("amount"),
		field.Int("balance_after"),
		field.Enum("reason").
			Values(
				"earned_by_streak",
				"manual_grant",
				"used_for_missed_day",
				"expired",
				"admin_adjustment",
			),
		field.Enum("source_type").Values("daily_mission", "streak", "admin"),
		field.UUID("source_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("applied_to_local_date"),
		field.String("timezone").NotEmpty(),
		field.String("idempotency_key").Optional().Nillable(),
	}
}

func (GraceDayLedger) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("grace_day_ledger").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (GraceDayLedger) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "idempotency_key").
			Unique().
			Annotations(entsql.IndexWhere("idempotency_key IS NOT NULL")),
		index.Fields("user_id", "applied_to_local_date"),
	}
}
