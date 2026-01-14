package cmd

import (
	"github.com/phantompunk/kata/internal/config"
	"github.com/spf13/cobra"
)

// validateLanguagePreRun validates and normalizes the language flag before command execution
func validateLanguagePreRun(cmd *cobra.Command, args []string) error {
	return validateLanguage()
}

// validateLanguage validates and normalizes the language global variable
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
