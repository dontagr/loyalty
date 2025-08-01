package handler

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"

	"github.com/dontagr/loyalty/internal/service/models"
	models2 "github.com/dontagr/loyalty/internal/store/models"
)

func (h *Handler) getBalance(c echo.Context) error {
	jwtUser := GetUserFromJWT(c)
	var waitGroup sync.WaitGroup

	var user *models2.User
	var userErr error
	waitGroup.Add(1)
	go func() {
		user, userErr = h.uService.GetUser(jwtUser.Login)
		waitGroup.Done()
	}()

	var withdrawal float64
	var withdrawalErr error
	waitGroup.Add(1)
	go func() {
		withdrawal, withdrawalErr = h.wService.GetTotalWithdrawal(jwtUser.ID)
		waitGroup.Done()
	}()

	waitGroup.Wait()
	if withdrawalErr != nil {
		h.log.Errorf("get total withdrawal failed: %v", withdrawalErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if userErr != nil {
		h.log.Errorf("get user failed: %v", userErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}

	return c.JSON(http.StatusOK, &models.ResponceWithdraw{
		Balance:    user.Balance,
		Withdrawal: withdrawal,
	})
}

func (h *Handler) postBalanceWithdraw(c echo.Context) error {
	requestWithdraw := &models.RequestWithdraw{}
	if err := c.Bind(requestWithdraw); err != nil {
		h.log.Errorf("request failed: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}
	//
	//jwtUser := GetUserFromJWT(c)
	//user, err := h.uService.GetUser(jwtUser.Login)
	//if err != nil {
	//	return err
	//}

	h.log.Infof("Withdraw %v", requestWithdraw)
	// 200 => Запрос на снятие успешно обработан
	// 401 => Пользователь не авторизован
	// 402 => На счету недостаточно средств
	// 422 => Неверный номер заказа
	// 500 => Внутренняя ошибка сервера

	return c.String(http.StatusNotImplemented, "Temporary handler stub.")
}

func (h *Handler) getWithdraw(c echo.Context) error {
	// 500 => Внутренняя ошибка сервера
	// 200 => Успешная обработка запроса
	// 204 => Нет ни одного списания
	// 401 => Пользователь не авторизован

	return c.String(http.StatusNotImplemented, "Temporary handler stub.")
}
