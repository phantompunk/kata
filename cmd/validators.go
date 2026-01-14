package cmd

import (
	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/config"
	"github.com/spf13/cobra"
)

// validateLanguagePreRun validates and normalizes the language flag before command execution
func validateLanguagePreRun(kata *app.App, language *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if *language == "" {
			*language = kata.Config.LanguageName()
		}

		canonical, err := config.NormalizeLanguage(*language)
		if err != nil {
			return err
		}

		*language = canonical
		return nil
	}
}
