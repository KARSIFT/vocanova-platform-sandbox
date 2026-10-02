package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserLearningPreferences stores current direction without rewriting onboarding.
type UserLearningPreferences struct{ ent.Schema }

func (UserLearningPreferences) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_learning_preferences", Checks: map[string]string{
		"learning_goal_valid": "learning_goal IN ('general','work','travel','study','conversation','exam')",
		"main_use_case_valid": "main_use_case IN ('daily_life','work','travel','study','social')",
		"revision_positive":   "revision > 0",
	}}}
}
func (UserLearningPreferences) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }
func (UserLearningPreferences) Fields() []ent.Field {
	return []ent.Field{field.UUID("user_id", uuid.UUID{}).Unique().Immutable(), field.String("learning_goal"), field.String("main_use_case"), field.Int64("revision").Positive()}
}
