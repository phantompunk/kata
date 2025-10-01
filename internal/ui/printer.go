package ui

import "fmt"

func PrintSuccess(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println("✔", msg)
}

func PrintInfo(message string) {
	fmt.Println("ℹ", message)
}

func PrintError(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println("✘", msg)
}

func Print(message string) {
	fmt.Println(message)
}

func PrintWarning(message string) {
	fmt.Println("⚠", message)
}

func PrintNextSteps(slug string) {
	tmpl := `
Next steps:
  • Start solving: kata solve %s
  • View details: kata show %s
  • Submit later: kata submit %s
`
	fmt.Printf(tmpl, slug, slug, slug)
}
