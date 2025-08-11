package main

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

const LEETCODE_URL = "https://leetcode.com/accounts/login/"

func LoginFunc(cmd *cobra.Command, args []string) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("could not read --force flag: %v", err)
	}

	opts := app.AppOptions{Force: force}
	return kata.Login(opts)
}
