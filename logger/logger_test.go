package logger

import (
	"go.uber.org/zap"
	"testing"
)

func Test_Logger(t *testing.T) {
	InitLogger(zap.InfoLevel.String())
	logger := GetSugaredLogger()
	logger.Debug("debug log..")
	logger.Info("info log..")
	logger.Error("error log..")
}
