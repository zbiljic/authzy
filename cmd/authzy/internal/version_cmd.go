package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zbiljic/authzy"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE:  printVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func printVersion(cmd *cobra.Command, args []string) error {
	_, err := fmt.Fprintf(os.Stdout, "%s\n", authzy.Version)
	return err
}
