package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dontagr/loyalty/internal/store/models"
)

const (
	searchOrderSQL = `SELECT id, user_id, status, accrual, create_dt FROM public.order WHERE id=$1`
	insertOrderSQL = `INSERT INTO public.order (id, user_id) VALUES ($1, $2);`
	listOrderSQL   = `SELECT id, user_id, status, accrual, create_dt  FROM public.order WHERE user_id = $1 ORDER BY create_dt DESC`
)

func (pg *PG) LockOrder() {
	pg.orderMX.Lock()
}

func (pg *PG) UnlockOrder() {
	pg.orderMX.Unlock()
}

func (pg *PG) GetOrder(orderID int64) (*models.Order, error) {
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

func (pg *PG) SaveOrder(orderID int64, userID int) error {
	_, err := pg.dbpool.Exec(context.Background(), insertOrderSQL, orderID, userID)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	return nil
}

func (pg *PG) GetListByUserId(userID int) ([]*models.Order, error) {
	rows, err := pg.dbpool.Query(context.Background(), listOrderSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении метрик: %w", err)
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
