package di

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/logger"
	"github.com/zbiljic/authzy/pkg/netutil"
)

var serverfx = fx.Invoke(NewServer)

type ServerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	OnErrorCh chan error

	Log        logger.Logger
	HTTPConfig *config.HTTPConfig
	Router     http.Handler `name:"api"`
}

func NewServer(p ServerParams) (*http.Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", p.HTTPConfig.Addr)
	if err != nil {
		return nil, fmt.Errorf("could not resolve TCP address: %w", err)
	}

	hostPort := addr.String()

	parts := strings.Split(hostPort, ":")
	if len(parts) == 2 && parts[0] == "" {
		host, err := netutil.PrivateIP()
		if err != nil {
			return nil, err
		}

		hostPort = fmt.Sprintf("%s:%s", host, parts[1])
	}

	server := &http.Server{
		Addr:         hostPort,
		Handler:      p.Router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				p.Log.Infof("starting HTTP server at: http://%s", hostPort)

				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					p.Log.Errorf("unable to start server: %v", err)
					p.OnErrorCh <- err
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Log.Infof("stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})

	return server, nil
}
