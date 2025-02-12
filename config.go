package main

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

//go:embed config_template.yml
var configTemplate string

var cfg Config

type Config struct {
	Workspace  string `yaml:"workspace"`
	configPath string
}

func ConfigFunc(cmd *cobra.Command, args []string) error {
	ensureConfig()
	configPath, err := OpenConfig()
	if err != nil {
		return err
	}

	fmt.Println("Wrote config file to:", configPath)
	return nil
}

func ensureConfig() (Config, error) {
	cfgPath, err := xdg.ConfigFile(filepath.Join("kata", "kata.yml"))
	if err != nil {
		fmt.Errorf("Config error")
		return cfg, err
	}
	cfg.configPath = cfgPath

	cfgDir := filepath.Dir(cfgPath)
	if dirErr := os.MkdirAll(cfgDir, os.ModePerm); dirErr != nil {
		fmt.Errorf("Could not create config directory")
		return cfg, err
	}

	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		err = writeConfigFile(cfgPath)
		if err != nil {
		}
	}

	cf, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Errorf("Error reading config file: %w", err)
	}
	err = yaml.Unmarshal(cf, &cfg)
	if err != nil {
		fmt.Errorf("Could not parse config file: %w", err)
	}
	return cfg, nil
}
func writeConfigFile(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return createConfigFile(path)
	} else if err != nil {
		return err
	}
	return nil
}

func createConfigFile(path string) error {
	tmpl := template.Must(template.New("config").Parse(configTemplate))
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()
	m := struct {
		Config Config
	}{Config: defaultConfig()}

	if err = tmpl.Execute(file, m); err != nil {
		return err
	}
	return nil
}

func defaultConfig() Config {
	return Config{
		Workspace: "/Users/rigo/Workspace/katas",
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
