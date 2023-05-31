// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

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

	// EventHooksTemplate adds support for generating event hooks
	EventHooksTemplate = parseT("template/event_hooks.tmpl")

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
