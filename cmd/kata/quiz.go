package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func QuizFunc(cmd *cobra.Command, args []string) error {
	question, err := kata.Questions.GetRandom()
	if err != nil {
		return err
	}
	problem := question.ToProblem(kata.Config.Workspace, kata.Config.Language)

	fmt.Println("Problem stubbed at", problem.SolutionPath)
	if kata.Config.OpenInEditor {
		openWithEditor(problem.SolutionPath)
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
