package main

import (
	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func QuizFunc(cmd *cobra.Command, args []string) error {
	return kata.Quiz(app.AppOptions{})
}
