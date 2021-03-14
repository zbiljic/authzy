package internal

import (
	"github.com/spf13/viper"
)

//nolint
const (
	// exitCodeOK is a 0 exit code.
	exitCodeOK = iota
	// exitCodeErr is a generic exit code, for all non-special errors.
	exitCodeErr
)

var (
	globalConfig = "" // Config flag set via command line
	// WHEN YOU ADD NEXT GLOBAL FLAG, MAKE SURE TO ALSO UPDATE PERSISTENT FLAGS, FLAG CONSTANTS AND UPDATE FUNC.
)

var (
	configSections = map[string]func(string) string{
		"global": nil,
	}
	// WHEN YOU ADD NEXT GLOBAL CONFIG SECTION, MAKE SURE TO ALSO UPDATE TESTS, ETC.
)

var _ error = initConfigSections()

func initConfigSections() error {
	// DO NOT EDIT - builds section functions
	for k := range configSections {
		localK := k
		configSections[localK] = func(key string) string {
			return localK + "." + key
		}
	}
	return nil
}

func updateGlobals() {
	globalSection := configSections["global"]
	globalConfig = viper.GetString(globalSection("config"))
}
