package bootstrap

import (
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/service/jwt"
	"github.com/dontagr/loyalty/internal/service/order"
	"github.com/dontagr/loyalty/internal/service/transport"
	"github.com/dontagr/loyalty/internal/service/user"
	"github.com/dontagr/loyalty/internal/service/withdrawal"
)

var Service = fx.Options(
	fx.Provide(
		user.NewUserService,
		jwt.NewJWTService,
		order.NewOrderService,
		transport.NewHTTPManager,
		withdrawal.NewWithdrawalService,
	),
)
