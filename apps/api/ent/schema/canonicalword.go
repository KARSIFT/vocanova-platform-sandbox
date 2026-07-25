package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CanonicalWord struct{ ent.Schema }

func (CanonicalWord) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "canonical_words"}}
}
func (CanonicalWord) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }

func (CanonicalWord) Fields() []ent.Field {
	return []ent.Field{
		field.String("text").NotEmpty(),
		field.String("normalized_text").NotEmpty(),
		field.Enum("word_type").
			Values("word", "phrase", "phrasal_verb", "idiom", "collocation").
			Default("word"),
		field.String("language_code").NotEmpty().Default("en"),
		field.Enum("status").
			Values("draft", "active", "archived").
			Default("draft"),
		field.Enum("difficulty_level").
			Values("a1", "a2", "b1", "b2", "c1", "unknown").
			Optional().Nillable(),
		field.Int("frequency_rank").Optional().Nillable(),
	}
}

func (CanonicalWord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("meanings", WordMeaning.Type),
	}
}

func (CanonicalWord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("language_code", "normalized_text").Unique(),
	}
}
