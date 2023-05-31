// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package crdbx

import (
	"net/url"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
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
func MustViperFlags(v *viper.Viper, _ *pflag.FlagSet) {
	v.MustBindEnv("crdb.host")
	v.MustBindEnv("crdb.params")
	v.MustBindEnv("crdb.user")
	v.MustBindEnv("crdb.password")
	v.MustBindEnv("crdb.uri")
	v.MustBindEnv("crdb.connections.max_open")
	v.MustBindEnv("crdb.connections.max_idle")
	v.MustBindEnv("crdb.connections.max_lifetime")

	v.SetDefault("crdb.host", "localhost:26257")
	v.SetDefault("crdb.connections.max_open", defaultMaxOpenConns)
	v.SetDefault("crdb.connections.max_idle", defaultMaxIdleConns)
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
