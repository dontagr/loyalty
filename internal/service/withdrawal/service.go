package withdrawal

import (
	"github.com/dontagr/loyalty/internal/store"
)

type WithdrawalService struct {
	// TODO use interface
	pg *store.PG
}

func NewWithdrawalService(pg *store.PG) *WithdrawalService {
	return &WithdrawalService{pg: pg}
}

func (w *WithdrawalService) GetTotalWithdrawal(userID int) (float64, error) {
	return w.pg.GetTotalWithdrawal(userID)
}
