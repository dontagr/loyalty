package withdrawal

import (
	"fmt"
	"net/http"

	intError "github.com/dontagr/loyalty/internal/service/error"
	"github.com/dontagr/loyalty/internal/service/models"
	"github.com/dontagr/loyalty/internal/store"
	models2 "github.com/dontagr/loyalty/internal/store/models"
)

type WithdrawalService struct {
	// TODO use interface
	pg *store.PG
}

func NewWithdrawalService(pg *store.PG) *WithdrawalService {
	return &WithdrawalService{pg: pg}
}

func (w *WithdrawalService) GetTotalWithdrawal(userID int) (float64, error) {
	return w.pg.GetTotalWithdrawal(userID)
}

func (w *WithdrawalService) SaveWithdraw(reqW *models.RequestWithdraw, user *models2.User) *intError.CustomError {
	if user.Balance-reqW.Sum < 0 {
		return intError.NewCustomError(http.StatusPaymentRequired, "На счету недостаточно средств", nil)
	}

	w.pg.LockWithdrawal()
	defer w.pg.UnlockWithdrawal()

	withdraw, err := w.pg.GetWithdraw(reqW.Order)
	if err != nil {
		return intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", err)
	}
	if withdraw.ID != "" {
		return intError.NewCustomError(http.StatusUnprocessableEntity, "Неверный номер заказа", nil)
	}

	err = w.pg.SaveWithdraw(models2.Withdrawal{ID: reqW.Order, Withdrawal: reqW.Sum, UserID: user.ID})
	if err != nil {
		return intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", err)
	}

	return nil
}

func (w *WithdrawalService) GetListByUser(user *models2.User) ([]*models2.Withdrawal, *intError.CustomError) {
	list, err := w.pg.GetWithdrawalListByUserID(user.ID)
	if err != nil {
		return nil, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed get list order: %v", err))
	}

	return list, nil
}
