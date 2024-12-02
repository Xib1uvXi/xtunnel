package xlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

func InitLogger(lvl zapcore.Level) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableCaller = true
	// set log output
	cfg.OutputPaths = []string{"stdout"}
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	logger, _ := cfg.Build()

	return logger
}

func InitJsonLogger(lvl zapcore.Level) *zap.Logger {
	eConfig := zap.NewProductionEncoderConfig()
	eConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	eConfig.EncodeTime = qyTimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(eConfig), os.Stdout, zap.NewAtomicLevelAt(lvl))

	return zap.New(core)
}

type normalLevelEnable struct {
	flagLevel zapcore.Level
}

func (c normalLevelEnable) Enabled(lvl zapcore.Level) bool {
	return lvl >= c.flagLevel && lvl < zap.ErrorLevel
}

func newLogOptions() []zap.Option {
	return []zap.Option{
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
}

func getOutPath(targetDir, logFileName string) (out string) {
	if logFileName == "" {
		return
	}

	if !pathExists(targetDir) {
		if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
			panic(err.Error() + "; dir=" + targetDir)
		}
	}

	return filepath.Join(targetDir, logFileName)
}

func getErrOutPath(targetDir, logFileName string) (out string) {
	if logFileName == "" {
		return
	}
	return getOutPath(targetDir, "err_"+logFileName)
}

// Determine if the path file exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func qyTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}
