package config

import (
	"os"
)

// Param represents config parameters.
type config struct {
	// BackendEndpoint is an endpoint for backend storage.
	BackendEndpoint string
}

// Config is set from the environment variables.
var Config = &config{}

// Load loads into Config from environment values.
func Load() {
	if v := os.Getenv("BINREP_BACKEND_ENDPOINT"); v != "" {
		Config.BackendEndpoint = v
	}
}
