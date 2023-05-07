package entx

import (
	"embed"
	"text/template"

	"entgo.io/ent/entc/gen"
)

var (
	// PrefixedIDTemplate adds support for generating the methods needed to convert a prefixed ID to it's backend type
	PrefixedIDTemplate = parseT("template/id_prefixes.tmpl")

	// FederationTemplate adds support for generating the required output to support gql federation
	FederationTemplate = parseT("template/gql_federation.tmpl")

	// NamespacedDataWhereFuncsTemplate adds support for generating <T>WhereInput filters for schema types using the NamespacedData mixin
	NamespacedDataWhereFuncsTemplate = parseT("template/namespaceddata_where_funcs.tmpl")

	// TemplateFuncs contains the extra template functions used by entx.
	TemplateFuncs = template.FuncMap{}

	// MixinTemplates includes all templates for extending ent to support entx mixins.
	MixinTemplates = []*gen.Template{
		NamespacedDataWhereFuncsTemplate,
	}

	//go:embed template/*
	_templates embed.FS
)

func parseT(path string) *gen.Template {
	return gen.MustParse(gen.NewTemplate(path).
		Funcs(TemplateFuncs).
		ParseFS(_templates, path))
}
