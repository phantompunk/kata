package render

import (
	"fmt"

	"github.com/phantompunk/kata/internal/domain"
)

type RenderStrategy int

const (
	SkipExisting RenderStrategy = iota
	ForceOverwrite
)

type RenderResult struct {
	DirectoryCreated string
	FilesCreated     []string
	FilesUpdated     []string
	FilesSkipped     []string
	TestSupported    bool
}

func NewRenderResult() *RenderResult {
	return &RenderResult{
		FilesCreated: []string{},
		FilesUpdated: []string{},
		FilesSkipped: []string{},
	}
}

func (r *RenderResult) RecordDirectoryCreated(path domain.Path) {
	r.DirectoryCreated = path.DisplayPath()
}

func (r *RenderResult) RecordFileCreated(path domain.Path) {
	r.FilesCreated = append(r.FilesCreated, path.Basename())
}

func (r *RenderResult) RecordFileUpdated(path domain.Path) {
	r.FilesUpdated = append(r.FilesUpdated, path.Basename())
}

func (r *RenderResult) RecordFileSkipped(path domain.Path) {
	r.FilesSkipped = append(r.FilesSkipped, path.Basename())
}

func (r *RenderResult) RecordAllSkipped() {
	r.FilesSkipped = append(r.FilesSkipped, "All Files")
}

func (r *RenderResult) RecordTestUnsupported() {
	r.TestSupported = false
}

type DirectoryPath string

func NewDirectoryPath(path string) (DirectoryPath, error) {
	if path == "" {
		return "", fmt.Errorf("directory path cannot be empty")
	}
	return DirectoryPath(path), nil
}

func (d DirectoryPath) String() string {
	return string(d)
}
