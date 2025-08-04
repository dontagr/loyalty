package handler

import (
	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/service/jwt"
	"github.com/dontagr/loyalty/internal/service/order"
	"github.com/dontagr/loyalty/internal/service/user"
	"github.com/dontagr/loyalty/internal/service/withdrawal"
)

type (
	Handler struct {
		log      *zap.SugaredLogger
		uService *user.UserService
		oService *order.OrderService
		wService *withdrawal.WithdrawalService
		jwt      *jwt.JWTService
	}
)

func NewHandler(
	uService *user.UserService,
	oService *order.OrderService,
	wService *withdrawal.WithdrawalService,
	log *zap.SugaredLogger,
	jwtService *jwt.JWTService,
) *Handler {
	h := &Handler{
		log:      log,
		uService: uService,
		oService: oService,
		wService: wService,
		jwt:      jwtService,
	}

	return h
}
