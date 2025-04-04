package leetcode

import (
	"bytes"
	"io"
	"net/http"
)

// mockRoundTripper implements http.RoundTripper for testing.
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req), nil
}

func mockResponse(response string, client *http.Client) *http.Client {
	client.Transport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(response))),
				Header:     make(http.Header),
			}
		},
	}
	return client
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
