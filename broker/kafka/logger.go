package kafka

import "github.com/go-kratos/kratos/v2/log"

type Logger struct {
	logger *log.Helper
}

func (l *Logger) Printf(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

type ErrorLogger struct {
	logger *log.Helper
}

func (l *ErrorLogger) Printf(msg string, args ...interface{}) {
	l.logger.Errorf(msg, args...)
}
