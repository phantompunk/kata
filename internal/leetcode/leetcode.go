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

	_ "github.com/browserutils/kooky/browser/all"
	"github.com/phantompunk/kata/internal/models"
)

const (
	baseURL    = "https://leetcode.com/graphql"
	problemURL = "https://leetcode.com/problems/%s/"
	loginURL   = "https://leetcode.com/accounts/login/"
	checkURL   = "https://leetcode.com/submissions/detail/%s/check/"
	testURL    = "https://leetcode.com/problems/%s/interpret_solution/"
	submitURL  = "https://leetcode.com/problems/%s/submit/"

	// Maximum number of attempts to check test status
	MaxTestAttempts = 10
	// Interval between test status checks
	TestPollInterval = 500 * time.Millisecond
)

var (
	ErrQuestionNotFound = errors.New("no matching question found")
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrBuildingRequest  = errors.New("not able to build request")
)

const (
	queryUserAuth = `
	query globalData {
		userStatus {
			isSignedIn
			username
		}
	}`

	queryQuestionDetails = `
		query questionEditorData($titleSlug: String!) {
			question(titleSlug: $titleSlug) {
				questionId
				content
				titleSlug
				title
				difficulty
				exampleTestcaseList
				codeSnippets {
					langSlug
					code
				}
			}
		}`
)

// Service struct represents the LeetCode API client.
type Service struct {
	client  *http.Client
	baseUrl string
	session string
	token   string
}

// Option is a functional option type for configuring the Service.
type Option func(*Service)

// WithHTTPClient sets the HTTP client for the Service.
func WithHTTPClient(client *http.Client) Option {
	return func(s *Service) {
		s.client = client
	}
}

// WithCookies sets the session and csrf cookies for the Service.
func WithCookies(session, csrf string) Option {
	return func(s *Service) {
		s.session = session
		s.token = csrf
	}
}

// SetCookies sets the session and csrf cookies for the Service.
func (s *Service) SetCookies(session, csrf string) {
	s.session = session
	s.token = csrf
}

func New(opts ...Option) (*Service, error) {
	lcs := &Service{baseUrl: baseURL}

	for _, opt := range opts {
		opt(lcs)
	}

	if lcs.client == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create cookie jar: %w", err)
		}

		lcs.client = &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		}
	}
	return lcs, nil
}

func (s *Service) setClientCookies(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL for cookie setting: %w", err)
	}

	cookies := []*http.Cookie{
		{Name: "csrftoken", Value: s.token, Path: "/"},
		{Name: "LEETCODE_SESSION", Value: s.session, Path: "/"},
	}
	s.client.Jar.SetCookies(u, cookies)
	return nil
}

type Map map[string]string

func (lc *Service) doRequest(ctx context.Context, method, url string, body io.Reader, customHeaders Map) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("origin", "https://leetcode.com")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36") // Updated User-Agent
	req.Header.Set("x-csrftoken", lc.token)

	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	if err := lc.setClientCookies(url); err != nil {
		return nil, fmt.Errorf("failed to set cookies for request: %w", err)
	}

	res, err := lc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", res.StatusCode)
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return responseBody, nil
}

// More Auth -> are we authenticated?
func (lc *Service) Ping() (bool, error) {
	data, err := json.Marshal(Request{Query: queryUserAuth})
	if err != nil {
		return false, err
	}

	body, err := lc.doRequest(context.Background(), "POST", lc.baseUrl, bytes.NewBuffer(data), nil)
	if err != nil {
		return false, fmt.Errorf("failed to do request: %w", err)
	}

	var response AuthResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Data.UserStatus.IsSignedIn, nil
}

func (lc *Service) GetUsername() (string, error) {
	data, err := json.Marshal(Request{Query: queryUserAuth})
	if err != nil {
		return "", err
	}

	body, err := lc.doRequest(context.Background(), "POST", lc.baseUrl, bytes.NewBuffer(data), nil)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}

	var response AuthResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Data.UserStatus.Username, nil
}

func (lc *Service) Fetch(name string) (*models.Question, error) {
	variables := map[string]any{"titleSlug": name}
	data, err := json.Marshal(Request{Query: queryQuestionDetails, Variables: variables})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	body, err := lc.doRequest(context.Background(), "POST", lc.baseUrl, bytes.NewReader(data), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Data.Question == nil {
		return nil, ErrQuestionNotFound
	}

	return response.Data.Question, nil
}

func (lc *Service) Test(problem *models.Problem, language, snippet string) (string, error) {
	url := fmt.Sprintf(testURL, problem.TitleSlug)
	contents := strings.ReplaceAll(snippet, "\t", "    ") // Consistent 4 spaces

	variables := map[string]any{
		"lang":        models.LangName[language],
		"question_id": problem.QuestionID,
		"typed_code":  contents,
		"data_input":  problem.TestCases,
	}

	data, err := json.Marshal(variables)
	if err != nil {
		return "", fmt.Errorf("failed to marshal test request: %w", err)
	}

	headers := map[string]string{
		"referer": fmt.Sprintf(problemURL, problem.TitleSlug),
	}

	fmt.Printf("Testing on server")
	body, err := lc.doRequest(context.Background(), http.MethodPost, url, bytes.NewReader(data), headers)
	if err != nil {
		return "", fmt.Errorf("test submission failed: %w", err)
	}

	var response TestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal test response: %w", err)
	}

	if response.InterpretID == "" {
		return "", errors.New("interpret_id not found in test response")
	}

	return fmt.Sprintf(checkURL, response.InterpretID), nil
}

func (lc *Service) Submit(problem *models.Problem, language, snippet string) (string, error) {
	url := fmt.Sprintf(submitURL, problem.TitleSlug)
	contents := strings.ReplaceAll(snippet, "\t", "    ") // Consistent 4 spaces

	variables := map[string]any{
		"lang":        models.LangName[language],
		"question_id": problem.QuestionID,
		"typed_code":  contents,
	}

	data, err := json.Marshal(variables)
	if err != nil {
		return "", fmt.Errorf("failed to marshal test request: %w", err)
	}

	headers := map[string]string{
		"referer": fmt.Sprintf(problemURL, problem.TitleSlug),
	}

	fmt.Printf("✓ Submitting solution to Leetcode")
	body, err := lc.doRequest(context.Background(), http.MethodPost, url, bytes.NewReader(data), headers)
	if err != nil {
		return "", fmt.Errorf("test submission failed: %w", err)
	}

	var response TestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal test response: %w", err)
	}

	if response.SubmissionID == "" {
		return "", errors.New("submission_id not found in test response")
	}

	return fmt.Sprintf(checkURL, response.SubmissionID), nil
}

func (lc *Service) CheckTestStatus(callbackUrl string) (*TestResponse, error) {
	// fmt.Printf("Checking test status at %s\n", callbackUrl)
	body, err := lc.doRequest(context.Background(), "GET", callbackUrl, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check test status: %w", err)
	}
	// fmt.Printf("Response check body: %s\n", string(body))

	var response TestResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal test status response: %w", err)
	}

	return &response, nil
}

func (lc *Service) PollTestStatus(testStatusUrl string) (*TestResponse, error) {
	for i := range MaxTestAttempts {
		res, err := lc.CheckTestStatus(testStatusUrl)
		if err != nil {
			return nil, fmt.Errorf("checking test status attempt %d failed: %w", i+1, err)
		}
		// fmt.Println("Test status:", res.State)

		if res.State == "SUCCESS" || res.State == "FAILED" {
			return res, nil
		}

		fmt.Print(".")
		time.Sleep(TestPollInterval)
	}

	return nil, fmt.Errorf("test status check timed out after %d attempts", MaxTestAttempts)
}
