package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const LEETCODE_URL = "https://leetcode.com/accounts/login/"

func LoginFunc(cmd *cobra.Command, args []string) {
	if err := kata.RefreshCookies(); err != nil {
		fmt.Println("Could not authenticate using browser cookies")
		fmt.Printf("Please login manually at %s\n", LEETCODE_URL)
		return
	}

	fmt.Println("Session cookies are valid. No need to refresh.")
	// return nil
}
