package worker

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/config"
	"github.com/dontagr/loyalty/internal/service/transport"
	"github.com/dontagr/loyalty/internal/store"
	"github.com/dontagr/loyalty/internal/store/models"
)

type Updater struct {
	cfg       *config.Config
	log       *zap.SugaredLogger
	workers   int
	interval  int
	pg        *store.PG
	transport transport.Transport
}

func NewUpdater(cfg *config.Config, pg *store.PG, transport *transport.HTTPManager, log *zap.SugaredLogger, lc fx.Lifecycle) *Updater {
	u := &Updater{
		cfg:       cfg,
		log:       log,
		workers:   cfg.Service.WorkerLimit,
		interval:  cfg.Service.UpdaterInterval,
		pg:        pg,
		transport: transport,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go u.Handle()

			return nil
		},
	})

	return u
}

func (u *Updater) Handle() {
	jobs := make(chan *models.Order, u.workers)
	for w := 1; w <= u.workers; w++ {
		go u.worker(w, jobs)
	}

	for {
		time.Sleep(time.Duration(u.interval) * time.Second)

		processing, err := u.pg.GetListForProcessing()
		if err != nil {
			u.log.Errorf("failed to get order list: %v", err)
			continue
		}

		for _, order := range processing {
			jobs <- order
		}
	}
}

func (s *Updater) worker(w int, jobs chan *models.Order) {
	s.log.Infof("worker %d runing", w)
	for row := range jobs {
		request, err := s.transport.NewRequest(row.ID, w)
		if err != nil {
			s.log.Errorf("worker %d request error code:%d message:%s : %v", w, err.Code, err.Message, err.Err)

			if err.Code == http.StatusTooManyRequests {
				time.Sleep(time.Duration(60) * time.Second)
			}
			return
		}

		order := &models.Order{ID: row.ID, Accrual: &request.Accrual}
		order.SetStatusFromStr(request.Status)

		if order.Status == models.StatusNew {
			continue
		}

		s.pg.UpdateOrder(order)
	}
}
