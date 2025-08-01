package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dontagr/loyalty/internal/store/models"
)

const (
	searchTotalWithdrawalSQL = `SELECT SUM(withdrawal) FROM public.withdrawal WHERE user_id=$1`
	insertWithdrawalSQL      = `INSERT INTO public.withdrawal (id, user_id, withdrawal) VALUES ($1, $2, $3);`
	searchWithdrawalSQL      = `SELECT id, user_id, withdrawal, create_dt FROM public.withdrawal WHERE id=$1`
	listWithdrawalSQL        = `SELECT id, user_id, withdrawal, create_dt  FROM public.withdrawal WHERE user_id = $1 ORDER BY create_dt DESC`
)

func (pg *PG) LockWithdrawal() {
	pg.withdrawalMX.Lock()
}

func (pg *PG) UnlockWithdrawal() {
	pg.withdrawalMX.Unlock()
}

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

func (pg *PG) GetWithdraw(orderID string) (*models.Withdrawal, error) {
	var withdrawal models.Withdrawal
	err := pg.dbpool.QueryRow(context.Background(), searchWithdrawalSQL, orderID).Scan(
		&withdrawal.ID,
		&withdrawal.UserID,
		&withdrawal.Withdrawal,
		&withdrawal.CreateDateTime,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return &models.Withdrawal{}, nil
	}
	if err != nil {
		return nil, err
	}

	return &withdrawal, nil
}

func (pg *PG) SaveWithdraw(withdrawal models.Withdrawal) error {
	tx, txErr := pg.dbpool.Begin(context.Background())
	if txErr != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", txErr)
	}
	defer func(txErr *error) {
		if *txErr != nil {
			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
				pg.log.Errorf("ошибка отката транзакции: %v", rollbackErr)
			}
		} else {
			if commitErr := tx.Commit(context.Background()); commitErr != nil {
				pg.log.Errorf("ошибка при коммите транзакции: %v", commitErr)
			}
		}
	}(&txErr)

	_, err := pg.dbpool.Exec(context.Background(), insertWithdrawalSQL, withdrawal.ID, withdrawal.UserID, withdrawal.Withdrawal)
	if err != nil {
		txErr = err
		return fmt.Errorf("ошибка при создания списания: %w", err)
	}
	_, err = pg.dbpool.Exec(context.Background(), decreaseUserBalanceSQL, withdrawal.Withdrawal, withdrawal.UserID)
	if err != nil {
		txErr = err
		return fmt.Errorf("ошибка при обновлении пользователя: %w", err)
	}

	return nil
}

func (pg *PG) GetWithdrawalListByUserID(userID int) ([]*models.Withdrawal, error) {
	rows, err := pg.dbpool.Query(context.Background(), listWithdrawalSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении списаний: %w", err)
	}
	defer rows.Close()

	var result []*models.Withdrawal
	for rows.Next() {
		withdrawal := new(models.Withdrawal)
		err := rows.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.Withdrawal, &withdrawal.CreateDateTime)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании списания: %w", err)
		}

		result = append(result, withdrawal)
	}

	return result, nil
}
