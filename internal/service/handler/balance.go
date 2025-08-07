package handler

import (
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"

	"github.com/dontagr/loyalty/internal/service/models"
	storeModel "github.com/dontagr/loyalty/internal/store/models"
)

func (h *Handler) GetBalance(c echo.Context) error {
	jwtUser := h.jwt.GetUser(c)
	var waitGroup sync.WaitGroup
	var user *storeModel.User
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
		Balance:    float64(user.Balance) / 100,
		Withdrawal: withdrawal,
	})
}

func (h *Handler) PostBalanceWithdraw(c echo.Context) error {
	requestWithdraw := &models.RequestWithdraw{}
	if err := c.Bind(requestWithdraw); err != nil {
		h.log.Errorf("request failed: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}
	if err := c.Validate(requestWithdraw); err != nil {
		h.log.Errorf("validation failed: %v", err)
		if strings.Contains(err.Error(), "algLuna") {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Неверный номер заказа")
		}

		return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}

	jwtUser := h.jwt.GetUser(c)
	intErr := h.wService.SaveWithdraw(requestWithdraw, h.uService, jwtUser.Login)
	if intErr != nil {
		if intErr.Err != nil {
			h.log.Infof(intErr.Error())
		}

		return echo.NewHTTPError(h.convertCustomErrorToServerCode(intErr.Code), intErr.Message)
	}

	return c.JSON(http.StatusOK, "Запрос на снятие успешно обработан")
}

func (h *Handler) GetWithdraw(c echo.Context) error {
	list, intErr := h.wService.GetListByUser(h.jwt.GetUser(c))
	if intErr != nil {
		if intErr.Err != nil {
			h.log.Infof(intErr.Error())
		}

		return echo.NewHTTPError(h.convertCustomErrorToServerCode(intErr.Code), intErr.Message)
	}

	if len(list) == 0 {
		return c.NoContent(http.StatusNoContent) // Нет данных для ответа
	}

	return c.JSON(http.StatusOK, list)
}
