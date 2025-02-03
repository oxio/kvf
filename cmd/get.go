package cmd

import (
	"errors"
	"fmt"
	"github.com/oxio/kvf/pkg/kvf"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	var defaultVal *string
	var skipMissingFiles *bool

	cmd := &cobra.Command{
		Use:   "get <file1> [<file2> <file3> ...] <key> [--default|-d value] [--skip-missing-files|-m]",
		Short: "Gets a value from the key-value file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments")
			}

			files := args[:len(args)-1]
			key := args[len(args)-1]

			var foundItem *kvf.Item
			for _, file := range files {
				repo := kvf.NewFileRepo(file, *skipMissingFiles)

				currentItem, err := repo.Get(key)
				if err != nil {
					if errors.Is(err, kvf.ErrItemNotFound) {
						continue
					}
					return err
				}
				foundItem = currentItem
			}

			if foundItem != nil {
				cmd.Print(foundItem.Val)
				return nil
			}

			if cmd.Flag(defaultValFlag).Changed {
				cmd.Print(*defaultVal)
				return nil
			}

			return errors.New("key not found")
		},
	}

	defaultVal = cmd.Flags().StringP(
		defaultValFlag,
		defaultValShortFlag,
		"",
		"This value will be returned if the key is not found in the provided file(s).",
	)
	skipMissingFiles = cmd.Flags().BoolP(
		skipMissingFilesFlag,
		skipMissingFilesShortFlag,
		false,
		"Do not issue \"no such file or directory\" error on missing or inaccessible files. Should only be used"+
			" with multiple files or in combination with the \"--default\" flag.",
	)

	return cmd
}

func init() {
	rootCmd.AddCommand(newGetCmd())
}
