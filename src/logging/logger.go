package logging

import (
	"os"
	"time"

	"orla-alert/solte.lab/src/config"
	"orla-alert/solte.lab/src/logging/gelf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config *config.Logging) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(config.LogLevel))
	if err != nil {
		return nil, err
	}

	cw := zapcore.Lock(os.Stdout)
	je := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "init_timestamp",
		LevelKey:       "log_level",
		NameKey:        "log_name",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	zapCore := zapcore.NewCore(je, cw, level)

	zapCore = zapcore.NewSamplerWithOptions(zapCore, time.Second, 100, 100)

	logger := zap.New(
		zapCore,
		zap.AddCaller(), zap.AddStacktrace(zapcore.PanicLevel),
	)

	return logger, nil
}

func NewGelfLogger(config *config.Logging) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(config.LogLevel))
	if err != nil {
		return nil, err
	}

	zapCore, err := gelf.NewCore(
		gelf.Addr(config.RemoteAddr),
		gelf.Port(config.RemotePort),
		gelf.Protocol(config.RemoteProtocol),
	)

	if err != nil {
		panic(err)
	}

	var logger = zap.New(
		zapCore,
		zap.AddCaller(),
		zap.AddStacktrace(zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return zapCore.Enabled(level)
		})),
	)

	return logger, nil
}
