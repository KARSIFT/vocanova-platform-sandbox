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

// UserSettings is the per-user settings row. DOC-05 §6 schema-complete shape;
// the P4 package reads/writes only timezone and daily_review_target (D01).
// One row per user. No public Settings API/UI is built in P4.
type UserSettings struct{ ent.Schema }

func (UserSettings) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "user_settings",
			Checks: map[string]string{
				"daily_review_target_in_range": "daily_review_target >= 5 AND daily_review_target <= 100",
				"review_interval_preset_valid": "review_interval_preset IN ('vocanova_default', 'wordup_like', 'custom')",
				"app_language_valid":           "app_language ~ '^[A-Za-z]{2,8}$'",
				"timezone_nonempty":            "char_length(timezone) > 0",
			},
		},
	}
}

func (UserSettings) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (UserSettings) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Unique(),
		field.String("timezone").Default("UTC"),
		field.Int("daily_review_target").Default(20).Min(5).Max(100),
		field.String("review_interval_preset").Default("vocanova_default"),
		field.Bool("notifications_enabled").Default(true),
		field.Bool("marketing_emails_enabled").Default(false),
		field.String("app_language").Default("en"),
	}
}

func (UserSettings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("user_settings").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (UserSettings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}
