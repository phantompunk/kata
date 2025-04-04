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
