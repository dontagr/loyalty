package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/config"
	intError "github.com/dontagr/loyalty/internal/service/error"
	"github.com/dontagr/loyalty/internal/service/transport/models"
)

type HTTPManager struct {
	urlPattern string
	client     *http.Client
	log        *zap.SugaredLogger
	cfg        *config.Config
}

func NewHTTPManager(cfg *config.Config, log *zap.SugaredLogger) *HTTPManager {
	return &HTTPManager{urlPattern: "http://%s/api/orders/%s", log: log, client: &http.Client{}, cfg: cfg}
}

func (h *HTTPManager) NewRequest(orderID string, w int) (*models.OrderResponse, *intError.CustomError) {
	req, err := http.NewRequest("GET", fmt.Sprintf(h.urlPattern, h.cfg.CalculateSystem.URI, orderID), nil)
	if err != nil {
		return nil, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("creating request: %v", err))
	}

	var resp *http.Response
	var netErr *net.OpError
	var errSend error
	var orderResponse *models.OrderResponse
	for i := 0; i < 3; i++ {
		resp, errSend = h.client.Do(req)
		if errSend == nil {
			defer func(Body io.ReadCloser, w int) {
				err := Body.Close()
				if err != nil {
					h.log.Errorf("worker %d failed close body %v", w, err)
				}
			}(resp.Body, w)
			if resp.StatusCode != http.StatusOK {
				return nil, intError.NewCustomError(resp.StatusCode, "ошибка от системы расчета", fmt.Errorf("sending data: %v", errSend))
			}

			orderResponse = new(models.OrderResponse)
			err = json.NewDecoder(resp.Body).Decode(orderResponse)
			if err != nil {
				return nil, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("decoding response: %v", err))
			}
			break
		}
		err := resp.Body.Close()
		if err != nil {
			h.log.Errorf("worker %d failed close body %v", w, err)
		}

		if errors.As(errSend, &netErr) {
			h.log.Warnf("worker %d connection error we try №%d", w, i+1)
			time.Sleep(5 * time.Second)
		} else {
			return nil, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("sending data: %v", errSend))
		}
	}

	if errSend != nil {
		return nil, intError.NewCustomError(http.StatusInternalServerError, "Внутренняя ошибка сервера", fmt.Errorf("sending data: %v", errSend))
	}

	h.log.Infof("worker %d request success full", w)

	return orderResponse, nil
}
