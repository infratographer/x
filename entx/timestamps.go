// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package entx

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// TimestampsMixin defines an interface of a Mixin that provides the created_at and updated_at timestamp fields
type TimestampsMixin interface {
	ent.Mixin
	CreatedAtAnnotations(...schema.Annotation) TimestampsMixin
	UpdatedAtAnnotations(...schema.Annotation) TimestampsMixin
}

type timestampsMixin struct {
	mixin.Schema
	createdAnnotations []schema.Annotation
	updatedAnnotations []schema.Annotation
}

// NewTimestampMixin returns a Mixin that provides the created_at and updated_at timestamp fields
func NewTimestampMixin() TimestampsMixin {
	return timestampsMixin{
		createdAnnotations: []schema.Annotation{
			entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			entgql.OrderField("CREATED_AT"),
		},
		updatedAnnotations: []schema.Annotation{
			entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			entgql.OrderField("UPDATED_AT"),
		},
	}
}

func (m timestampsMixin) CreatedAtAnnotations(ants ...schema.Annotation) TimestampsMixin {
	m.createdAnnotations = ants
	return m
}

func (m timestampsMixin) UpdatedAtAnnotations(ants ...schema.Annotation) TimestampsMixin {
	m.updatedAnnotations = ants
	return m
}

// Fields provides the created_at and updated_at fields
func (m timestampsMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Annotations(m.createdAnnotations...),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Immutable().
			Annotations(m.updatedAnnotations...),
	}
}

// Indexes provides indexes on both created_at and updated_at fields
func (timestampsMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Fields("updated_at"),
	}
}
