package leetcode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/phantompunk/kata/internal/models"
)

const BASE_URL = "https://leetcode.com/graphql"
const CHECK_URL = "https://leetcode.com/submissions/detail/%s/check/"

var (
	ErrNotFound = errors.New("no matching question found")
)

type Service struct {
	client  *http.Client
	baseUrl string
	session string
	token   string
}

type Option func(*Service)

func WithCookies(session, csrf string) Option {
	return func(s *Service) {
		s.session = session
		s.token = csrf
	}
}

func (s *Service) SetCookies(session, csrf string) {
	s.session = session
	s.token = csrf
}

func (c *Service) Set(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

func New(opts ...Option) *Service {
	client := http.DefaultClient
	lcs := &Service{baseUrl: BASE_URL, client: client}

	for _, opt := range opts {
		opt(lcs)
	}

	return lcs
}

var queryUserStreak string = `query getStreakCounter { streakCounter { currentDayCompleted } }`

// More Auth -> are we authenticated?
func (lc *Service) Ping() (bool, error) {
	data, err := json.Marshal(models.Request{Query: queryUserStreak})
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", lc.baseUrl, bytes.NewBuffer(data))
	if err != nil {
		return false, err
	}
	req.Header.Set("referer", "https://leetcode.com/problemset/")
	req.Header.Set("origin", "https://leetcode.com")
	req.Header.Set("content-type", "application/json")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return false, fmt.Errorf("Error creating cookie jar: %v", err)
	}
	cookies := []*http.Cookie{
		{Name: "csrftoken", Value: lc.token},
		{Name: "LEETCODE_SESSION", Value: lc.session},
	}
	jar.SetCookies(req.URL, cookies)
	lc.client.Jar = jar

	res, err := lc.client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	// convert response to a question
	body, err := io.ReadAll(res.Body)
	var response models.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, fmt.Errorf("Error unmarshalling response: %w", err)
	}

	if response.Data.StreakCounter == nil {
		return false, nil
	}
	return true, nil
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
	body, err := json.Marshal(models.Request{Query: queryQuestionDetails, Variables: variables})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", lc.baseUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := lc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// convert response to a question
	body, err = io.ReadAll(res.Body)
	var response models.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Data.Question == nil {
		return nil, ErrNotFound
	}

	return response.Data.Question, nil
}

func (lc *Service) Test(snippet string) (string, error) {
	// {
	//   "lang": "golang",
	//   "question_id": "1",
	//   "typed_code": "func twoSum(nums []int, target int) []int {\n    diffMap := make(map[int]int)\n\n    for i, num := range nums {\n        diff := target - num\n\n        if val, exists := diffMap[diff]; exists {\n            return []int{i, val}\n        }\n\n        diffMap[num] = i\n    }\n    return nil\n}",
	//   "data_input": "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6"
	// }
	// code, err := strconv.Unquote(`"` + snippet + `"`)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to unquote snippet: %w", err)
	// }
	fmt.Println("Code", snippet)

	variables := map[string]any{
		"lang":        "golang",
		"question_id": "1",
		"typed_code":  snippet,
		"data_input":  "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6",
	}

	data, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}

	fmt.Println("Sending", string(data))

	req, err := http.NewRequest(http.MethodPost, "https://leetcode.com/problems/two-sum/interpret_solution/", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Referer", "https://leetcode.com/problemset/")
	req.Header.Set("Origin", "https://leetcode.com")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-csrf-token", lc.token)
	req.Header.Set("User-Agent", "Mozilla/6.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
		// return false, fmt.Errorf("Error creating cookie jar: %v", err)
	}
	cookies := []*http.Cookie{
		{Name: "csrftoken", Value: lc.token},
		{Name: "LEETCODE_SESSION", Value: lc.session},
	}
	jar.SetCookies(req.URL, cookies)
	lc.client.Jar = jar

	res, err := lc.client.Do(req)
	if err != nil {
		return "", err
		// return nil, err
	}
	defer res.Body.Close()

	//		{
	//	    "interpret_id": "runcode_1743738879.496263_11Hix5eo20",
	//	    "test_case": "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6"
	//	}
	body, err := io.ReadAll(res.Body)
	fmt.Println("Raw Test Resp", string(body))
	var response models.TestResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling response: %w", err)
	}

	fmt.Println("URL", response.InterpretID)
	return fmt.Sprintf(CHECK_URL, response.InterpretID), nil
}

func (lc *Service) CheckTestStatus(callbackUrl string) (*models.TestResponse, error) {
	req, err := http.NewRequest(http.MethodGet, lc.baseUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("referer", "https://leetcode.com/problemset/")
	req.Header.Set("origin", "https://leetcode.com")
	req.Header.Set("content-type", "application/json")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating cookie jar: %v", err)
	}
	cookies := []*http.Cookie{
		{Name: "csrftoken", Value: lc.token},
		{Name: "LEETCODE_SESSION", Value: lc.session},
	}
	jar.SetCookies(req.URL, cookies)
	lc.client.Jar = jar

	res, err := lc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// convert response to a question
	body, err := io.ReadAll(res.Body)
	var response models.TestResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling response: %w", err)
	}

	return &response, nil
}
