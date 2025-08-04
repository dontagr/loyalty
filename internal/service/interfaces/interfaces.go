package interfaces

import "github.com/dontagr/loyalty/internal/store/models"

type (
	UserStore interface {
		Lock()
		Unlock()
		GetUser(login string) (*models.User, error)
		SaveUser(login string, passwordHash string) error
	}
	OrderStore interface {
		Lock()
		Unlock()
		GetOrder(orderID string) (*models.Order, error)
		SaveOrder(orderID string, userID int) error
		GetListByUserID(userID int) ([]*models.Order, error)
		GetListForProcessing() ([]*models.Order, error)
		UpdateOrder(order *models.Order) error
		BlockOrder(orderID string) bool
		UnblockOrder(orderID string) bool
	}
	WithdrawalStore interface {
		Lock()
		Unlock()
		GetTotalWithdrawal(userID int) (float64, error)
		GetWithdraw(orderID string) (*models.Withdrawal, error)
		SaveWithdraw(withdrawal models.Withdrawal) error
		GetWithdrawalListByUserID(userID int) ([]*models.Withdrawal, error)
	}
)
