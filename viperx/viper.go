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

package viperx

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// MustBindFlag provides a wrapper around the viper bindings that panics if an error occurs
func MustBindFlag(v *viper.Viper, name string, flag *pflag.Flag) {
	err := v.BindPFlag(name, flag)
	if err != nil {
		panic(err)
	}
}
