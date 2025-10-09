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
  • Test solution:    kata test %s 
  • Submit solution:  kata submit %s
`
	fmt.Printf(tmpl, slug, slug)
}
