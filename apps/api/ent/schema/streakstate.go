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

// StreakState is the per-user rolling streak state. One row per user.
// status is active / at_risk / broken — derived at read time from the
// daily_mission_snapshots history. Owned by gamification.
type StreakState struct{ ent.Schema }

func (StreakState) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "streak_states",
			Checks: map[string]string{
				"longest_ge_current": "longest_streak_count >= current_streak_count",
				"counts_nonnegative": "current_streak_count >= 0 AND longest_streak_count >= 0",
			},
		},
	}
}

func (StreakState) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (StreakState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Unique(),
		field.Int("current_streak_count").Default(0).Min(0),
		field.Int("longest_streak_count").Default(0).Min(0),
		field.Time("last_completed_local_date").Optional().Nillable(),
		field.Time("last_activity_local_date").Optional().Nillable(),
		field.String("timezone").NotEmpty(),
		field.Enum("status").Values("active", "at_risk", "broken").Default("active"),
	}
}

func (StreakState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("streak_state").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (StreakState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}
