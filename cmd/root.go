package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kvf",
	Short: "Simple Key-Value storage tool",
}

// SetVersion sets the version for the root command.
// It should be called from main after the version is injected via ldflags.
func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() {
	_ = rootCmd.Execute()
}
