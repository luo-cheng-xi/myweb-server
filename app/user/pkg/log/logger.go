package log

import (
	"go.uber.org/zap"
	"log"
)

func NewLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	return logger
}
