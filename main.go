package main

import (
	"runtime/debug"

	"github.com/oxio/kvf/cmd"
)

// version is set via ldflags during build: -X main.version=<version>
var version = "dev"

func init() {
	// If version wasn't set via ldflags, try to get it from build info
	// This allows `go install` users to see a meaningful version
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					if len(setting.Value) >= 7 {
						version = setting.Value[:7] // short commit hash
					} else {
						version = setting.Value
					}
					break
				}
			}
		}
	}
}

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
