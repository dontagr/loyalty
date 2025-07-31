package user

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	intError "github.com/dontagr/loyalty/internal/service/error"
	"github.com/dontagr/loyalty/internal/service/jwt"
	"github.com/dontagr/loyalty/internal/store"
	"github.com/dontagr/loyalty/internal/store/models"
)

type UserService struct {
	// TODO use interface
	pg         *store.PG
	jwtService *jwt.JWTService
}

func NewUserService(pg *store.PG, jwtService *jwt.JWTService) *UserService {
	return &UserService{pg: pg, jwtService: jwtService}
}

func (u *UserService) Lock() {
	u.pg.LockUser()
}

func (u *UserService) Unlock() {
	u.pg.UnlockUser()
}

func (u *UserService) HasLogin(login string) (bool, error) {
	user, err := u.pg.GetUser(login)
	if err != nil {
		return false, err
	}

	return user.Login == login, nil
}

func (u *UserService) GetUser(login string) (*models.User, error) {
	return u.pg.GetUser(login)
}

func (u *UserService) SignUp(login string, password string) (string, error) {
	passHash, err := u.generatePassHash(password)
	if err != nil {
		return "", err
	}

	err = u.pg.SaveUser(login, passHash)
	if err != nil {
		return "", err
	}

	user, err := u.pg.GetUser(login)
	if err != nil {
		return "", err
	}

	jwtHash, err := u.jwtService.GetJWT(user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("failed create jwt: %v", err)
	}

	return jwtHash, nil
}

func (u *UserService) SignIn(password string, user *models.User) (string, *intError.CustomError) {
	valid, intErrors := u.CompareHashAndPassword(user, password)
	if !valid {
		return "", intErrors
	}

	jwtHash, err := u.jwtService.GetJWT(user.ID, user.Login)
	if err != nil {
		return "", intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed create jwt: %v", err))
	}

	return jwtHash, nil
}

func (u *UserService) generatePassHash(password string) (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %v", err)
	}

	return string(passHash), nil
}

func (u *UserService) CompareHashAndPassword(user *models.User, password string) (bool, *intError.CustomError) {
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return false, intError.NewCustomError(http.StatusUnauthorized, "Неверная пара логин/пароль", nil)
	}

	return true, nil
}
