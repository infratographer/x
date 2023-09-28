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
	"encoding/json"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// MarshalRawMessage provides a graphql.Marshaler for json.RawMessage
func MarshalRawMessage(t json.RawMessage) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		s, _ := t.MarshalJSON()
		_, _ = io.WriteString(w, string(s))
	})
}

// UnmarshalRawMessage provides a graphql.Unmarshaler for json.RawMessage
func UnmarshalRawMessage(v interface{}) (json.RawMessage, error) {
	switch j := v.(type) {
	case []byte:
		return json.RawMessage(j), nil
	case map[string]interface{}:
		js, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		return json.RawMessage(js), nil
	default:
		// Attempt to cast it as a fall back but return an error if it fails
		js, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		return json.RawMessage(js), nil
	}
}
