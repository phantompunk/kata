package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/domain"
)

const (
	baseUrl            = "https://leetcode.com"
	graphQLEndpoint    = baseUrl + "/graphql/"
	testEndpoint       = baseUrl + "/problems/%s/interpret_solution/"
	submitEndpoint     = baseUrl + "/problems/%s/submit/"
	submissionEndpoint = baseUrl + "/submissions/detail/%s/check/"
)

var (
	ErrRequestFailed = errors.New("request failed")
)

type Client interface {
	// FetchQuestion fetches a question by its slug.
	FetchQuestion(ctx context.Context, slug string) (*Question, error)

	SubmitTest(ctx context.Context, problem *domain.Problem, snippet string) (string, error)
	SubmitSolution(ctx context.Context, problem *domain.Problem, snippet string) (string, error)
	CheckSubmissionResult(ctx context.Context, url string) (*SubmissionResult, error)

	// GetUserProfile(ctx context.Context) (UserProfile, error)
	// GetUserStats(ctx context.Context) (UserStats, error)
	// GetUserName() string
}

type SubmissionResult struct {
	State      string
	Answer     bool
	Result     string
	Runtime    string
	RuntimeMsg string
	Memory     string
	MemoryMsg  string
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

func WithSession(session config.Session) Options {
	return func(lc *LeetCodeClient) {
		lc.sessionID = session.SessionToken
		lc.csrfToken = session.CsrfToken
	}
}

// NewClient creates a new LeetCode client with the provided options.
func NewClient(httpClient *http.Client) *LeetCodeClient {
	return &LeetCodeClient{
		client: httpClient,
	}
}

func NewLC(opts ...Options) *LeetCodeClient {
	jar, _ := cookiejar.New(nil)
	lc := &LeetCodeClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
	}

	for _, opt := range opts {
		opt(lc)
	}

	lc.initialize()
	return lc
}

func (lc *LeetCodeClient) initialize() {
	if lc.sessionID == "" || lc.csrfToken == "" {
		return
	}

	cookies := []*http.Cookie{
		{Name: "csrftoken", Value: lc.csrfToken, Path: "/"},
		{Name: "LEETCODE_SESSION", Value: lc.sessionID, Path: "/"},
	}

	u, _ := url.Parse(baseUrl)
	lc.client.Jar.SetCookies(u, cookies)
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
	res, err := lc.graphQLRequest(ctx, query, variables, nil)
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

func (lc *LeetCodeClient) SubmitTest(ctx context.Context, problem *domain.Problem, snippet string) (string, error) {
	payload := map[string]any{
		"lang":        problem.Language.TemplateName(),
		"question_id": problem.ID,
		"typed_code":  strings.ReplaceAll(snippet, "\t", "    "), // Consistent 4 spaces
		"data_input":  problem.Testcases,
	}

	url := fmt.Sprintf(testEndpoint, problem.Slug)
	res, err := lc.Submit(ctx, url, problem, payload)
	if err != nil {
		return "", err
	}

	return res.InterpretID, nil
}

func (lc *LeetCodeClient) SubmitSolution(ctx context.Context, problem *domain.Problem, snippet string) (string, error) {
	payload := map[string]any{
		"lang":        problem.Language.TemplateName(),
		"question_id": problem.ID,
		"typed_code":  strings.ReplaceAll(snippet, "\t", "    "), // Consistent 4 spaces
	}

	url := fmt.Sprintf(submitEndpoint, problem.Slug)
	res, err := lc.Submit(ctx, url, problem, payload)
	if err != nil {
		return "", err
	}

	return res.GetSubmissionID(), nil
}

func (lc *LeetCodeClient) CheckSubmissionResult(ctx context.Context, submissionId string) (*SubmissionResult, error) {
	url := fmt.Sprintf(submissionEndpoint, submissionId)
	resp, err := lc.makeRequest(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: handle errors
	if resp.StatusCode != http.StatusOK {
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response SubmissionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal submission response: %w", err)
	}

	return response.ToResult(), nil
}

func (lc *LeetCodeClient) Submit(ctx context.Context, url string, problem *domain.Problem, payload map[string]any) (*SubmitResponse, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"referer": fmt.Sprintf(problemURL, problem.Slug),
	}

	resp, err := lc.makeRequest(ctx, "POST", url, bytes.NewBuffer(data), headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response SubmitResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test response: %w", err)
	}

	return &response, nil
}

func (lc *LeetCodeClient) graphQLRequest(ctx context.Context, query string, variables map[string]any, headers Map) ([]byte, error) {
	reqBody := map[string]any{
		"query":     query,
		"variables": variables,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := lc.makeRequest(ctx, "POST", graphQLEndpoint, bytes.NewBuffer(data), headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

func (lc *LeetCodeClient) makeRequest(ctx context.Context, method, Surl string, body io.Reader, headers Map) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, Surl, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36") // Updated User-Agent
	req.Header.Set("Origin", "https://leetcode.com")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("X-Csrftoken", lc.csrfToken)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return lc.client.Do(req)
}
