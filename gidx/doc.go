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
