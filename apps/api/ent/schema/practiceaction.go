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

type PracticeAction struct{ ent.Schema }

func (PracticeAction) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "practice_actions"}}
}
func (PracticeAction) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}} }
func (PracticeAction) Fields() []ent.Field {
	return []ent.Field{field.UUID("session_id", uuid.UUID{}).Immutable(), field.UUID("user_id", uuid.UUID{}).Immutable(), field.Enum("operation").Values("start", "action").Immutable(), field.String("idempotency_key").NotEmpty().Immutable(), field.String("client_action_id").NotEmpty().Immutable(), field.String("fingerprint").Immutable(), field.JSON("action", json.RawMessage{}).Immutable(), field.JSON("result", json.RawMessage{}).Immutable(), field.UUID("meaning_id", uuid.UUID{}).Optional().Nillable().Immutable(), field.Bool("correct").Optional().Nillable().Immutable(), field.Time("created_at").Immutable()}
}
func (PracticeAction) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "operation", "idempotency_key").Unique(), index.Fields("session_id", "operation", "client_action_id").Unique(), index.Fields("user_id", "meaning_id", "created_at").Annotations(entsql.IndexWhere("correct = false"))}
}
