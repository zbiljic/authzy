package internal

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/di"
	"github.com/zbiljic/authzy/pkg/logger"
	"github.com/zbiljic/authzy/pkg/logger/zlogger"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return execWithConfig(cmd, start)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func start(conf *config.Config) error {
	log, err := zlogger.New(conf.Logger)
	if err != nil {
		return fmt.Errorf("error creating logger: %w", err)
	}

	onErrorCh := make(chan error)
	defer close(onErrorCh)

	app := fx.New(
		fx.Supply(conf),
		fx.Logger(di.NewFxLogger(log)),
		fx.Provide(func() logger.Logger { return log }),
		fx.Provide(func() chan error { return onErrorCh }),
		di.Module,
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Error(err)
		return err
	}

	log.Info("started")

	// Wait for termination or interrupt signals.
	// Block main goroutine until it is interrupted.
	select {
	case sig := <-app.Done():
		log.Debugf("received signal: %v", sig)
	case err := <-onErrorCh:
		log.Debugf("lifecycle error: %v", err)
	}

	log.Info("shutting down")

	if err := app.Stop(ctx); err != nil {
		log.Errorf("shutdown: %v", err)
		return err
	}

	return nil
}
