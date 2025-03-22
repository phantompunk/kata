package mocks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/phantompunk/kata/internal/models"
)

// mockRoundTripper implements http.RoundTripper for testing.
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req), nil
}

func newMockClient(response string) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(response))),
					Header:     make(http.Header),
				}
			},
		},
	}
}

type MockQuestionModel struct{}

func (m *MockQuestionModel) GetBySlug(name, language string) (*models.Question, error) {
	return &models.Question{}, fmt.Errorf("no records found")
}
