package app

import (
	"context"
	"testing"
)

type mockService struct{}

func (m *mockService) GetQuizQuestion(ctx context.Context) string { return "Two Sum" }

func TestQuizQuestion(t *testing.T) {
	service := &mockService{}
	question := service.GetQuizQuestion(context.Background())

	expected := "Two Sum"
	if question != expected {
		t.Errorf("Expected question %q, but got %q", expected, question)
	}
}
