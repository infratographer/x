// Copyright 2022 The Infratographer Authors
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

package crdbx

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.infratographer.com/x/viperx"
)

const (
	defaultHost            string        = "localhost:26257"
	defaultMaxOpenConns    int           = 25
	defaultMaxIdleConns    int           = 25
	defaultMaxConnLifetime time.Duration = 5 * 60 * time.Second
)

// Config is used to configure a new cockroachdb connection
type Config struct {
	Name        string `mapstructure:"name"`
	Host        string `mapstructure:"host"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Params      string `mapstructure:"params"`
	URI         string `mapstructure:"uri"`
	Connections struct {
		MaxOpen     int           `mapstructure:"max_open"`
		MaxIdle     int           `mapstructure:"max_idle"`
		MaxLifetime time.Duration `mapstructure:"max_lifetime"`
	}
}

// GetURI returns the connection URI, if a config URI is provided that will be
// returned, otherwise the host, user, password, and params will be put together
// to make a URI that is returned.
func (c Config) GetURI() string {
	if c.URI != "" {
		return c.URI
	}

	u := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(c.User, c.Password),
		Host:     c.Host,
		Path:     c.Name,
		RawQuery: c.Params,
	}

	return u.String()
}

// MustViperFlags returns the cobra flags and viper config to prevent code duplication
// and help provide consistent flags across the applications
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.String("crdb.host", "", fmt.Sprintf("Database host. (default %q)", defaultHost))
	viperx.MustBindFlag(v, "crdb.host", flags.Lookup("crdb.host"))
	v.MustBindEnv("crdb.host")

	flags.String("crdb.params", "", "Database connection parameters.")
	viperx.MustBindFlag(v, "crdb.params", flags.Lookup("crdb.params"))
	v.MustBindEnv("crdb.params")

	flags.String("crdb.user", "", "Database user.")
	viperx.MustBindFlag(v, "crdb.user", flags.Lookup("crdb.user"))
	v.MustBindEnv("crdb.user")
	v.SetDefault("crdb.host", defaultHost)

	flags.String("crdb.password", "", "Database password.")
	viperx.MustBindFlag(v, "crdb.password", flags.Lookup("crdb.password"))
	v.MustBindEnv("crdb.password")

	flags.String("crdb.uri", "", "Database connection URI.")
	viperx.MustBindFlag(v, "crdb.uri", flags.Lookup("crdb.uri"))
	v.MustBindEnv("crdb.uri")

	flags.Int("crdb.connections.max_open", 0, fmt.Sprintf("Database max open connections. (default %d)", defaultMaxOpenConns))
	viperx.MustBindFlag(v, "crdb.connections.max_open", flags.Lookup("crdb.connections.max_open"))
	v.MustBindEnv("crdb.connections.max_open")
	v.SetDefault("crdb.connections.max_open", defaultMaxOpenConns)

	flags.Int("crdb.connections.max_idle", 0, fmt.Sprintf("Database max idle connections. (default %d)", defaultMaxIdleConns))
	viperx.MustBindFlag(v, "crdb.connections.max_idle", flags.Lookup("crdb.connections.max_idle"))
	v.MustBindEnv("crdb.connections.max_idle")
	v.SetDefault("crdb.connections.max_idle", defaultMaxIdleConns)

	flags.Duration("crdb.connections.max_lifetime", 0, fmt.Sprintf("Database max connection lifetime. (default %s)", defaultMaxConnLifetime))
	viperx.MustBindFlag(v, "crdb.connections.max_lifetime", flags.Lookup("crdb.connections.max_lifetime"))
	v.MustBindEnv("crdb.connections.max_lifetime")
	v.SetDefault("crdb.connections.max_lifetime", defaultMaxConnLifetime)
}

// ConfigFromArgs returns a crdbx.Config from the provided viper-provided
// flags.
func ConfigFromArgs(v *viper.Viper, dbName string) Config {
	cfg := Config{
		Name:     dbName,
		Host:     v.GetString("crdb.host"),
		User:     v.GetString("crdb.user"),
		Password: v.GetString("crdb.password"),
		Params:   v.GetString("crdb.params"),
		URI:      v.GetString("crdb.uri"),
	}

	cfg.Connections.MaxOpen = v.GetInt("crdb.connections.max_open")
	cfg.Connections.MaxIdle = v.GetInt("crdb.connections.max_idle")
	cfg.Connections.MaxLifetime = v.GetDuration("crdb.connections.max_lifetime")

	return cfg
}
