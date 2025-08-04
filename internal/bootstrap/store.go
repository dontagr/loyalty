package bootstrap

import (
	"go.uber.org/fx"

	"github.com/dontagr/loyalty/internal/service/interfaces"
	"github.com/dontagr/loyalty/internal/store/order"
	"github.com/dontagr/loyalty/internal/store/user"
	"github.com/dontagr/loyalty/internal/store/withdrawal"
)

var Store = fx.Options(
	fx.Provide(
		fx.Annotate(
			order.NewOrder,
			fx.As(new(interfaces.OrderStore)),
		),
		fx.Annotate(
			user.NewUser,
			fx.As(new(interfaces.UserStore)),
		),
		fx.Annotate(
			withdrawal.NewWithdrawal,
			fx.As(new(interfaces.WithdrawalStore)),
		),
	),
	fx.Invoke(
		func(interfaces.OrderStore) {},
		func(interfaces.UserStore) {},
		func(interfaces.WithdrawalStore) {},
	),
)
