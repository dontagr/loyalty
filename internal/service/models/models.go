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
		Order string  `json:"order" validate:"required,number,algLuna"`
		Sum   float64 `json:"sum" validate:"required,floatGtZero"`
	}
	ResponceWithdraw struct {
		Balance    float64 `json:"current"`
		Withdrawal float64 `json:"withdrawn"`
	}
)
