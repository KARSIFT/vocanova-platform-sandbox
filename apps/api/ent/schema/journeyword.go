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

type JourneyWord struct{ ent.Schema }

func (JourneyWord) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "journey_words"}}
}
func (JourneyWord) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }

func (JourneyWord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("journey_situation_id", uuid.UUID{}),
		field.UUID("meaning_id", uuid.UUID{}),
		field.Int("relevance_score").Min(1).Max(100).Default(50),
		field.Int("display_order").Optional().Nillable(),
		field.Bool("is_core").Default(false),
	}
}

func (JourneyWord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("journey_situation", JourneySituation.Type).Ref("journey_words").Field("journey_situation_id").Unique().Required(),
		edge.From("meaning", WordMeaning.Type).Ref("journey_words").Field("meaning_id").Unique().Required(),
	}
}

func (JourneyWord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("journey_situation_id", "meaning_id").Unique(),
	}
}
