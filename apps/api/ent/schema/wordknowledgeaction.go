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

// WordKnowledgeAction retains only replay fingerprints, not deleted note text.
type WordKnowledgeAction struct{ ent.Schema }

func (WordKnowledgeAction) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "word_knowledge_actions"}}
}
func (WordKnowledgeAction) Fields() []ent.Field {
	return []ent.Field{field.UUID("id", uuid.UUID{}).Default(uuid.New).Annotations(entsql.DefaultExpr("gen_random_uuid()")), field.UUID("user_id", uuid.UUID{}).Immutable(), field.String("idempotency_key").NotEmpty().MaxLen(128).Immutable(), field.String("fingerprint").MaxLen(64).Immutable(), field.Time("created_at").Default(time.Now).Immutable()}
}
func (WordKnowledgeAction) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "idempotency_key").Unique()}
}
