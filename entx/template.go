// Copyright 2023 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entx

import (
	"embed"
	"strings"
	"text/template"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"github.com/mitchellh/mapstructure"
)

var (
	// FederationTemplate adds support for generating the required output to support gql federation
	FederationTemplate = parseT("template/gql_federation.tmpl")

	// EventHooksTemplate adds support for generating event hooks
	EventHooksTemplate = parseT("template/event_hooks.tmpl")

	// PaginationTemplate adds support for adding the nodes field to relay connections
	PaginationTemplate = parseT("template/pagination.tmpl")

	// TemplateFuncs contains the extra template functions used by entx.
	TemplateFuncs = template.FuncMap{
		"contains":                          strings.Contains,
		"hasNonSensitiveAdditionalSubjects": hasNonSensitiveAdditionalSubjects,
	}

	//go:embed template/*
	_templates embed.FS
)

func parseT(path string) *gen.Template {
	funcMap := entgql.TemplateFuncs

	for k, v := range TemplateFuncs {
		funcMap[k] = v
	}

	return gen.MustParse(gen.NewTemplate(path).
		Funcs(funcMap).
		ParseFS(_templates, path))
}

func hasNonSensitiveAdditionalSubjects(node *gen.Type) bool {
	for _, f := range node.Fields {
		if !f.Sensitive() {
			if ann0, ok := f.Annotations[EventsHookAnnotationName]; ok {
				ann := EventsHookAnnotation{}
				if err := mapstructure.Decode(ann0, &ann); err == nil {
					if ann.IsAdditionalSubjectField || ann.AdditionalSubjectRelation != "" {
						return true
					}
				}
			}
		}
	}

	return false
}
