package order

import (
	"fmt"
	"net/http"

	intError "github.com/dontagr/loyalty/internal/service/error"
	"github.com/dontagr/loyalty/internal/store"
	"github.com/dontagr/loyalty/internal/store/models"
)

type OrderService struct {
	// TODO use interface
	pg *store.PG
}

func NewOrderService(pg *store.PG) *OrderService {
	return &OrderService{pg: pg}
}

func (o *OrderService) Lock() {
	o.pg.LockOrder()
}

func (o *OrderService) Unlock() {
	o.pg.UnlockOrder()
}

func (o *OrderService) CreateOrder(orderId int64, user *models.User) (bool, *intError.CustomError) {
	order, err := o.pg.GetOrder(orderId)
	if err != nil {
		return false, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed get order: %v", err))
	}
	if order.UserId != 0 && order.UserId != user.Id {
		return false, intError.NewCustomError(http.StatusConflict, "Номер заказа уже был загружен другим пользователем", nil)
	}
	if order.UserId != 0 && order.UserId == user.Id {
		return false, nil
	}

	err = o.pg.SaveOrder(orderId, user.Id)
	if err != nil {
		return false, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed save order: %v", err))
	}

	return true, nil
}

func (o *OrderService) GetListByUser(user *models.User) ([]*models.Order, *intError.CustomError) {
	list, err := o.pg.GetListByUserId(user.Id)
	if err != nil {
		return nil, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed get list order: %v", err))
	}

	return list, nil
}
