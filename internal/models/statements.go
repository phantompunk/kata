package models

var gQLQueryQuestion string = `query questionEditorData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    content
    titleSlug
    title
    difficulty
    codeSnippets {
      langSlug
      code
    }
  }
}`

var gQLQueryStreak string = `query getStreakCounter {
  streakCounter {
    currentDayCompleted
    daysSkipped
  }
}`




