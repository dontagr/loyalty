package withdrawal

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/faultTolerance/pgretry"
	"github.com/dontagr/loyalty/internal/store/models"
)

const (
	searchTotalWithdrawalSQL = `SELECT SUM(withdrawal) FROM public.withdrawal WHERE user_id=$1`
	insertWithdrawalSQL      = `INSERT INTO public.withdrawal (id, user_id, withdrawal) VALUES ($1, $2, $3);`
	searchWithdrawalSQL      = `SELECT id, user_id, withdrawal, create_dt FROM public.withdrawal WHERE id=$1`
	listWithdrawalSQL        = `SELECT id, user_id, withdrawal, create_dt  FROM public.withdrawal WHERE user_id = $1 ORDER BY create_dt DESC`
	decreaseUserBalanceSQL   = `UPDATE public.user SET balance=balance-$1 WHERE ID=$2`
	createWithdrawalTable    = `
CREATE TABLE IF NOT EXISTS public."withdrawal" (
	id bigint NOT NULL,
	user_id bigint NOT NULL,
	withdrawal double precision DEFAULT NUll,
	create_dt timestamptz DEFAULT NOW() NOT NULL,
	CONSTRAINT withdrawal_pk PRIMARY KEY (id),
	CONSTRAINT withdrawal_id_idx UNIQUE (user_id,id)
);
`
)

type Withdrawal struct {
	mx     sync.RWMutex
	dbpool *pgretry.PgxRetry
	log    *zap.SugaredLogger
}

func NewWithdrawal(log *zap.SugaredLogger, dbpool *pgretry.PgxRetry, lc fx.Lifecycle) *Withdrawal {
	withdrawal := Withdrawal{
		dbpool: dbpool,
		log:    log,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return withdrawal.addShema(ctx)
		},
	})

	return &withdrawal
}

func (w *Withdrawal) addShema(ctx context.Context) error {
	_, err := w.dbpool.Exec(ctx, createWithdrawalTable)

	return err
}

func (w *Withdrawal) Lock() {
	w.mx.Lock()
}

func (w *Withdrawal) Unlock() {
	w.mx.Unlock()
}

func (w *Withdrawal) GetTotalWithdrawal(userID int) (float64, error) {
	var withdrawal *float64
	err := w.dbpool.QueryRow(context.Background(), searchTotalWithdrawalSQL, userID).Scan(&withdrawal)
	if errors.Is(err, pgx.ErrNoRows) || withdrawal == nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return *withdrawal, nil
}

func (w *Withdrawal) GetWithdraw(orderID string) (*models.Withdrawal, error) {
	var withdrawal models.Withdrawal
	err := w.dbpool.QueryRow(context.Background(), searchWithdrawalSQL, orderID).Scan(
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

func (w *Withdrawal) SaveWithdraw(withdrawal models.Withdrawal) error {
	tx, txErr := w.dbpool.Begin(context.Background())
	if txErr != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", txErr)
	}
	defer func(txErr *error) {
		if *txErr != nil {
			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
				w.log.Errorf("ошибка отката транзакции: %v", rollbackErr)
			}
		} else {
			if commitErr := tx.Commit(context.Background()); commitErr != nil {
				w.log.Errorf("ошибка при коммите транзакции: %v", commitErr)
			}
		}
	}(&txErr)

	_, err := w.dbpool.Exec(context.Background(), insertWithdrawalSQL, withdrawal.ID, withdrawal.UserID, withdrawal.Withdrawal)
	if err != nil {
		txErr = err
		return fmt.Errorf("ошибка при создания списания: %w", err)
	}
	_, err = w.dbpool.Exec(context.Background(), decreaseUserBalanceSQL, withdrawal.Withdrawal, withdrawal.UserID)
	if err != nil {
		txErr = err
		return fmt.Errorf("ошибка при обновлении пользователя: %w", err)
	}

	return nil
}

func (w *Withdrawal) GetWithdrawalListByUserID(userID int) ([]*models.Withdrawal, error) {
	rows, err := w.dbpool.Query(context.Background(), listWithdrawalSQL, userID)
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
