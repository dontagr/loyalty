package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/dontagr/loyalty/internal/service/models"
)

func (h *Handler) getBalance(c echo.Context) error {
	fmt.Println(123)
	// 200 => Текущий баланс пользователя
	// 401 => Пользователь не авторизован
	// 500 => Внутренняя ошибка сервера

	return c.String(http.StatusNotImplemented, "Temporary handler stub.")
}

func (h *Handler) postBalanceWithdraw(c echo.Context) error {
	requestWithdraw := &models.RequestWithdraw{}
	if err := c.Bind(requestWithdraw); err != nil {
		h.log.Errorf("request failed: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}

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
