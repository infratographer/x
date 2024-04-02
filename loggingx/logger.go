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

	if cfg.DisableStacktrace {
		lgrCfg.DisableStacktrace = true
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
