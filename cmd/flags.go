package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func checkDateFlags(cmd *cobra.Command) error {

	fromFlag, err := cmd.Flags().GetTime("from")
	if err != nil {
		return err
	}

	toFlag, err := cmd.Flags().GetTime("to")
	if err != nil {
		return err
	}

	if fromFlag.After(toFlag) {
		return fmt.Errorf("flag values `from` may not be after `to`")
	}

	return nil
}
