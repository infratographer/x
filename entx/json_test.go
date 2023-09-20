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
	"reflect"
	"testing"
)

func TestUnmarshalRawMessage(t *testing.T) {
	tests := []struct {
		name    string
		arg     interface{}
		want    json.RawMessage
		wantErr bool
	}{{
		name: "map",
		arg:  map[string]any{"a": true},
		want: json.RawMessage(`{"a":true}`),
	}, {
		name: "bytes",
		arg:  []byte{'"', 'a', '"'},
		want: json.RawMessage(`"a"`),
	}, {
		// In practice, this is the way graphql Unmarshal is processing input like {json: "a"}:
		name: "string",
		arg:  "a",
		want: json.RawMessage(`"a"`),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRawMessage(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalRawMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalRawMessage() = %s, want %s", got, tt.want)
			}
		})
	}
}
