package authzy

// Updated by linker flags during build.
var (
	// The git commit that was compiled. This will be filled in by the compiler.
	GitCommit   string
	GitDescribe string

	version = "dev"

	Version = func() string {
		if GitDescribe != "" {
			return GitDescribe
		}

		return version
	}()
)
