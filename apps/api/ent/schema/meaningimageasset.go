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

type MeaningImageAsset struct{ ent.Schema }

func (MeaningImageAsset) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "meaning_image_assets"}}
}
func (MeaningImageAsset) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}} }
func (MeaningImageAsset) Fields() []ent.Field {
	return []ent.Field{field.UUID("meaning_id", uuid.UUID{}).Immutable(), field.String("meaning_revision").NotEmpty().Immutable(), field.String("sha256").NotEmpty().Immutable(), field.String("alt").NotEmpty().Immutable(), field.Int("width").Positive().Immutable(), field.Int("height").Positive().Immutable(), field.Bool("approved").Default(false), field.Time("created_at").Immutable()}
}
func (MeaningImageAsset) Indexes() []ent.Index {
	return []ent.Index{index.Fields("meaning_id", "meaning_revision").Unique()}
}

func (MeaningImageAsset) Edges() []ent.Edge {
	return []ent.Edge{edge.From("meaning", WordMeaning.Type).Ref("images").Field("meaning_id").Unique().Required()}
}
