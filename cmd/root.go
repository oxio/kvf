package cmd

import (
	"io"

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

// SetOut sets the output writer for the root command.
func SetOut(out io.Writer) {
	rootCmd.SetOut(out)
}

func Execute() {
	_ = rootCmd.Execute()
}
