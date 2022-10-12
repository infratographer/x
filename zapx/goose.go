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

package zapx

import (
	"strings"

	"go.uber.org/zap"
)

// GooseLogger provides the methods needed for a logger in the goose package.
// This wraps the zap logger and in some cases formats data to ensure it prints
// out nicely.
type GooseLogger struct {
	logger *zap.SugaredLogger
}

// NewGooseLogger returns a new GooseLogger from a zap logger
func NewGooseLogger(logger *zap.SugaredLogger) *GooseLogger {
	return &GooseLogger{logger: logger}
}

func (l *GooseLogger) Fatal(v ...interface{})                 { l.logger.Fatal(v...) }                   //nolint:revive
func (l *GooseLogger) Fatalf(format string, v ...interface{}) { l.logger.Fatalf(l.clean(format), v...) } //nolint:revive
func (l *GooseLogger) Print(v ...interface{})                 { l.logger.Info(v...) }                    //nolint:revive
func (l *GooseLogger) Println(v ...interface{})               { l.logger.Infoln(v...) }                  //nolint:revive
func (l *GooseLogger) Printf(format string, v ...interface{}) { l.logger.Infof(l.clean(format), v...) }  //nolint:revive

func (l *GooseLogger) clean(str string) string {
	str = strings.TrimPrefix(str, "goose: ")
	return strings.TrimSuffix(str, "\n")
}
