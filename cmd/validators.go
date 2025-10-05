package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/ui"
)

func validateL() (string, error) {
	if language == "" {
		language = kata.Config.LanguageName()
	}

	if !kata.Settings.IsSupportedLanguage(language) {
		ui.PrintError("language %q not supported", language)
		return language, fmt.Errorf("language %q is not supported", language)
	}

	return language, nil
}

func validateLanguage() error {
	if language == "" {
		language = kata.Config.LanguageName()
	}

	if !kata.Settings.IsSupportedLanguage(language) {
		return fmt.Errorf("language %q is not supported", language)
	}

	return nil
}
