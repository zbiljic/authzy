package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zbiljic/authzy"
	"github.com/zbiljic/authzy/pkg/config"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   authzy.AppName,
	Short: "Authzy is an Identity and User Management system",
	Long:  `Authzy is an Identity and User Management system`,
}

func init() {
	defineFlagsGlobal(rootCmd.PersistentFlags())

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return registerBefore(cmd)
	}
}

func registerBefore(cmd *cobra.Command) error {
	// Bind global flags to viper config.
	err := bindFlagsGlobal(cmd)
	if err != nil {
		return err
	}

	// Update global flags (if anything changed from other sources).
	updateGlobals()

	return nil
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if cmd, err := rootCmd.ExecuteC(); err != nil {
		if strings.Contains(err.Error(), "arg(s)") || strings.Contains(err.Error(), "usage") {
			cmd.Usage() //nolint:errcheck
		}

		os.Exit(exitCodeErr)
	}
}

func execWithConfig(_ *cobra.Command, fn func(conf *config.Config) error) error {
	conf, err := config.LoadConfig(globalConfig)
	if err != nil {
		return fmt.Errorf("Failed to load configuration: %v", err)
	}

	return fn(conf)
}
