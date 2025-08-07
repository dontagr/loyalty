package interfaces

import (
	"github.com/jackc/pgx/v5"

	"github.com/dontagr/loyalty/internal/store/models"
)

type (
	UserStore interface {
		GetUser(login string, params ...bool) (*models.User, error)
		SaveUser(login string, passwordHash string) error
	}
	OrderStore interface {
		GetOrder(orderID string) (*models.Order, error)
		SaveOrder(orderID string, userID int) error
		GetListByUserID(userID int) ([]*models.Order, error)
		GetListForProcessing() ([]*models.Order, error)
		UpdateOrder(order *models.Order) error
		BlockOrder(orderID string) bool
		UnblockOrder(orderID string) bool
	}
	WithdrawalStore interface {
		BeginTX() (pgx.Tx, error)
		RollbackTX(tx pgx.Tx)
		CommitTX(tx pgx.Tx)
		GetTotalWithdrawal(userID int) (float64, error)
		GetWithdraw(orderID string) (*models.Withdrawal, error)
		SaveWithdraw(withdrawal models.Withdrawal) error
		GetWithdrawalListByUserID(userID int) ([]*models.Withdrawal, error)
	}
)
