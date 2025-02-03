package cmd

import (
	"fmt"
	"github.com/oxio/kvf/pkg/kvf"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <file1> [<file2> <file3> ...] <key> <value>",
		Short: "Sets a value to a key-value file",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) < 3 {
				return fmt.Errorf("not enough arguments")
			}

			files := args[:len(args)-2]
			key := args[len(args)-2]
			value := args[len(args)-1]

			for _, file := range files {
				repo := kvf.NewFileRepo(file, false)
				item, err := kvf.NewItem(key, value)
				if err != nil {
					return err
				}

				err = repo.Set(item)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(newSetCmd())
}
