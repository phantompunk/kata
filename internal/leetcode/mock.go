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

func mockResponse(status int, response string, client *http.Client) *http.Client {
	client.Jar, _ = cookiejar.New(nil)
	client.Transport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(bytes.NewReader([]byte(response))),
				Header:     make(http.Header),
				Request:    req,
			}
		},
	}
	return client
}

func newTestService() *Service {
	cookiejar, _ := cookiejar.New(nil)
	mockClient := &http.Client{
		Jar:       cookiejar,
		Transport: &mockRoundTripper{},
	}
	session, csrf := "abc123", "csrf123"
	lc, _ := New(WithHTTPClient2(mockClient), WithCookies2(session, csrf))
	return lc
}

type Responder struct {
	Status int
	Body   string
}

func (r *Responder) SetResponse(status int, body string) {
	r.Status = status
	r.Body = body
}

func (r *Responder) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: r.Status,
		Body:       io.NopCloser(bytes.NewReader([]byte(r.Body))),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func newTestClient(responder *Responder) *LeetCodeClient {
	client := &http.Client{}
	client.Transport = responder
	return NewClient(client)
}
