package bootstrap

import (
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/store"
)

var Store = fx.Options(
	fx.Provide(
		store.RegisterStorePG,
	),
	fx.Invoke(func(*store.PG) {}),
)
