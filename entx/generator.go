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
	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// Extension is an implementation of entc.Extension that adds all the templates
// that entx needs.
type Extension struct {
	entc.DefaultExtension

	templates []*gen.Template

	gqlSchemaHooks []entgql.SchemaHook
}

// ExtensionOption allow for control over the behavior of the generator
type ExtensionOption func(*Extension) error

// WithFederation adds support for graphql federation by adding the Entity interface
// to all types, as well as removing the node() and nodes() query calls.
func WithFederation() ExtensionOption {
	return func(ex *Extension) error {
		ex.templates = append(ex.templates, FederationTemplate)
		ex.gqlSchemaHooks = append(ex.gqlSchemaHooks, removeNodeGoModel, removeNodeQueries, setPageInfoShareable)

		return nil
	}
}

// WithConnectionNodes adds the templates for adding nodes to relay connections
func WithConnectionNodes() ExtensionOption {
	return func(ex *Extension) error {
		ex.templates = append(ex.templates, PaginationTemplate, CollectionTemplate)
		ex.gqlSchemaHooks = append(ex.gqlSchemaHooks, addNodesToConnections)

		return nil
	}
}

// WithJSONScalar adds the JSON scalar definition
func WithJSONScalar() ExtensionOption {
	return func(ex *Extension) error {
		ex.gqlSchemaHooks = append(ex.gqlSchemaHooks, addJSONScalar)
		return nil
	}
}

// WithEventHooks adds the templates for generating event hooks
func WithEventHooks() ExtensionOption {
	return func(ex *Extension) error {
		ex.templates = append(ex.templates, EventHooksTemplate)

		return nil
	}
}

// NewExtension returns an entc Extension that allows the entx package to generate
// the schema changes and templates needed to function
func NewExtension(opts ...ExtensionOption) (*Extension, error) {
	e := &Extension{
		templates:      []*gen.Template{},
		gqlSchemaHooks: []entgql.SchemaHook{},
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	return e, nil
}

// Templates of the extension
func (e *Extension) Templates() []*gen.Template {
	return e.templates
}

// GQLSchemaHooks of the extension to seamlessly edit the final gql interface.
func (e *Extension) GQLSchemaHooks() []entgql.SchemaHook {
	return e.gqlSchemaHooks
}

var _ entc.Extension = (*Extension)(nil)
