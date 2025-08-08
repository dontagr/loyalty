package transport

import (
	error2 "github.com/dontagr/loyalty/internal/service/customerror"
	"github.com/dontagr/loyalty/internal/service/transport/models"
)

type (
	Transport interface {
		NewRequest(orderID string, w int) (*models.OrderResponse, *error2.CustomError)
	}
)
