package models

type (
	RequestUser struct {
		Login    string `json:"login" validate:"required,alphanum|email"`
		Password string `json:"password" validate:"required"`
	}
	RequestOrder struct {
		ID string `validate:"required,number,algLuna"`
	}
	RequestWithdraw struct {
		Order string `json:"order" validate:"required,alphanum|algLuna"`
		Sum   int    `json:"sum" validate:"required"`
	}
)
