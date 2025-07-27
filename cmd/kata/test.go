package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func TestFunc(cmd *cobra.Command, args []string) error {

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

	status, err := kata.TestSolution(name, language)
	if status == "" {
		return fmt.Errorf("failed to submit test %v", err.Error())
	}

	fmt.Println(status)
	return nil
}
