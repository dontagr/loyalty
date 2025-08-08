package jwt

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/dontagr/loyalty/internal/config"
	"github.com/dontagr/loyalty/internal/store/models"
)

type (
	JWTService struct {
		key string
	}
	JWTAuth struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		jwt.RegisteredClaims
	}
)

func NewJWTService(cnf *config.Config) *JWTService {
	return &JWTService{key: cnf.Security.Key}
}

func (j *JWTService) GetJWT(ID int, Login string) (string, error) {
	claims := &JWTAuth{
		ID,
		Login,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	hash, err := token.SignedString([]byte(j.key))
	if err != nil {
		return "", err
	}

	return hash, err
}

func (j *JWTService) GetJWTEchoConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return &JWTAuth{} },
		SigningKey:    []byte(j.key),
		TokenLookup:   "header:Authorization",
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Пользователь не аутентифицирован"})
		},
	}
}

func (j *JWTService) GetUser(c echo.Context) *models.User {
	jwtUser := c.Get("user").(*jwt.Token)
	claims := jwtUser.Claims.(*JWTAuth)

	return &models.User{
		ID:    claims.ID,
		Login: claims.Login,
	}
}

func (j *JWTService) GetMiddleware(jwtConfig echojwt.Config) echo.MiddlewareFunc {
	return echojwt.WithConfig(jwtConfig)
}
