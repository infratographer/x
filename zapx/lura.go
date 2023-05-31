// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

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
