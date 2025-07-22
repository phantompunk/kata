package main

import (
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/spf13/cobra"
)

const LEETCODE_URL = "https://leetcode.com/accounts/login/"

func LoginFunc(cmd *cobra.Command, args []string) error {
	kata, err := app.New()
	if err != nil {
		return err
	}

	isNewCookie := false
	if kata.Config.SessionToken == "" || kata.Config.CsrfToken == "" {
		sessionToken, csrfToken, expires, err := leetcode.RefreshCookies()
		if err != nil {
			return err
		}
		kata.UpdateConfig(sessionToken, csrfToken, expires)
		isNewCookie = true
	}

	isValid, err := kata.CheckSession()
	if err != nil {
		return fmt.Errorf("Error: %v\nPlease log in at %s", err.Error(), LEETCODE_URL)
	}

	if !isValid {
		return fmt.Errorf("Session cookies are invalid. Please log in at %s using chrome or chromium browser and try again", LEETCODE_URL)
	}

	if isNewCookie {
		err = kata.Config.Update()
		if err != nil {
			return fmt.Errorf("failed to update config file %v", err)
		}
	}
	fmt.Println("Successfully logged in to LeetCode. Session is valid.")

	return nil
}
