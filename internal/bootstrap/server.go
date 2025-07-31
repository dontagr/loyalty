package bootstrap

import (
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/httpserver"
)

var Server = fx.Options(
	fx.Provide(httpserver.NewServer),
	fx.Invoke(func(*httpserver.HTTPServer) {}),
)
