package log

import "go.uber.org/zap"

func NewLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}
