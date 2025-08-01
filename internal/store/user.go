package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dontagr/loyalty/internal/store/models"
)

const (
	searchUserSQL        = `SELECT id, login, password, balance FROM public.user WHERE login=$1`
	insertUserSQL        = `INSERT INTO public.user (login, password) VALUES ($1, $2);`
	updateUserBalanceSQL = `UPDATE public.user SET balance=$1 WHERE ID=$2`
)

func (pg *PG) LockUser() {
	pg.userMX.Lock()
}

func (pg *PG) UnlockUser() {
	pg.userMX.Unlock()
}

func (pg *PG) GetUser(login string) (*models.User, error) {
	var user models.User
	err := pg.dbpool.QueryRow(context.Background(), searchUserSQL, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return &models.User{}, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (pg *PG) SaveUser(login string, passwordHash string) error {
	_, err := pg.dbpool.Exec(context.Background(), insertUserSQL, login, passwordHash)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	return nil
}
