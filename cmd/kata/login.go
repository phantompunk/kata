package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const LEETCODE_URL = "https://leetcode.com/accounts/login/"

func LoginFunc(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	language, err := cmd.Flags().GetString("language")
	if err != nil || language == "" {
		language = kata.Config.Language
	}

	if kata.Config.IsSessionValid() && !force {
		fmt.Println("You are already logged in")
		return nil
	}

	if err := kata.RefreshCookies(); err != nil {
		return fmt.Errorf("Could not authenticate using browser cookies: %v\nPlease login manually at %s", err, LEETCODE_URL)
	}

	if !kata.CheckSession() {
		return fmt.Errorf("Session cookies are invalid.\nPlease login manually at %s", LEETCODE_URL)
	}

	fmt.Println("Successfully logged in using browser cookies.")
	return nil
}
