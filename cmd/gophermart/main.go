package main

import (
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/bootstrap"
)

func main() {
	fx.New(
		bootstrap.Server,
		bootstrap.Config,
		bootstrap.Logger,
		bootstrap.Postgres,
		bootstrap.Store,
		bootstrap.Route,
		bootstrap.Service,
		bootstrap.Worker,
	).Run()
}
