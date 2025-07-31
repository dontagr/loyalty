package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/config"
	"github.com/dontagr/loyalty/internal/faultTolerance/pgretry"
)

var Postgres = fx.Options(
	fx.Provide(
		newPostgresConnect,
		pgretry.NewPgxRetry,
	),
)

func newPostgresConnect(cfg *config.Config, lc fx.Lifecycle) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), cfg.DataBase.DatabaseDsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			dbpool.Close()
			return nil
		},
	})

	return dbpool, nil
}
