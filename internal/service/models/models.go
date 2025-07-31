package models

type (
	RequestUser struct {
		Login    string `json:"login" validate:"required,alphanum|email"`
		Password string `json:"password" validate:"required"`
	}
	RequestOrder struct {
		Id string `validate:"required,number,algLuna"`
	}
)
