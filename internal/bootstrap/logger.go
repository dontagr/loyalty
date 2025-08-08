package bootstrap

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dontagr/loyalty/internal/config"
)

var Logger = fx.Options(
	fx.Provide(newLogger),
	fx.WithLogger(func(log *zap.SugaredLogger) fxevent.Logger {
		zap := fxevent.ZapLogger{Logger: log.Desugar()}
		zap.UseLogLevel(zapcore.DebugLevel)

		return &zap
	}),
)

func newLogger(lc fx.Lifecycle, cnf *config.Config) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()

	lvl, err := zap.ParseAtomicLevel(cnf.Log.LogLevel)
	if err != nil {
		return nil, err
	}

	cfg.EncoderConfig.TimeKey = "dt"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("01/02 03:04:05PM")
	cfg.Level = lvl
	logger, err := cfg.Build()

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			_ = logger.Sync()
			return nil
		},
	})

	return logger.Sugar(), err
}
