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

package echozap

import (
	"io"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

var _ echo.Logger = (*ZapLogger)(nil)

// ZapLogger implements echo's Logger interface.
type ZapLogger struct {
	logger *zap.SugaredLogger
}

// Output statically returns os.Stderr as zap doesn't expose the underlying writer.
func (l *ZapLogger) Output() io.Writer {
	return os.Stderr
}

// SetOutput does nothing, as zaps underlying writer is unable to be changed.
func (l *ZapLogger) SetOutput(_ io.Writer) {

}

// Prefix returns an empty string as zap doesn't expose the underlying named logger.
func (l *ZapLogger) Prefix() string {
	return ""
}

// SetPrefix updates named logger with the provided prefix.
func (l *ZapLogger) SetPrefix(p string) {
	l.logger = l.logger.Named(p)
}

// Level returns an equivalent echo log level.
func (l *ZapLogger) Level() log.Lvl {
	switch l.logger.Level() {
	case zap.DebugLevel:
		return log.DEBUG
	case zap.InfoLevel:
		return log.INFO
	case zap.WarnLevel:
		return log.WARN
	default:
		return log.ERROR
	}
}

// SetLevel does nothing as zap doesn't expose a way to change the level for a logger.
func (l *ZapLogger) SetLevel(_ log.Lvl) {
}

// SetHeader does nothing.
func (l *ZapLogger) SetHeader(_ string) {

}

// Print implements echo's Logger interface.
func (l *ZapLogger) Print(i ...interface{}) {
	l.logger.Info(i...)
}

// Printf implements echo's Logger interface.
func (l *ZapLogger) Printf(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Printj implements echo's Logger interface.
func (l *ZapLogger) Printj(j log.JSON) {
	var args []interface{}

	for k, v := range j {
		args = append(args, k, v)
	}

	l.logger.Info(args...)
}

// Debug implements echo's Logger interface.
func (l *ZapLogger) Debug(i ...interface{}) {
	l.logger.Debug(i...)
}

// Debugf implements echo's Logger interface.
func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Debugj implements echo's Logger interface.
func (l *ZapLogger) Debugj(j log.JSON) {
	var args []interface{}

	for k, v := range j {
		args = append(args, k, v)
	}

	l.logger.Debugw("json", args...)
}

// Info implements echo's Logger interface.
func (l *ZapLogger) Info(i ...interface{}) {
	l.logger.Info(i...)
}

// Infof implements echo's Logger interface.
func (l *ZapLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Infoj implements echo's Logger interface.
func (l *ZapLogger) Infoj(j log.JSON) {
	var args []interface{}

	for k, v := range j {
		args = append(args, k, v)
	}

	l.logger.Infow("json", args...)
}

// Warn implements echo's Logger interface.
func (l *ZapLogger) Warn(i ...interface{}) {
	l.logger.Warn(i...)
}

// Warnf implements echo's Logger interface.
func (l *ZapLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Warnj implements echo's Logger interface.
func (l *ZapLogger) Warnj(j log.JSON) {
	var args []interface{}

	for k, v := range j {
		args = append(args, k, v)
	}

	l.logger.Warnw("json", args...)
}

// Error implements echo's Logger interface.
func (l *ZapLogger) Error(i ...interface{}) {
	l.logger.Error(i...)
}

// Errorf implements echo's Logger interface.
func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Errorj implements echo's Logger interface.
func (l *ZapLogger) Errorj(j log.JSON) {
	var args []interface{}

	for k, v := range j {
		args = append(args, k, v)
	}

	l.logger.Errorw("json", args...)
}

// Fatal implements echo's Logger interface.
func (l *ZapLogger) Fatal(i ...interface{}) {
	l.logger.Fatal(i...)
}

// Fatalj implements echo's Logger interface.
func (l *ZapLogger) Fatalj(j log.JSON) {
	var args []interface{}

	for k, v := range j {
		args = append(args, k, v)
	}

	l.logger.Fatalw("json", args...)
}

// Fatalf implements echo's Logger interface.
func (l *ZapLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// Panic implements echo's Logger interface.
func (l *ZapLogger) Panic(i ...interface{}) {
	l.logger.Panic(i...)
}

// Panicj implements echo's Logger interface.
func (l *ZapLogger) Panicj(j log.JSON) {
	var args []interface{}

	for k, v := range j {
		args = append(args, k, v)
	}

	l.logger.Panicw("json", args...)
}

// Panicf implements echo's Logger interface.
func (l *ZapLogger) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args...)
}

// Middleware returns a new echo middleware with the provided logger.
func (l *ZapLogger) Middleware(options ...MiddlewareOption) echo.MiddlewareFunc {
	return Middleware(l.logger.Desugar(), options...)
}

// NewLogger creatse a new ZapLogger which implements echo.Logger.
func NewLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: logger.Sugar(),
	}
}
