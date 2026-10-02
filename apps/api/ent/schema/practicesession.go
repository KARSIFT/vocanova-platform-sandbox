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

// PracticeSession is repeatable self-study, independent of SRS and lesson completion.
// Explicit migration owns composite ownership and progress constraints.
type PracticeSession struct{ ent.Schema }

func (PracticeSession) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "practice_sessions"}}
}
func (PracticeSession) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }
func (PracticeSession) Fields() []ent.Field {
	return []ent.Field{field.UUID("user_id", uuid.UUID{}).Immutable(), field.Enum("mode").Values("typed_recall", "listening_choice", "mistakes").Immutable(), field.String("lesson_key").Default("").Immutable(), field.JSON("snapshot", json.RawMessage{}).Immutable(), field.Int("current_step").Default(0), field.Int("total_steps"), field.Int("revision").Default(0), field.JSON("feedback", json.RawMessage{}).Optional(), field.Int("first_answers_correct").Default(0), field.Int("questions_answered").Default(0), field.Enum("status").Values("in_progress", "completed").Default("in_progress"), field.Time("completed_at").Optional().Nillable()}
}
func (PracticeSession) Indexes() []ent.Index {
	return []ent.Index{index.Fields("id", "user_id").Unique(), index.Fields("user_id", "updated_at", "id")}
}
