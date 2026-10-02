package schema

import (
	"encoding/json"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LessonSession stores a versioned guided-practice snapshot, not SRS mastery.
// The explicit migration owns the composite owner FK and progress checks.
type LessonSession struct{ ent.Schema }

func (LessonSession) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "lesson_sessions"}}
}
func (LessonSession) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }
func (LessonSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Immutable(), field.String("lesson_key").NotEmpty().Immutable(), field.String("lesson_version").Immutable(),
		field.JSON("snapshot", json.RawMessage{}).Immutable(), field.Int("current_step").Default(0), field.Int("total_steps"), field.Int("revision").Default(0),
		field.JSON("feedback", json.RawMessage{}).Optional(), field.Int("first_answers_correct").Default(0), field.Int("questions_answered").Default(0),
		field.Enum("status").Values("in_progress", "completed").Default("in_progress"), field.Time("completed_at").Optional().Nillable(),
	}
}
func (LessonSession) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "lesson_key").Unique(), index.Fields("id", "user_id").Unique()}
}
