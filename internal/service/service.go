package service

import (
	"errors"
	"net/url"
	"strings"

	"shortlink/internal/shortcode"
	"shortlink/internal/store"
)

const (
	retriesAtBaseLength = 5
	maxURLLength        = 2048
)

type Service struct {
	store      store.Store
	baseURL    string
	codeLength int
}

func New(s store.Store, baseURL string, codeLength int) *Service {
	if codeLength <= 0 {
		codeLength = 6
	}
	return &Service{
		store:      s,
		baseURL:    strings.TrimRight(baseURL, "/"),
		codeLength: codeLength,
	}
}

func (s *Service) Encode(rawURL string) (string, error) {
	normalized, err := s.normalizeURL(rawURL)
	if err != nil {
		return "", err
	}

	attempts := 0
	gen := func() (string, error) {
		attempts++
		length := s.codeLength
		if attempts > retriesAtBaseLength {
			length++
		}
		return shortcode.Generate(length)
	}

	code, err := s.store.SaveURL(normalized, gen)
	if err != nil {
		return "", err
	}

	return s.baseURL + "/" + code, nil
}

func (s *Service) Decode(shortURL string) (string, error) {
	code, err := s.extractCode(shortURL)
	if err != nil {
		return "", err
	}

	original, err := s.store.Lookup(code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}

	return original, nil
}

func (s *Service) normalizeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" || len(rawURL) > maxURLLength {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	if !u.IsAbs() || u.Host == "" {
		return "", ErrInvalidURL
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrInvalidURL
	}
	u.Scheme = scheme

	return u.String(), nil
}

func (s *Service) extractCode(shortURL string) (string, error) {
	shortURL = strings.TrimSpace(shortURL)
	if shortURL == "" {
		return "", ErrInvalidCode
	}

	code := shortURL
	if strings.Contains(shortURL, "/") {
		u, err := url.Parse(shortURL)
		if err != nil {
			return "", ErrInvalidCode
		}
		code = strings.Trim(u.Path, "/")
	}

	if code == "" || !isValidCode(code) {
		return "", ErrInvalidCode
	}

	return code, nil
}

func isValidCode(code string) bool {
	for _, r := range code {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		default:
			return false
		}
	}
	return true
}
