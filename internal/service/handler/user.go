package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/dontagr/loyalty/internal/service/models"
)

func (h *Handler) signUp(c echo.Context) error {
	requestUser, echoError := h.getRequestUser(c)
	if echoError != nil {
		return echoError
	}

	h.uService.Lock()
	defer h.uService.Unlock()
	hasLogin, err := h.uService.HasLogin(requestUser.Login)
	if err != nil {
		h.log.Errorf("has login error: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if hasLogin {
		return echo.NewHTTPError(http.StatusConflict, "Логин уже занят")
	}

	jwtHash, err := h.uService.SignUp(requestUser.Login, requestUser.Password)
	if err != nil {
		h.log.Errorf("failed registration: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}

	c.Response().Header().Set("Authorization", jwtHash)

	return c.JSON(http.StatusOK, "Пользователь успешно зарегистрирован и аутентифицирован")
}

func (h *Handler) signIn(c echo.Context) error {
	requestUser, echoError := h.getRequestUser(c)
	if echoError != nil {
		return echoError
	}

	user, err := h.uService.GetUser(requestUser.Login)
	if err != nil {
		h.log.Errorf("get user error: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	if user.Login == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Неверная пара логин/пароль")
	}

	jwtHash, intErr := h.uService.SignIn(requestUser.Password, user)
	if intErr != nil {
		if intErr.Err != nil {
			h.log.Infof(intErr.Error())
		}

		return echo.NewHTTPError(intErr.Code, intErr.Message)
	}

	c.Response().Header().Set("Authorization", jwtHash)

	return c.JSON(http.StatusOK, "Пользователь успешно аутентифицирован")
}

func (h *Handler) getRequestUser(c echo.Context) (*models.RequestUser, *echo.HTTPError) {
	requestUser := &models.RequestUser{}
	if err := c.Bind(requestUser); err != nil {
		h.log.Errorf("request failed: %v", err)

		return nil, echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}
	if err := c.Validate(requestUser); err != nil {
		h.log.Errorf("validation failed: %v", err)

		return nil, echo.NewHTTPError(http.StatusBadRequest, "Неверный формат запроса")
	}

	return requestUser, nil
}
