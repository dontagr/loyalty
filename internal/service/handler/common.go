package handler

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/service/customerror"
	"github.com/dontagr/loyalty/internal/service/jwt"
	"github.com/dontagr/loyalty/internal/service/order"
	"github.com/dontagr/loyalty/internal/service/user"
	"github.com/dontagr/loyalty/internal/service/withdrawal"
)

type (
	Handler struct {
		log      *zap.SugaredLogger
		uService *user.Service
		oService *order.Service
		wService *withdrawal.Service
		jwt      *jwt.JWTService
	}
)

func NewHandler(
	uService *user.Service,
	oService *order.Service,
	wService *withdrawal.Service,
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

func (h *Handler) convertCustomErrorToServerCode(code int) int {
	switch code {
	case customerror.Internal:
		return http.StatusInternalServerError
	case customerror.Unprocessable:
		return http.StatusUnprocessableEntity
	case customerror.Payment:
		return http.StatusPaymentRequired
	case customerror.Unauthorized:
		return http.StatusUnauthorized
	case customerror.Conflict:
		return http.StatusConflict
	default:
		return 0
	}
}
