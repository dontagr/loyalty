package handler

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/config"
	"github.com/dontagr/loyalty/internal/httpserver"
	jwt2 "github.com/dontagr/loyalty/internal/service/jwt"
	"github.com/dontagr/loyalty/internal/service/order"
	"github.com/dontagr/loyalty/internal/service/user"
	"github.com/dontagr/loyalty/internal/service/withdrawal"
	"github.com/dontagr/loyalty/internal/store/models"
)

type (
	Handler struct {
		log      *zap.SugaredLogger
		uService *user.UserService
		oService *order.OrderService
		wService *withdrawal.WithdrawalService
	}
	CustomValidator struct {
		validator *validator.Validate
	}
)

func NewHandler(
	cfg *config.Config,
	server *httpserver.HTTPServer,
	uService *user.UserService,
	oService *order.OrderService,
	wService *withdrawal.WithdrawalService,
	log *zap.SugaredLogger,
	lc fx.Lifecycle,
) *Handler {
	h := &Handler{
		log:      log,
		uService: uService,
		oService: oService,
		wService: wService,
	}

	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(jwt2.JWTAuth) },
		SigningKey:    []byte(cfg.Security.Key),
		TokenLookup:   "header:Authorization",
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Пользователь не аутентифицирован"})
		},
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			validate := validator.New()
			err := validate.RegisterValidation("algLuna", AlgLunaValidator)
			if err != nil {
				return err
			}
			server.Master.Validator = &CustomValidator{validator: validate}

			g := server.Master.Group("/api/user")

			g.POST("/register", h.signUp)
			g.POST("/login", h.signIn)

			g.GET("/orders", h.getOrder, echojwt.WithConfig(jwtConfig))
			g.POST("/orders", h.createOrder, echojwt.WithConfig(jwtConfig))

			g.GET("/withdrawals", h.getWithdraw, echojwt.WithConfig(jwtConfig))
			g.GET("/balance", h.getBalance, echojwt.WithConfig(jwtConfig))
			g.POST("/balance/withdraw", h.postBalanceWithdraw, echojwt.WithConfig(jwtConfig))
			server.Master.GET("/balance", h.getBalance, echojwt.WithConfig(jwtConfig))

			return nil
		},
	})

	return h
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func AlgLunaValidator(fl validator.FieldLevel) bool {
	switch v := fl.Field(); v.Kind() {
	case reflect.String:
		number := v.String()

		return AlgLuna(number)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number := strconv.FormatInt(v.Int(), 10)

		return AlgLuna(number)
	default:
		return false
	}
}

func AlgLuna(number string) bool {
	var sum int
	nDigits := len(number)
	isSecond := false

	for i := nDigits - 1; i >= 0; i-- {
		dchar := number[i]
		if !unicode.IsDigit(rune(dchar)) {
			return false
		}

		digit, _ := strconv.Atoi(string(dchar))
		if isSecond {
			digit = digit * 2
		}
		if digit > 9 {
			digit = digit - 9
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}

func GetUserFromJWT(c echo.Context) *models.User {
	jwtUser := c.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*jwt2.JWTAuth)

	return &models.User{
		ID:    claims.ID,
		Login: claims.Login,
	}
}
