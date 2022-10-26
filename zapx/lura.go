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

import "go.uber.org/zap"

// LuraLogger provides the methods needed for a logger from lura. This wraps the
// zap logger methods and provides all the required Lura methods.
type LuraLogger struct {
	logger *zap.SugaredLogger
}

// NewLuraLogger returns a new LuraLogger from a zap.SugaredLogger
func NewLuraLogger(logger *zap.SugaredLogger) *LuraLogger {
	return &LuraLogger{logger: logger}
}

func (l *LuraLogger) Debug(v ...interface{})    { l.logger.Debug(v...) } //nolint:revive
func (l *LuraLogger) Info(v ...interface{})     { l.logger.Info(v...) }  //nolint:revive
func (l *LuraLogger) Warning(v ...interface{})  { l.logger.Warn(v...) }  //nolint:revive
func (l *LuraLogger) Error(v ...interface{})    { l.logger.Error(v...) } //nolint:revive
func (l *LuraLogger) Critical(v ...interface{}) { l.logger.Error(v...) } //nolint:revive
func (l *LuraLogger) Fatal(v ...interface{})    { l.logger.Fatal(v...) } //nolint:revive
