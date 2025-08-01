package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

const (
	searchTotalWithdrawalSQL = `SELECT SUM(withdrawal) FROM public.withdrawal WHERE user_id=$1`
)

func (pg *PG) GetTotalWithdrawal(userID int) (float64, error) {
	var withdrawal *float64
	err := pg.dbpool.QueryRow(context.Background(), searchTotalWithdrawalSQL, userID).Scan(&withdrawal)
	if errors.Is(err, pgx.ErrNoRows) || withdrawal == nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return *withdrawal, nil
}
