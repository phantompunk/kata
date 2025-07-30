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

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
	"github.com/phantompunk/kata/internal/models"
)

const baseURL = "https://leetcode.com/graphql"
const loginURL = "https://leetcode.com/accounts/login/"
const checkURL = "https://leetcode.com/submissions/detail/%s/check/"
const testURL = "https://leetcode.com/problems/%s/interpret_solution/"

var (
	ErrNotFound = errors.New("no matching question found")
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
			Jar: jar,
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

func (lc *Service) doRequest(method, url string, body []byte, customHeaders map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
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

	fmt.Println("URL:", url)

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

// RefreshCookies fetches the session and csrf cookies from the browser.
func RefreshCookies() (string, string, time.Time, error) {
	var sessionCookie *kooky.Cookie
	var csrfCookie *kooky.Cookie

	cookiesSeq := kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(".leetcode.com"), kooky.Name("LEETCODE_SESSION")).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "LEETCODE_SESSION" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		return "", "", time.Time{}, fmt.Errorf("Failed to find LEETCODE_SESSION cookie in any browser.\nLog in at %s using a supported browser (e.g. Chrome, Chromium, Safari)", loginURL)
	}

	cookiesSeq = kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("csrftoken")).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "csrftoken" {
			csrfCookie = cookie
			break
		}
	}
	if csrfCookie == nil {
		return "", "", time.Time{}, fmt.Errorf("Failed to find csrftoken cookie in any browser.\nLog in at %s using a supported browser (e.g. Chrome, Chromium, Safari)", loginURL)
	}

	fmt.Println("Session cookie expires at", sessionCookie.Expires)
	fmt.Println("Csrf cookie expires at", csrfCookie.Expires)

	return sessionCookie.Value, csrfCookie.Value, sessionCookie.Expires, nil
}

// More Auth -> are we authenticated?
func (lc *Service) Ping() (bool, error) {
	data, err := json.Marshal(models.Request{Query: queryUserStreak})
	if err != nil {
		return false, fmt.Errorf("failed to marshal request data: %w", err)
	}

	headers := map[string]string{
		"referer": "https://leetcode.com/problemset/",
	}

	body, err := lc.doRequest("POST", lc.baseUrl, data, headers)
	if err != nil {
		return false, fmt.Errorf("failed to do request: %w", err)
	}

	var response models.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Data.StreakCounter != nil, nil
}

var queryQuestionDetails string = `query questionEditorData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    content
    titleSlug
    title
    difficulty
    codeSnippets {
      langSlug
      code
    }
  }
}`

func (lc *Service) Fetch(name string) (*models.Question, error) {
	variables := map[string]any{"titleSlug": name}
	data, err := json.Marshal(models.Request{Query: queryQuestionDetails, Variables: variables})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	body, err := lc.doRequest("POST", lc.baseUrl, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	var response models.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Data.Question == nil {
		return nil, ErrNotFound
	}

	return response.Data.Question, nil
}

func (lc *Service) Test(question *models.Question, language, snippet string) (string, error) {
	url := fmt.Sprintf(testURL, question.TitleSlug)
	contents := strings.ReplaceAll(snippet, "\t", "    ") // Consistent 4 spaces

	// The data_input is hardcoded here. In a real scenario, you'd want to fetch
	// the test cases for the question or allow the user to provide them.
	variables := map[string]any{
		"lang":        models.LangName[language],
		"question_id": question.ID,
		"typed_code":  contents,
		"data_input":  "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6", // :TODO: Make dynamic
	}

	data, err := json.Marshal(variables)
	if err != nil {
		return "", fmt.Errorf("failed to marshal test request: %w", err)
	}

	headers := map[string]string{
		"referer": fmt.Sprintf("https://leetcode.com/problemset/%s/description", question.TitleSlug),
	}

	body, err := lc.doRequest(http.MethodPost, url, data, headers)
	if err != nil {
		return "", fmt.Errorf("test submission failed: %w", err)
	}

	var response models.TestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal test response: %w", err)
	}

	if response.InterpretID == "" {
		return "", errors.New("interpret_id not found in test response")
	}

	return fmt.Sprintf(checkURL, response.InterpretID), nil
}

func (lc *Service) CheckTestStatus(callbackUrl string) (*models.TestResponse, error) {
	headers := map[string]string{"referer": "https://leetcode.com/problemset/"}
	fmt.Println("Callback URL:", callbackUrl)
	body, err := lc.doRequest("GET", callbackUrl, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to check test status: %w", err)
	}

	var response models.TestResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal test status response: %w", err)
	}

	return &response, nil
}
