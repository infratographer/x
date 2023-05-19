package entx

import (
	"embed"
	"strings"
	"text/template"

	"entgo.io/ent/entc/gen"
)

var (
	// FederationTemplate adds support for generating the required output to support gql federation
	FederationTemplate = parseT("template/gql_federation.tmpl")

	// PubsubHooksTemplate adds support for generating pubsub hooks
	PubsubHooksTemplate = parseT("template/pubsub_hooks.tmpl")

	// TemplateFuncs contains the extra template functions used by entx.
	TemplateFuncs = template.FuncMap{
		"contains": strings.Contains,
	}

	//go:embed template/*
	_templates embed.FS
)

func parseT(path string) *gen.Template {
	return gen.MustParse(gen.NewTemplate(path).
		Funcs(TemplateFuncs).
		ParseFS(_templates, path))
}
