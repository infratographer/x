// Copyright 2022 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Looking around at echo loggers there's github.com/ziflex/lecho which provides a
// generic interface to zerolog as middleware and reduces the overall code footprint
// the Infratographer project maintains relative to zap. This is a good thing.
package loggingx

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.infratographer.com/x/versionx"
)

// InitZeroLogger returns a logger based on the config
func InitZeroLogger(appName string, cfg Config) *zerolog.Logger {
	return InitLoggerWithWriter(appName, cfg, os.Stdout)
}

func InitLoggerWithWriter(appName string, cfg Config, out io.Writer) *zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.Debug {
		zerolog.TimeFieldFormat = time.RFC3339
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	fields := map[string]interface{}{
		"app":     appName,
		"version": versionx.BuildDetails().Version,
	}

	if cfg.Pretty {
		zerolog.TimeFieldFormat = time.RFC3339
		out = zerolog.ConsoleWriter{
			NoColor:    true,
			Out:        out,
			TimeFormat: time.RFC3339,
		}
	}
	ctx := zerolog.New(out).
		With().
		Fields(fields).
		Timestamp()

	zl := ctx.Logger()

	return &zl
}
