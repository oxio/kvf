package main

import "github.com/oxio/kvf/cmd"

// version is set via ldflags during build: -X main.version=<version>
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
