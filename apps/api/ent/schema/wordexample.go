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

type WordExample struct{ ent.Schema }

func (WordExample) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "word_examples"}}
}
func (WordExample) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }

func (WordExample) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("meaning_id", uuid.UUID{}),
		field.String("example_text").NotEmpty(),
		field.Int("example_order").Positive(),
		field.Enum("difficulty_level").
			Values("a1", "a2", "b1", "b2", "c1", "unknown").
			Optional().Nillable(),
		field.String("situation_label").Optional().Nillable(),
		field.Enum("status").
			Values("draft", "active", "archived").
			Default("draft"),
	}
}

func (WordExample) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("meaning", WordMeaning.Type).Ref("examples").Field("meaning_id").Unique().Required(),
	}
}

func (WordExample) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("meaning_id", "example_order").Unique(),
	}
}
