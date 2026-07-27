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

// UserOnboardingProfile is the per-user onboarding submission record. DOC-05
// §6 schema-complete shape. One row per user. P5 (VOC-031) is the first
// milestone to read or write this table; T01 will persist the five onboarding
// answers and the T00 seed function will use `daily_review_target` to seed
// `user_settings` per VOC-031-D04.
type UserOnboardingProfile struct{ ent.Schema }

func (UserOnboardingProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "user_onboarding_profiles",
			Checks: map[string]string{
				"english_level_valid":   "english_level IN ('a1', 'a2', 'b1', 'b2', 'unknown')",
				"native_language_set":   "char_length(native_language) > 0",
				"learning_goal_valid":   "learning_goal IN ('general', 'work', 'travel', 'study', 'conversation', 'exam')",
				"main_use_case_valid":   "main_use_case IN ('daily_life', 'work', 'travel', 'study', 'social')",
				"daily_review_in_range": "daily_review_target >= 5 AND daily_review_target <= 100",
			},
		},
	}
}

func (UserOnboardingProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (UserOnboardingProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Unique(),
		field.String("english_level").Default("unknown"),
		field.String("native_language"),
		field.String("learning_goal"),
		field.String("main_use_case"),
		field.Int("daily_review_target").Min(5).Max(100),
		field.Time("completed_at").Immutable(),
	}
}

func (UserOnboardingProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("user_onboarding_profile").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (UserOnboardingProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}
