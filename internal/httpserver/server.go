package httpserver

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/dontagr/loyalty/internal/config"
)

type HTTPServer struct {
	Master *echo.Echo
}

func NewServer(cfg *config.Config, log *zap.SugaredLogger, lc fx.Lifecycle, shutdowner fx.Shutdowner) *HTTPServer {
	mainServer := echo.New()

	mainServer.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogMethod:       true,
		LogStatus:       true,
		LogError:        true,
		LogResponseSize: true,
		LogLatency:      true,
		HandleError:     true,
		LogHeaders:      []string{echo.HeaderContentType, echo.HeaderContentEncoding, echo.HeaderAcceptEncoding},
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.Infow("Request", "Method", v.Method, "URI", v.URI, "Status", v.Status, "Duration", v.Latency, "ResponseSize", v.ResponseSize, "Headers", v.Headers)
			} else {
				log.Errorw(v.Error.Error(), "Method", v.Method, "URI", v.URI, "Status", v.Status, "Duration", v.Latency, "ResponseSize", v.ResponseSize, "Headers", v.Headers)
			}

			return nil
		},
	}))
	mainServer.Use(middleware.Decompress())
	mainServer.Use(middleware.Gzip())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Infof("starting HTTP server. Bind: %s", cfg.HTTPServer.BindAddress)
			go func() {
				if err := mainServer.Start(cfg.HTTPServer.BindAddress); err != nil && err != http.ErrServerClosed {
					log.Errorf("failed to start HTTP Server: %v", err)
					_ = shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return mainServer.Shutdown(ctx)
		},
	})

	return &HTTPServer{
		Master: mainServer,
	}
}
