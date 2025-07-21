package main

import (
	"context"
	"fmt"
	"time"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/config"
	"github.com/spf13/cobra"
)

const LEETCODE_URL = "https://leetcode.com/accounts/login/"

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

	isNewCookie := false
	if kata.Config.SessionToken == "" || kata.Config.CsrfToken == "" {
		err = refreshCookies(kata.Config)
		if err != nil {
			// fmt.Println("No session cookies found. Please log in to LeetCode using a supported browser.")
			// return fmt.Errorf("%v. Please log in at %s and try again", err, LEETCODE_URL)
			return err
		}
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

// try my key
// loggedIn, err := kata.CheckSession()
// if err != nil {
// 	fmt.Println("Tried getting cookie")
// 	return fmt.Errorf("ping err: %v. Please log in at %s and try again", err.Error(), LEETCODE_URL)
// }
//
// if loggedIn {
// 	fmt.Println("Logged in to LeetCode. Session is valid.")
// 	return nil
// }
//
// if !loggedIn {
// 	fmt.Println("failed to login")
// 	err := refreshCookies(kata.Config)
// 	if err != nil {
// 		return fmt.Errorf("%v. Please log in at %s and try again", err, LEETCODE_URL)
// 	}
// 	retry, err := kata.CheckSession()
// 	if err != nil {
// 		return fmt.Errorf("ping err: %v. Please log in at %s and try again", err.Error(), LEETCODE_URL)
// 	}
// 	if !retry {
// 		return fmt.Errorf("Session cookies are invalid. Please log in at %s using chrome or chromium browser and try again", LEETCODE_URL)
// 	}
// 	err = kata.Config.Update()
// 	if err != nil {
// 		return fmt.Errorf("failed to update config file %v", err)
// 	}
// }
//
// // :TODO:
// // open browser
// // print message then wait until keypress
// // try cookies again
// // return or error out
// // Open LC in browser
// // openbrowser(models.API_URL)
//
// if isNewCookie {
// 	fmt.Println("Set Leetcode session cookies to setting")
// 	err = kata.Config.Update()
// 	if err != nil {
// 		return fmt.Errorf("failed to update config file %v", err)
// 	}
// }
// 	return nil
// }

// :TODO: Cookie fetching logic to leetcode package
func refreshCookies(cfg *config.Config) error {
	var sessionCookie *kooky.Cookie
	var csrfCookie *kooky.Cookie

	cookiesSeq := kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(".leetcode.com"), kooky.Name("LEETCODE_SESSION")).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "LEETCODE_SESSION" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		// return fmt.Errorf("Failed to find LEETCODE_SESSION cookie in any browser.\nPlease log in at %s first", LEETCODE_URL)
		return fmt.Errorf("Failed to find LEETCODE_SESSION cookie in any browser.\nLog in at %s using a supported browser (e.g. Chrome, Chromium, Safari)", LEETCODE_URL)
	}

	cookiesSeq = kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("csrftoken")).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "csrftoken" {
			csrfCookie = cookie
			break
		}
	}
	if csrfCookie == nil {
		return fmt.Errorf("Failed to find csrftoken cookie in any browser.\nLog in at %s using a supported browser (e.g. Chrome, Chromium, Safari)", LEETCODE_URL)
	}

	fmt.Println("Session cookie expires at", sessionCookie.Expires)
	fmt.Println("Csrf cookie expires at", csrfCookie.Expires)

	cfg.CsrfToken = csrfCookie.Value
	cfg.SessionToken = sessionCookie.Value
	cfg.SessionExpires = sessionCookie.Expires
	return nil
}
