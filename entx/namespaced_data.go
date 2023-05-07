package entx

import (
	"encoding/json"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var (
	namespaceMinLength = 5
	namespaceMaxLength = 64
)

// NamespacedDataMixin defines an ent Mixin that captures raw json associated with a namespace.
type NamespacedDataMixin struct {
	mixin.Schema
}

// Fields provides the namespace and data fields used in this mixin.
func (m NamespacedDataMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Text("namespace").
			NotEmpty().
			MinLen(namespaceMinLength).
			MaxLen(namespaceMaxLength).
			Annotations(
				entgql.OrderField("NAMESPACE"),
			),
		field.JSON("data", json.RawMessage{}).
			Annotations(
				entgql.Type("JSON"),
				Annotation{IsNamespacedDataJSONField: true},
			),
	}
}
