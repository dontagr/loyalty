package order

import (
	"fmt"
	"net/http"

	"github.com/dontagr/loyalty/internal/service/customerror"
	"github.com/dontagr/loyalty/internal/service/interfaces"
	"github.com/dontagr/loyalty/internal/store/models"
)

type OrderService struct {
	store interfaces.OrderStore
}

func NewOrderService(store interfaces.OrderStore) *OrderService {
	return &OrderService{store: store}
}

func (o *OrderService) Lock() {
	o.store.Lock()
}

func (o *OrderService) Unlock() {
	o.store.Unlock()
}

func (o *OrderService) CreateOrder(orderID string, user *models.User) (bool, *customerror.CustomError) {
	order, err := o.store.GetOrder(orderID)
	if err != nil {
		return false, customerror.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed get order: %v", err))
	}
	if order.UserID != 0 && order.UserID != user.ID {
		return false, customerror.NewCustomError(http.StatusConflict, "Номер заказа уже был загружен другим пользователем", nil)
	}
	if order.UserID != 0 && order.UserID == user.ID {
		return false, nil
	}

	err = o.store.SaveOrder(orderID, user.ID)
	if err != nil {
		return false, customerror.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed save order: %v", err))
	}

	return true, nil
}

func (o *OrderService) GetListByUser(user *models.User) ([]*models.Order, *customerror.CustomError) {
	list, err := o.store.GetListByUserID(user.ID)
	if err != nil {
		return nil, customerror.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed get list order: %v", err))
	}

	return list, nil
}
