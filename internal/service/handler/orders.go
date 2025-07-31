package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/dontagr/loyalty/internal/service/models"
	models2 "github.com/dontagr/loyalty/internal/store/models"
)

func (h *Handler) createOrder(c echo.Context) error {
	requestOrder, echoErr := h.getOrderBody(c)
	if echoErr != nil {
		return echoErr
	}
	if err := c.Validate(requestOrder); err != nil {
		if strings.Contains(err.Error(), "algLuna") {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Неверный формат номера заказа")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}

	order, echoErr := h.createOrderStoreModel(requestOrder)
	if echoErr != nil {
		return echoErr
	}

	h.oService.Lock()
	defer h.oService.Unlock()
	success, intErr := h.oService.CreateOrder(order.Id, GetUserFromJWT(c))
	if intErr != nil {
		if intErr.Err != nil {
			h.log.Infof(intErr.Error())
		}

		return echo.NewHTTPError(intErr.Code, intErr.Message)
	}
	if !success {
		return c.JSON(http.StatusOK, "Номер заказа уже был загружен этим пользователем")
	}

	return c.JSON(http.StatusAccepted, "Новый номер заказа принят в обработку")
}

func (h *Handler) getOrder(c echo.Context) error {
	list, intErr := h.oService.GetListByUser(GetUserFromJWT(c))
	if intErr != nil {
		if intErr.Err != nil {
			h.log.Infof(intErr.Error())
		}

		return echo.NewHTTPError(intErr.Code, intErr.Message)
	}

	fmt.Println(list)

	// 500 => Внутренняя ошибка сервера
	// 200 => Список загруженных номеров заказов
	// 204 => Нет данных для ответа
	// 401 => Пользователь не авторизован

	return c.JSON(http.StatusOK, list)
}

func (h *Handler) getOrderBody(c echo.Context) (*models.RequestOrder, *echo.HTTPError) {
	requestOrder := &models.RequestOrder{}
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.log.Errorf("failed to read body: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}
	requestOrder.Id = string(body)

	return requestOrder, nil
}

func (h *Handler) createOrderStoreModel(order *models.RequestOrder) (*models2.Order, *echo.HTTPError) {
	id, err := strconv.ParseInt(order.Id, 10, 64)
	if err != nil {
		h.log.Errorf("failed to convert str to int: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}

	return &models2.Order{Id: id}, nil
}
