package order

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
	searchOrderSQL                 = `SELECT id, user_id, status, accrual, create_dt FROM public.order WHERE id=$1`
	insertOrderSQL                 = `INSERT INTO public.order (id, user_id) VALUES ($1, $2);`
	updateOrderStatusSQL           = `UPDATE public.order SET status=$1 WHERE id=$2;`
	updateOrderStatusAndAccrualSQL = `UPDATE public.order SET status=$1, accrual=$2 WHERE id=$3;`
	listOrderSQL                   = `SELECT id, user_id, status, accrual, create_dt  FROM public.order WHERE user_id = $1 ORDER BY create_dt DESC`
	listOrderForProcessingSQL      = `SELECT id  FROM public.order WHERE status IN (0, 1)`
	increaseUserBalanceSQL         = `UPDATE public.user SET balance=balance+$1 WHERE ID=$2`
	createOrderTable               = `
CREATE TABLE IF NOT EXISTS public."order" (
	id bigint NOT NULL,
	user_id bigint NOT NULL,
	accrual double precision DEFAULT NUll,
	status int2 DEFAULT 0 NOT NULL,
	create_dt timestamptz DEFAULT NOW() NOT NULL,
	CONSTRAINT order_pk PRIMARY KEY (id),
	CONSTRAINT order_id_idx UNIQUE (user_id,id)
);
`
)

type Order struct {
	mx     sync.RWMutex
	dbpool *pgretry.PgxRetry
	log    *zap.SugaredLogger
}

func NewOrder(log *zap.SugaredLogger, dbpool *pgretry.PgxRetry, lc fx.Lifecycle) *Order {
	order := Order{
		dbpool: dbpool,
		log:    log,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return order.addShema(ctx)
		},
	})

	return &order
}

func (o *Order) addShema(ctx context.Context) error {
	_, err := o.dbpool.Exec(ctx, createOrderTable)

	return err
}

func (o *Order) Lock() {
	o.mx.Lock()
}

func (o *Order) Unlock() {
	o.mx.Unlock()
}

func (o *Order) GetOrder(orderID string) (*models.Order, error) {
	var order models.Order
	err := o.dbpool.QueryRow(context.Background(), searchOrderSQL, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.CreateDateTime,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return &models.Order{}, nil
	}
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *Order) SaveOrder(orderID string, userID int) error {
	_, err := o.dbpool.Exec(context.Background(), insertOrderSQL, orderID, userID)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении заказа: %w", err)
	}

	return nil
}

func (o *Order) GetListByUserID(userID int) ([]*models.Order, error) {
	rows, err := o.dbpool.Query(context.Background(), listOrderSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении заказов: %w", err)
	}
	defer rows.Close()

	var result []*models.Order
	for rows.Next() {
		order := new(models.Order)
		err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.CreateDateTime)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании заказа: %w", err)
		}

		result = append(result, order)
	}

	return result, nil
}

func (o *Order) GetListForProcessing() ([]*models.Order, error) {
	rows, err := o.dbpool.Query(context.Background(), listOrderForProcessingSQL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении заказов: %w", err)
	}
	defer rows.Close()

	var result []*models.Order
	for rows.Next() {
		order := new(models.Order)
		err := rows.Scan(&order.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании заказа: %w", err)
		}

		result = append(result, order)
	}

	return result, nil
}

func (o *Order) UpdateOrder(order *models.Order) error {
	oldOrder, err := o.GetOrder(order.ID)
	if err != nil {
		return err
	}

	o.Lock()
	defer o.Unlock()
	if order.Status == models.StatusProcessing && oldOrder.Status != models.StatusInvalid && oldOrder.Status != models.StatusProcessed {
		_, err := o.dbpool.Exec(context.Background(), updateOrderStatusSQL, order.Status, order.ID)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении заказа: %w", err)
		}

		return nil
	}

	if order.Status == models.StatusInvalid {
		_, err := o.dbpool.Exec(context.Background(), updateOrderStatusSQL, order.Status, order.ID)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении заказа: %w", err)
		}

		return nil
	}

	if order.Status == models.StatusProcessed && *order.Accrual > 0 {
		tx, txErr := o.dbpool.Begin(context.Background())
		if txErr != nil {
			return fmt.Errorf("ошибка начала транзакции: %w", txErr)
		}
		defer func(txErr *error) {
			if *txErr != nil {
				if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
					o.log.Errorf("ошибка отката транзакции: %v", rollbackErr)
				}
			} else {
				if commitErr := tx.Commit(context.Background()); commitErr != nil {
					o.log.Errorf("ошибка при коммите транзакции: %v", commitErr)
				}
			}
		}(&txErr)

		_, err := o.dbpool.Exec(context.Background(), updateOrderStatusAndAccrualSQL, order.Status, order.Accrual, order.ID)
		if err != nil {
			txErr = err
			return fmt.Errorf("ошибка при обновлении заказа: %w", err)
		}
		_, err = o.dbpool.Exec(context.Background(), increaseUserBalanceSQL, order.Accrual, oldOrder.UserID)
		if err != nil {
			txErr = err
			return fmt.Errorf("ошибка при обновлении пользователя: %w", err)
		}

		return nil
	}

	return fmt.Errorf("update order has failed order %v", order)
}
