package internal

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Place for all the flags for all commands

var globalFlags = map[string]func(*pflag.FlagSet){
	"config": func(flags *pflag.FlagSet) {
		flags.StringP("config", "c", "", "Configuration file to use")
	},
}

func defineFlagsGlobal(flags *pflag.FlagSet) {
	for _, ffn := range globalFlags {
		ffn(flags)
	}
}

func bindFlagsGlobal(cmd *cobra.Command) error {
	globalSection := configSections["global"]
	for k := range globalFlags {
		key := globalSection(k)
		err := viper.BindPFlag(key, cmd.Flag(k))
		if err != nil {
			return err
		}
	}
	return nil
}
