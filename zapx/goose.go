package zapx

import (
	"strings"

	"go.uber.org/zap"
)

type GooseLogger struct {
	logger *zap.SugaredLogger
}

func NewGooseLogger(logger *zap.SugaredLogger) *GooseLogger {
	return &GooseLogger{logger: logger}
}

func (l *GooseLogger) Fatal(v ...interface{})                 { l.logger.Fatal(v...) }
func (l *GooseLogger) Fatalf(format string, v ...interface{}) { l.logger.Fatalf(l.clean(format), v...) }
func (l *GooseLogger) Print(v ...interface{})                 { l.logger.Info(v...) }
func (l *GooseLogger) Println(v ...interface{})               { l.logger.Infoln(v...) }
func (l *GooseLogger) Printf(format string, v ...interface{}) { l.logger.Infof(l.clean(format), v...) }

func (l *GooseLogger) clean(str string) string {
	str = strings.TrimPrefix(str, "goose: ")
	return strings.TrimSuffix(str, "\n")
}
