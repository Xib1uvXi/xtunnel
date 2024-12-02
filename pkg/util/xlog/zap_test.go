package xlog

import (
	"go.uber.org/zap"
	"testing"
)

func TestInitLogger(t *testing.T) {
	log := InitLogger(zap.InfoLevel)

	log.Debug("debug")
	log.Error("error")

	log2 := InitJsonLogger(zap.InfoLevel)
	log2.Error("debug")
}
