package di

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // Comment this line to disable pprof endpoint.
	"strings"

	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/logger"
)

var debugfx = fx.Invoke(NewDebug)

type DebugParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	OnErrorCh chan error

	Log         logger.Logger
	DebugConfig *config.DebugConfig
}

func NewDebug(p DebugParams) error {
	addr, err := net.ResolveTCPAddr("tcp", p.DebugConfig.Addr)
	if err != nil {
		return fmt.Errorf("could not resolve TCP address: %w", err)
	}

	pprofHostPort := addr.String()
	parts := strings.Split(pprofHostPort, ":")
	if len(parts) == 2 && parts[0] == "" {
		pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if p.DebugConfig.Enabled {
				go func() {
					p.Log.Infof("starting pprof HTTP server at: http://%s/debug/pprof/", pprofHostPort)

					if err := http.ListenAndServe(pprofHostPort, nil); err != nil && err != http.ErrServerClosed {
						p.Log.Errorf("unable to start debug server: %v", err)
						p.OnErrorCh <- err
					}
				}()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return nil
}
