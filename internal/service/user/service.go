package user

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/dontagr/loyalty/internal/service/customerror"
	"github.com/dontagr/loyalty/internal/service/interfaces"
	"github.com/dontagr/loyalty/internal/service/jwt"
	"github.com/dontagr/loyalty/internal/store/models"
)

type UserService struct {
	store      interfaces.UserStore
	jwtService *jwt.JWTService
}

func NewUserService(store interfaces.UserStore, jwtService *jwt.JWTService) *UserService {
	return &UserService{store: store, jwtService: jwtService}
}

func (u *UserService) Lock() {
	u.store.Lock()
}

func (u *UserService) Unlock() {
	u.store.Unlock()
}

func (u *UserService) HasLogin(login string) (bool, error) {
	user, err := u.store.GetUser(login)
	if err != nil {
		return false, err
	}

	return user.Login == login, nil
}

func (u *UserService) GetUser(login string) (*models.User, error) {
	return u.store.GetUser(login)
}

func (u *UserService) SignUp(login string, password string) (string, error) {
	passHash, err := u.generatePassHash(password)
	if err != nil {
		return "", err
	}

	err = u.store.SaveUser(login, passHash)
	if err != nil {
		return "", err
	}

	user, err := u.store.GetUser(login)
	if err != nil {
		return "", err
	}

	jwtHash, err := u.jwtService.GetJWT(user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("failed create jwt: %v", err)
	}

	return jwtHash, nil
}

func (u *UserService) SignIn(password string, user *models.User) (string, *customerror.CustomError) {
	valid, cError := u.CompareHashAndPassword(user, password)
	if !valid {
		return "", cError
	}

	jwtHash, err := u.jwtService.GetJWT(user.ID, user.Login)
	if err != nil {
		return "", customerror.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("failed create jwt: %v", err))
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

func (u *UserService) CompareHashAndPassword(user *models.User, password string) (bool, *customerror.CustomError) {
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return false, customerror.NewCustomError(http.StatusUnauthorized, "Неверная пара логин/пароль", nil)
	}

	return true, nil
}
