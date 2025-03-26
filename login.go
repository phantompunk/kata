package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

type SessionKey struct {
	CsrfToken    string
	SessionToken string
}

func LoginFunc(cmd *cobra.Command, args []string) error {
	kata, err := app.New()
	if err != nil {
		return err
	}

	sessionKey := SessionKey{}
	cookies := kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("LEETCODE_SESSION"))
	if len(cookies) == 0 {
		return fmt.Errorf("failed to find LEETCODE_SESSION cookie in any browser")
	}
	sessionKey.SessionToken = cookies[0].Value[32:]

	cookies = kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("csrftoken"))
	if len(cookies) == 0 {
		return fmt.Errorf("failed to find csrftoken cookie in any browser")
	}
	sessionKey.CsrfToken = cookies[0].Value[32:]

	loggedIn, err := kata.Questions.Ping(sessionKey.SessionToken, sessionKey.CsrfToken)
	if err != nil {
		return fmt.Errorf("ping err: %v", err.Error())
	}

	if !loggedIn {
		return fmt.Errorf("Session cookies are invalid trying logging in via browser-must use chrome or chromium browsers")
	}

	// :TODO:
	// open browser
	// print message then wait until keypress
	// try cookies again
	// return or error out
	// Open LC in browser
	// openbrowser(models.API_URL)

	err = kata.Config.UpdateSessionToken(sessionKey.SessionToken, sessionKey.CsrfToken)
	if err != nil {
		return fmt.Errorf("failed to update config file %v", err)
	}
	fmt.Println("Set Leetcode session cookies to setting")
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
