package leetcode

import (
	"bytes"
	"io"
	"net/http"
	"net/http/cookiejar"
)

// mockRoundTripper implements http.RoundTripper for testing.
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req), nil
}

func mockResponse(response string, client *http.Client) *http.Client {
	client.Jar, _ = cookiejar.New(nil)
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

func newTestService() *Service {
	cookiejar, _ := cookiejar.New(nil)
	mockClient := &http.Client{Jar: cookiejar}
	session, csrftoken := "abc123", "csrf123"
	lc, _ := New(WithHTTPClient(mockClient), WithCookies(session, csrftoken))
	return lc
}
// func newMockClient(response string) *http.Client {
// 	return &http.Client{
// 		Transport: &mockRoundTripper{
// 			roundTripFunc: func(req *http.Request) *http.Response {
// 				return &http.Response{
// 					StatusCode: http.StatusOK,
// 					Body:       io.NopCloser(bytes.NewReader([]byte(response))),
// 					Header:     make(http.Header),
// 				}
// 			},
// 		},
// 	}
// }
