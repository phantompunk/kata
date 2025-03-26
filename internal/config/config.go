package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/adrg/xdg"
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
)

var configTemplate string

var cfg Config

type Config struct {
	Workspace    string   `yaml:"workspace"`
	Language     string   `yaml:"language"`
	OpenInEditor bool     `yaml:"openInEditor"`
	SessionToken string   `yaml:"sessionToken"`
	CsrfToken    string   `yaml:"csrfToken"`
	Tracks       []string `yaml:"tracks"`
	configPath   string
}

func ConfigFunc(cmd *cobra.Command, args []string) error {
	EnsureConfig()
	configPath, err := OpenConfig()
	if err != nil {
		return err
	}

	fmt.Println("Wrote config file to:", configPath)
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
	}
}

func findEditor() string {
	term, found := os.LookupEnv("EDITOR")
	if found && isCmdAvailable(term) {
		return term
	} else if isCmdAvailable("nvim") {
		return "nvim"
	} else if isCmdAvailable("vim") {
		return "vim"
	} else if isCmdAvailable("vi") {
		return "vi"
	} else {
		return "nano"
	}
}

func isCmdAvailable(name string) bool {
	cmd := exec.Command("command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func OpenConfig() (string, error) {
	cfp, err := xdg.ConfigFile(filepath.Join("kata", "kata.yml"))
	if err != nil {
		return "", err
	}

	editor := findEditor()
	command := exec.Command(editor, cfp)
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		return "", err
	}
	return cfp, nil
}

func (c *Config) UpdateSessionToken(sessionToken, csrfToken string) error {
	c.SessionToken = sessionToken
	c.CsrfToken = csrfToken

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	cfgPath, _ := getConfigPath()
	err = os.WriteFile(cfgPath, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
