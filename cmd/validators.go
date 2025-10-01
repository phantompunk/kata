package cmd

import (
	"fmt"

	"github.com/phantompunk/kata/internal/config"
)

func ValidateLanguage(language string) error {
	validator := config.NewConfigValidator()
	if !validator.IsSupportedLanguage(language) {
		return fmt.Errorf("language %q is not supported", language)
	}
	return nil
}
