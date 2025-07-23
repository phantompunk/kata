package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func SubmitFunc(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("problem")
	if err != nil || name == "" {
		cmd.Usage()
		fmt.Println()
		return fmt.Errorf("Use \"kata download --problem two-sum\" to download and stub a problem")
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil || language == "" {
		language = kata.Config.Language
	}

	fmt.Printf("Submitting solution for %s for %s\n", name, language)

	return nil
}
