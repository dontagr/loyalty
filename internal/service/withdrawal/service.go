package withdrawal

import (
	"fmt"
	"net/http"

	"github.com/dontagr/loyalty/internal/service/customerror"
	"github.com/dontagr/loyalty/internal/service/interfaces"
	"github.com/dontagr/loyalty/internal/service/models"
	storeModel "github.com/dontagr/loyalty/internal/store/models"
)

type WithdrawalService struct {
	store interfaces.WithdrawalStore
}

func NewWithdrawalService(store interfaces.WithdrawalStore) *WithdrawalService {
	return &WithdrawalService{store: store}
}

func (w *WithdrawalService) GetTotalWithdrawal(userID int) (float64, error) {
	return w.store.GetTotalWithdrawal(userID)
}

func (w *WithdrawalService) SaveWithdraw(reqW *models.RequestWithdraw, user *storeModel.User) *customerror.CustomError {
	if user.Balance-reqW.Sum < 0 {
		return customerror.NewCustomError(http.StatusPaymentRequired, "На счету недостаточно средств", nil)
	}

	w.store.Lock()
	defer w.store.Unlock()

	withdraw, err := w.store.GetWithdraw(reqW.Order)
	if err != nil {
		return customerror.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", err)
	}
	if withdraw.ID != "" {
		return customerror.NewCustomError(http.StatusUnprocessableEntity, "Неверный номер заказа", nil)
	}

	err = w.store.SaveWithdraw(storeModel.Withdrawal{ID: reqW.Order, Withdrawal: reqW.Sum, UserID: user.ID})
	if err != nil {
		return customerror.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", err)
	}

	return nil
}

func (w *WithdrawalService) GetListByUser(user *storeModel.User) ([]*storeModel.Withdrawal, *customerror.CustomError) {
	list, err := w.store.GetWithdrawalListByUserID(user.ID)
	if err != nil {
		return nil, customerror.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed get list order: %v", err))
	}

	return list, nil
}
