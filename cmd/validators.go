package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/ui"
)

func validateLanguage() (string, error) {
	if language == "" {
		language = kata.Config.LanguageName()
	}

	if !kata.Settings.IsSupportedLanguage(language) {
		ui.PrintError("language %q not supported", language)
		return language, fmt.Errorf("language %q is not supported", language)
	}

	return language, nil
}
