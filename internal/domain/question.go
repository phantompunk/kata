package domain

type QuestionStat struct {
	ID         string
	Title      string
	Difficulty string
	LangStatus map[string]bool
}
