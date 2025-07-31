package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dontagr/loyalty/internal/config"
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
