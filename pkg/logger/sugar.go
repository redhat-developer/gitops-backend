package logger

import (
	"log"

	"go.uber.org/zap"
)

func MakeLogger() Logger {
	logger, _ := zap.NewProduction()
	defer func() {
		err := logger.Sync() // flushes buffer, if any
		if err != nil {
			log.Fatal(err)
		}
	}()
	return logger.Sugar()
}
