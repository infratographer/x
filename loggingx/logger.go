// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package loggingx

import (
	"go.uber.org/zap"

	"go.infratographer.com/x/versionx"
)

// InitLogger returns a logger based on the config
func InitLogger(appName string, cfg Config) *zap.SugaredLogger {
	lgrCfg := zap.NewProductionConfig()
	if cfg.Pretty {
		lgrCfg = zap.NewDevelopmentConfig()
	}

	if cfg.Debug {
		lgrCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		lgrCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	l, err := lgrCfg.Build()
	if err != nil {
		panic(err)
	}

	return l.Sugar().With(
		"app", appName,
		"version", versionx.BuildDetails().Version,
	)
}
