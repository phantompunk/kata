package app

type Problem interface {
	GetSolutionPath(string, string) string
}
