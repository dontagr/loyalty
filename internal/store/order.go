package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dontagr/loyalty/internal/store/models"
)

const (
	searchOrderSQL                 = `SELECT id, user_id, status, accrual, create_dt FROM public.order WHERE id=$1`
	insertOrderSQL                 = `INSERT INTO public.order (id, user_id) VALUES ($1, $2);`
	updateOrderStatusSQL           = `UPDATE public.order SET status=$1 WHERE id=$2;`
	updateOrderStatusAndAccrualSQL = `UPDATE public.order SET status=$1, accrual=$2 WHERE id=$3;`
	listOrderSQL                   = `SELECT id, user_id, status, accrual, create_dt  FROM public.order WHERE user_id = $1 ORDER BY create_dt DESC`
	listOrderForProcessingSQL      = `SELECT id  FROM public.order WHERE status IN (0, 1)`
)

func (pg *PG) LockOrder() {
	pg.orderMX.Lock()
}

func (pg *PG) UnlockOrder() {
	pg.orderMX.Unlock()
}

func (pg *PG) GetOrder(orderID string) (*models.Order, error) {
	var order models.Order
	err := pg.dbpool.QueryRow(context.Background(), searchOrderSQL, orderID).Scan(
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

func (pg *PG) SaveOrder(orderID string, userID int) error {
	_, err := pg.dbpool.Exec(context.Background(), insertOrderSQL, orderID, userID)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении заказа: %w", err)
	}

	return nil
}

func (pg *PG) GetListByUserID(userID int) ([]*models.Order, error) {
	rows, err := pg.dbpool.Query(context.Background(), listOrderSQL, userID)
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

func (pg *PG) GetListForProcessing() ([]*models.Order, error) {
	rows, err := pg.dbpool.Query(context.Background(), listOrderForProcessingSQL)
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

func (pg *PG) UpdateOrder(order *models.Order) error {
	oldOrder, err := pg.GetOrder(order.ID)
	if err != nil {
		return err
	}

	pg.LockOrder()
	defer pg.UnlockOrder()
	if order.Status == models.StatusProcessing && oldOrder.Status != models.StatusInvalid && oldOrder.Status != models.StatusProcessed {
		_, err := pg.dbpool.Exec(context.Background(), updateOrderStatusSQL, order.Status, order.ID)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении заказа: %w", err)
		}

		return nil
	}

	if order.Status == models.StatusInvalid {
		_, err := pg.dbpool.Exec(context.Background(), updateOrderStatusSQL, order.Status, order.ID)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении заказа: %w", err)
		}
	}

	pg.LockUser()
	defer pg.UnlockUser()
	if order.Status == models.StatusProcessed && *order.Accrual > 0 {
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

		_, err := pg.dbpool.Exec(context.Background(), updateOrderStatusAndAccrualSQL, order.Status, order.Accrual, order.ID)
		if err != nil {
			txErr = err
			return fmt.Errorf("ошибка при обновлении заказа: %w", err)
		}
		_, err = pg.dbpool.Exec(context.Background(), updateUserBalanceSQL, order.Accrual, oldOrder.UserID)
		if err != nil {
			txErr = err
			return fmt.Errorf("ошибка при обновлении пользователя: %w", err)
		}
	}

	return fmt.Errorf("update order has failed order %v", order)
}
