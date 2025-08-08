package bootstrap

import (
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/worker"
)

var Worker = fx.Options(
	fx.Provide(
		worker.NewUpdater,
	),
	fx.Invoke(func(*worker.Updater) {}),
)
