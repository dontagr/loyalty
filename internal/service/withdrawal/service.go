package withdrawal

import (
	"fmt"

	"github.com/dontagr/loyalty/internal/service/customerror"
	"github.com/dontagr/loyalty/internal/service/interfaces"
	"github.com/dontagr/loyalty/internal/service/models"
	"github.com/dontagr/loyalty/internal/service/user"
	storeModel "github.com/dontagr/loyalty/internal/store/models"
)

type Service struct {
	store interfaces.WithdrawalStore
}

func NewWithdrawalService(store interfaces.WithdrawalStore) *Service {
	return &Service{store: store}
}

func (w *Service) GetTotalWithdrawal(userID int) (float64, error) {
	return w.store.GetTotalWithdrawal(userID)
}

func (w *Service) SaveWithdraw(reqW *models.RequestWithdraw, userService *user.Service, login string) *customerror.CustomError {
	tx, txErr := w.store.BeginTX()
	if txErr != nil {
		return customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", fmt.Errorf("failed get userDTO %v", txErr))
	}
	defer func(txErr *error) {
		if *txErr != nil {
			w.store.RollbackTX(tx)
		} else {
			w.store.CommitTX(tx)
		}
	}(&txErr)

	userDTO, err := userService.GetTxUser(tx, login)
	if err != nil {
		txErr = err
		return customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", fmt.Errorf("failed get userDTO %v", err))
	}

	sum := int(reqW.Sum * 100)
	if userDTO.Balance-sum < 0 {
		txErr = fmt.Errorf("на счету недостаточно средств")
		return customerror.NewCustomError(customerror.Payment, "На счету недостаточно средств", nil)
	}

	withdraw, err := w.store.GetWithdraw(reqW.Order)
	if err != nil {
		txErr = err
		return customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", err)
	}
	if withdraw.ID != "" {
		txErr = fmt.Errorf("неверный номер заказа")
		return customerror.NewCustomError(customerror.Unprocessable, "Неверный номер заказа", nil)
	}

	err = w.store.SaveWithdraw(tx, storeModel.Withdrawal{ID: reqW.Order, Withdrawal: sum, UserID: userDTO.ID})
	if err != nil {
		txErr = err
		return customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", err)
	}

	return nil
}

func (w *Service) GetListByUser(user *storeModel.User) ([]*storeModel.Withdrawal, *customerror.CustomError) {
	list, err := w.store.GetWithdrawalListByUserID(user.ID)
	if err != nil {
		return nil, customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", fmt.Errorf("failed get list order: %v", err))
	}

	return list, nil
}
