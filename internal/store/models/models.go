package models

import (
	"encoding/json"
	"time"
)

type (
	User struct {
		ID           int    `json:"id"`
		Login        string `json:"login"`
		PasswordHash string `json:"password"`
		Balance      float64
	}
	Order struct {
		ID             string      `json:"number"`
		UserID         int         `json:"-"`
		Status         OrderStatus `json:"status"`
		Accrual        *float64    `json:"accrual,omitempty"`
		CreateDateTime time.Time   `json:"uploaded_at"`
	}
	Withdrawal struct {
		ID             string    `json:"order"`
		UserID         int       `json:"-"`
		Withdrawal     float64   `json:"sum"`
		CreateDateTime time.Time `json:"processed_at"`
	}
	OrderStatus int
)

const (
	StatusNew OrderStatus = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)

var statusToString = map[OrderStatus]string{
	StatusNew:        "NEW",
	StatusProcessing: "PROCESSING",
	StatusInvalid:    "INVALID",
	StatusProcessed:  "PROCESSED",
}
var stringToStatus = map[string]OrderStatus{
	"REGISTERED": StatusNew,
	"PROCESSING": StatusProcessing,
	"INVALID":    StatusInvalid,
	"PROCESSED":  StatusProcessed,
}

func (o *Order) SetStatusFromStr(status string) {
	if intStatus, exists := stringToStatus[status]; exists {
		o.Status = intStatus
	}
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
