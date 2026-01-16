package app

import (
	"context"
	"fmt"

	"github.com/phantompunk/kata/internal/browser"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/leetcode"
)

type SessionService struct {
	client      leetcode.Client
	config      *config.Config
	confService *config.ConfigService
}

func NewSessionService(cfg *config.Config, client leetcode.Client, config *config.ConfigService) *SessionService {
	return &SessionService{config: cfg, client: client, confService: config}
}

func (s *SessionService) CheckSession(ctx context.Context) error {
	if !s.config.HasValidSession() {
		s.confService.ClearSession()
		return ErrInvalidSession
	}

	valid, err := s.client.IsAuthenticated(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping leetcode service: %w", err)
	}

	if !valid {
		s.confService.ClearSession()
		return ErrInvalidSession
	}

	return nil
}

func (s *SessionService) RefreshFromBrowser() error {
	cookies, err := browser.GetCookies()
	if err != nil {
		return fmt.Errorf("failed to get browser cookies: %v", err)
	}

	sessionID, hasSession := cookies["LEETCODE_SESSION"]
	csrfToken, hasCSRF := cookies["csrftoken"]

	if !hasSession || !hasCSRF {
		return fmt.Errorf("required leetcode cookies not found in browser")
	}
	session := config.NewSession(sessionID, csrfToken)

	s.client.SetSession(session)
	return s.confService.UpdateSession(session)
}

func (s *SessionService) ValidateSession(ctx context.Context) (string, error) {
	userStatus, err := s.client.GetUserStatus(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to ping leetcode service: %w", err)
	}

	if userStatus.Username == "" {
		return "", ErrInvalidSession
	}

	s.confService.SaveUsername(userStatus.Username)
	s.confService.SavePremiumStatus(userStatus.IsPremium)
	s.config.IsPremium = userStatus.IsPremium

	return userStatus.Username, nil
}
