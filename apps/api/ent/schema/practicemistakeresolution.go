package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Exact source events are resolved: a later mistake cannot be cleared by an old session.
type PracticeMistakeResolution struct{ ent.Schema }

func (PracticeMistakeResolution) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "practice_mistake_resolutions"}}
}
func (PracticeMistakeResolution) Fields() []ent.Field {
	return []ent.Field{field.UUID("id", uuid.UUID{}).Default(uuid.New).Annotations(entsql.DefaultExpr("gen_random_uuid()")), field.UUID("user_id", uuid.UUID{}).Immutable(), field.Enum("source_kind").Values("lesson", "review", "practice").Immutable(), field.UUID("source_id", uuid.UUID{}).Immutable(), field.UUID("meaning_id", uuid.UUID{}).Immutable(), field.UUID("session_id", uuid.UUID{}).Immutable(), field.Time("resolved_at").Immutable()}
}
func (PracticeMistakeResolution) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "source_kind", "source_id").Unique()}
}
