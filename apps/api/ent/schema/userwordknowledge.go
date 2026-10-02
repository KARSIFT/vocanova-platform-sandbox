package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// UserWordKnowledge is a private self-assessment, never scheduling or mastery.
type UserWordKnowledge struct{ ent.Schema }

func (UserWordKnowledge) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_word_knowledge"}}
}
func (UserWordKnowledge) Fields() []ent.Field {
	return []ent.Field{field.UUID("id", uuid.UUID{}).Default(uuid.New).Annotations(entsql.DefaultExpr("gen_random_uuid()")), field.UUID("user_id", uuid.UUID{}).Immutable(), field.UUID("meaning_id", uuid.UUID{}).Immutable(), field.Bool("self_reported_known").Default(false), field.Text("note").Default(""), field.Time("updated_at").Default(time.Now)}
}
func (UserWordKnowledge) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "meaning_id").Unique()}
}
