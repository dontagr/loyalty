package pgretry

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

const (
	DurationFirstIteration  = 1
	DurationSecondIteration = 3
	DurationThirdIteration  = 5
)

type PgxRetry struct {
	dbpool   *pgxpool.Pool
	duration []int
	log      *zap.SugaredLogger
}

func NewPgxRetry(conn *pgxpool.Pool, log *zap.SugaredLogger) *PgxRetry {
	if conn == nil {
		return nil
	}

	return &PgxRetry{
		dbpool:   conn,
		duration: []int{DurationFirstIteration, DurationSecondIteration, DurationThirdIteration},
		log:      log,
	}
}

func (pgr *PgxRetry) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	iter := atomic.Int32{}
	operation := func() (pgconn.CommandTag, error) {
		defer iter.Add(1)
		tag, err := pgr.dbpool.Exec(ctx, sql, arguments...)

		if err != nil {
			end := time.Now()
			duration := end.Sub(start)
			var connectErr *pgconn.ConnectError
			if errors.As(err, &connectErr) {
				log.Debugf("ошибка подключения к базе; Пробуем еще раз, прошло времени: %v сек, итерация %v", duration.Seconds(), iter.Load()+1)

				return tag, backoff.RetryAfter(pgr.duration[iter.Load()])
			}

			log.Debugf("ошибка фатальна; Прошло времени: %v сек, итерация %v", duration.Seconds(), iter.Load()+1)
			return tag, backoff.Permanent(err)
		}

		return tag, nil
	}

	opt := backoff.ExponentialBackOff{
		InitialInterval:     time.Duration(pgr.duration[iter.Load()]) * time.Second,
		RandomizationFactor: 0,
		Multiplier:          1,
		MaxInterval:         time.Duration(10) * time.Second,
	}

	return backoff.Retry(ctx, operation, backoff.WithBackOff(&opt), backoff.WithMaxTries(3))
}

func (pgr *PgxRetry) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	row := pgr.dbpool.QueryRow(ctx, sql, args...)

	return row
}

func (pgr *PgxRetry) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := pgr.dbpool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении SQL: %w", err)
	}

	return rows, nil
}

func (pgr *PgxRetry) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := pgr.dbpool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}

	return tx, nil
}

func (pgr *PgxRetry) Ping(ctx context.Context) error {
	err := pgr.dbpool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ошибка пинга: %w", err)
	}

	return nil
}
