package editor

import (
	"os"
	"os/exec"
	"runtime"
)

func Open(path string) error {
	editor, args := findEditor()
	cmdArgs := append(args, path)

	cmd := exec.Command(editor, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func findEditor() (string, []string) {
	// 1. Check $VISUAL (GUI editors)
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor, nil
	}

	// 2. Check $EDITOR (terminal editors)
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor, nil
	}

	// 3. Platform-specific defaults
	return platformDefault()
}

func platformDefault() (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return darwinDefault()
	case "windows":
		return windowsDefault()
	default:
		return linuxDefault()
	}
}

func darwinDefault() (string, []string) {
	editors := []string{"code", "subl", "nano", "vim"}
	for _, e := range editors {
		if _, err := exec.LookPath(e); err == nil {
			return e, nil
		}
	}
	// Fallback: open with default text editor (TextEdit)
	return "open", []string{"-t"}
}

func windowsDefault() (string, []string) {
	if _, err := exec.LookPath("code"); err == nil {
		return "code", nil
	}
	return "notepad", nil
}

func linuxDefault() (string, []string) {
	editors := []string{"code", "nano", "vim", "vi"}
	for _, e := range editors {
		if _, err := exec.LookPath(e); err == nil {
			return e, nil
		}
	}
	return "vi", nil
}
