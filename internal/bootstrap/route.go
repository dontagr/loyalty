package bootstrap

import (
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/service/handler"
)

var Route = fx.Options(
	fx.Provide(handler.NewHandler),
	fx.Invoke(func(handler *handler.Handler) {}),
)
