package config

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/adrg/xdg"
	"github.com/go-yaml/yaml"
	"github.com/phantompunk/kata/pkg/editor"
)

var (
	ErrUnsupportedLanguage = errors.New("language not supported")
)

type ConfigService struct {
	repository ConfigRepository
	validator  ConfigValidator
}

func New() (*ConfigService, error) {
	repo, err := NewConfigRepository()
	if err != nil {
		return nil, err
	}

	return NewConfigService(*repo, *NewConfigValidator()), nil
}

func NewConfigService(repo ConfigRepository, validator ConfigValidator) *ConfigService {
	return &ConfigService{
		repository: repo,
		validator:  validator,
	}
}

func (s *ConfigService) EnsureConfig() (*Config, error) {
	exists, err := s.repository.Exists()
	if err != nil {
		return nil, err
	}

	if !exists {
		defaultCfg := s.createDefault()
		if err := s.repository.Save(defaultCfg); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return defaultCfg, nil
	}

	cfg, err := s.repository.Load()
	if err != nil {
		return nil, err
	}

	if err := s.validator.ValidateWithFallback(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (s *ConfigService) EditConfig() error {
	cfg, err := s.EnsureConfig()
	if err != nil {
		return err
	}

	backup := NewConfigBackup(cfg)
	configPath := s.repository.GetPath()
	if err := editor.Open(configPath); err != nil {
		return fmt.Errorf("failed to open config in editor: %w", err)
	}

	editedCfg, err := s.repository.Load()
	if err != nil {
		return s.handleEditFailure(backup, err)
	}

	if err := s.validator.ValidateWithFallback(editedCfg); err != nil {
		return s.handleEditFailure(backup, err)
	}

	return nil
}

func (s *ConfigService) UpdateSession(session Session) error {
	cfg, err := s.repository.Load()
	if err != nil {
		return err
	}

	cfg.Session = session
	return s.repository.Save(cfg)
}

func (s *ConfigService) SaveUsername(username string) error {
	cfg, err := s.repository.Load()
	if err != nil {
		return err
	}

	cfg.Username = username
	return s.repository.Save(cfg)
}

func (s *ConfigService) ClearSession() error {
	cfg, err := s.repository.Load()
	if err != nil {
		return err
	}

	cfg.Session = cfg.Session.Clear()
	return s.repository.Save(cfg)
}

func (s *ConfigService) GetPath() string {
	return s.repository.path
}

func (s *ConfigService) GetWarnings() []string {
	return s.validator.warnings
}

func (s *ConfigService) IsSupportedLanguage(language string) bool {
	return s.validator.IsSupportedLanguage(language)
}

func (s *ConfigService) handleEditFailure(backup *ConfigBackup, validationErr error) error {
	if err := s.repository.Restore(backup); err != nil {
		return fmt.Errorf("failed to restore config after validation error: %v; original validation error: %w", err, validationErr)
	}
	return fmt.Errorf("config validation failed: %w; changes have been reverted", validationErr)
}

func (s *ConfigService) createDefault() *Config {
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	workspace, _ := NewWorkspace(filepath.Join(homeDir, "katas"))
	language, _ := NewLanguage("python3")

	return &Config{
		workspace:    workspace,
		language:     language,
		OpenInEditor: false,
		Verbose:      false,
		Session:      Session{},
		Username:     "",
		Tracks:       []string{},
	}
}

type ConfigRepository struct {
	path string
}

func NewConfigRepository() (*ConfigRepository, error) {
	path, err := xdg.ConfigFile(filepath.Join("kata", "kata.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}
	return &ConfigRepository{path: path}, nil
}

func (r *ConfigRepository) GetPath() string {
	return r.path
}

func (r *ConfigRepository) Exists() (bool, error) {
	_, err := os.Stat(r.path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, nil
}

func (r *ConfigRepository) Load() (*Config, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func (r *ConfigRepository) Save(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(r.path), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(r.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (r *ConfigRepository) Restore(backup *ConfigBackup) error {
	return r.Save(backup.Config)
}

func (r *ConfigRepository) SaveWithTemplate(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(r.path), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tmpl := template.Must(template.New("config").Parse(configTemplate))
	file, err := os.Create(r.path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	return tmpl.Execute(file, cfg)
}

type ConfigValidator struct {
	warnings []string
}

func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{warnings: []string{}}
}

func (v *ConfigValidator) Validate(c *Config) error {
	if c.workspace == "" {
		return errors.New("workspace is not set")
	}

	if c.language == "" {
		return errors.New("language is not set")
	}

	if !v.IsSupportedLanguage(c.LanguageName()) {
		return fmt.Errorf("language %q is not supported: %w", c.LanguageName(), ErrUnsupportedLanguage)
	}

	session := c.Session
	if (session.SessionToken == "") != (session.CsrfToken == "") {
		return errors.New("both sessionToken and csrfToken must be set or unset")
	}

	return nil
}

func (v *ConfigValidator) ValidateWithFallback(c *Config) error {
	if c.workspace == "" {
		return errors.New("workspace is not set")
	}

	if c.language == "" {
		result := NewLanguageWithFallback("")
		c.language = result.Language
		v.warnings = append(v.warnings, result.Warning)
	}

	if !v.IsSupportedLanguage(c.LanguageName()) {
		result := NewLanguageWithFallback(c.LanguageName())
		v.warnings = append(v.warnings, result.Warning)
		c.language = result.Language
	}

	session := c.Session
	if (session.SessionToken == "") != (session.CsrfToken == "") {
		return errors.New("both sessionToken and csrfToken must be set or unset")
	}

	return nil
}

func (v *ConfigValidator) IsSupportedLanguage(lang string) bool {
	normalized := normalizeLanguage(lang)
	return normalized != ""
}

var supportedLanguages = map[string]string{
	"go":         "golang",
	"golang":     "golang",
	"javascript": "javascript",
	"js":         "javascript",
	"python":     "python",
	"python3":    "python",
	"py":         "python",
	"typescript": "typescript",
	"ts":         "typescript",
	// "rust":       "rust",
	// "c":          "c",
	// "csharp":     "csharp",
	// "c#":         "csharp",
	// "cpp":        "cpp",
	// "c++":        "cpp",
	// "java":       "java",
	// "ruby":       "ruby",
	// "swift":      "swift",
	// "kotlin":     "kotlin",
	// "scala":      "scala",
	// "php":        "php",
}

func normalizeLanguage(lang string) string {
	normalized, ok := supportedLanguages[strings.ToLower(strings.TrimSpace(lang))]
	if !ok {
		return ""
	}
	return normalized
}

func GetSupportedLanguages() map[string]string {
	return supportedLanguages
}

func NormalizeLanguage(language string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(language))

	canonical, ok := supportedLanguages[normalized]
	if !ok {
		return "", fmt.Errorf("language %q is not supported: %w", language, ErrUnsupportedLanguage)
	}
	return canonical, nil
}
