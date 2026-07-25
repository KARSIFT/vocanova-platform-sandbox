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

type UsageNote struct{ ent.Schema }

func (UsageNote) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "usage_notes"}}
}
func (UsageNote) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }

func (UsageNote) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("meaning_id", uuid.UUID{}),
		field.Enum("note_type").
			Values("collocation", "register", "common_mistake", "grammar", "pronunciation", "other"),
		field.String("note_text").NotEmpty(),
		field.Int("note_order").Positive(),
		field.Enum("status").
			Values("draft", "active", "archived").
			Default("draft"),
	}
}

func (UsageNote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("meaning", WordMeaning.Type).Ref("usage_notes").Field("meaning_id").Unique().Required(),
	}
}

func (UsageNote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("meaning_id", "note_order").Unique(),
	}
}
