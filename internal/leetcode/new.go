package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

const (
	baseUrl         = "https://leetcode.com"
	graphQLEndpoint = baseUrl + "/graphql"
)

var (
	ErrRequestFailed = errors.New("request failed")
)

type Client interface {
	// FetchQuestion fetches a question by its slug.
	FetchQuestion(ctx context.Context, slug string) (*Question, error)

	// SubmitTest(ctx context.Context, problem Problem) (string, error)
	//
	// SubmitSolution(ctx context.Context, problem Problem) (string, error)
	//
	// CheckSubmissionStatus(ctx context.Context, url string) (TestResult, error)
	//
	// WaitForSubmissionResult(ctx context.Context, url string) (TestResult, error)
	//
	// GetUserProfile(ctx context.Context) (UserProfile, error)
	//
	// GetUserStats(ctx context.Context) (UserStats, error)
	//
	// GetUserName() string
}

type LeetCodeClient struct {
	client    *http.Client
	sessionID string
	csrfToken string
}

type Options func(*LeetCodeClient)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Options {
	return func(lc *LeetCodeClient) {
		lc.client = httpClient
	}
}

// WithCookies sets the session ID and CSRF token for authentication.
func WithCookies(sessionID, csrfToken string) Options {
	return func(lc *LeetCodeClient) {
		lc.sessionID = sessionID
		lc.csrfToken = csrfToken
	}
}

// NewClient creates a new LeetCode client with the provided options.
func NewClient(httpClient *http.Client) *LeetCodeClient {
	return &LeetCodeClient{
		client: httpClient,
	}
}

func NewLC(opts ...Options) *LeetCodeClient {
	lc := &LeetCodeClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(lc)
	}

	return lc
}

func (lc *LeetCodeClient) FetchQuestion(ctx context.Context, slug string) (*Question, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := `
		query questionEditorData($titleSlug: String!) {
			question(titleSlug: $titleSlug) {
				questionId
				content
				titleSlug
				title
				difficulty
				metaData
				exampleTestcaseList
				codeSnippets {
					langSlug
					code
				}
			}
		}
	`

	variables := map[string]any{"titleSlug": slug}
	res, err := lc.graphQLRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var response QuestionReponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}

	if response.Data.Question.ID == "" {
		return nil, ErrQuestionNotFound
	}

	return &response.Data.Question, nil
}

func (lc *LeetCodeClient) graphQLRequest(ctx context.Context, query string, variables map[string]any) ([]byte, error) {
	reqBody := map[string]any{
		"query":     query,
		"variables": variables,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := lc.makeRequest(ctx, "POST", graphQLEndpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

func (lc *LeetCodeClient) makeRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("origin", "https://leetcode.com")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36") // Updated User-Agent
	req.Header.Set("x-csrftoken", lc.csrfToken)

	if lc.sessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "LEETCODE_SESSION",
			Value: lc.sessionID,
		})
	}

	if lc.csrfToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "csrftoken",
			Value: lc.csrfToken,
		})
		req.Header.Set("X-CSRFToken", lc.csrfToken)
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	return lc.client.Do(req)
}
