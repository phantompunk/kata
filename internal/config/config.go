package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/adrg/xdg"
	"github.com/go-yaml/yaml"
	"github.com/phantompunk/kata/internal/editor"
	"github.com/spf13/cobra"
)

var configTemplate string

var cfg Config

type Config struct {
	Workspace    string   `yaml:"workspace"`
	Language     string   `yaml:"language"`
	OpenInEditor bool     `yaml:"openInEditor"`
	Verbose      bool     `yaml:"verbose"`
	SessionToken string   `yaml:"sessionToken"`
	CsrfToken    string   `yaml:"csrfToken"`
	Username     string   `yaml:"username"`
	Tracks       []string `yaml:"tracks"`
	configPath   string
}

func ConfigFunc(cmd *cobra.Command, args []string) error {
	EnsureConfig()
	configPath, err := OpenConfig()
	if err != nil {
		return err
	}

	fmt.Println("✓ Opening config file:", configPath)
	return nil
}

func getConfigPath() (string, error) {
	return xdg.ConfigFile(filepath.Join("kata", "kata.yml"))
}

func EnsureConfig() (Config, error) {
	cfgPath, err := getConfigPath()
	if err != nil {
		return cfg, fmt.Errorf("Config error")
	}

	if dirErr := os.MkdirAll(filepath.Dir(cfgPath), os.ModePerm); dirErr != nil {
		return cfg, fmt.Errorf("Could not create config directory")
	}

	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		err = createConfigFile(cfgPath, defaultConfig())
		if err != nil {
			return cfg, fmt.Errorf("Error creating a default config")
		}
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return cfg, fmt.Errorf("Error reading config file: %w", err)
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("Could not parse config file: %w", err)
	}

	return cfg, nil
}

func createConfigFile(path string, cfg Config) error {
	tmpl := template.Must(template.New("config").Parse(configTemplate))
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, cfg)
}

func defaultConfig() Config {
	return Config{
		Workspace:    "/Users/rigo/Workspace/katas",
		Language:     "python",
		OpenInEditor: false,
		Verbose:      false,
	}
}

func OpenConfig() (string, error) {
	cfp, err := xdg.ConfigFile(filepath.Join("kata", "kata.yml"))
	if err != nil {
		return "", err
	}

	err = editor.OpenWithEditor(cfp)
	if err == nil {
		return cfp, err
	}

	return cfp, nil
}

func (c *Config) Update() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, os.ModePerm)
}

func (c *Config) UpdateSession(sessionToken, csrfToken string) error {
	c.SessionToken = sessionToken
	c.CsrfToken = csrfToken
	return c.Update()
}

func (c *Config) SaveUsername(username string) error {
	c.Username = username
	return c.Update()
}

func (c *Config) IsSessionValid() bool {
	return c.CsrfToken != "" && c.SessionToken != ""
}

func (c *Config) ClearSession() error {
	c.SessionToken = ""
	c.CsrfToken = ""
	return c.Update()
}
