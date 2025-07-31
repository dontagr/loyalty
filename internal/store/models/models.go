package models

import (
	"encoding/json"
	"time"
)

type (
	User struct {
		Id           int    `json:"id"`
		Login        string `json:"login"`
		PasswordHash string `json:"password"`
	}
	Order struct {
		Id             int64       `json:"id"`
		UserId         int         `json:"user_id"`
		Status         OrderStatus `json:"status"`
		Accrual        *int        `json:"accrual,omitempty"`
		CreateDateTime time.Time   `json:"uploaded_at"`
	}
	OrderStatus int
)

const (
	StatusNEW OrderStatus = iota
	StatusPROCESSING
	StatusInvalid
	StatusProcessed
)

var statusToString = map[OrderStatus]string{
	StatusNEW:        "NEW",
	StatusPROCESSING: "PROCESSING",
	StatusInvalid:    "INVALID",
	StatusProcessed:  "PROCESSED",
}

func (status OrderStatus) String() string {
	if str, exists := statusToString[status]; exists {
		return str
	}

	return "UNKNOWN"
}

func (status OrderStatus) MarshalJSON() ([]byte, error) {
	str := status.String()

	return json.Marshal(str)
}

func (u *User) Unpack() (string, string) {
	return u.Login, u.PasswordHash
}
