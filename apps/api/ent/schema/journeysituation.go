package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type JourneySituation struct{ ent.Schema }

func (JourneySituation) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "journey_situations"}}
}
func (JourneySituation) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }

func (JourneySituation) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").NotEmpty(),
		field.String("title").NotEmpty(),
		field.String("short_description").NotEmpty(),
		field.Enum("level_band").
			Values("a1_a2", "a2_b1", "b1_b2", "mixed").
			Optional().Nillable(),
		field.Enum("category").
			Values("daily_life", "travel", "work", "study", "social"),
		field.Enum("status").
			Values("draft", "active", "archived").
			Default("draft"),
		field.Int("display_order").Positive(),
	}
}

func (JourneySituation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("journey_words", JourneyWord.Type),
	}
}

func (JourneySituation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("status", "display_order"),
	}
}
