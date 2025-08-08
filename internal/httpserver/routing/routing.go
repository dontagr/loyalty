package routing

import (
	"fmt"

	"github.com/dontagr/loyalty/internal/httpserver"
	"github.com/dontagr/loyalty/internal/service/checkup"
	"github.com/dontagr/loyalty/internal/service/handler"
	"github.com/dontagr/loyalty/internal/service/jwt"
)

func InitRouting(
	server *httpserver.HTTPServer,
	jwt *jwt.JWTService,
	handler *handler.Handler,
) error {
	var err error
	jwtConfig := jwt.GetJWTEchoConfig()
	server.Master.Validator, err = checkup.NewCustomValidator()
	if err != nil {
		return fmt.Errorf("failed create validator %v", err)
	}

	g := server.Master.Group("/api/user")
	g.POST("/register", handler.SignUp)
	g.POST("/login", handler.SignIn)
	g.GET("/orders", handler.GetOrder, jwt.GetMiddleware(jwtConfig))
	g.POST("/orders", handler.CreateOrder, jwt.GetMiddleware(jwtConfig))
	g.GET("/withdrawals", handler.GetWithdraw, jwt.GetMiddleware(jwtConfig))
	g.GET("/balance", handler.GetBalance, jwt.GetMiddleware(jwtConfig))
	g.POST("/balance/withdraw", handler.PostBalanceWithdraw, jwt.GetMiddleware(jwtConfig))

	return nil
}
