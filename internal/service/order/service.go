package order

import (
	"fmt"

	"github.com/dontagr/loyalty/internal/service/customerror"
	"github.com/dontagr/loyalty/internal/service/interfaces"
	"github.com/dontagr/loyalty/internal/store/models"
)

type Service struct {
	store interfaces.OrderStore
}

func NewOrderService(store interfaces.OrderStore) *Service {
	return &Service{store: store}
}

func (o *Service) CreateOrder(orderID string, user *models.User) (bool, *customerror.CustomError) {
	order, err := o.store.GetOrder(orderID)
	if err != nil {
		return false, customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", fmt.Errorf("failed get order: %v", err))
	}
	if order.UserID != 0 && order.UserID != user.ID {
		return false, customerror.NewCustomError(customerror.Conflict, "Номер заказа уже был загружен другим пользователем", nil)
	}
	if order.UserID != 0 && order.UserID == user.ID {
		return false, nil
	}

	err = o.store.SaveOrder(orderID, user.ID)
	if err != nil {
		return false, customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", fmt.Errorf("failed save order: %v", err))
	}

	return true, nil
}

func (o *Service) GetListByUser(user *models.User) ([]*models.Order, *customerror.CustomError) {
	list, err := o.store.GetListByUserID(user.ID)
	if err != nil {
		return nil, customerror.NewCustomError(customerror.Internal, "Внутренняя ошибка сервера", fmt.Errorf("failed get list order: %v", err))
	}

	return list, nil
}
