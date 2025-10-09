package domain

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

type Problem struct {
	ID            string
	Title         string
	Slug          string
	DirName       string
	Content       string
	Code          string
	Difficulty    string
	FunctionName  string
	Testcases     []string
	Status        string
	LastAttempted string
	Language      Language
	DirectoryPath Path
	FileSet       []ProblemFile // Solution, Test, Readme
}

func (p *Problem) SolutionPath() string { return p.FileSet[0].Path.String() }
func (p *Problem) SolutionExists() bool { return p.FileSet[0].Path.Exists() }

func (p *Problem) GetID() int {
	id, _ := strconv.Atoi(p.ID)
	return id
}

type Language struct {
	slug          string //ts
	displayName   string //Typescript
	templateName  string //typescript
	testTemplate  string //tsjest
	fileExtension string //.ts
	testExtension string //.test.ts
}

func NewProgrammingLanguage(name string) Language {
	solution, test := resolveTemplateNames(name)
	displayName, slug, extension, testext := resolveLanguageMetadata(name)
	return Language{
		slug:          slug,
		displayName:   displayName,
		templateName:  solution,
		testTemplate:  test,
		fileExtension: extension,
		testExtension: testext,
	}
}

func (l Language) Slug() string          { return l.slug }
func (l Language) Extension() string     { return l.fileExtension }
func (l Language) DisplayName() string   { return l.displayName }
func (l Language) TemplateName() string  { return l.templateName }
func (l Language) TestTemplate() string  { return l.testTemplate }
func (l Language) TestExtension() string { return l.testExtension }

type ProblemFile struct {
	Type     FileType //solution
	Path     Path     //katas/ts/two_sum.ts
	Language Language // Typescript
}

func NewProblemFileSet(baseName string, lang Language, directory Path) []ProblemFile {
	return []ProblemFile{
		{
			Type:     SolutionFile,
			Path:     directory.Join(fmt.Sprintf("%s%s", baseName, lang.Extension())),
			Language: lang,
		},
		{
			Type:     TestFile,
			Path:     directory.Join(fmt.Sprintf("%s%s", baseName, lang.TestExtension())),
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

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

func resolveTemplateNames(slug string) (string, string) {
	switch slug {
	case "go", "golang":
		return "golang", "gotest"
	case "python", "python3":
		return "python3", "pytest"
	case "js", "javascript":
		return "javascript", "jest"
	case "ts", "typescript":
		return "typescript", "jest-ts"
	case "java":
		return "java", ""
	case "csharp", "c#":
		return "csharp", ""
	case "cpp", "c++":
		return "cpp", ""
	case "c":
		return "c", ""
	case "rust":
		return "rust", ""
	case "ruby":
		return "ruby", ""
	case "swift":
		return "swift", ""
	case "kotlin":
		return "kotlin", ""
	case "scala":
		return "scala", ""
	case "php":
		return "php", ""
	default:
		// For languages without specific templates, use generic fallbacks
		// Solution template will work, but test template may not exist
		return "solution", ""
	}
}

func resolveLanguageMetadata(slug string) (string, string, string, string) {
	switch slug {
	case "go", "golang":
		return "Go", "go", ".go", "_test.go"
	case "python", "python3":
		return "Python", "python", ".py", "_test.py"
	case "js", "javascript":
		return "JavaScript", "javascript", ".js", ".test.js"
	case "ts", "typescript":
		return "TypeScript", "typescript", ".ts", ".test.ts"
	case "rust":
		return "Rust", "rust", ".rs", ""
	case "c":
		return "C", "c", ".c", ""
	case "csharp", "c#":
		return "C#", "csharp", ".cs", ""
	case "cpp", "c++":
		return "C++", "cpp", ".cpp", ""
	case "java":
		return "Java", "java", ".java", ""
	case "ruby":
		return "Ruby", "ruby", ".rb", ""
	case "swift":
		return "Swift", "swift", ".swift", ""
	case "kotlin":
		return "Kotlin", "kotlin", ".kt", ""
	case "scala":
		return "Scala", "scala", ".scala", ""
	case "php":
		return "PHP", "php", ".php", ""
	default:
		return slug, slug, slug, slug
	}
}
