package test

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// MakeTestLogger creates a logger than uses zaptest.
func MakeTestLogger(t *testing.T) *zap.SugaredLogger {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	return logger.Sugar()
}
