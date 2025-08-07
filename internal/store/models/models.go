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
		Balance      int
	}
	Order struct {
		ID             string      `json:"number"`
		UserID         int         `json:"-"`
		Status         OrderStatus `json:"status"`
		Accrual        *int        `json:"accrual,omitempty"`
		CreateDateTime time.Time   `json:"uploaded_at"`
	}
	Withdrawal struct {
		ID             string    `json:"order"`
		UserID         int       `json:"-"`
		Withdrawal     int       `json:"sum"`
		CreateDateTime time.Time `json:"processed_at"`
	}
	OrderStatus string
)

const (
	StatusNew        = "new"
	StatusProcessing = "processing"
	StatusInvalid    = "invalid"
	StatusProcessed  = "processed"
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

func (o *Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		Accrual float64 `json:"accrual,omitempty"`
		*Alias
	}{
		Accrual: func() float64 {
			if o.Accrual != nil {
				return float64(*o.Accrual) / 100
			}
			return 0.0
		}(),
		Alias: (*Alias)(o),
	})
}

func (w *Withdrawal) MarshalJSON() ([]byte, error) {
	type Alias Withdrawal
	return json.Marshal(&struct {
		Withdrawal float64 `json:"sum"`
		*Alias
	}{
		Withdrawal: float64(w.Withdrawal) / 100,
		Alias:      (*Alias)(w),
	})
}
