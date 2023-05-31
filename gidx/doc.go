// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

// Package gidx creates and parses Infratographer-based Global IDs. These IDs
// are in the format of <resource type prefix>-<resource id> and are provided
// by the PrefixedID type. The IDs are intended to follow the Relay Global
// Object Identification Specification (https://relay.dev/graphql/objectidentification.htm),
// ensuring global uniqueness and the ability to resolve object types with only
// their IDs. The 7 character prefix includes the application and object type,
// allowing for easy identification of object types from the ID.
//
// Prefixed IDs use a 7 character prefix, with the first 4 characters
// representing the application and the next 3 characters representing the
// object type. This allows for easy identification of the object type from the
// ID. For example, the instance-api application might use the 4 character
// prefix inst and have an object type of instance. The 3 character code for
// instance might be anc, resulting in a prefix of instanc. An instance ID might
// then look like instanc-myrandomidvalue.
package gidx
