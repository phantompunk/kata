package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/config"
	"github.com/spf13/cobra"
)

const LEETCODE_URL = "https://leetcode.com"

type SessionKey struct {
	CsrfToken    string
	SessionToken string
	Expires      time.Time
}

func LoginFunc(cmd *cobra.Command, args []string) error {
	kata, err := app.New()
	if err != nil {
		return err
	}

	// useOrReplaceCookies
	// if tokens are missing
	// -> Please login using supported browser
	// if tokens are found
	// -> Validate they are not expired
	//    if expired
	//	  -> Please login using supported browser
	//	  else validate
	//		    if not valid
	//				-> Please login using supported browser

	var isNewCookie bool
	if kata.Config.SessionToken == "" || kata.Config.CsrfToken == "" {
		// if either empty fetch fresh tokens
		err := refreshCookies(kata.Config)
		if err != nil {
			return fmt.Errorf("%v. Please log in at %s and try again", err, LEETCODE_URL)
		}
		isNewCookie = true
	}

	// try my key
	loggedIn, err := kata.CheckSession()
	if err != nil {
		return fmt.Errorf("ping err: %v. Please log in at %s and try again", err.Error(), LEETCODE_URL)
	}

	if loggedIn {
		fmt.Println("Logged in to LeetCode. Session is valid.")
		return nil
	}

	if !loggedIn {
		err := refreshCookies(kata.Config)
		if err != nil {
			return fmt.Errorf("%v. Please log in at %s and try again", err, LEETCODE_URL)
		}
		retry, err := kata.CheckSession()
		if err != nil {
			return fmt.Errorf("ping err: %v. Please log in at %s and try again", err.Error(), LEETCODE_URL)
		}
		if !retry {
			return fmt.Errorf("Session cookies are invalid. Please log in at %s using chrome or chromium browser and try again", LEETCODE_URL)
		}
		err = kata.Config.Update()
		if err != nil {
			return fmt.Errorf("failed to update config file %v", err)
		}
	}

	// :TODO:
	// open browser
	// print message then wait until keypress
	// try cookies again
	// return or error out
	// Open LC in browser
	// openbrowser(models.API_URL)

	if isNewCookie {
		fmt.Println("Set Leetcode session cookies to setting")
		err = kata.Config.Update()
		if err != nil {
			return fmt.Errorf("failed to update config file %v", err)
		}
	}
	return nil
}

// :TODO: Cookie fetching logic to leetcode package
func refreshCookies(cfg *config.Config) error {
	cookies := kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("LEETCODE_SESSION"))
	if len(cookies) == 0 {
		return fmt.Errorf("failed to find LEETCODE_SESSION cookie in any browser. Please log in at %s first", LEETCODE_URL)
	}
	cfg.SessionToken = cookies[0].Value[32:]

	cookies = kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("csrftoken"))
	if len(cookies) == 0 {
		return fmt.Errorf("failed to find csrftoken cookie in any browser. Please log in at %s first", LEETCODE_URL)
	}
	cfg.CsrfToken = cookies[0].Value[32:]
	return nil
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println("Fail to open browser")
	}
}
