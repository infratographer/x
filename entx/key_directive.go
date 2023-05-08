package entx

import (
	"entgo.io/contrib/entgql"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphKeyDirective returns an entgql.Directive for setting the @key field on
// a graphql type
func GraphKeyDirective(fields string) entgql.Annotation {
	return entgql.Directives(keyDirective(fields))
}

func keyDirective(fields string) entgql.Directive {
	var args []*ast.Argument
	if fields != "" {
		args = append(args, &ast.Argument{
			Name: "fields",
			Value: &ast.Value{
				Raw:  fields,
				Kind: ast.StringValue,
			},
		})
	}

	return entgql.NewDirective("key", args...)
}
