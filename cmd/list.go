package cmd

import (
	"github.com/spf13/cobra"
)

func ListFunc(cmd *cobra.Command, args []string) error {
	return kata.ListQuestions()
}
