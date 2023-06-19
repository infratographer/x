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

package echojwtx

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

const (
	// DefaultOIDCJWKSRemoteTimeout defines the default timeout for fetching the OIDC JWKS file.
	DefaultOIDCJWKSRemoteTimeout = 5 * time.Second
)

// MustViperFlags adds jwks-uri to the provided flagset and binds to viper jwks.uri.
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("oidc", true, "use oidc auth")
	viperx.MustBindFlag(v, "oidc.enabled", flags.Lookup("oidc"))

	flags.String("oidc-aud", "", "expected audience on OIDC JWT")
	viperx.MustBindFlag(v, "oidc.audience", flags.Lookup("oidc-aud"))

	flags.String("oidc-issuer", "", "expected issuer of OIDC JWT")
	viperx.MustBindFlag(v, "oidc.issuer", flags.Lookup("oidc-issuer"))

	flags.Duration("oidc-jwks-remote-timeout", DefaultOIDCJWKSRemoteTimeout, "timeout for remote JWKS fetching")
	viperx.MustBindFlag(v, "oidc.jwks.remote-timeout", flags.Lookup("oidc-jwks-remote-timeout"))
}
