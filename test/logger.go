package test

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// MakeTestLogger creates a logger than uses zaptest.
func MakeTestLogger() logger.Logger {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	return logger.Sugar()
}
