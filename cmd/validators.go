package cmd

import (
	"github.com/phantompunk/kata/internal/config"
)

func validateLanguage() error {
	if language == "" {
		language = kata.Config.LanguageName()
	}

	canonical, err := config.NormalizeLanguage(language)
	if err != nil {
		return err
	}

	language = canonical
	return nil
}
