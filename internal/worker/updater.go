package worker

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/config"
	"github.com/dontagr/loyalty/internal/service/interfaces"
	"github.com/dontagr/loyalty/internal/service/transport"
	"github.com/dontagr/loyalty/internal/store/models"
)

type Updater struct {
	cfg       *config.Config
	log       *zap.SugaredLogger
	workers   int
	interval  int
	store     interfaces.OrderStore
	transport transport.Transport
}

func NewUpdater(cfg *config.Config, store interfaces.OrderStore, transport *transport.HTTPManager, log *zap.SugaredLogger, lc fx.Lifecycle) *Updater {
	u := &Updater{
		cfg:       cfg,
		log:       log,
		workers:   cfg.Service.WorkerLimit,
		interval:  cfg.Service.UpdaterInterval,
		store:     store,
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

func (upd *Updater) Handle() {
	jobs := make(chan *models.Order, upd.workers)
	for w := 1; w <= upd.workers; w++ {
		go upd.worker(w, jobs)
	}

	for {
		time.Sleep(time.Duration(upd.interval) * time.Second)
		upd.log.Infof("start planing")

		processing, err := upd.store.GetListForProcessing()
		if err != nil {
			upd.log.Errorf("failed to get order list: %v", err)
			continue
		}

		upd.log.Infof("start send")
		for _, order := range processing {
			jobs <- order
		}
		upd.log.Infof("finish send")
	}
}

func (upd *Updater) worker(w int, jobs chan *models.Order) {
	upd.log.Infof("worker %d runing", w)
	for row := range jobs {
		upd.orderProcess(row, w)
	}
}

func (upd *Updater) orderProcess(row *models.Order, w int) {
	if !upd.store.BlockOrder(row.ID) {
		return
	}
	defer upd.store.UnblockOrder(row.ID)

	request, err := upd.transport.NewRequest(row.ID, w)
	if err != nil {
		upd.log.Errorf("worker %d request orderID:%s error code:%d message:%s : %v", w, row.ID, err.Code, err.Message, err.Err)

		if err.Code == http.StatusTooManyRequests {
			time.Sleep(time.Duration(60) * time.Second)
		}
		return
	}

	order := &models.Order{ID: row.ID, Accrual: &request.Accrual}
	order.SetStatusFromStr(request.Status)

	if order.Status == models.StatusNew {
		return
	}

	er := upd.store.UpdateOrder(order)
	if er != nil {
		upd.log.Errorf("worker %d update failed: %v", w, er)
	}
}
