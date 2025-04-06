package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func DownloadFunc(cmd *cobra.Command, args []string) error {
	kata, err := app.New()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("problem")
	if err != nil || name == "" {
		cmd.Usage()
		fmt.Println()
		return fmt.Errorf("Use \"kata download --problem two-sum\" to download and stub a problem")
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil || language == "" {
		language = kata.Config.Language
	}

	open, err := cmd.Flags().GetBool("open")
	if err != nil || !open && kata.Config.OpenInEditor {
		open = kata.Config.OpenInEditor
	}

	// fetch code snippet, save question, update functionName, return
	question, err := kata.FetchQuestion(name, language)
	if err != nil && question == nil {
		return fmt.Errorf("Problem %q not found, use the correct title slug", name)
	}

	problem := question.ToProblem(kata.Config.Workspace, language)
	err = kata.StubProblem(problem)
	if err != nil {
		return err
	}

	fmt.Println("Problem stubbed at", problem.SolutionPath)
	if open {
		openWithEditor(filepath.Join(kata.Config.Workspace, problem.SolutionPath))
	}
	return nil
}

func openWithEditor(pathToFile string) error {
	textEditor := findTextEditor()

	command := exec.Command(textEditor, pathToFile)
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := os.Chdir(filepath.Dir(pathToFile))
	err = command.Run()
	if err != nil {
		return err
	}
	return nil
}

func findTextEditor() string {
	if isCommandAvailable("nvim") {
		return "nvim"
	} else if isCommandAvailable("vim") {
		return "vim"
	} else if isCommandAvailable("nano") {
		return "nano"
	} else if isCommandAvailable("editor") {
		return "editor"
	} else {
		return "vi"
	}
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
