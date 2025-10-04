package domain

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Problem struct {
	Title         string
	Slug          string
	Content       string
	Code          string
	FunctionName  string
	Language      Language
	DirectoryPath Path
	FileSet       []ProblemFile // Solution, Test, Readme
}

func (p *Problem) SolutionPath() string { return p.FileSet[0].Path.String() }

type Language struct {
	slug          string //ts
	displayName   string //typescript
	templateName  string //typescript
	testTemplate  string //tsjest
	fileExtension string //.ts
}

func NewProgrammingLanguage(slug string) Language {
	solution, test := resolveTemplateNames(slug)
	displayName, extension := resolveLanguageMetadata(slug)
	return Language{
		slug:          slug,
		displayName:   displayName,
		templateName:  solution,
		testTemplate:  test,
		fileExtension: extension,
	}
}

func (l Language) Slug() string         { return l.slug }
func (l Language) Extension() string    { return l.fileExtension }
func (l Language) DisplayName() string  { return l.displayName }
func (l Language) TemplateName() string { return l.templateName }
func (l Language) TestTemplate() string { return l.testTemplate }

type ProblemFile struct {
	Type     FileType //solution
	Path     Path     //katas/ts/two_sum.ts
	Language Language // Typescript
}

func NewProblemFileSet(slug string, lang Language, directory Path) []ProblemFile {
	return []ProblemFile{
		{
			Type:     SolutionFile,
			Path:     directory.Join(fmt.Sprintf("%s%s", slug, lang.Extension())),
			Language: lang,
		},
		{
			Type:     TestFile,
			Path:     directory.Join(fmt.Sprintf("%s_test%s", slug, lang.Extension())),
			Language: lang,
		},
		{
			Type:     ReadmeFile,
			Path:     directory.Join("readme.md"),
			Language: lang,
		},
	}
}

type Path string

func NewDirectoryPath(path string) (Path, error) {
	if path == "" {
		return "", fmt.Errorf("directory cannot be empty")
	}
	return Path(path), nil
}

func NewFilePath(path string) (Path, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	return Path(path), nil
}

func (f Path) String() string {
	return string(f)
}

func (f Path) Basename() string {
	return filepath.Base(string(f))
}

func (f Path) Exists() bool {
	_, err := os.Stat(string(f))
	return err == nil
}

func (f Path) Dir() string {
	return filepath.Dir(string(f))
}

func (p Path) Join(parts ...string) Path {
	allParts := append([]string{string(p)}, parts...)
	return Path(filepath.Join(allParts...))
}

func (p Path) DisplayPath() string {
	path := string(p)
	if usr, err := user.Current(); err == nil {
		homeDir := usr.HomeDir
		if strings.HasPrefix(path, homeDir) {
			return "~" + path[len(homeDir):]
		}
	}
	return path
}

type FileType string

const (
	SolutionFile FileType = "solution"
	TestFile     FileType = "test"
	ReadmeFile   FileType = "readme"
)

func resolveTemplateNames(slug string) (string, string) {
	switch slug {
	case "go", "golang":
		return "golang", "gotest"
	case "python", "python3":
		return "python", "pytest"
	// case "js", "javascript":
	// 	return "javascript", "jest"
	// case "ts", "typescript":
	// 	return "typescript", "tsjest"
	default:
		return "solution", "test"
	}
}

func resolveLanguageMetadata(slug string) (string, string) {
	switch slug {
	case "go", "golang":
		return "Go", ".go"
	case "python", "python3":
		return "Python", ".py"
	case "js", "javascript":
		return "JavaScript", ".js"
	case "ts", "typescript":
		return "TypeScript", ".ts"
	case "rust":
		return "Rust", ".rs"
	case "cpp", "c++":
		return "C++", ".cpp"
	default:
		return slug, slug
	}
}
