package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
)

var configTemplate string

type Config struct {
	workspace    Workspace `yaml:"workspace"`
	language     Language  `yaml:"language"`
	OpenInEditor bool      `yaml:"openInEditor"`
	Verbose      bool      `yaml:"verbose"`
	Session      Session   `yaml:"session,inline"`
	Username     string    `yaml:"username"`
	Tracks       []string  `yaml:"tracks"`
}

func (c *Config) WorkspacePath() string { return c.workspace.String() }
func (c *Config) LanguageName() string  { return c.language.String() }
func (c *Config) HasValidSession() bool { return c.Session.IsValid() }

func (c Config) MarshalYAML() (any, error) {
	return map[string]any{
		"workspace":    c.workspace.String(),
		"language":     c.language.String(),
		"openInEditor": c.OpenInEditor,
		"verbose":      c.Verbose,
		"sessionToken": c.Session.SessionToken,
		"csrfToken":    c.Session.CsrfToken,
		"username":     c.Username,
		"tracks":       c.Tracks,
	}, nil
}

func (c *Config) UnmarshalYAML(unmarshal func(any) error) error {
	var raw struct {
		Workspace    string   `yaml:"workspace"`
		Language     string   `yaml:"language"`
		OpenInEditor bool     `yaml:"openInEditor"`
		Verbose      bool     `yaml:"verbose"`
		SessionToken string   `yaml:"sessionToken"`
		CsrfToken    string   `yaml:"csrfToken"`
		Username     string   `yaml:"username"`
		Tracks       []string `yaml:"tracks"`
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	// Validate and set value objects
	ws, err := NewWorkspace(raw.Workspace)
	if err != nil {
		return err
	}

	c.workspace = ws
	c.language = Language(raw.Language)
	c.OpenInEditor = raw.OpenInEditor
	c.Verbose = raw.Verbose
	c.Session = NewSession(raw.SessionToken, raw.CsrfToken)
	c.Username = raw.Username
	c.Tracks = raw.Tracks

	return nil
}

type Workspace string

func NewWorkspace(path string) (Workspace, error) {
	if path == "" {
		return "", errors.New("workspace path cannot be empty")
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("workspace must be an absolute path, got %q", path)
	}
	return Workspace(path), nil
}

func (w Workspace) String() string {
	return string(w)
}

type Language string

const DefaultLanguage = "go"

type LanguageResult struct {
	Language Language
	Warning  string
}

func NewLanguage(lang string) (Language, error) {
	if lang == "" {
		return "", errors.New("language cannot be empty")
	}

	normalized := normalizeLanguage(lang)
	if normalized == "" {
		return "", fmt.Errorf("unsupported language: %q", lang)
	}

	return Language(normalized), nil
}

func NewLanguageWithFallback(lang string) LanguageResult {
	if lang == "" {
		return LanguageResult{
			Language: DefaultLanguage,
			Warning:  fmt.Sprintf("language was empty, using default: %s", DefaultLanguage),
		}
	}

	normalized := normalizeLanguage(lang)
	if normalized == "" {
		return LanguageResult{
			Language: DefaultLanguage,
			Warning:  fmt.Sprintf("language %q is not supported, using default: %s", lang, DefaultLanguage),
		}
	}

	return LanguageResult{
		Language: Language(normalized),
		Warning:  "",
	}
}

func (l Language) String() string {
	return string(l)
}

type Session struct {
	SessionToken string `yaml:"sessionToken"`
	CsrfToken    string `yaml:"csrfToken"`
}

func NewSession(sessionToken, csrfToken string) Session {
	return Session{
		SessionToken: sessionToken,
		CsrfToken:    csrfToken,
	}
}

func (s Session) IsValid() bool {
	return s.SessionToken != "" && s.CsrfToken != ""
}

func (s Session) Clear() Session {
	return Session{}
}

type ConfigBackup struct {
	Config *Config
}

func NewConfigBackup(cfg *Config) *ConfigBackup {
	// Deep copy to avoid mutations
	backup := &Config{
		workspace:    cfg.workspace,
		language:     cfg.language,
		OpenInEditor: cfg.OpenInEditor,
		Verbose:      cfg.Verbose,
		Session:      cfg.Session,
		Username:     cfg.Username,
		Tracks:       slices.Clone(cfg.Tracks),
	}
	return &ConfigBackup{Config: backup}
}
