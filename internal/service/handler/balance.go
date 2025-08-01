package handler

import (
	"net/http"
	"strings"
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
	if err := c.Validate(requestWithdraw); err != nil {
		h.log.Errorf("validation failed: %v", err)
		if strings.Contains(err.Error(), "algLuna") {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Неверный номер заказа")
		}

		return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}

	jwtUser := GetUserFromJWT(c)
	h.uService.Lock()
	defer h.uService.Unlock()
	user, err := h.uService.GetUser(jwtUser.Login)
	if err != nil {
		h.log.Errorf("get user failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}

	intErr := h.wService.SaveWithdraw(requestWithdraw, user)
	if intErr != nil {
		if intErr.Err != nil {
			h.log.Infof(intErr.Error())
		}

		return echo.NewHTTPError(intErr.Code, intErr.Message)
	}

	return c.JSON(http.StatusOK, "Запрос на снятие успешно обработан")
}

func (h *Handler) getWithdraw(c echo.Context) error {
	list, intErr := h.wService.GetListByUser(GetUserFromJWT(c))
	if intErr != nil {
		if intErr.Err != nil {
			h.log.Infof(intErr.Error())
		}

		return echo.NewHTTPError(intErr.Code, intErr.Message)
	}

	if len(list) == 0 {
		return c.NoContent(http.StatusNoContent) // Нет данных для ответа
	}

	return c.JSON(http.StatusOK, list)
}
