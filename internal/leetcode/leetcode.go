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
	"strings"
	"time"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
	"github.com/phantompunk/kata/internal/models"
)

const BASE_URL = "https://leetcode.com/graphql"
const LOGIN_URL = "https://leetcode.com/accounts/login/"
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

// :TODO: Cookie fetching logic to leetcode package
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
		// return fmt.Errorf("Failed to find LEETCODE_SESSION cookie in any browser.\nPlease log in at %s first", LEETCODE_URL)
		return "", "", time.Time{}, fmt.Errorf("Failed to find LEETCODE_SESSION cookie in any browser.\nLog in at %s using a supported browser (e.g. Chrome, Chromium, Safari)", LOGIN_URL)
	}

	cookiesSeq = kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("csrftoken")).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "csrftoken" {
			csrfCookie = cookie
			break
		}
	}
	if csrfCookie == nil {
		return "", "", time.Time{}, fmt.Errorf("Failed to find csrftoken cookie in any browser.\nLog in at %s using a supported browser (e.g. Chrome, Chromium, Safari)", LOGIN_URL)
	}

	fmt.Println("Session cookie expires at", sessionCookie.Expires)
	fmt.Println("Csrf cookie expires at", csrfCookie.Expires)

	return sessionCookie.Value, csrfCookie.Value, sessionCookie.Expires, nil
}

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

func (lc *Service) Test(question *models.Question, language, snippet string) (string, error) {

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
	// fmt.Printf("CodeR: %v", snippet)
	url := fmt.Sprintf("https://leetcode.com/problems/%s/interpret_solution/", question.TitleSlug)
	// fmt.Println("Code", snippet)
	contents := strings.ReplaceAll(snippet, "\t", "    ") // 4 spaces

	variables := map[string]any{
		"lang":        models.GetLangName(language),
		"question_id": question.ID,
		"typed_code":  contents,
		"data_input":  "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6",
	}

	data, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}

	// fmt.Println("Sending", string(data))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("referer", fmt.Sprintf("https://leetcode.com/problemset/%s/description", question.TitleSlug))
	req.Header.Set("origin", "https://leetcode.com")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-csrftoken", lc.token)
	req.Header.Set("user-agent", "Mozilla/6.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
	// fmt.Println("Token", lc.token)
	// fmt.Println("Session", lc.session)
	// cookie := fmt.Sprintf("csrftoken=%s; LEETCODE_SESSION=%s", lc.token, lc.session)
	// req.Header.Set("Cookie", cookie)
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

	//	{
	//	    "interpret_id": "runcode_1743738879.496263_11Hix5eo20",
	//	    "test_case": "[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6"
	//	}
	body, err := io.ReadAll(res.Body)
	var response models.TestResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling response: %w", err)
	}

	fmt.Println("URL", response.InterpretID)
	return fmt.Sprintf(CHECK_URL, response.InterpretID), nil
}

func (lc *Service) CheckTestStatus(callbackUrl string) (*models.TestResponse, error) {
	req, err := http.NewRequest(http.MethodGet, callbackUrl, nil)
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
	// fmt.Println("RAW Check Resp", string(body))
	var response models.TestResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling response: %w", err)
	}

	return &response, nil
}
