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
	"errors"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/vektah/gqlparser/v2/ast"
)

// Skipping err113 linting since these errors are returned during generation and not runtime
//
//nolint:goerr113
var (
	removeNodeGoModel = func(_ *gen.Graph, s *ast.Schema) error {
		n, ok := s.Types["Node"]
		if !ok {
			return errors.New("failed to find node interface in schema")
		}

		dirs := ast.DirectiveList{}

		for _, d := range n.Directives {
			switch d.Name {
			case "goModel":
				continue
			default:
				dirs = append(dirs, d)
			}
		}
		n.Directives = dirs

		return nil
	}

	removeNodeQueries = func(_ *gen.Graph, s *ast.Schema) error {
		q, ok := s.Types["Query"]
		if !ok {
			return errors.New("failed to find query definition in schema")
		}

		fields := ast.FieldList{}

		for _, f := range q.Fields {
			switch f.Name {
			case "node":
			case "nodes":
				continue
			default:
				fields = append(fields, f)
			}
		}
		q.Fields = fields

		return nil
	}

	setPageInfoShareable = func(_ *gen.Graph, s *ast.Schema) error {
		q, ok := s.Types["PageInfo"]
		if !ok {
			return nil
		}

		q.Directives = append(q.Directives, &ast.Directive{Name: "shareable"})

		return nil
	}

	addJSONScalar = func(_ *gen.Graph, s *ast.Schema) error {
		s.Types["JSON"] = &ast.Definition{
			Kind:        ast.Scalar,
			Description: "A valid JSON string.",
			Name:        "JSON",
		}
		return nil
	}
)

// import string mutations from entc
var (
	_ entc.Extension = (*Extension)(nil)
)
