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

type WordMeaning struct{ ent.Schema }

func (WordMeaning) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "word_meanings"}}
}
func (WordMeaning) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }

func (WordMeaning) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("word_id", uuid.UUID{}),
		field.Enum("part_of_speech").
			Values("noun", "verb", "adjective", "adverb", "preposition", "conjunction", "interjection", "pronoun", "determiner", "phrase", "idiom", "phrasal_verb", "collocation", "other"),
		field.String("short_definition").NotEmpty(),
		field.String("learner_definition").Optional().Nillable(),
		field.Int("meaning_order").Positive(),
		field.Enum("status").
			Values("draft", "active", "archived").
			Default("draft"),
		field.Enum("difficulty_level").
			Values("a1", "a2", "b1", "b2", "c1", "unknown").
			Optional().Nillable(),
	}
}

func (WordMeaning) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("word", CanonicalWord.Type).Ref("meanings").Field("word_id").Unique().Required(),
		edge.To("examples", WordExample.Type),
		edge.To("usage_notes", UsageNote.Type),
		edge.To("journey_words", JourneyWord.Type),
		edge.To("user_words", UserWord.Type),
		edge.To("review_attempts", ReviewAttempt.Type),
		edge.To("selected_option_attempts", ReviewAttempt.Type),
	}
}

func (WordMeaning) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("word_id", "meaning_order").Unique(),
	}
}
